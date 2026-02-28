// exam03_slices_maps.go — EXAMEN 03 : "Le Carnet d'Ordres Fantome"
// Ce code compile sans erreur mais contient 4 bugs de slices/maps.
// Lancer avec : go run phase1-bootcamp/exams/exam03_slices_maps.go

package main

import "fmt"

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Order struct {
	Symbol   string
	Side     string
	Price    float64
	Quantity int
}

func (o Order) String() string {
	return fmt.Sprintf("%s %s x%d @ $%.2f", o.Symbol, o.Side, o.Quantity, o.Price)
}

// ---------------------------------------------------------------------------
// BUG #1 : addOrder passe la slice par valeur.
// append() peut allouer un nouveau backing array.
// La slice originale dans main() n'est JAMAIS modifiee.
// Correction : retourner la nouvelle slice, ou accepter un *[]Order.
// ---------------------------------------------------------------------------
func addOrder(book []Order, o Order) {
	book = append(book, o) // Modification locale uniquement — perdue a la sortie
}

// ---------------------------------------------------------------------------
// BUG #2 : Le range loop capture la variable d'iteration par reference.
// Toutes les entrees du snapshot pointent vers la meme adresse memoire
// (la variable 'o' de la boucle), qui contient le dernier ordre apres la boucle.
// ---------------------------------------------------------------------------
func buildSnapshot(book []Order) []*Order {
	var snapshot []*Order
	for _, o := range book {
		snapshot = append(snapshot, &o) // BUG : &o = adresse de la var de boucle
	}
	return snapshot
}

// ---------------------------------------------------------------------------
// BUG #3 : La map prices n'est pas initialisee avec make().
// var prices map[string]float64 => prices == nil
// Ecrire dans une nil map provoque un panic en runtime.
// ---------------------------------------------------------------------------
func updatePrices(symbols []string, newPrices []float64) map[string]float64 {
	var prices map[string]float64 // BUG : nil map !
	for i, symbol := range symbols {
		if i < len(newPrices) {
			prices[symbol] = newPrices[i] // PANIC ici
		}
	}
	return prices
}

// ---------------------------------------------------------------------------
// BUG #4 : filterByMinQty retourne des *Order pointant vers les originaux.
// Quand on modifie les Prix dans le filtre (simulation d'un ajustement),
// on modifie aussi le carnet original. Un filtre ne doit pas avoir
// d'effets de bord sur la source.
// Note : dans cet exam, la "modification" est simulee pour montrer le probleme.
// ---------------------------------------------------------------------------
func filterByMinQty(book []Order, minQty int) []*Order {
	var result []*Order
	for i := range book {
		if book[i].Quantity >= minQty {
			result = append(result, &book[i]) // Pointe vers l'original
		}
	}
	// Simulation : le filtre "marque" les ordres filtrés (effet de bord)
	for _, o := range result {
		o.Price = o.Price * 1.001 // BUG : modifie le carnet original !
	}
	return result
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// --- Test BUG #1 : addOrder ---
	fmt.Println("--- Test addOrder ---")
	var book []Order
	addOrder(book, Order{"AAPL", "BUY",  189.50, 100})
	addOrder(book, Order{"AAPL", "SELL", 190.00, 200})
	addOrder(book, Order{"AAPL", "BUY",  188.00, 50})
	fmt.Printf("Carnet apres ajout : %d ordres\n", len(book)) // Attendu: 3, Obtenu: 0
	for _, o := range book {
		fmt.Printf("  %v\n", o)
	}
	fmt.Println()

	// Pour les tests suivants, on initialise le carnet correctement
	book2 := []Order{
		{"AAPL", "BUY",  189.50, 100},
		{"AAPL", "SELL", 190.00, 200},
		{"AAPL", "BUY",  188.00, 50},
	}

	// --- Test BUG #2 : buildSnapshot ---
	fmt.Println("--- Test buildSnapshot ---")
	snapshot := buildSnapshot(book2)
	fmt.Printf("Snapshot : %d ordres\n", len(snapshot))
	for i, o := range snapshot {
		// Attendu: chaque pointeur pointe vers un ordre different
		// Obtenu : tous pointent vers le dernier ordre (AAPL BUY x50 @ $188.00)
		fmt.Printf("  Snapshot[%d] : %v\n", i, *o)
	}
	fmt.Println()

	// --- Test BUG #3 : updatePrices ---
	fmt.Println("--- Test updatePrices ---")
	symbols := []string{"AAPL", "MSFT"}
	newPrices := []float64{190.00, 416.00}
	prices := updatePrices(symbols, newPrices) // PANIC probable ici
	fmt.Printf("Prix mis a jour : AAPL=%.2f MSFT=%.2f\n",
		prices["AAPL"], prices["MSFT"])
	fmt.Println()

	// --- Test BUG #4 : filterByMinQty ---
	fmt.Println("--- Test filterByMinQty ---")
	book3 := []Order{
		{"AAPL", "BUY",  189.50, 100},
		{"AAPL", "SELL", 190.00, 200},
		{"AAPL", "BUY",  188.00, 50},  // < 100, doit etre filtre
	}

	fmt.Printf("Carnet avant filtre : %d ordres\n", len(book3))
	for _, o := range book3 {
		fmt.Printf("  Prix original : %.2f\n", o.Price)
	}

	filtered := filterByMinQty(book3, 100)
	fmt.Printf("Apres filtre (min 100) : %d ordres\n", len(filtered))

	fmt.Printf("Carnet original intact  : %d ordres\n", len(book3))
	for _, o := range book3 {
		// BUG : les prix ont ete modifies meme dans le carnet ORIGINAL
		fmt.Printf("  Prix apres filtre : %.2f (devrait etre identique a avant)\n", o.Price)
	}
}
