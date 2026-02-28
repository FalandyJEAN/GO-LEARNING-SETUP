// ================================================================================
// LECON 09 — Mutex et Synchronisation
// ================================================================================
// COMMENT EXECUTER :  go run phase3-concurrency/lessons/lesson03_mutex.go
// COMMENT VERIFIER : go run -race phase3-concurrency/lessons/lesson03_mutex.go
// OBJECTIF         :  Proteger les ressources partagees contre les data races
// APRES CETTE LECON : Tu peux attaquer les 4 examens de Phase 3 !
// ================================================================================

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ============================================================
// PARTIE 1 : PROBLEME — DATA RACE
// ============================================================
// Quand plusieurs goroutines lisent ET ecrivent la meme variable
// sans synchronisation, le resultat est indefini.
// En HFT, ca peut causer des pertes massives.

// PortefeuilleNonSafe : acces concurrent non protege (BUGGE)
type PortefeuilleNonSafe struct {
	positions map[string]int
}

func (p *PortefeuilleNonSafe) AjouterPosition(symbol string, qty int) {
	p.positions[symbol] += qty // DATA RACE si appele depuis plusieurs goroutines
}

// ============================================================
// PARTIE 2 : SOLUTION 1 — sync.Mutex
// ============================================================
// Mutex = Mutual Exclusion.
// Une seule goroutine peut tenir le verrou a la fois.
// Les autres attendent.
//
// Lock()    : prendre le verrou (bloquer si deja pris)
// Unlock()  : liberer le verrou
// TOUJOURS utiliser defer mu.Unlock() pour garantir la liberation.

type Portefeuille struct {
	mu        sync.Mutex     // le verrou
	positions map[string]int // donnee protegee
	pnl       float64
}

func NewPortefeuille() *Portefeuille {
	return &Portefeuille{
		positions: make(map[string]int),
	}
}

func (p *Portefeuille) AjouterPosition(symbol string, qty int) {
	p.mu.Lock()         // prendre le verrou
	defer p.mu.Unlock() // liberer quoi qu'il arrive (meme si panic)
	p.positions[symbol] += qty
}

func (p *Portefeuille) ObtenirPosition(symbol string) int {
	p.mu.Lock()         // toujours locker, meme pour la lecture !
	defer p.mu.Unlock()
	return p.positions[symbol]
}

func (p *Portefeuille) ValeurTotale(prixMap map[string]float64) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := 0.0
	for sym, qty := range p.positions {
		if px, ok := prixMap[sym]; ok {
			total += float64(qty) * px
		}
	}
	return total
}

// ============================================================
// PARTIE 3 : SOLUTION 2 — sync.RWMutex
// ============================================================
// RWMutex optimise les cas ou on lit beaucoup, ecrit peu.
// Plusieurs goroutines peuvent LIRE en meme temps.
// L'ecriture necessite un acces exclusif.
//
// RLock()/RUnlock()  : verrou en lecture (partage)
// Lock()/Unlock()    : verrou en ecriture (exclusif)

type CarnetOrdres struct {
	rwmu sync.RWMutex
	bids map[float64]int // prix -> quantite
	asks map[float64]int
}

func NewCarnetOrdres() *CarnetOrdres {
	return &CarnetOrdres{
		bids: make(map[float64]int),
		asks: make(map[float64]int),
	}
}

// AjouterOrdre : ecriture -> Lock exclusif
func (c *CarnetOrdres) AjouterOrdre(estBid bool, prix float64, qty int) {
	c.rwmu.Lock()
	defer c.rwmu.Unlock()
	if estBid {
		c.bids[prix] += qty
	} else {
		c.asks[prix] += qty
	}
}

// MeilleurBid : lecture seule -> RLock partage
// Plusieurs goroutines peuvent appeler ca simultanement
func (c *CarnetOrdres) MeilleurBid() float64 {
	c.rwmu.RLock() // lecture seulement
	defer c.rwmu.RUnlock()
	meilleur := 0.0
	for prix := range c.bids {
		if prix > meilleur {
			meilleur = prix
		}
	}
	return meilleur
}

// ============================================================
// PARTIE 4 : SOLUTION 3 — sync/atomic (pour les cas simples)
// ============================================================
// Pour les compteurs et flags simples, atomic est plus rapide que Mutex.
// N'utilise PAS de verrou -> utilise des instructions CPU atomiques.

type Metriques struct {
	nbOrdres     atomic.Int64
	nbTrades     atomic.Int64
	volumeTotal  atomic.Int64
}

func (m *Metriques) NouvelOrdre() {
	m.nbOrdres.Add(1) // atomique, thread-safe, sans mutex
}

func (m *Metriques) NouveauTrade(volume int64) {
	m.nbTrades.Add(1)
	m.volumeTotal.Add(volume)
}

