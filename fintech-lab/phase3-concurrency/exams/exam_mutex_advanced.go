// exam_mutex_advanced.go — EXAMEN 3-D : "Les Pieges du Mutex"
// Deadlock intermittent et data race sur le OrderCache.
//
// Lancer : go run phase3-concurrency/exams/exam_mutex_advanced.go
// Avec race detector : go run -race phase3-concurrency/exams/exam_mutex_advanced.go

package main

import (
	"fmt"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Order struct {
	ID     uint64
	Symbol string
	Price  float64
	Qty    int64
	Filled int64
}

// ---------------------------------------------------------------------------
// OrderCache — Cache thread-safe des ordres actifs
// ---------------------------------------------------------------------------

// OrderCache maintient un index des ordres en memoire pour acces O(1).
// Concu pour etre utilise par des goroutines concurrentes.
type OrderCache struct {
	mu   sync.RWMutex
	data map[uint64]*Order
	hits int64 // nombre de cache hits
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		data: make(map[uint64]*Order),
	}
}

// Get retourne un ordre par son ID. Thread-safe en lecture.
func (c *OrderCache) Get(id uint64) (*Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	o, ok := c.data[id]
	if ok {
		c.hits++ // BUG #1 : ecriture sur c.hits avec seulement RLock !
		// RLock ne protege pas les ecritures concurrentes.
		// Plusieurs goroutines peuvent incrementer hits simultanement => data race.
	}
	return o, ok
}

// UpdateOrder met a jour un ordre existant.
// BUG #1 : RLock utilise pour une ECRITURE. Corruption silencieuse garantie.
func (c *OrderCache) UpdateOrder(o *Order) {
	c.mu.RLock()         // BUG #1 : devrait etre Lock() pour une ecriture !
	defer c.mu.RUnlock() // BUG #1 correspondant
	c.data[o.ID] = o     // Ecriture dans la map avec seulement un RLock !
}

// Add ajoute un ordre. Thread-safe.
func (c *OrderCache) Add(o *Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[o.ID] = o
}

// Remove supprime un ordre du cache.
// BUG #2 : Remove appelle logDeletion() qui tente d'acquerir le meme Lock.
// Deadlock garanti : la goroutine attend un lock qu'elle tient deja.
func (c *OrderCache) Remove(id uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.data[id]; !ok {
		return false
	}
	delete(c.data, id)

	// BUG #2 : logDeletion() appelle c.mu.Lock() en interne.
	// sync.Mutex n'est PAS reentrant en Go => DEADLOCK.
	c.logDeletion(id)
	return true
}

// logDeletion enregistre la suppression d'un ordre.
// PROBLEME : cette methode acquiert le lock, mais elle est appellee
// depuis Remove() qui tient deja le lock => deadlock.
func (c *OrderCache) logDeletion(id uint64) {
	c.mu.Lock()         // BUG #2 : DEADLOCK — lock deja tenu par Remove()
	defer c.mu.Unlock()
	fmt.Printf("  [LOG] Ordre #%d supprime du cache (size: %d)\n", id, len(c.data))
}

// GetOrCreate retourne un ordre existant ou en cree un nouveau.
// BUG #3 : Double unlock. Si createFn() appelle c.Add() qui lock/unlock,
//          et que le defer final unloque aussi, le comportement est incorrect.
func (c *OrderCache) GetOrCreate(id uint64, createFn func() *Order) *Order {
	c.mu.Lock()

	if o, ok := c.data[id]; ok {
		c.mu.Unlock() // Premier unlock : OK si l'ordre existe
		return o
	}

	// L'ordre n'existe pas, on le cree
	// BUG #3 : On tient le lock et on appelle createFn qui peut appeler c.Add()
	// c.Add() va tenter de locker => DEADLOCK
	newOrder := createFn()

	c.data[newOrder.ID] = newOrder
	c.mu.Unlock() // Deuxieme unlock : OK dans ce cas
	return newOrder
}

// Hits retourne le nombre de cache hits (approximatif a cause du bug #1).
func (c *OrderCache) Hits() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits
}

// Size retourne le nombre d'ordres en cache.
func (c *OrderCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// ---------------------------------------------------------------------------
// Simulation de charge concurrente
// ---------------------------------------------------------------------------

func main() {
	cache := NewOrderCache()

	// Pre-remplir le cache
	for i := uint64(1); i <= 100; i++ {
		cache.Add(&Order{ID: i, Symbol: "AAPL", Price: 189.50, Qty: 100})
	}

	var wg sync.WaitGroup

	// --- Goroutines readers (simulant le market data feed) ---
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				id := uint64(i%100 + 1)
				if o, ok := cache.Get(id); ok {
					_ = o.Price // Utiliser la valeur pour eviter optimisation
				}
				time.Sleep(time.Microsecond)
			}
		}(r)
	}

	// --- Goroutines writers (simulant les executions de trades) ---
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				id := uint64(i%100 + 1)
				if o, ok := cache.Get(id); ok {
					o.Filled += 10
					cache.UpdateOrder(o) // BUG #1 : race condition ici
				}
				time.Sleep(time.Microsecond)
			}
		}(w)
	}

	// --- Goroutine manager (suppression d'ordres remplis) ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := uint64(1); i <= 10; i++ {
			time.Sleep(5 * time.Millisecond)
			// BUG #2 : Remove -> logDeletion -> Lock => DEADLOCK
			removed := cache.Remove(i)
			if removed {
				fmt.Printf("  Ordre #%d retire du cache\n", i)
			}
		}
	}()

	// --- Test GetOrCreate ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := uint64(200); i <= 210; i++ {
			id := i
			// BUG #3 : createFn appelle cache.Add() qui tente de locker
			// alors que GetOrCreate tient deja le lock => DEADLOCK
			order := cache.GetOrCreate(id, func() *Order {
				newO := &Order{ID: id, Symbol: "MSFT", Price: 415.20, Qty: 50}
				// BUG #3 : ceci ne doit PAS appeler cache.Add() depuis l'interieur
				// (on est deja dans le lock de GetOrCreate)
				return newO
			})
			_ = order
		}
	}()

	// Attendre avec timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("\nToutes les goroutines terminees.")
		fmt.Printf("Cache size : %d ordres\n", cache.Size())
		fmt.Printf("Cache hits : %d\n", cache.Hits())
	case <-time.After(5 * time.Second):
		fmt.Println("\nTIMEOUT (5s) : DEADLOCK detecte !")
		fmt.Println("Identifie et corrige les 3 bugs de mutex.")
		fmt.Println("Explique le bug #4 (starvation) en commentaire.")
	}
}
