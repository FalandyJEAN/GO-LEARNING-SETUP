// ================================================================================
// LECON 04 — Pointeurs
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson04_pointers.go
// OBJECTIF         :  Comprendre & et * — le sujet le plus piege en entretien
// PROCHAINE LECON  :  lesson05_interfaces.go
// ================================================================================
// NOTE : Les pointeurs sont AU COEUR de l'exam01_syntax.go
//        BUG #4 de l'examen vient d'une mauvaise utilisation des pointeurs.
//        Apres cette lecon, tu pourras deboguer cet examen !
// ================================================================================

package main

import "fmt"

// ============================================================
// PARTIE 1 : QU'EST-CE QU'UN POINTEUR ?
// ============================================================
// Un pointeur est une variable qui contient une ADRESSE MEMOIRE.
// Au lieu de stocker la valeur directement, il pointe vers elle.
//
//   Variable normale :  prix = 142.50  (stocke la valeur)
//   Pointeur         :  pPrix = 0xc000 (stocke l'adresse de prix)
//
// Deux operateurs :
//   & = "prendre l'adresse de" (ampersand)
//   * = "dereferencer" = acceder a la valeur a cette adresse

func main() {

	// --- Creation d'un pointeur ---
	fmt.Println("=== POINTEURS DE BASE ===")

	prix := 142.50           // variable normale, valeur directe
	pPrix := &prix           // pPrix = adresse memoire de prix

	fmt.Printf("prix           = %f\n", prix)
	fmt.Printf("adresse de prix = %p\n", pPrix)   // ex: 0xc0000b4010
	fmt.Printf("*pPrix (valeur) = %f\n", *pPrix)  // 142.50

	// Modifier via le pointeur
	*pPrix = 145.00 // on ecrit a l'adresse -> modifie prix !
	fmt.Printf("\nApres *pPrix = 145.00 :\n")
	fmt.Printf("prix  = %f\n", prix)   // 145.00 (modifie !)
	fmt.Printf("*pPrix = %f\n", *pPrix) // 145.00

	// ============================================================
	// PARTIE 2 : POURQUOI LES POINTEURS ?
	// ============================================================
	// En Go, les arguments de fonctions sont COPIES.
	// Sans pointeur, modifier un parametre ne change pas l'original.
	// AVEC pointeur, on travaille directement sur l'original.

	fmt.Println("\n=== COPIE vs POINTEUR ===")

	// Exemple 1 : SANS pointeur (mauvais pour modifier)
	valeur := 100.0
	fmt.Printf("Avant essaiModifierSansPointeur : %.1f\n", valeur)
	essaiModifierSansPointeur(valeur)
	fmt.Printf("Apres essaiModifierSansPointeur : %.1f\n", valeur) // toujours 100 !

	// Exemple 2 : AVEC pointeur (correct)
	fmt.Printf("\nAvant modifierAvecPointeur : %.1f\n", valeur)
	modifierAvecPointeur(&valeur) // on passe l'ADRESSE
	fmt.Printf("Apres modifierAvecPointeur : %.1f\n", valeur) // maintenant 200 !

	// ============================================================
	// PARTIE 3 : LE PIEGE DE L'EXAM01 (BUG #4)
	// ============================================================
	// C'est exactement le bug que tu vas corriger dans exam01_syntax.go
	// La fonction applyDiscount doit modifier le prix d'un ordre.
	// Voici le bug et sa correction :

	fmt.Println("\n=== PIEGE EXAM01 — REASSIGNATION DE POINTEUR ===")

	prixOrdre := 100.0
	ptr := &prixOrdre

	fmt.Println("--- VERSION BUGGEE ---")
	fmt.Printf("Avant : prixOrdre = %.1f\n", prixOrdre)
	applyDiscountBugge(ptr, 0.10)
	fmt.Printf("Apres : prixOrdre = %.1f\n", prixOrdre) // toujours 100 !

	fmt.Println("\n--- VERSION CORRECTE ---")
	fmt.Printf("Avant : prixOrdre = %.1f\n", prixOrdre)
	applyDiscountCorrige(ptr, 0.10)
	fmt.Printf("Apres : prixOrdre = %.1f\n", prixOrdre) // maintenant 90 !

	// ============================================================
	// PARTIE 4 : POINTEURS ET STRUCTS
	// ============================================================
	// Les structs sont souvent manipules via des pointeurs.
	// Go simplifie l'acces aux champs : pas besoin d'ecrire (*o).Field

	fmt.Println("\n=== POINTEURS ET STRUCTS ===")

	type Position struct {
		Symbol   string
		Quantity int
		AvgPrice float64
	}

	pos := Position{Symbol: "AAPL", Quantity: 1000, AvgPrice: 142.50}
	pPos := &pos // pointeur vers pos

	// Pour acceder aux champs via un pointeur :
	// Syntaxe longue  : (*pPos).Symbol
	// Syntaxe courte  : pPos.Symbol  (Go fait la dereference automatiquement)
	fmt.Println("Symbol via pointeur :", pPos.Symbol) // Go = (*pPos).Symbol
	fmt.Printf("Qty via pointeur    : %d\n", pPos.Quantity)

	// Modifier via pointeur
	pPos.Quantity = 500 // modifie l'original !
	fmt.Printf("Quantity apres modif: %d\n", pos.Quantity) // 500

	// ============================================================
	// PARTIE 5 : POINTEUR NIL
	// ============================================================
	// Un pointeur non initialise vaut NIL (equivalent de null/None).
	// Dereferencer un nil pointer = PANIC (crash du programme).
	// TOUJOURS verifier si un pointeur est nil avant de l'utiliser !

	fmt.Println("\n=== POINTEUR NIL ===")

	var p *float64 // nil pointer
	fmt.Printf("p est nil : %v\n", p == nil) // true

	// CECI CRASHERAIT : fmt.Println(*p) -> panic: nil pointer dereference

	// Pattern correct : toujours verifier
	if p != nil {
		fmt.Println("Valeur :", *p)
	} else {
		fmt.Println("Pointeur nil, pas d'acces")
	}

	// ============================================================
	// PARTIE 6 : QUAND UTILISER DES POINTEURS ?
	// ============================================================
	// UTILISER un pointeur si :
	//   1. Tu veux MODIFIER la valeur originale dans une fonction
	//   2. Le struct est GRAND (eviter de copier beaucoup de donnees)
	//   3. Tu veux exprimer "cette valeur peut etre absente" (nil)
	//
	// PAS besoin de pointeur si :
	//   1. Le struct est PETIT (int, float, petit struct)
	//   2. La fonction n'a pas besoin de modifier l'original
	//   3. Tu passes des types de base (int, string)

	fmt.Println("\n=== RESUME POINTEURS ===")
	fmt.Println("& = prendre l'adresse d'une variable")
	fmt.Println("* = acceder a la valeur pointee")
	fmt.Println("*ptr = newVal : modifier la valeur (correct)")
	fmt.Println("ptr = &newVal : changer le pointeur (le bug de l'exam !)")
	fmt.Println("Toujours verifier : if ptr != nil { ... }")

	fmt.Println("\n=== FIN LECON 04 ===")
	fmt.Println("Prochaine etape : lesson05_interfaces.go")
	fmt.Println("Tu es pret pour exam01_syntax.go apres les lecons 01-04 !")
}

