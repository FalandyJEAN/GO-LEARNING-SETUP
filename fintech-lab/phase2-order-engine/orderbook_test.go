// orderbook_test.go — Tests unitaires du matching engine.
// Lancer avec : go test ./phase2-order-engine/ -v
// Avec couverture : go test ./phase2-order-engine/ -v -cover
// Benchmark  : go test ./phase2-order-engine/ -bench=. -benchmem

package main

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers de test
// ---------------------------------------------------------------------------

// mustSubmit soumet un ordre et fait echouer le test si erreur.
func mustSubmit(t *testing.T, gw *Gateway, o *Order) []Trade {
	t.Helper()
	trades, err := gw.Submit(o)
	if err != nil {
		t.Fatalf("Submit inattendu: %v", err)
	}
	return trades
}

func newTestGateway() (*Gateway, *TradeLog) {
	log := NewTradeLog()
	gw := NewGateway([]string{"AAPL", "MSFT"}, log)
	return gw, log
}

// ---------------------------------------------------------------------------
// Tests de base
// ---------------------------------------------------------------------------

// TestNoMatch verifie qu'un ordre sans contrepartie reste dans le book.
func TestNoMatch(t *testing.T) {
	gw, _ := newTestGateway()

	buy := NewLimitOrder("AAPL", Buy, 189.00, 100)
	trades := mustSubmit(t, gw, buy)

	if len(trades) != 0 {
		t.Errorf("attendu 0 trades, obtenu %d", len(trades))
	}
	if buy.Status != StatusOpen {
		t.Errorf("attendu StatusOpen, obtenu %s", buy.Status)
	}
}

// TestFullMatch verifie un match complet entre un acheteur et un vendeur.
func TestFullMatch(t *testing.T) {
	gw, log := newTestGateway()

	sell := NewLimitOrder("AAPL", Sell, 189.00, 100)
	mustSubmit(t, gw, sell)

	buy := NewLimitOrder("AAPL", Buy, 189.50, 100)
	trades := mustSubmit(t, gw, buy)

	if len(trades) != 1 {
		t.Fatalf("attendu 1 trade, obtenu %d", len(trades))
	}

	trade := trades[0]
	if trade.Quantity != 100 {
		t.Errorf("quantite attendue: 100, obtenu: %d", trade.Quantity)
	}
	// Le prix est celui du vendeur (ordre passif)
	if trade.Price != 189.00 {
		t.Errorf("prix attendu: 189.00, obtenu: %.2f", trade.Price)
	}
	if sell.Status != StatusFilled {
		t.Errorf("sell status attendu: FILLED, obtenu: %s", sell.Status)
	}
	if buy.Status != StatusFilled {
		t.Errorf("buy status attendu: FILLED, obtenu: %s", buy.Status)
	}
	if log.Count() != 1 {
		t.Errorf("trade log attendu: 1, obtenu: %d", log.Count())
	}
}

// TestPartialMatch verifie un match partiel.
func TestPartialMatch(t *testing.T) {
	gw, _ := newTestGateway()

	sell := NewLimitOrder("AAPL", Sell, 189.00, 50) // Vend 50
	mustSubmit(t, gw, sell)

	buy := NewLimitOrder("AAPL", Buy, 189.50, 100) // Veut 100
	trades := mustSubmit(t, gw, buy)

	if len(trades) != 1 {
		t.Fatalf("attendu 1 trade, obtenu %d", len(trades))
	}
	if trades[0].Quantity != 50 {
		t.Errorf("quantite attendue: 50, obtenu: %d", trades[0].Quantity)
	}
	if sell.Status != StatusFilled {
		t.Errorf("sell: attendu FILLED, obtenu %s", sell.Status)
	}
	// L'acheteur n'a recu que 50 sur 100 => PARTIAL, reste dans le book
	if buy.Status != StatusPartial {
		t.Errorf("buy: attendu PARTIAL, obtenu %s", buy.Status)
	}
	if buy.Filled != 50 {
		t.Errorf("buy.Filled attendu: 50, obtenu: %d", buy.Filled)
	}
}

// TestPricePriority verifie que le meilleur prix est execute en premier.
func TestPricePriority(t *testing.T) {
	gw, _ := newTestGateway()

	// Deux vendeurs : l'un moins cher doit etre execute en premier
	sellExpensive := NewLimitOrder("AAPL", Sell, 191.00, 100)
	sellCheap := NewLimitOrder("AAPL", Sell, 190.00, 100)

	mustSubmit(t, gw, sellExpensive) // Soumis en premier
	mustSubmit(t, gw, sellCheap)     // Soumis en second mais moins cher

	buy := NewLimitOrder("AAPL", Buy, 191.00, 100)
	trades := mustSubmit(t, gw, buy)

	if len(trades) != 1 {
		t.Fatalf("attendu 1 trade, obtenu %d", len(trades))
	}
	// Doit matcher avec le vendeur le moins cher (190.00)
	if trades[0].Price != 190.00 {
		t.Errorf("attendu prix 190.00, obtenu %.2f — Price Priority violation!", trades[0].Price)
	}
	if trades[0].SellOrderID != sellCheap.ID {
		t.Errorf("attendu match avec sellCheap (#%d), obtenu #%d", sellCheap.ID, trades[0].SellOrderID)
	}
}

