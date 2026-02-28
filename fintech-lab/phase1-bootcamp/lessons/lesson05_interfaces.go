// ================================================================================
// LECON 05 — Interfaces
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson05_interfaces.go
// OBJECTIF         :  Comprendre le polymorphisme Go style (sans heritage)
// PROCHAINE LECON  :  lesson06_slices_maps.go
// ================================================================================
// NOTE : Cette lecon prepare exam02_interfaces.go
// ================================================================================

package main

import (
	"fmt"
	"math"
)

// ============================================================
// PARTIE 1 : QU'EST-CE QU'UNE INTERFACE ?
// ============================================================
// Une interface definit un CONTRAT : "tout type qui a ces methodes
// satisfait cette interface".
//
// En Go, l'implementation est IMPLICITE.
// Il n'y a pas de "implements MonInterface" comme en Java.
// Si un type a toutes les methodes de l'interface -> il la satisfait.

// Instrument financier : peut etre calcule et affiche
type Instrument interface {
	NomSymbol() string    // retourne le ticker
	PrixTheorique() float64 // retourne le prix theorique
	Description() string  // retourne une description
}

// ============================================================
// PARTIE 2 : IMPLEMENTER UNE INTERFACE
// ============================================================
// On cree plusieurs types qui implementent tous Instrument.

// Action (Stock)
type Action struct {
	Symbol    string
	PrixSpot  float64
	Dividende float64
}

func (a Action) NomSymbol() string       { return a.Symbol }
func (a Action) PrixTheorique() float64  { return a.PrixSpot }
func (a Action) Description() string {
	return fmt.Sprintf("Action %s @ %.2f (div: %.2f)", a.Symbol, a.PrixSpot, a.Dividende)
}

// Option (Call ou Put)
type Option struct {
	Symbol    string
	SousJacent float64 // prix du sous-jacent
	Strike    float64  // prix d'exercice
	Maturite  float64  // en annees (ex: 0.5 = 6 mois)
	EstCall   bool     // true = Call, false = Put
}

func (o Option) NomSymbol() string { return o.Symbol }

// Formule de Black-Scholes simplifiee (approximation pour l'exemple)
func (o Option) PrixTheorique() float64 {
	// Valeur intrinseque (version tres simplifiee)
	if o.EstCall {
		return math.Max(0, o.SousJacent-o.Strike)
	}
	return math.Max(0, o.Strike-o.SousJacent)
}

func (o Option) Description() string {
	typeOption := "PUT"
	if o.EstCall {
		typeOption = "CALL"
	}
	return fmt.Sprintf("Option %s %s Strike=%.0f Maturite=%.1fY Prix=%.2f",
		typeOption, o.Symbol, o.Strike, o.Maturite, o.PrixTheorique())
}

// Future
type Future struct {
	Symbol    string
	PrixSpot  float64
	TauxInteret float64 // taux sans risque
	Maturite  float64   // en annees
}

func (f Future) NomSymbol() string { return f.Symbol }

// Prix futur = Spot * e^(r*T) (formule simplifiee)
func (f Future) PrixTheorique() float64 {
	return f.PrixSpot * math.Exp(f.TauxInteret*f.Maturite)
}

func (f Future) Description() string {
	return fmt.Sprintf("Future %s Spot=%.2f r=%.2f%% T=%.1fY -> Prix=%.2f",
		f.Symbol, f.PrixSpot, f.TauxInteret*100, f.Maturite, f.PrixTheorique())
}

// ============================================================
// PARTIE 3 : UTILISER L'INTERFACE
// ============================================================
// Une fonction qui prend une Interface peut recevoir N'IMPORTE
// quel type qui satisfait cette interface.
// C'est le polymorphisme Go.

func afficherInstrument(inst Instrument) {
	fmt.Printf("  [%s] Prix=%.4f | %s\n",
		inst.NomSymbol(),
		inst.PrixTheorique(),
		inst.Description())
}

func calculerValeurPortefeuille(instruments []Instrument, quantites []int) float64 {
	total := 0.0
	for i, inst := range instruments {
		total += inst.PrixTheorique() * float64(quantites[i])
	}
	return total
}

