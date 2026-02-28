// exam01_syntax.go — EXAMEN 01 : "Le Premier Ordre qui Crashe"
// ATTENTION : Ce code compile SANS erreur mais contient des bugs.
// Ton objectif : trouver et corriger les 4 problemes.
// NE PAS modifier les signatures des fonctions ou la structure generale.

package main

import "fmt"

// Side represente le cote d'un ordre boursier.
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// Order represente un ordre soumis au systeme de matching.
type Order struct {
	ID       string
	Symbol   string
	Side     Side
	Price    float64
	Quantity int
}

// validateOrder verifie qu'un ordre est valide avant envoi au marche.
// Retourne nil si l'ordre est valide, une erreur sinon.
func validateOrder(o Order) error {
	if o.Price <= 0 {
		return fmt.Errorf("prix invalide (%.2f)", o.Price)
	}
	if o.Quantity < 0 {
		return fmt.Errorf("quantite invalide (%d)", o.Quantity)
	}
	return nil
}

// calcPnL calcule le Profit & Loss d'une position.
//   - symbol : le ticker (ex: "AAPL")
//   - isLong : true si position longue (achat), false si courte (vente a decouvert)
//   - qty    : nombre d'actions
//   - entry  : prix d'entree
//   - exit   : prix de sortie
//
// Pour une position LONG  : P&L = (exit - entry) * qty
// Pour une position SHORT : P&L = (entry - exit) * qty
func calcPnL(symbol string, isLong bool, qty int, entry, exit float64) float64 {
	var pnl float64
	if isLong {
		pnl = (entry - exit) * float64(qty) // BUG #3 ICI
	} else {
		pnl = (entry - exit) * float64(qty)
	}
	return pnl
}

// applyDiscount applique une remise de courtage sur le prix d'un ordre.
// La remise est en pourcentage (ex: 0.1 = 10%).
// La fonction modifie le prix directement.
func applyDiscount(o *Order, discountPct float64) {
	discounted := o.Price * (1 - discountPct)
	o = &Order{  // BUG #4 ICI
		ID:       o.ID,
		Symbol:   o.Symbol,
		Side:     o.Side,
		Price:    discounted,
		Quantity: o.Quantity,
	}
	_ = o // evite l'erreur "declared and not used"
}

func main() {
	orders := []Order{
		{ID: "ORD-001", Symbol: "AAPL", Side: Buy,  Price: 189.50, Quantity: 100},
		{ID: "ORD-002", Symbol: "AAPL", Side: Sell, Price: 0.00,   Quantity: 50},
		{ID: "ORD-003", Symbol: "MSFT", Side: Buy,  Price: 415.20, Quantity: -50}, // BUG #2 non intercepte
		{ID: "ORD-004", Symbol: "MSFT", Side: Sell, Price: 415.20, Quantity: 200},
	}

	fmt.Println("--- Validation des ordres ---")
	for _, o := range orders {
		err := validateOrder(o)
		if err != nil {
			fmt.Printf("[ERR] %s : %v\n", o.ID, err)
		} else {
			fmt.Printf("[OK]  %s : %s %s x%d @ $%.2f\n",
				o.ID, o.Side, o.Symbol, o.Quantity, o.Price)
		}
	}

	fmt.Println()
	fmt.Println("--- Calcul P&L ---")

	// Position LONG AAPL : achat a 189.50, revente a 192.00
	// 100 actions => P&L attendu = (192.00 - 189.50) * 100 = +$250.00
	pnlLong := calcPnL("AAPL", true, 100, 189.50, 192.00)
	fmt.Printf("Position LONG  AAPL : entree=189.50, sortie=192.00 => P&L = %+.2f\n",
		pnlLong * 100) // ATTENTION a ce calcul

	// Position SHORT MSFT : vente a 415.20, rachat a 410.00
	// 100 actions => P&L attendu = (415.20 - 410.00) * 100 = +$520.00
	pnlShort := calcPnL("MSFT", false, 100, 415.20, 410.00)
	fmt.Printf("Position SHORT MSFT : entree=415.20, sortie=410.00 => P&L = %+.2f\n",
		pnlShort * 100) // ATTENTION a ce calcul
}
