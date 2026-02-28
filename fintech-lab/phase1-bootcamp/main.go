// main.go — Script de validation de l'environnement Go
// Objectif : confirmer que Go est correctement installe et configure.
// Commande : go run phase1-bootcamp/main.go

package main

import (
	"fmt"
	"math"
	"runtime"
	"time"
)

// Order represente un ordre de bourse simplifie.
// C'est la structure de donnees centrale de tout notre projet.
type Order struct {
	ID        string
	Symbol    string
	Side      string  // "BUY" ou "SELL"
	Price     float64
	Quantity  int
	Timestamp time.Time
}

// String implemente l'interface fmt.Stringer.
// En Go, si un type a une methode String() string, fmt l'utilise automatiquement.
func (o Order) String() string {
	return fmt.Sprintf("[%s] %s %s x%d @ %.2f",
		o.ID, o.Side, o.Symbol, o.Quantity, o.Price)
}

// calcMidPrice calcule le prix median entre bid et ask.
// Retourne deux valeurs : le resultat et une erreur potentielle.
// C'est le pattern idiomatique Go pour la gestion d'erreurs.
func calcMidPrice(bid, ask float64) (float64, error) {
	if bid <= 0 || ask <= 0 {
		return 0, fmt.Errorf("prix invalide: bid=%.2f ask=%.2f", bid, ask)
	}
	if bid >= ask {
		return 0, fmt.Errorf("bid (%.2f) doit etre inferieur a ask (%.2f)", bid, ask)
	}
	return (bid + ask) / 2.0, nil
}

func main() {
	fmt.Println("================================================================================")
	fmt.Println("  FINTECH LAB — Validation de l'environnement")
	fmt.Println("================================================================================")
	fmt.Println()

	// --- [1] Informations systeme ---
	fmt.Printf("Go Version   : %s\n", runtime.Version())
	fmt.Printf("OS / Arch    : %s / %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPU Logiques : %d\n", runtime.NumCPU())
	fmt.Println()

	// --- [2] Types de base et zero values ---
	// En Go, chaque type a une "zero value" par defaut (pas de null).
	var orderID string    // zero value = ""
	var price float64     // zero value = 0.0
	var quantity int      // zero value = 0
	var isActive bool     // zero value = false

	fmt.Println("[Types de base - Zero Values]")
	fmt.Printf("  string  : %q\n", orderID)
	fmt.Printf("  float64 : %v\n", price)
	fmt.Printf("  int     : %v\n", quantity)
	fmt.Printf("  bool    : %v\n", isActive)
	fmt.Println()

	// --- [3] Struct et methode String() ---
	order := Order{
		ID:        "ORD-001",
		Symbol:    "AAPL",
		Side:      "BUY",
		Price:     189.50,
		Quantity:  100,
		Timestamp: time.Now(),
	}
	fmt.Println("[Struct Order]")
	fmt.Printf("  %v\n", order)
	fmt.Println()

	// --- [4] Slice et boucle range ---
	portfolio := []Order{
		{ID: "ORD-001", Symbol: "AAPL", Side: "BUY",  Price: 189.50, Quantity: 100},
		{ID: "ORD-002", Symbol: "MSFT", Side: "SELL", Price: 415.20, Quantity: 50},
		{ID: "ORD-003", Symbol: "TSLA", Side: "BUY",  Price: 172.80, Quantity: 200},
	}
	fmt.Println("[Portfolio - Slice d'ordres]")
	totalValue := 0.0
	for i, o := range portfolio {
		notional := o.Price * float64(o.Quantity)
		totalValue += notional
		fmt.Printf("  [%d] %v | Notional: $%.2f\n", i, o, notional)
	}
	fmt.Printf("  Total Notional : $%.2f\n", totalValue)
	fmt.Println()

	// --- [5] Map ---
	bidAsk := map[string][2]float64{
		"AAPL": {189.45, 189.55},
		"MSFT": {415.10, 415.30},
		"TSLA": {172.75, 172.85},
	}
	fmt.Println("[Carnet d'ordres - Bid/Ask]")
	for symbol, prices := range bidAsk {
		mid, err := calcMidPrice(prices[0], prices[1])
		if err != nil {
			fmt.Printf("  %s : ERREUR - %v\n", symbol, err)
			continue
		}
		spread := prices[1] - prices[0]
		fmt.Printf("  %s : Bid=%.2f | Ask=%.2f | Mid=%.4f | Spread=%.4f\n",
			symbol, prices[0], prices[1], mid, spread)
	}
	fmt.Println()

	// --- [6] Pointeurs ---
	// Comprendre les pointeurs est critique pour la performance en Go.
	px := 42
	ptr := &px       // ptr pointe vers px
	*ptr = 100       // modifie px via le pointeur
	fmt.Println("[Pointeurs]")
	fmt.Printf("  Valeur de px    : %d\n", px)
	fmt.Printf("  Adresse de px   : %p\n", ptr)
	fmt.Printf("  Valeur via *ptr : %d\n", *ptr)
	fmt.Println()

	// --- [7] Gestion d'erreurs ---
	fmt.Println("[Gestion d'erreurs - Pattern Go]")
	testCases := [][2]float64{
		{100.0, 100.5},  // valide
		{0.0, 100.5},    // bid invalide
		{101.0, 100.0},  // bid > ask
	}
	for _, tc := range testCases {
		mid, err := calcMidPrice(tc[0], tc[1])
		if err != nil {
			fmt.Printf("  ERR [bid=%.1f ask=%.1f] : %v\n", tc[0], tc[1], err)
		} else {
			fmt.Printf("  OK  [bid=%.1f ask=%.1f] : mid=%.4f\n", tc[0], tc[1], mid)
		}
	}
	fmt.Println()

	// --- [8] Math de base (utile pour la finance) ---
	fmt.Println("[Calculs financiers de base]")
	volatility := 0.25
	spotPrice  := 189.50
	strike     := 190.00
	timeToExp  := 30.0 / 365.0
	_ = math.Sqrt(volatility * timeToExp) // On va en avoir besoin en Phase 2
	fmt.Printf("  Spot: %.2f | Strike: %.2f | Vol: %.0f%% | T: %.4f ans\n",
		spotPrice, strike, volatility*100, timeToExp)
	fmt.Println()

	fmt.Println("================================================================================")
	fmt.Println("  ENVIRONNEMENT OK — Tu peux passer a l'Examen 01.")
	fmt.Println("  Lis : phase1-bootcamp/exams/exam01_instructions.txt")
	fmt.Println("================================================================================")
}
