// exam_channel_deadlock.go — EXAMEN 3-C : "Le Deadlock du Pipeline"
// Ce code deadlocke ou panique selon le chemin d'execution.
//
// Lancer : go run phase3-concurrency/exams/exam_channel_deadlock.go
// Attendu : "fatal error: all goroutines are asleep - deadlock!"

package main

import (
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Order struct {
	ID     int
	Symbol string
	Price  float64
	Qty    int
}

type ValidatedOrder struct {
	Order
	Valid bool
}

type EnrichedOrder struct {
	ValidatedOrder
	Notional float64
}

// ---------------------------------------------------------------------------
// BUG #4 : nilChannel — Receiver depuis un nil channel bloque pour toujours
// ---------------------------------------------------------------------------

// drainChannel lit depuis un channel. Si ch est nil, bloque indefiniment.
func drainChannel(ch <-chan int) {
	// BUG #4 : Si ch == nil, ce range ne se termine JAMAIS.
	// Un nil channel bloque toutes les operations (send, receive, range).
	for v := range ch {
		fmt.Printf("  Valeur recue : %d\n", v)
	}
}

// ---------------------------------------------------------------------------
// Pipeline Stage 1 : Validation
// ---------------------------------------------------------------------------

// validateOrders valide les ordres et les envoie sur le channel de sortie.
// BUG #1 : Le channel 'out' a une capacite de 1. Si le consommateur est lent,
//          l'envoi bloque et le goroutine est suspendu indefiniment.
func validateOrders(orders []Order) chan ValidatedOrder {
	out := make(chan ValidatedOrder, 1) // BUG #1 : buffer trop petit !

	go func() {
		defer close(out)
		for _, o := range orders {
			validated := ValidatedOrder{
				Order: o,
				Valid: o.Price > 0 && o.Qty > 0,
			}
			// Si le consommateur ne lit pas assez vite, ceci bloque
			// quand le buffer (capacite 1) est plein.
			out <- validated
		}
	}()

	return out
}

// ---------------------------------------------------------------------------
// Pipeline Stage 2 : Enrichissement
// ---------------------------------------------------------------------------

// enrichOrders ajoute des donnees supplementaires (notional, etc.)
// BUG #2 : Le select n'a pas de case de timeout.
//          Si 'in' ne produit pas de donnees, bloque indefiniment.
func enrichOrders(in <-chan ValidatedOrder) chan EnrichedOrder {
	out := make(chan EnrichedOrder, 1) // BUG #1 bis

	go func() {
		defer close(out)
		for {
			// BUG #2 : Select sans timeout ni default.
			// Si 'in' est vide ET jamais ferme, bloque ici pour toujours.
			select {
			case v, ok := <-in:
				if !ok {
					return // channel ferme, on sort
				}
				enriched := EnrichedOrder{
					ValidatedOrder: v,
					Notional:       v.Price * float64(v.Qty),
				}
				out <- enriched
			// MANQUE : case <-time.After(5 * time.Second): return
			}
		}
	}()

	return out
}

// ---------------------------------------------------------------------------
// Pipeline Stage 3 : Execution
// ---------------------------------------------------------------------------

// executeOrders simule l'envoi des ordres au marche.
// BUG #3 : closeChannel est appele deux fois si execute() est appele
//          plusieurs fois avec le meme done channel.
func executeOrders(in <-chan EnrichedOrder, done chan struct{}) int {
	count := 0
	for o := range in {
		if o.Valid {
			count++
			fmt.Printf("  [EXEC] Order#%d %s x%d @ $%.2f (notional: $%.2f)\n",
				o.ID, o.Symbol, o.Qty, o.Price, o.Notional)
		} else {
			fmt.Printf("  [SKIP] Order#%d invalide\n", o.ID)
		}
	}
	// BUG #3 : Si executeOrders est appele deux fois avec le meme 'done',
	// close() sera appele deux fois => PANIC: close of closed channel
	close(done)
	return count
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	fmt.Println("=== Pipeline de traitement d'ordres ===")
	fmt.Println()

	// Creer 20 ordres (le buffer de 1 sera rapidement sature)
	orders := make([]Order, 20)
	for i := range orders {
		orders[i] = Order{
			ID:     i + 1,
			Symbol: "AAPL",
			Price:  189.50 + float64(i)*0.01,
			Qty:    100,
		}
	}
	// Ajouter un ordre invalide
	orders[5] = Order{ID: 6, Symbol: "AAPL", Price: -1.0, Qty: 100}

	done := make(chan struct{})

	// Lancer le pipeline
	validated := validateOrders(orders)
	enriched := enrichOrders(validated)

	// Executer en arriere-plan
	var result int
	go func() {
		result = executeOrders(enriched, done)
	}()

	// Attendre la fin
	select {
	case <-done:
		fmt.Printf("\nPipeline termine. %d ordres executes.\n", result)
	case <-time.After(2 * time.Second):
		fmt.Println("\nTIMEOUT : Le pipeline a bloque !")
		fmt.Println("Diagnostique les bugs et corrige le code.")
		return
	}

	// --- Test BUG #3 : Double close ---
	fmt.Println("\n=== Test Double Close ===")
	doneCh := make(chan struct{})

	// Simuler deux appels qui ferment le meme channel
	closeChannel := func(ch chan struct{}) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("PANIC interceptee : %v\n", r)
				fmt.Println("=> BUG #3 : double close d'un channel")
			}
		}()
		close(ch)
	}

	closeChannel(doneCh) // Premier close : OK
	closeChannel(doneCh) // Deuxieme close : PANIC (interceptee pour l'exemple)

	// --- Test BUG #4 : Nil channel ---
	fmt.Println("\n=== Test Nil Channel ===")
	var nilCh chan int // == nil

	// Ceci bloquerait indefiniment si on ne met pas de timeout
	go func() {
		fmt.Println("Tentative de lecture depuis nil channel...")
		drainChannel(nilCh) // BUG #4 : ne retournera jamais
		fmt.Println("(ce message n'apparaitra jamais)")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("La goroutine ci-dessus est bloquee indefiniment sur nil channel.")
	fmt.Println("\nCorrige les 4 bugs. Le programme doit se terminer proprement.")
}
