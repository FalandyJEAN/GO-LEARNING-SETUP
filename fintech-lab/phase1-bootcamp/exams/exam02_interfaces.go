// exam02_interfaces.go — EXAMEN 02 : "L'Interface qui Ment"
// Ce code compile sans erreur mais contient 4 bugs runtime/logiques.
// Lancer avec : go run phase1-bootcamp/exams/exam02_interfaces.go

package main

import "fmt"

// ---------------------------------------------------------------------------
// Types de base
// ---------------------------------------------------------------------------

// Validator est l'interface centrale du systeme de validation.
type Validator interface {
	Validate() error
	String() string
}

// Pricer est implementee par tout ordre qui a un prix calculable.
type Pricer interface {
	Price() float64
	Fee() float64
}

// LimitOrder est un ordre a cours limite.
type LimitOrder struct {
	Symbol   string
	Side     string
	Price_   float64
	Quantity int
}

// StopOrder est un ordre stop-loss.
// Il implemente partiellement Validator : Validate() est defini,
// mais String() est volontairement absent pour forcer la type assertion dans processOrder().
// Note : pour le compiler, on ajoute une String() minimale qui ne sert pas le bon objectif.
type StopOrder struct {
	Symbol    string
	StopPrice float64
}

func (s StopOrder) Validate() error {
	if s.StopPrice <= 0 {
		return fmt.Errorf("stop price invalide: %.2f", s.StopPrice)
	}
	return nil
}

func (s StopOrder) String() string {
	return fmt.Sprintf("STOP %s @ $%.2f", s.Symbol, s.StopPrice)
}

// ---------------------------------------------------------------------------
// Implementation de Validator pour LimitOrder
// ---------------------------------------------------------------------------

func (o LimitOrder) Validate() error {
	if o.Price_ <= 0 {
		return fmt.Errorf("prix invalide: %.2f", o.Price_)
	}
	if o.Quantity <= 0 {
		return fmt.Errorf("quantite invalide: %d", o.Quantity)
	}
	return nil
}

func (o LimitOrder) String() string {
	return fmt.Sprintf("%s x%d @ $%.2f", o.Symbol, o.Quantity, o.Price_)
}

// ---------------------------------------------------------------------------
// BUG #4 : Fee() est definie sur LimitOrder (valeur), mais dans processOrder()
// on passe un *LimitOrder. La methode Fee() est appelee sur une copie et
// retourne toujours 0 car la logique est dans le mauvais receiver.
// De plus, la formule de fee est incorrecte.
//
// Fee = 0.2% du notional (Price * Quantity)
// Exemple : 189.50 * 100 * 0.002 = $37.90
// ---------------------------------------------------------------------------
func (o LimitOrder) Fee() float64 {
	// BUG : le calcul est fait sur o (copie), et la formule est fausse.
	// Ici on multiplie Price par Fee au lieu de Price * Quantity * 0.002
	return o.Price_ * 0.002 // Manque * float64(o.Quantity)
}

func (o LimitOrder) Price() float64 {
	return o.Price_
}

// ---------------------------------------------------------------------------
// BUG #1 : SetDefaults utilise un value receiver.
// Elle travaille sur une COPIE de l'ordre, pas sur l'original.
// La modification de Quantity est perdue apres le retour de la fonction.
// ---------------------------------------------------------------------------
func (o LimitOrder) SetDefaults() {
	if o.Quantity == 0 {
		o.Quantity = 1
	}
}

// ---------------------------------------------------------------------------
// getDefaultValidator retourne un Validator pour les ordres vides.
// BUG #3 : Cette fonction retourne une interface non-nil contenant un
// pointeur nil. Le test "validator != nil" passera, mais appeler
// validator.Validate() provoquera un nil pointer dereference.
// ---------------------------------------------------------------------------
func getDefaultValidator() Validator {
	var o *LimitOrder = nil
	return o // BUG : interface{type=*LimitOrder, value=nil} != nil
}

// ---------------------------------------------------------------------------
// processOrder traite un Validator et tente d'en extraire les frais.
// BUG #2 : La type assertion n'utilise pas la forme securisee (v, ok).
// Si v n'est pas un *LimitOrder (ex: StopOrder), le programme panique.
// ---------------------------------------------------------------------------
func processOrder(v Validator) {
	err := v.Validate()
	if err != nil {
		fmt.Printf("  Validation echouee : %v\n", err)
		return
	}

	// BUG #2 : panic si v est un StopOrder !
	lo := v.(*LimitOrder)
	fmt.Printf("  [OK]  %s | fee=$%.2f\n", lo.String(), lo.Fee())
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// --- Test BUG #1 : SetDefaults ---
	fmt.Println("--- Test SetDefaults ---")
	order := LimitOrder{Symbol: "AAPL", Side: "BUY", Price_: 189.50, Quantity: 0}
	fmt.Printf("Avant : Quantity=%d\n", order.Quantity)
	order.SetDefaults()
	fmt.Printf("Apres : Quantity=%d\n", order.Quantity) // Attendu: 1, Obtenu: 0
	fmt.Println()

	// --- Test BUG #2 : Type assertion ---
	fmt.Println("--- Test Type Assertion ---")
	validators := []Validator{
		&LimitOrder{Symbol: "AAPL", Side: "BUY", Price_: 189.50, Quantity: 100},
		StopOrder{Symbol: "MSFT", StopPrice: 410.00}, // Pas un *LimitOrder !
	}
	for _, v := range validators {
		processOrder(v) // Panique sur le StopOrder
	}
	fmt.Println()

	// --- Test BUG #3 : Nil trap ---
	fmt.Println("--- Test Nil Trap ---")
	validator := getDefaultValidator()
	if validator != nil {
		// On pense que c'est safe, mais ce n'est pas le cas !
		err := validator.Validate() // Nil pointer dereference !
		if err != nil {
			fmt.Printf("Erreur de validation : %v\n", err)
		}
	} else {
		fmt.Println("Validator correctement absent (nil detecte)")
	}
	fmt.Println()

	// --- Test BUG #4 : Fee incorrecte ---
	fmt.Println("--- Validation Complete ---")
	orders := []*LimitOrder{
		{Symbol: "AAPL", Side: "BUY",  Price_: 189.50, Quantity: 100},
		{Symbol: "MSFT", Side: "SELL", Price_: 415.20, Quantity: 50},
	}
	for _, o := range orders {
		if err := o.Validate(); err != nil {
			fmt.Printf("[ERR] %s : %v\n", o.Symbol, err)
			continue
		}
		// Attendu AAPL: fee = 189.50 * 100 * 0.002 = $37.90
		// Obtenu  AAPL: fee = 189.50 * 0.002 = $0.379 (sans la quantite)
		fmt.Printf("[OK]  %s x%d @ $%.2f | Fee: $%.2f\n",
			o.Symbol, o.Quantity, o.Price_,
			o.Fee()) // BUG #4 : valeur incorrecte
	}
}
