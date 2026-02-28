// ================================================================================
// LECON 03 — Structs et Methodes
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson03_structs_methods.go
// OBJECTIF         :  Creer des types personnalises avec des comportements
// PROCHAINE LECON  :  lesson04_pointers.go
// ================================================================================

package main

import (
	"errors"
	"fmt"
)

// ============================================================
// PARTIE 1 : DEFINIR UN STRUCT
// ============================================================
// Un struct est un type personnalise qui regroupe des champs.
// C'est l'equivalent d'une classe (sans heritage) en Java/Python.
// Syntaxe : type NomStruct struct { champ1 type1; champ2 type2 }

// OrderSide represente le sens d'un ordre (achat ou vente)
type OrderSide string

// On utilise des constantes typees pour les valeurs possibles
const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

// Order represente un ordre de bourse
type Order struct {
	ID       int64     // identifiant unique
	Symbol   string    // ticker (ex: "AAPL")
	Side     OrderSide // BUY ou SELL
	Quantity int       // nombre d'actions
	Price    float64   // prix limite
	Filled   int       // quantite deja executee
}

// ============================================================
// PARTIE 2 : CONSTRUCTEUR
// ============================================================
// Go n'a pas de constructeur built-in.
// La convention est une fonction NewXxx qui retourne (Type, error).

func NewOrder(id int64, symbol string, side OrderSide, qty int, price float64) (*Order, error) {
	// Validation des parametres
	if qty <= 0 {
		return nil, errors.New("quantite doit etre > 0")
	}
	if price <= 0 {
		return nil, errors.New("prix doit etre > 0")
	}
	if symbol == "" {
		return nil, errors.New("symbol ne peut pas etre vide")
	}

	return &Order{ // & = retourner un POINTEUR vers le struct (voir lecon 04)
		ID:       id,
		Symbol:   symbol,
		Side:     side,
		Quantity: qty,
		Price:    price,
		Filled:   0, // zero value par defaut
	}, nil
}

// ============================================================
// PARTIE 3 : METHODES
// ============================================================
// Une methode est une fonction attachee a un type (le "receiver").
//
// Il y a deux types de receivers :
//
//   Value receiver   : func (o Order) NomMethode() ...
//     -> Travaille sur une COPIE du struct
//     -> Ne peut PAS modifier l'original
//     -> Utiliser pour les methodes de lecture (getters)
//
//   Pointer receiver : func (o *Order) NomMethode() ...
//     -> Travaille sur le struct ORIGINAL
//     -> PEUT modifier l'original
//     -> Utiliser pour les methodes qui modifient l'etat

// Value receiver : methode de lecture (n'a pas besoin de modifier l'ordre)
func (o Order) ValeurTotale() float64 {
	return float64(o.Quantity) * o.Price
}

// Value receiver : affichage
func (o Order) String() string {
	return fmt.Sprintf("Ordre[%d] %s %s %d@%.2f (filled: %d)",
		o.ID, o.Side, o.Symbol, o.Quantity, o.Price, o.Filled)
}

// Value receiver : est-ce que l'ordre est completement execute ?
func (o Order) EstComplet() bool {
	return o.Filled >= o.Quantity
}

// Value receiver : quantite restante a executer
func (o Order) QuantiteRestante() int {
	return o.Quantity - o.Filled
}

// Pointer receiver : MODIFIE le struct (execute une partie de l'ordre)
func (o *Order) Executer(quantite int) error {
	if quantite <= 0 {
		return errors.New("quantite d'execution doit etre > 0")
	}
	if quantite > o.QuantiteRestante() {
		return fmt.Errorf("impossible d'executer %d : seulement %d disponible",
			quantite, o.QuantiteRestante())
	}
	o.Filled += quantite // MODIFIE le champ Filled de l'ordre original
	return nil
}

// ============================================================
// PARTIE 4 : STRUCT IMBRIQUES
// ============================================================
// Les structs peuvent contenir d'autres structs.

type Trade struct {
	ID       int64
	BuyerID  int64
	SellerID int64
	Symbol   string
	Quantity int
	Price    float64
}