// ============================================================
// PARTIE 5 : sync.Once — initialisation unique
// ============================================================
// Garantit qu'une fonction n'est executee qu'UNE SEULE FOIS,
// meme si appelee depuis plusieurs goroutines. (Singleton pattern)

type ConfigMarche struct {
	Symboles []string
	MaxOrdre int
}

var (
	configInstance *ConfigMarche
	configOnce    sync.Once
)

func ObtenirConfig() *ConfigMarche {
	configOnce.Do(func() {
		// Sera execute UNE SEULE FOIS, peu importe les goroutines
		fmt.Println("  [Config] Initialisation unique...")
		configInstance = &ConfigMarche{
			Symboles: []string{"AAPL", "MSFT", "GOOGL"},
			MaxOrdre: 10_000,
		}
	})
	return configInstance
}

// ============================================================
// MAIN
// ============================================================

func main() {

	// --- Mutex ---
	fmt.Println("=== PORTEFEUILLE THREAD-SAFE (sync.Mutex) ===")

	portefeuille := NewPortefeuille()
	var wg sync.WaitGroup

	// 100 goroutines ajoutent des positions en meme temps
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			portefeuille.AjouterPosition("AAPL", 10)
		}()
	}
	wg.Wait()

	fmt.Printf("Position AAPL finale : %d (attendu: 1000)\n",
		portefeuille.ObtenirPosition("AAPL"))

	// --- RWMutex ---
	fmt.Println("\n=== CARNET D'ORDRES (sync.RWMutex) ===")

	carnet := NewCarnetOrdres()
	var wgCarnet sync.WaitGroup

	// 10 goroutines ecrivent
	for i := 0; i < 10; i++ {
		wgCarnet.Add(1)
		prix := 142.50 - float64(i)*0.01
		go func(p float64) {
			defer wgCarnet.Done()
			carnet.AjouterOrdre(true, p, 100)
		}(prix)
	}

	// 5 goroutines lisent en meme temps (ok avec RWMutex)
	for i := 0; i < 5; i++ {
		wgCarnet.Add(1)
		go func() {
			defer wgCarnet.Done()
			_ = carnet.MeilleurBid()
		}()
	}

	wgCarnet.Wait()
	fmt.Printf("Meilleur bid : %.2f\n", carnet.MeilleurBid())

	// --- Atomic ---
	fmt.Println("\n=== METRIQUES ATOMIQUES ===")

	metriques := &Metriques{}
	var wgMetr sync.WaitGroup

	for i := 0; i < 500; i++ {
		wgMetr.Add(1)
		go func() {
			defer wgMetr.Done()
			metriques.NouvelOrdre()
			metriques.NouveauTrade(100)
		}()
	}
	wgMetr.Wait()

	fmt.Printf("Ordres  : %d (attendu: 500)\n", metriques.nbOrdres.Load())
	fmt.Printf("Trades  : %d (attendu: 500)\n", metriques.nbTrades.Load())
	fmt.Printf("Volume  : %d (attendu: 50000)\n", metriques.volumeTotal.Load())

	// --- sync.Once ---
	fmt.Println("\n=== sync.Once (initialisation unique) ===")
	var wgOnce sync.WaitGroup
	for i := 0; i < 5; i++ {
		wgOnce.Add(1)
		go func() {
			defer wgOnce.Done()
			cfg := ObtenirConfig()
			_ = cfg // utiliser la config
		}()
	}
	wgOnce.Wait()
	fmt.Println("  Config creee une seule fois, meme avec 5 goroutines")

	// ============================================================
	// RESUME
	// ============================================================
	// 1. sync.Mutex
	//      mu.Lock() / defer mu.Unlock()
	//      -> Un seul acces a la fois (lecture ET ecriture)
	//
	// 2. sync.RWMutex
	//      mu.RLock() / defer mu.RUnlock()  (lectures simultanees ok)
	//      mu.Lock()  / defer mu.Unlock()   (ecriture exclusive)
	//      -> Optimise pour beaucoup de lectures, peu d'ecritures
	//
	// 3. sync/atomic
	//      atomic.Int64 : Add(), Load(), Store(), Swap(), CompareAndSwap()
	//      -> Pour compteurs simples, plus rapide que Mutex
	//
	// 4. sync.Once
	//      once.Do(func() { ... })
	//      -> Execute UNE SEULE FOIS, thread-safe (Singleton)
	//
	// 5. TOUJOURS defer mu.Unlock() pour eviter les deadlocks !

	fmt.Println("\n=== FIN LECON 09 ===")
	fmt.Println("Tu peux maintenant attaquer les examens Phase 3 !")
	fmt.Println("  exam_data_race.go")
	fmt.Println("  exam_goroutine_leak.go")
	fmt.Println("  exam_channel_deadlock.go")
	fmt.Println("  exam_mutex_advanced.go")
}
