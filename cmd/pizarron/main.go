// Command pizarron levanta el pizarrón ligero v1: 1 binario, stdlib only,
// text/plain, RAM-first. Agentes: /next /read /write (/append opcional).
// Personal (solo lectura): / /stream /list /healthz.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"pizarron/internal/pizarron"
	"pizarron/web"
)

// Version se inyecta en release vía: -ldflags "-X main.Version=v0.1.0".
var Version = "dev"

const (
	defaultAddr         = "127.0.0.1:8000"
	defaultLogName      = "pizarron.log"
	logFilePermission   = 0o644
	maxBodyBytes        = 1 << 20 // 1 MiB por nodo
	streamPingInterval  = 15 * time.Second
	contentTypeText     = "text/plain; charset=utf-8"
	contentTypeHTML     = "text/html; charset=utf-8"
	contentTypeStream   = "text/event-stream"
	contentTypeJSON     = "application/json"
)

// config agrupa lo que los handlers necesitan sin variables globales regadas.
type config struct {
	store     *pizarron.Store
	logPath   string
	viewToken string
	version   string
	startedAt time.Time
	logMu     sync.Mutex
}

func main() {
	showVersion := flag.Bool("version", false, "imprime la versión y sale")
	flag.Parse()
	if *showVersion {
		fmt.Println(Version)
		return
	}

	cfg := &config{
		store:     pizarron.New(),
		logPath:   envOr("PIZARRON_LOG", defaultLogName),
		viewToken: os.Getenv("VIEW_TOKEN"),
		version:   Version,
		startedAt: time.Now(),
	}
	if err := cfg.replay(); err != nil {
		log.Fatalf("replay de %q: %v", cfg.logPath, err)
	}

	addr := envOr("PIZARRON_ADDR", defaultAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/next", cfg.handleNext)
	mux.HandleFunc("/read", cfg.handleRead)
	mux.HandleFunc("/write", cfg.handleWrite)
	mux.HandleFunc("/append", cfg.handleAppend)
	mux.HandleFunc("/list", cfg.handleList)
	mux.HandleFunc("/healthz", cfg.handleHealth)
	mux.HandleFunc("/stream", cfg.handleStream)
	mux.HandleFunc("/", cfg.handleViewer)
	log.Printf("pizarron %s en http://%s (log=%s)", cfg.version, addr, cfg.logPath)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// --- núcleo: agentes ---

func (c *config) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeText(w, fmt.Sprintf("%d\n", c.store.Next()))
}

func (c *config) handleRead(w http.ResponseWriter, r *http.Request) {
	id, ok := queryID(r)
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	body, found := c.store.Read(id)
	if !found {
		http.Error(w, "id not found", http.StatusNotFound)
		return
	}
	if strings.Contains(r.Header.Get("Accept"), contentTypeJSON) {
		w.Header().Set("Content-Type", contentTypeJSON)
		fmt.Fprintf(w, "{\"id\":%d,\"body\":%s}\n", id, jsonQuote(body))
		return
	}
	writeText(w, body)
}

func (c *config) handleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := queryID(r)
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		log.Printf("write id=%d: error leyendo cuerpo: %v", id, err)
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if len(body) > maxBodyBytes {
		http.Error(w, "too large", http.StatusRequestEntityTooLarge)
		return
	}
	c.store.Write(id, string(body))
	c.persist(id, string(body))
	writeText(w, "OK\n")
}

func (c *config) handleAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, ok := queryID(r)
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	fragment, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		log.Printf("append id=%d: error leyendo cuerpo: %v", id, err)
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	merged, err := c.store.Append(id, string(fragment), maxBodyBytes)
	if err != nil {
		http.Error(w, "too large", http.StatusRequestEntityTooLarge)
		return
	}
	c.persist(id, merged)
	writeText(w, "OK\n")
}

// --- solo lectura: personal ---

func (c *config) handleList(w http.ResponseWriter, r *http.Request) {
	from, _ := strconv.ParseUint(r.URL.Query().Get("from"), 10, 64)
	to, _ := strconv.ParseUint(r.URL.Query().Get("to"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	var out strings.Builder
	for _, entry := range c.store.List(from, to, limit) {
		fmt.Fprintf(&out, "%d\t%d\t%s\n", entry.ID, entry.Size, entry.Preview)
	}
	writeText(w, out.String())
}

func (c *config) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeText(w, fmt.Sprintf("ok N=%d version=%s uptime=%s\n",
		c.store.Last(), c.version, time.Since(c.startedAt).Round(time.Second)))
}

func (c *config) handleStream(w http.ResponseWriter, r *http.Request) {
	if !c.authorizedViewer(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", contentTypeStream)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no stream", http.StatusInternalServerError)
		return
	}
	events, unsubscribe := c.store.Subscribe()
	defer unsubscribe()
	fmt.Fprintf(w, "event: hello\ndata: %d\n\n", c.store.Last())
	flusher.Flush()
	disconnected := r.Context().Done()
	heartbeat := time.NewTicker(streamPingInterval)
	defer heartbeat.Stop()
	for {
		select {
		case <-disconnected:
			return
		case id := <-events:
			fmt.Fprintf(w, "event: write\ndata: %d\n\n", id)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func (c *config) handleViewer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !c.authorizedViewer(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	page, err := web.FS.ReadFile("viewer.html")
	if err != nil {
		log.Printf("viewer: embed roto: %v", err)
		http.Error(w, "viewer unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeHTML)
	w.Write(page)
}

func (c *config) authorizedViewer(r *http.Request) bool {
	return c.viewToken == "" || r.URL.Query().Get("token") == c.viewToken
}

// --- persistencia append-only (background, nunca bloquea el hot path) ---

func (c *config) persist(id uint64, body string) {
	go func() {
		c.logMu.Lock()
		defer c.logMu.Unlock()
		file, err := os.OpenFile(c.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePermission)
		if err != nil {
			log.Printf("persist id=%d: no se pudo abrir %q: %v", id, c.logPath, err)
			return
		}
		defer file.Close()
		if _, err := fmt.Fprintf(file, "%d\t%d\t%d\n%s\n.\n", id, time.Now().Unix(), len(body), body); err != nil {
			log.Printf("persist id=%d: error escribiendo: %v", id, err)
		}
	}()
}

// replay reconstruye la RAM desde el log. Un bloque corrupto se salta con
// aviso; si el archivo no existe, se arranca vacío (no es error).
func (c *config) replay() error {
	raw, err := os.ReadFile(c.logPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, block := range strings.Split(string(raw), "\n.\n") {
		if block == "" {
			continue
		}
		head, body, found := strings.Cut(block, "\n")
		if !found {
			log.Printf("replay: bloque sin cabecera, saltado (%d bytes)", len(block))
			continue
		}
		fields := strings.Split(head, "\t")
		id, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			log.Printf("replay: cabecera inválida %q, saltada", head)
			continue
		}
		c.store.Restore(id, body)
	}
	return nil
}

// --- helpers ---

func queryID(r *http.Request) (uint64, bool) {
	raw := r.URL.Query().Get("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if raw == "" || err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

func writeText(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", contentTypeText)
	io.WriteString(w, body)
}

func jsonQuote(text string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, char := range text {
		switch char {
		case '"':
			out.WriteString("\\\"")
		case '\\':
			out.WriteString("\\\\")
		case '\n':
			out.WriteString("\\n")
		case '\r':
			out.WriteString("\\r")
		case '\t':
			out.WriteString("\\t")
		default:
			out.WriteRune(char)
		}
	}
	out.WriteByte('"')
	return out.String()
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