// ============================================================
// PARTIE 4 : TYPE ASSERTION
// ============================================================
// Parfois on a une Interface mais on veut acceder aux champs
// specifiques d'un type concret. On utilise une type assertion.
//
//   concret, ok := interface.(TypeConcret)
//   Si l'interface contient bien un TypeConcret -> ok = true

func analyserInstrument(inst Instrument) {
	switch v := inst.(type) {
	case Action:
		fmt.Printf("  Action detectee : dividende = %.2f\n", v.Dividende)
	case Option:
		if v.EstCall {
			fmt.Println("  Option Call detectee")
		} else {
			fmt.Println("  Option Put detectee")
		}
	case Future:
		fmt.Printf("  Future detecte : taux = %.2f%%\n", v.TauxInteret*100)
	default:
		fmt.Println("  Type inconnu")
	}
}

// ============================================================
// PARTIE 5 : INTERFACE COMPOSEE
// ============================================================
// On peut combiner des interfaces.

type InstrumentNegociable interface {
	Instrument
	EstLiquide() bool // ajoute une methode supplementaire
}

// Action implementera aussi InstrumentNegociable
func (a Action) EstLiquide() bool {
	return a.PrixSpot > 1.0 // simplification : liquide si prix > 1$
}

func main() {

	// --- Creation des instruments ---
	apple := Action{Symbol: "AAPL", PrixSpot: 142.50, Dividende: 0.96}
	callOption := Option{Symbol: "AAPL_C145", SousJacent: 142.50, Strike: 145.0, Maturite: 0.5, EstCall: true}
	spxFuture := Future{Symbol: "ES1!", PrixSpot: 4500.0, TauxInteret: 0.05, Maturite: 0.25}

	fmt.Println("=== POLYMORPHISME : meme fonction, types differents ===")

	// afficherInstrument accepte n'importe quel Instrument
	afficherInstrument(apple)
	afficherInstrument(callOption)
	afficherInstrument(spxFuture)

	// --- Slice d'interfaces ---
	fmt.Println("\n=== PORTEFEUILLE (slice d'interfaces) ===")

	// On peut mettre des types DIFFERENTS dans une slice d'interface
	portefeuille := []Instrument{apple, callOption, spxFuture}
	quantites := []int{1000, 10, 5}

	for _, inst := range portefeuille {
		afficherInstrument(inst)
	}

	valeurTotale := calculerValeurPortefeuille(portefeuille, quantites)
	fmt.Printf("\nValeur totale portefeuille : %.2f USD\n", valeurTotale)

	// --- Type switch (type assertion) ---
	fmt.Println("\n=== TYPE ASSERTION (type switch) ===")
	for _, inst := range portefeuille {
		fmt.Printf("Analyse de %s :\n", inst.NomSymbol())
		analyserInstrument(inst)
	}

	// --- Type assertion simple ---
	fmt.Println("\n=== TYPE ASSERTION SIMPLE ===")
	var inst Instrument = apple

	// Tenter de convertir en Action
	if action, ok := inst.(Action); ok {
		fmt.Printf("C'est une Action ! Dividende = %.2f\n", action.Dividende)
	}

	// Tenter de convertir en Option (va echouer)
	if _, ok := inst.(Option); !ok {
		fmt.Println("Ce n'est pas une Option.")
	}

	// --- Interface composee ---
	fmt.Println("\n=== INTERFACE COMPOSEE ===")

	// Action satisfait InstrumentNegociable (elle a EstLiquide())
	var negociable InstrumentNegociable = apple
	fmt.Printf("AAPL est liquide : %v\n", negociable.EstLiquide())
	fmt.Printf("Prix AAPL       : %.2f\n", negociable.PrixTheorique())

	// ============================================================
	// RESUME — Ce que tu dois retenir
	// ============================================================
	// 1. type MonInterface interface { Methode1() ...; Methode2() ... }
	// 2. Pas de "implements" : si le type a les methodes -> il satisfait l'interface
	// 3. Une interface peut contenir des types differents -> polymorphisme
	// 4. Type assertion : concret, ok := monInterface.(TypeConcret)
	// 5. Type switch : switch v := i.(type) { case Type1: ... }
	// 6. L'interface vide (interface{} ou any) accepte n'importe quel type

	fmt.Println("\n=== FIN LECON 05 ===")
	fmt.Println("Prochaine etape : lesson06_slices_maps.go")
}
