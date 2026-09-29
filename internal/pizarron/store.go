// Package pizarron implementa el núcleo en memoria del pizarrón ligero:
// IDs secuenciales atómicos, nodos de texto plano y notificaciones de
// escritura no bloqueantes. Sin I/O, sin HTTP: la persistencia y el
// transporte viven fuera (cmd/pizarron) para que esto sea trivial de probar.
package pizarron

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	// DefaultPreviewLength trunca el preview de /list para no saturar la red.
	DefaultPreviewLength = 140
	// DefaultListLimit y MaxListLimit acotan el coste de /list.
	DefaultListLimit = 100
	MaxListLimit     = 500
	// SubscriberBufferSize evita que un viewer lento bloquee a los agentes.
	SubscriberBufferSize = 64
)

// ErrTooLarge indica que un nodo superaría el límite configurado.
var ErrTooLarge = errors.New("body too large")

// Entry es una fila de /list: ID, tamaño real y preview legible.
type Entry struct {
	ID      uint64
	Size    int
	Preview string
}

// Store es el pizarrón en RAM. Seguro para uso concurrente.
// El cero-valor no es válido: usar New.
type Store struct {
	items map[uint64]string
	next  atomic.Uint64

	subsMu sync.Mutex
	subs   map[chan uint64]struct{}
}

// New crea un Store vacío.
func New() *Store {
	return &Store{
		items: make(map[uint64]string),
		subs:  make(map[chan uint64]struct{}),
	}
}

// Next reserva el siguiente ID secuencial. Consume el ID aunque nunca
// se escriba (como las secuencias SQL): los huecos son normales.
func (s *Store) Next() uint64 {
	return s.next.Add(1)
}

// Last devuelve el último ID reservado (0 si ninguno).
func (s *Store) Last() uint64 {
	return s.next.Load()
}

// Restore carga un nodo existente (replay del log) y eleva la marca
// de agua del contador para que Next nunca colisione.
func (s *Store) Restore(id uint64, body string) {
	if id == 0 {
		return
	}
	s.items[id] = body
	for {
		current := s.next.Load()
		if id <= current || s.next.CompareAndSwap(current, id) {
			return
		}
	}
}

// Read devuelve el cuerpo del nodo o false si el ID no existe.
func (s *Store) Read(id uint64) (string, bool) {
	body, found := s.items[id]
	return body, found
}

// Write guarda (o sobrescribe) el nodo y notifica a los suscriptores.
func (s *Store) Write(id uint64, body string) {
	s.items[id] = body
	s.broadcast(id)
}

// Append concatena fragment al nodo con separador "\n".
// Si el ID no existe, lo crea. Devuelve ErrTooLarge sin mutar nada
// cuando el resultado superaría maxBody.
func (s *Store) Append(id uint64, fragment string, maxBody int) (string, error) {
	previous := s.items[id]
	merged := previous
	if merged != "" {
		merged += "\n"
	}
	merged += fragment
	if len(merged) > maxBody {
		return "", ErrTooLarge
	}
	s.items[id] = merged
	s.broadcast(id)
	return merged, nil
}

// List recorre [from, to] devolviendo hasta limit nodos existentes.
// from==0 significa "los últimos limit hasta to"; to==0 significa Last().
func (s *Store) List(from, to uint64, limit int) []Entry {
	if limit <= 0 || limit > MaxListLimit {
		limit = DefaultListLimit
	}
	last := s.next.Load()
	if to == 0 || to > last {
		to = last
	}
	if from == 0 && to > uint64(limit) {
		from = to - uint64(limit) + 1
	}
	if from == 0 {
		from = 1
	}
	entries := make([]Entry, 0, limit)
	for id := from; id <= to && len(entries) < limit; id++ {
		body, found := s.items[id]
		if !found {
			continue
		}
		entries = append(entries, Entry{
			ID:      id,
			Size:    len(body),
			Preview: Preview(body, DefaultPreviewLength),
		})
	}
	return entries
}

// Subscribe devuelve un canal con los IDs escritos y la función para
// desuscribirse. El canal nunca bloquea a los escritores: si el lector
// va lento, se descarta el aviso (el dato sigue en Read).
func (s *Store) Subscribe() (chan uint64, func()) {
	events := make(chan uint64, SubscriberBufferSize)
	s.subsMu.Lock()
	s.subs[events] = struct{}{}
	s.subsMu.Unlock()
	unsubscribe := func() {
		s.subsMu.Lock()
		delete(s.subs, events)
		s.subsMu.Unlock()
	}
	return events, unsubscribe
}

func (s *Store) broadcast(id uint64) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for events := range s.subs {
		select {
		case events <- id:
		default:
		}
	}
}

// Preview aplana un cuerpo a una línea de hasta maxRunes caracteres.
func Preview(body string, maxRunes int) string {
	flat := strings.ReplaceAll(body, "\n", " ")
	flat = strings.ReplaceAll(flat, "\t", " ")
	if len(flat) > maxRunes {
		return flat[:maxRunes]
	}
	return flat
}
