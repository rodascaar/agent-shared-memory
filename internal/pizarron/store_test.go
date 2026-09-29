package pizarron

import (
	"fmt"
	"sync"
	"testing"
)

func TestNextIsSequentialAndUnique(t *testing.T) {
	store := New()
	const writers = 20
	const perWriter = 50
	seen := make(chan uint64, writers*perWriter)
	var group sync.WaitGroup
	for w := 0; w < writers; w++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for i := 0; i < perWriter; i++ {
				seen <- store.Next()
			}
		}()
	}
	group.Wait()
	close(seen)
	unique := map[uint64]bool{}
	for id := range seen {
		if unique[id] {
			t.Fatalf("ID duplicado: %d", id)
		}
		unique[id] = true
	}
	if len(unique) != writers*perWriter {
		t.Fatalf("esperaba %d IDs, obtuvo %d", writers*perWriter, len(unique))
	}
	if store.Last() != writers*perWriter {
		t.Fatalf("Last=%d, esperaba %d", store.Last(), writers*perWriter)
	}
}

func TestNextGapIsNormal(t *testing.T) {
	store := New()
	reserved := store.Next()
	if _, found := store.Read(reserved); found {
		t.Fatal("un ID reservado sin Write no debe ser legible")
	}
}

func TestWriteReadRoundtrip(t *testing.T) {
	store := New()
	id := store.Next()
	store.Write(id, "hallazgo")
	body, found := store.Read(id)
	if !found || body != "hallazgo" {
		t.Fatalf("roundtrip roto: %q %v", body, found)
	}
	if _, found := store.Read(id + 1); found {
		t.Fatal("leer ID inexistente debe devolver false")
	}
}

func TestAppendJoinsWithNewline(t *testing.T) {
	store := New()
	id := store.Next()
	if _, err := store.Append(id, "uno", 1024); err != nil {
		t.Fatal(err)
	}
	merged, err := store.Append(id, "dos", 1024)
	if err != nil || merged != "uno\ndos" {
		t.Fatalf("append roto: %q %v", merged, err)
	}
}

func TestAppendRejectsOversizeWithoutMutating(t *testing.T) {
	store := New()
	id := store.Next()
	store.Write(id, "base")
	if _, err := store.Append(id, "demasiado-largo", len("base")+2); err != ErrTooLarge {
		t.Fatalf("esperaba ErrTooLarge, obtuvo %v", err)
	}
	body, _ := store.Read(id)
	if body != "base" {
		t.Fatalf("el rechazo debe dejar el nodo intacto: %q", body)
	}
}

func TestListWindow(t *testing.T) {
	store := New()
	for i := 0; i < 5; i++ {
		id := store.Next()
		store.Write(id, fmt.Sprintf("nodo-%d", id))
	}
	entries := store.List(2, 4, 10)
	if len(entries) != 3 || entries[0].ID != 2 || entries[2].ID != 4 {
		t.Fatalf("ventana incorrecta: %+v", entries)
	}
	recent := store.List(0, 0, 2)
	if len(recent) != 2 || recent[0].ID != 4 || recent[1].ID != 5 {
		t.Fatalf("últimos N incorrectos: %+v", recent)
	}
	if got := store.List(0, 0, MaxListLimit+1); len(got) != 5 {
		t.Fatalf("el límite debe recortarse, no fallar: %d", len(got))
	}
}

func TestSubscribeReceivesWrites(t *testing.T) {
	store := New()
	events, unsubscribe := store.Subscribe()
	defer unsubscribe()
	id := store.Next()
	store.Write(id, "x")
	select {
	case got := <-events:
		if got != id {
			t.Fatalf("evento con ID incorrecto: %d", got)
		}
	default:
		t.Fatal("el suscriptor no recibió el write")
	}
}

func TestRestoreRaisesHighWaterMark(t *testing.T) {
	store := New()
	store.Restore(41, "viejo")
	store.Restore(0, "ignorado")
	if body, found := store.Read(41); !found || body != "viejo" {
		t.Fatal("restore no cargó el nodo")
	}
	if next := store.Next(); next != 42 {
		t.Fatalf("Next tras restore=%d, esperaba 42", next)
	}
}