// ============================================================
// FONCTIONS DE DEMONSTRATION
// ============================================================

// Sans pointeur : travaille sur une COPIE -> ne modifie pas l'original
func essaiModifierSansPointeur(v float64) {
	v = v * 2 // modifie la copie locale, pas l'original
	fmt.Printf("  (dans la fonction, v = %.1f)\n", v)
}

// Avec pointeur : travaille sur l'ORIGINAL -> modifie l'original
func modifierAvecPointeur(v *float64) {
	*v = *v * 2 // dereference + modifie la valeur a l'adresse
	fmt.Printf("  (dans la fonction, *v = %.1f)\n", *v)
}

// VERSION BUGGEE : reassigne le pointeur local (ne touche pas a l'original)
func applyDiscountBugge(prix *float64, discount float64) {
	// BUG : on cree une nouvelle variable et on fait pointer "prix" dessus.
	// Le pointeur local change, mais l'original n'est pas modifie !
	nouvPrix := *prix * (1 - discount)
	prix = &nouvPrix // BUG : modifie le pointeur LOCAL, pas la valeur originale
	_ = prix         // eviter "declared and not used"
}

// VERSION CORRECTE : modifie la VALEUR a l'adresse pointee
func applyDiscountCorrige(prix *float64, discount float64) {
	*prix = *prix * (1 - discount) // CORRECT : dereference + modification
}
