// exam_goroutine_leak.go — EXAMEN 3-B : "La Fuite de Goroutines"
// Ce code compile mais les goroutines ne se terminent jamais.
// Le programme ne se terminera PAS sans Ctrl+C.
//
// Lancer : go run phase3-concurrency/exams/exam_goroutine_leak.go

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Order struct {
	ID     int
	Symbol string
	Price  float64
}

type Result struct {
	OrderID int
	Success bool
	Message string
}

// ---------------------------------------------------------------------------
// OrderProcessor — Traite les ordres en arriere-plan
// ---------------------------------------------------------------------------

type OrderProcessor struct {
	results chan Result
	wg      sync.WaitGroup
}

func NewOrderProcessor() *OrderProcessor {
	return &OrderProcessor{
		results: make(chan Result, 100),
	}
}

// processOrders traite un batch d'ordres via une goroutine worker.
// BUG #1 + BUG #2 : Une nouvelle goroutine est creee a chaque appel.
// Ces goroutines lisent 'orders' jusqu'au bout, mais ensuite attendent
// sur 'results' qui n'a personne pour les lire. Elles bloquent.
func (p *OrderProcessor) processOrders(orders []Order) {
	orderCh := make(chan Order) // BUG #4 : channel non buffe, jamais ferme apres usage

	// BUG #1 : Cette goroutine n'a pas de mecanisme d'arret.
	// Si le channel n'est jamais ferme, elle bloque ici pour toujours.
	go func() {
		for o := range orderCh { // Bloque si orderCh jamais ferme
			result := Result{
				OrderID: o.ID,
				Success: o.Price > 0,
				Message: fmt.Sprintf("Processed %s @ %.2f", o.Symbol, o.Price),
			}
			p.results <- result
		}
		// Ce point n'est JAMAIS atteint si orderCh n'est pas ferme.
	}()

	// BUG #3 : wg.Add est appele APRES go func().
	// wg.Wait() pourrait retourner avant que les goroutines ci-dessous
	// soient toutes enregistrees.
	for _, o := range orders {
		o := o // Correction deja faite pour la closure — concentre-toi sur les autres bugs
		go func() {
			p.wg.Add(1) // BUG #3 : trop tard ! Add doit etre AVANT go func()
			defer p.wg.Done()
			orderCh <- o // Envoie l'ordre au worker
		}()
	}

	// BUG : le channel n'est jamais ferme apres l'envoi.
	// Le worker goroutine restera bloque sur "for o := range orderCh"
	// p.wg.Wait() ici ne garantit pas que tous les ordres sont envoyes
}

// collectResults lit les resultats du processor.
// BUG #2 : Cette goroutine tourne indefiniment — aucun signal d'arret.
func (p *OrderProcessor) collectResults(count int) []Result {
	var results []Result

	// BUG : Cette goroutine ne sait pas quand s'arreter.
	// Si 'count' resultats n'arrivent jamais, elle bloque pour toujours.
	// En prod, si processOrders bugue, cette goroutine "disparait" dans le void.
	go func() {
		for r := range p.results { // Jamais ferme = jamais termine
			results = append(results, r)
			fmt.Printf("  Resultat recu : Order#%d -> %s\n", r.OrderID, r.Message)
		}
	}()

	// Attendre "un peu" — non-deterministe !
	time.Sleep(200 * time.Millisecond)
	return results
}

// ---------------------------------------------------------------------------
// FeedSubscriber — S'abonne a des flux de donnees
// ---------------------------------------------------------------------------

// subscribeToFeed simule un abonnement a un flux de prix.
// BUG #1 bis : Goroutine sans fin, sans mecanisme d'arret.
// Appelee plusieurs fois => goroutines s'accumulent.
func subscribeToFeed(symbol string, prices <-chan float64) {
	go func() {
		// Cette goroutine tourne jusqu'a... quand ?
		// Si 'prices' n'est jamais ferme, elle tourne pour TOUJOURS.
		for price := range prices {
			fmt.Printf("  [%s] Nouveau prix : $%.2f\n", symbol, price)
		}
		// Ce fmt ne sera jamais imprime si prices n'est pas ferme.
		fmt.Printf("  [%s] Feed ferme\n", symbol)
	}()
	// La fonction retourne immediatement, mais la goroutine vit indefiniment.
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	fmt.Printf("Goroutines au depart : %d\n\n", runtime.NumGoroutine())

	// --- Partie 1 : OrderProcessor ---
	fmt.Println("=== Test OrderProcessor ===")
	processor := NewOrderProcessor()

	orders := []Order{
		{1, "AAPL", 189.50},
		{2, "MSFT", 415.20},
		{3, "TSLA", -1.0}, // Prix invalide
		{4, "GOOGL", 175.00},
	}

	// Appeler processOrders plusieurs fois simule plusieurs requetes clients
	for batch := 0; batch < 3; batch++ {
		processor.processOrders(orders)
		fmt.Printf("Batch %d soumis. Goroutines actives : %d\n", batch+1, runtime.NumGoroutine())
		time.Sleep(50 * time.Millisecond)
	}

	// --- Partie 2 : FeedSubscriber ---
	fmt.Println("\n=== Test FeedSubscriber ===")
	pricesAApl := make(chan float64, 10)
	pricesMSFT := make(chan float64, 10)

	subscribeToFeed("AAPL", pricesAApl)
	subscribeToFeed("MSFT", pricesMSFT)

	// Envoyer quelques prix
	pricesAApl <- 189.50
	pricesAApl <- 190.00
	pricesMSFT <- 415.20

	time.Sleep(50 * time.Millisecond)

	// BUG : Les channels ne sont jamais fermes.
	// Les goroutines subscribes restent bloquees indefiniment.
	// Si on ajoute : close(pricesAApl) close(pricesMSFT), les goroutines se terminent.

	// --- Diagnostic final ---
	fmt.Println("\n=== Diagnostic ===")
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("Goroutines apres traitement : %d\n", runtime.NumGoroutine())
	fmt.Println("(attendu : similaire au nombre initial)")
	fmt.Println("Si le nombre a augmente : goroutine leak detecte !")
	fmt.Println("\nLe programme ne se terminera PAS. Appuie sur Ctrl+C.")
	fmt.Println("Apres correction, il doit se terminer proprement.")

	// Ce point n'est jamais atteint car les goroutines bloquent main()
	// via les channels non fermes qui gardent le runtime en vie.
	select {} // Bloquer intentionnellement pour illustrer le probleme
}
