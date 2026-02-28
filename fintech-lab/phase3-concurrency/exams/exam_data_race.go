// exam_data_race.go — EXAMEN 3-A : "La Data Race Silencieuse"
// Ce code compile et peut sembler fonctionner normalement.
// Il contient 3 bugs de concurrence detectables avec : go run -race
//
// Lancer : go run phase3-concurrency/exams/exam_data_race.go
// CRITIQUE : go run -race phase3-concurrency/exams/exam_data_race.go

package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ---------------------------------------------------------------------------
// MarketDataCache — Cache de prix de marche, mis a jour par plusieurs sources
// ---------------------------------------------------------------------------

// MarketDataCache agregge les prix de plusieurs exchanges.
// BUG : Aucune synchronisation. Acces concurrents non proteges.
type MarketDataCache struct {
	prices      map[string]float64 // BUG #1 : pas de mutex
	updateCount int64              // BUG #2 : pas d'atomic
}

func NewMarketDataCache() *MarketDataCache {
	return &MarketDataCache{
		prices: make(map[string]float64),
	}
}

// UpdatePrice met a jour le prix d'un symbole depuis une source externe.
// Appelee par plusieurs goroutines simultanement.
func (c *MarketDataCache) UpdatePrice(symbol string, price float64) {
	c.prices[symbol] = price // BUG #1 : DATA RACE — ecriture non protegee
	c.updateCount++          // BUG #2 : DATA RACE — increment non atomique
}

// GetPrice retourne le dernier prix connu d'un symbole.
func (c *MarketDataCache) GetPrice(symbol string) float64 {
	return c.prices[symbol] // BUG #1 : DATA RACE — lecture concurrente avec ecriture
}

// GetUpdateCount retourne le nombre total de mises a jour.
func (c *MarketDataCache) GetUpdateCount() int64 {
	return c.updateCount // BUG #2 : DATA RACE — lecture concurrente
}

// ---------------------------------------------------------------------------
// Simulation de flux de donnees de marche
// ---------------------------------------------------------------------------

// exchangeFeed simule un flux de prix venant d'un exchange externe.
// Chaque exchange met a jour les prix aleatoirement.
func exchangeFeed(cache *MarketDataCache, exchangeName string, symbols []string) {
	// BUG #3 : Cette goroutine n'est pas synchronisee avec main().
	// main() peut se terminer avant que cette goroutine ait fini.
	// Les dernieres mises a jour sont silencieusement perdues.
	for i := 0; i < 50; i++ {
		symbol := symbols[rand.Intn(len(symbols))]
		basePrice := map[string]float64{
			"AAPL": 189.50,
			"MSFT": 415.20,
			"TSLA": 172.80,
			"GOOGL": 175.00,
		}[symbol]

		// Simuler une variation de prix (+/- 0.5%)
		variation := (rand.Float64() - 0.5) * basePrice * 0.005
		newPrice := basePrice + variation

		cache.UpdatePrice(symbol, newPrice)
	}
}

// priceMonitor lit les prix en continu pendant que les feeds ecrivent.
// C'est la race classique lecture/ecriture concurrente.
func priceMonitor(cache *MarketDataCache, symbols []string, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			for _, s := range symbols {
				_ = cache.GetPrice(s) // BUG #1 : race avec UpdatePrice
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	symbols := []string{"AAPL", "MSFT", "TSLA", "GOOGL"}
	cache := NewMarketDataCache()
	done := make(chan struct{})

	fmt.Println("Demarrage de 5 exchanges + 1 monitor...")

	// BUG #3 : Pas de WaitGroup. main() n'attend pas les goroutines.
	// Lancer 5 goroutines simulant 5 exchanges differents
	for i := 1; i <= 5; i++ {
		go exchangeFeed(cache, fmt.Sprintf("Exchange-%d", i), symbols)
	}

	// Lancer le monitor en parallele
	go priceMonitor(cache, symbols, done)

	// Attendre "un peu" — non-deterministe et insuffisant !
	// BUG #3 : time.Sleep ne garantit pas que toutes les goroutines ont fini.
	time.Sleep(100 * time.Millisecond)

	// Signaler au monitor d'arreter
	close(done)

	// Afficher les resultats finaux
	fmt.Println("\n--- Prix finaux ---")
	for _, s := range symbols {
		fmt.Printf("  %s : $%.4f\n", s, cache.GetPrice(s))
	}
	fmt.Printf("\nTotal updates : %d\n", cache.GetUpdateCount())
	fmt.Println("(attendu : ~250 updates = 5 exchanges x 50 iterations)")
	fmt.Println("\nLancer avec -race pour voir les data races !")
}