// TestFIFOPriority verifie la priorite temps (FIFO) a prix egal.
func TestFIFOPriority(t *testing.T) {
	gw, _ := newTestGateway()

	// Deux vendeurs au meme prix : le premier arrive doit etre execute en premier
	sell1 := NewLimitOrder("AAPL", Sell, 190.00, 100) // arrive T=1
	sell2 := NewLimitOrder("AAPL", Sell, 190.00, 100) // arrive T=2

	// Forcer un timestamp different pour garantir l'ordre
	sell2.Timestamp = sell1.Timestamp + 1000

	mustSubmit(t, gw, sell1)
	mustSubmit(t, gw, sell2)

	buy := NewLimitOrder("AAPL", Buy, 190.00, 100)
	trades := mustSubmit(t, gw, buy)

	if len(trades) != 1 {
		t.Fatalf("attendu 1 trade, obtenu %d", len(trades))
	}
	// Doit matcher avec sell1 (FIFO)
	if trades[0].SellOrderID != sell1.ID {
		t.Errorf("FIFO violation: attendu sell1 (#%d), obtenu #%d", sell1.ID, trades[0].SellOrderID)
	}
}

// TestMarketOrder verifie qu'un Market order s'execute au meilleur prix dispo.
func TestMarketOrder(t *testing.T) {
	gw, _ := newTestGateway()

	mustSubmit(t, gw, NewLimitOrder("AAPL", Sell, 190.00, 100))
	mustSubmit(t, gw, NewLimitOrder("AAPL", Sell, 191.00, 100))

	marketBuy := NewMarketOrder("AAPL", Buy, 100)
	trades := mustSubmit(t, gw, marketBuy)

	if len(trades) != 1 {
		t.Fatalf("attendu 1 trade, obtenu %d", len(trades))
	}
	if trades[0].Price != 190.00 {
		t.Errorf("Market order: attendu prix 190.00, obtenu %.2f", trades[0].Price)
	}
}

// TestCancelOrder verifie l'annulation d'un ordre.
func TestCancelOrder(t *testing.T) {
	gw, _ := newTestGateway()

	order := NewLimitOrder("AAPL", Buy, 189.00, 100)
	mustSubmit(t, gw, order)

	err := gw.Cancel("AAPL", order.ID)
	if err != nil {
		t.Fatalf("Cancel inattendu: %v", err)
	}
	if order.Status != StatusCancelled {
		t.Errorf("attendu CANCELLED, obtenu %s", order.Status)
	}

	// Verifier qu'un sell apres annulation ne matche pas avec l'ordre annule
	sell := NewLimitOrder("AAPL", Sell, 189.00, 100)
	trades := mustSubmit(t, gw, sell)
	if len(trades) != 0 {
		t.Errorf("l'ordre annule ne doit pas etre matche, obtenu %d trades", len(trades))
	}
}

// TestValidation verifie les rejets d'ordres invalides.
func TestValidation(t *testing.T) {
	gw, _ := newTestGateway()

	cases := []struct {
		name  string
		order *Order
	}{
		{"prix negatif", NewLimitOrder("AAPL", Buy, -1, 100)},
		{"quantite zero", NewLimitOrder("AAPL", Buy, 189.0, 0)},
		{"symbole inconnu", NewLimitOrder("GOOG", Buy, 150.0, 10)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gw.Submit(tc.order)
			if err == nil {
				t.Errorf("attendu une erreur pour %q, aucune obtenue", tc.name)
			}
		})
	}
}

// TestSpread verifie le calcul du spread.
func TestSpread(t *testing.T) {
	gw, _ := newTestGateway()

	mustSubmit(t, gw, NewLimitOrder("AAPL", Buy, 189.00, 100))
	mustSubmit(t, gw, NewLimitOrder("AAPL", Sell, 190.00, 100))

	spread, ok := gw.books["AAPL"].Spread()
	if !ok {
		t.Fatal("spread non disponible")
	}
	expected := 1.00
	if spread != expected {
		t.Errorf("spread attendu: %.2f, obtenu: %.2f", expected, spread)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks — Critique pour valider la performance en HFT
// Lancer : go test ./phase2-order-engine/ -bench=. -benchmem -count=5
// ---------------------------------------------------------------------------

// BenchmarkSubmitNoMatch mesure le cout de soumission sans match.
func BenchmarkSubmitNoMatch(b *testing.B) {
	gw, _ := newTestGateway()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		o := NewLimitOrder("AAPL", Buy, 189.00, 100)
		gw.Submit(o) //nolint
	}
}

// BenchmarkSubmitWithMatch mesure le cout d'un match complet.
func BenchmarkSubmitWithMatch(b *testing.B) {
	gw, _ := newTestGateway()

	// Pre-remplir le book avec des vendeurs
	for i := 0; i < 1000; i++ {
		gw.Submit(NewLimitOrder("AAPL", Sell, 190.00+float64(i)*0.01, 100)) //nolint
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Acheteur agressif qui matche avec le meilleur vendeur
		gw.Submit(NewLimitOrder("AAPL", Buy, 200.00, 100)) //nolint
		// Remettre un vendeur pour le prochain tour
		gw.Submit(NewLimitOrder("AAPL", Sell, 190.00, 100)) //nolint
	}
}