func (t Trade) String() string {
	return fmt.Sprintf("Trade[%d] %s %d@%.2f (buyer:%d seller:%d)",
		t.ID, t.Symbol, t.Quantity, t.Price, t.BuyerID, t.SellerID)
}

// ============================================================
// PARTIE 5 : STRUCT LITERAL (creer sans constructeur)
// ============================================================
// Parfois on cree un struct directement sans NewXxx.
// Pratique pour les tests ou les structs simples.

type RiskLimits struct {
	MaxPositionSize  int
	MaxOrderValue    float64
	MaxDailyLoss     float64
}

func main() {

	// --- Constructeur ---
	fmt.Println("=== CREATION D'ORDRES ===")

	ordre1, err := NewOrder(1001, "AAPL", SideBuy, 500, 142.50)
	if err != nil {
		fmt.Println("Erreur:", err)
		return
	}

	ordre2, err := NewOrder(1002, "MSFT", SideSell, 200, 385.00)
	if err != nil {
		fmt.Println("Erreur:", err)
		return
	}

	// La methode String() est appelee automatiquement par fmt
	fmt.Println(ordre1)
	fmt.Println(ordre2)

	// --- Methodes de lecture ---
	fmt.Println("\n=== METHODES DE LECTURE ===")
	fmt.Printf("Valeur ordre 1  : %.2f USD\n", ordre1.ValeurTotale())
	fmt.Printf("Reste a executer: %d actions\n", ordre1.QuantiteRestante())
	fmt.Printf("Est complet     : %v\n", ordre1.EstComplet())

	// --- Methode de modification ---
	fmt.Println("\n=== EXECUTION PARTIELLE ===")

	fmt.Println("Avant :", ordre1)

	err = ordre1.Executer(200) // execute 200 sur 500
	if err != nil {
		fmt.Println("Erreur:", err)
	}
	fmt.Println("Apres execution de 200 :", ordre1)
	fmt.Printf("Reste : %d\n", ordre1.QuantiteRestante())

	err = ordre1.Executer(300) // execute les 300 restants
	if err != nil {
		fmt.Println("Erreur:", err)
	}
	fmt.Println("Apres execution de 300 :", ordre1)
	fmt.Printf("Est complet : %v\n", ordre1.EstComplet())

	// Essayer d'executer trop -> erreur
	err = ordre1.Executer(1)
	if err != nil {
		fmt.Println("Erreur attendue:", err)
	}

	// --- Creation d'un Trade ---
	fmt.Println("\n=== CREATION D'UN TRADE ===")
	trade := Trade{
		ID:       9001,
		BuyerID:  ordre1.ID,
		SellerID: ordre2.ID,
		Symbol:   "AAPL",
		Quantity: 200,
		Price:    142.50,
	}
	fmt.Println(trade)

	// --- Struct literal simple ---
	fmt.Println("\n=== RISK LIMITS ===")
	limites := RiskLimits{
		MaxPositionSize: 10_000,
		MaxOrderValue:   1_000_000.0,
		MaxDailyLoss:    50_000.0,
	}
	fmt.Printf("Max position : %d actions\n", limites.MaxPositionSize)
	fmt.Printf("Max perte/j  : %.0f USD\n", limites.MaxDailyLoss)

	// --- Acceder aux champs ---
	fmt.Println("\n=== ACCES AUX CHAMPS ===")
	// On accede aux champs avec un point : variable.Champ
	fmt.Println("Symbol ordre 2 :", ordre2.Symbol)
	fmt.Println("Prix ordre 2   :", ordre2.Price)

	// ============================================================
	// RESUME — Ce que tu dois retenir
	// ============================================================
	// 1. type NomStruct struct { ... }          <- definition
	// 2. func NewXxx(...) (*Type, error) { ... } <- constructeur (convention)
	// 3. func (o Order) Methode() ...            <- value receiver (lecture)
	// 4. func (o *Order) Methode() ...           <- pointer receiver (modification)
	// 5. Toujours utiliser pointer receiver si la methode modifie l'etat
	// 6. acceder aux champs : monStruct.Champ

	fmt.Println("\n=== FIN LECON 03 ===")
	fmt.Println("Prochaine etape : lesson04_pointers.go")
	fmt.Println("IMPORTANT : la lecon 04 explique pourquoi on utilise & et * !")
}
