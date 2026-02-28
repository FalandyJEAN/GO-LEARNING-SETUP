// ================================================================================
// LECON 02 — Fonctions et Gestion d'Erreurs
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson02_functions_errors.go
// OBJECTIF         :  Ecrire des fonctions qui retournent des erreurs (pattern Go)
// PROCHAINE LECON  :  lesson03_structs_methods.go
// ================================================================================

package main

import (
	"errors"
	"fmt"
)

// ============================================================
// PARTIE 1 : FONCTION SIMPLE
// ============================================================
// Syntaxe : func nomFonction(param1 type1, param2 type2) typeRetour {
//
// En Go, les fonctions sont declarees HORS du main.
// Le main appelle les fonctions.

func calculerValeurOrdre(prix float64, quantite int) float64 {
	return prix * float64(quantite)
}

// ============================================================
// PARTIE 2 : RETOURS MULTIPLES
// ============================================================
// C'est LA fonctionnalite Go la plus importante.
// En Go, une fonction peut retourner PLUSIEURS valeurs.
// Le pattern standard : (resultat, error)
//
// Si l'operation reussit -> (valeur, nil)
// Si elle echoue        -> (zero_value, error)

func calculerPnL(prixEntree, prixSortie float64, quantite int, estLong bool) (float64, error) {

	// Validation des entrees (toujours valider !)
	if quantite <= 0 {
		return 0, errors.New("quantite doit etre positive")
	}
	if prixEntree <= 0 || prixSortie <= 0 {
		return 0, errors.New("les prix doivent etre positifs")
	}

	var pnl float64
	if estLong {
		// Position LONG : on a achete, on profite si le prix monte
		pnl = (prixSortie - prixEntree) * float64(quantite)
	} else {
		// Position SHORT : on a vendu, on profite si le prix baisse
		pnl = (prixEntree - prixSortie) * float64(quantite)
	}

	return pnl, nil // nil = pas d'erreur
}

// ============================================================
// PARTIE 3 : CREER DES ERREURS PERSONNALISEES
// ============================================================
// errors.New("message") cree une erreur simple.
// fmt.Errorf("...%w", err) cree une erreur avec contexte (wrapping).

func validerPrixMarche(prix float64, prixReference float64) error {
	if prix <= 0 {
		return errors.New("prix invalide : doit etre superieur a zero")
	}

	// Un prix qui s'ecarte de plus de 10% du prix de reference
	// est suspect (protection contre les "fat finger trades")
	ecart := (prix - prixReference) / prixReference
	if ecart > 0.10 || ecart < -0.10 {
		// fmt.Errorf permet d'inclure des variables dans le message
		return fmt.Errorf("prix %.2f trop eloigne du reference %.2f (ecart: %.1f%%)",
			prix, prixReference, ecart*100)
	}

	return nil
}

// ============================================================
// PARTIE 4 : DEFER
// ============================================================
// defer execute une instruction APRES la fin de la fonction,
// peu importe comment elle se termine (return normal ou erreur).
// Utile pour : fermer des connexions, liberer des ressources, logger.

func simulerTraitement(ordreID int) {
	fmt.Printf("\n[Ordre %d] Debut du traitement\n", ordreID)

	// defer s'execute en dernier, quand la fonction se termine
	defer fmt.Printf("[Ordre %d] Traitement termine (defer)\n", ordreID)

	fmt.Printf("[Ordre %d] Validation en cours...\n", ordreID)
	fmt.Printf("[Ordre %d] Matching en cours...\n", ordreID)
	// La ligne defer s'affiche ICI, apres tout le reste
}

// ============================================================
// PARTIE 5 : FONCTIONS VARIADIC (nombre variable d'arguments)
// ============================================================
// Utile pour des fonctions comme "calculer la moyenne de N prix"

func moyennePrix(prix ...float64) float64 {
	// "prix" est une slice (liste) de float64
	if len(prix) == 0 {
		return 0
	}
	total := 0.0
	for _, p := range prix {
		total += p
	}
	return total / float64(len(prix))
}

func main() {

	// --- Fonction simple ---
	fmt.Println("=== FONCTION SIMPLE ===")
	valeur := calculerValeurOrdre(142.50, 1000)
	fmt.Printf("Valeur ordre : %.2f USD\n", valeur)

	// ============================================================
	// GESTION D'ERREUR — le pattern le plus important en Go
	// ============================================================
	// Regarde la structure : resultat, err := maFonction(...)
	// Ensuite : if err != nil { ... }
	// C'est ce pattern que tu DOIS connaitre par coeur.
	// ============================================================

	fmt.Println("\n=== RETOURS MULTIPLES + GESTION ERREUR ===")

	// Cas 1 : ordre valide (position LONG)
	pnl, err := calculerPnL(100.0, 105.0, 500, true)
	if err != nil {
		fmt.Println("ERREUR:", err)
	} else {
		fmt.Printf("P&L position LONG : %.2f USD\n", pnl) // +2500
	}

	// Cas 2 : position SHORT
	pnl, err = calculerPnL(100.0, 95.0, 500, false)
	if err != nil {
		fmt.Println("ERREUR:", err)
	} else {
		fmt.Printf("P&L position SHORT: %.2f USD\n", pnl) // +2500
	}

	// Cas 3 : quantite invalide (declenchement de l'erreur)
	pnl, err = calculerPnL(100.0, 105.0, -10, true)
	if err != nil {
		fmt.Println("ERREUR capturee:", err) // affiche l'erreur
	} else {
		fmt.Printf("P&L : %.2f USD\n", pnl)
	}

	// NOTA : quand on ne veut pas utiliser une valeur de retour, on utilise _
	// Exemple : _, err = calculerPnL(...)
	// Le _ dit "j'ignore cette valeur". Tres courant en Go.

	// --- Validation avec erreur contextualisee ---
	fmt.Println("\n=== VALIDATION DE PRIX ===")

	prixRef := 142.50

	err = validerPrixMarche(142.80, prixRef)
	if err != nil {
		fmt.Println("Rejete:", err)
	} else {
		fmt.Println("Prix 142.80 : OK")
	}

	err = validerPrixMarche(165.00, prixRef) // +16% -> rejet
	if err != nil {
		fmt.Println("Rejete:", err)
	} else {
		fmt.Println("Prix 165.00 : OK")
	}

	// --- Defer ---
	fmt.Println("\n=== DEFER ===")
	simulerTraitement(42)

	// --- Variadic ---
	fmt.Println("\n=== FONCTION VARIADIC ===")
	vwap := moyennePrix(142.50, 142.48, 142.52, 142.51, 142.49)
	fmt.Printf("Prix moyen (VWAP simplifie): %.4f\n", vwap)

	// ============================================================
	// RESUME — Ce que tu dois retenir
	// ============================================================
	// 1. func nom(params) typeRetour { ... }
	// 2. func nom(params) (type1, type2) { ... }  <- retours multiples
	// 3. Le pattern standard : (valeur, error)
	//    - Si succes : return valeur, nil
	//    - Si echec  : return zero_value, errors.New("msg")
	// 4. Toujours verifier : if err != nil { ... }
	// 5. defer = execute a la fin de la fonction (cleanup)
	// 6. _ = ignorer une valeur de retour

	fmt.Println("\n=== FIN LECON 02 ===")
	fmt.Println("Prochaine etape : lesson03_structs_methods.go")
}
