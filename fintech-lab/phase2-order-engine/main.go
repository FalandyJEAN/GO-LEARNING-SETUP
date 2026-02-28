// main.go — Simulation du matching engine GoMatchEngine (GME).
// Lancer avec : go run ./phase2-order-engine/

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║     GoMatchEngine (GME) — Simulation Session     ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	// --- Initialisation ---
	tradeLog := NewTradeLog()
	gw := NewGateway([]string{"AAPL", "MSFT", "TSLA"}, tradeLog)

	// =========================================================================
	// SCENARIO 1 : Construction du carnet d'ordres AAPL
	// On soumet des ordres qui ne se croisent pas encore.
	// =========================================================================
	fmt.Println("━━━ SCENARIO 1 : Construction du carnet AAPL ━━━")
	fmt.Println()

	submitAndPrint(gw, NewLimitOrder("AAPL", Buy, 189.00, 200))
	submitAndPrint(gw, NewLimitOrder("AAPL", Buy, 189.50, 100))
	submitAndPrint(gw, NewLimitOrder("AAPL", Buy, 188.50, 300))
	submitAndPrint(gw, NewLimitOrder("AAPL", Sell, 190.50, 150))
	submitAndPrint(gw, NewLimitOrder("AAPL", Sell, 191.00, 200))
	submitAndPrint(gw, NewLimitOrder("AAPL", Sell, 190.00, 100))

	fmt.Println()
	if book, ok := gw.Book("AAPL"); ok {
		book.PrintBook(5)
	}

	bid, _ := gw.books["AAPL"].BestBid()
	ask, _ := gw.books["AAPL"].BestAsk()
	spread, _ := gw.books["AAPL"].Spread()
	fmt.Printf("\nTop of Book : Bid=$%.2f | Ask=$%.2f | Spread=$%.4f\n\n", bid, ask, spread)

	// =========================================================================
	// SCENARIO 2 : Matching — Un acheteur agressif croise le ask
	// Un BUY @ $191.00 va matcher contre les SELL @ $190.00 et $190.50
	// =========================================================================
	fmt.Println("━━━ SCENARIO 2 : Matching agressif (BUY @ $191.00 x250) ━━━")
	fmt.Println()

	aggressiveBuy := NewLimitOrder("AAPL", Buy, 191.00, 250)
	fmt.Printf("Soumission : %v\n", aggressiveBuy)

	trades, err := gw.Submit(aggressiveBuy)
	if err != nil {
		fmt.Printf("ERREUR : %v\n", err)
	} else {
		fmt.Printf("Status : %s | Filled: %d/%d\n", aggressiveBuy.Status, aggressiveBuy.Filled, aggressiveBuy.Quantity)
		fmt.Printf("Trades executes : %d\n", len(trades))
		for _, t := range trades {
			fmt.Printf("  => %v\n", t)
		}
	}

	fmt.Println()
	if book, ok := gw.Book("AAPL"); ok {
		book.PrintBook(5)
	}

	// =========================================================================
	// SCENARIO 3 : Market Order
	// Un vendeur urgence veut vendre 100 AAPL au meilleur prix disponible.
	// =========================================================================
	fmt.Println()
	fmt.Println("━━━ SCENARIO 3 : Market Order SELL x100 ━━━")
	fmt.Println()

	marketSell := NewMarketOrder("AAPL", Sell, 100)
	fmt.Printf("Soumission : %v\n", marketSell)

	trades, err = gw.Submit(marketSell)
	if err != nil {
		fmt.Printf("ERREUR : %v\n", err)
	} else {
		fmt.Printf("Status : %s | Filled: %d/%d\n", marketSell.Status, marketSell.Filled, marketSell.Quantity)
		for _, t := range trades {
			fmt.Printf("  => %v\n", t)
		}
	}

	// =========================================================================
	// SCENARIO 4 : Annulation d'ordre
	// =========================================================================
	fmt.Println()
	fmt.Println("━━━ SCENARIO 4 : Annulation d'un ordre ━━━")
	fmt.Println()

	cancelTarget := NewLimitOrder("AAPL", Buy, 188.50, 500)
	submitAndPrint(gw, cancelTarget)

	time.Sleep(1 * time.Millisecond) // Simuler du temps

	err = gw.Cancel("AAPL", cancelTarget.ID)
	if err != nil {
		fmt.Printf("Echec annulation ordre #%d : %v\n", cancelTarget.ID, err)
	} else {
		fmt.Printf("Ordre #%d annule. Status : %s\n", cancelTarget.ID, cancelTarget.Status)
	}

	// =========================================================================
	// SCENARIO 5 : Validation — Rejets d'ordres invalides
	// =========================================================================
	fmt.Println()
	fmt.Println("━━━ SCENARIO 5 : Ordres invalides ━━━")
	fmt.Println()

	invalidOrders := []*Order{
		NewLimitOrder("AAPL", Buy, -10.0, 100),    // Prix negatif
		NewLimitOrder("AAPL", Buy, 189.50, 0),      // Quantite nulle
		NewLimitOrder("UNKNOWN", Buy, 100.0, 50),   // Symbole inconnu
		NewLimitOrder("AAPL", "HOLD", 189.50, 100), // Side invalide
	}

	for _, o := range invalidOrders {
		_, err := gw.Submit(o)
		if err != nil {
			fmt.Printf("  [REJETE] %v\n", err)
		}
	}

	// =========================================================================
	// RESUME FINAL
	// =========================================================================
	fmt.Println()
	fmt.Println("━━━ RESUME DE SESSION ━━━")
	fmt.Println()
	tradeLog.PrintSummary()
	fmt.Println()

	bids, asks := gw.books["AAPL"].Depth()
	fmt.Printf("AAPL Book Depth : %d bids | %d asks\n", bids, asks)
}

// submitAndPrint est un helper pour soumettre un ordre et afficher le resultat.
func submitAndPrint(gw *Gateway, o *Order) {
	trades, err := gw.Submit(o)
	if err != nil {
		fmt.Printf("  [ERR] %v\n", err)
		return
	}
	if len(trades) > 0 {
		for _, t := range trades {
			fmt.Printf("  [MATCH] %v\n", t)
		}
	} else {
		fmt.Printf("  [BOOK]  Ordre #%d ajoute : %v\n", o.ID, o)
	}
}
