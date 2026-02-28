// ================================================================================
// LECON 01 — Variables et Types
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson01_variables_types.go
// OBJECTIF         :  Comprendre les types de base Go et les declarer correctement
// PROCHAINE LECON  :  lesson02_functions_errors.go
// ================================================================================

package main

import "fmt"

func main() {

	// ============================================================
	// PARTIE 1 : DECLARER UNE VARIABLE
	// ============================================================
	// En Go, il y a deux facons de declarer une variable.

	// Facon 1 : explicite avec "var"
	// Syntaxe : var nomVariable type = valeur
	var prix float64 = 142.50 // prix d'une action Apple
	var ticker string = "AAPL"
	var quantite int = 1000
	var actif bool = true

	fmt.Println("=== DECLARATION EXPLICITE ===")
	fmt.Println("Ticker  :", ticker)
	fmt.Println("Prix    :", prix)
	fmt.Println("Quantite:", quantite)
	fmt.Println("Actif   :", actif)

	// ============================================================
	// Facon 2 : implicite avec ":=" (le plus courant en pratique)
	// Go devine le type automatiquement. C'est ce que tu verras
	// dans 90% du code Go professionnel.
	// ============================================================
	prixBid := 142.48  // le meilleur prix acheteur (bid)
	prixAsk := 142.52  // le meilleur prix vendeur (ask)
	spread := prixAsk - prixBid

	fmt.Println("\n=== DECLARATION IMPLICITE (:=) ===")
	fmt.Println("Bid   :", prixBid)
	fmt.Println("Ask   :", prixAsk)
	fmt.Printf("Spread: %.4f\n", spread) // %.4f = 4 decimales

	// ============================================================
	// PARTIE 2 : LES TYPES DE BASE
	// ============================================================
	// Go a des types entiers de taille fixe. En HFT c'est important
	// car on ne veut pas gaspiller de memoire.
	//
	//   int8   : -128 a 127
	//   int16  : -32768 a 32767
	//   int32  : -2 147 483 648 a 2 147 483 647
	//   int64  : environ -9.2 * 10^18 a 9.2 * 10^18  <-- le plus utilise
	//   int    : taille dependante de l'OS (64 bits sur machine 64 bits)
	//
	//   float32 : ~7 chiffres significatifs (eviter pour les prix !)
	//   float64 : ~15 chiffres significatifs <-- toujours utiliser celui-ci

	var ordreID int64 = 1_000_000_000 // 1 milliard d'ordres possibles
	var volume int32 = 5_000          // quantite echangee
	var prixTrade float64 = 142.4912  // prix d'execution precis

	// ATTENTION : en finance, on n'utilise PAS float pour l'argent en prod.
	// On utilise des entiers (prix en centimes ou en "basis points").
	// Ici on utilise float64 pour apprendre, c'est suffisant.

	fmt.Println("\n=== TYPES ENTIERS ET FLOTTANTS ===")
	fmt.Printf("Ordre ID : %d\n", ordreID)
	fmt.Printf("Volume   : %d\n", volume)
	fmt.Printf("Prix     : %.4f\n", prixTrade)

	// ============================================================
	// PARTIE 3 : ZERO VALUES
	// ============================================================
	// En Go, toute variable non initialisee a une "zero value".
	// C'est une securite — pas de valeur garbage comme en C.
	//
	//   int     -> 0
	//   float64 -> 0.0
	//   string  -> "" (chaine vide)
	//   bool    -> false

	var prixNonInit float64 // zero value = 0.0
	var nomNonInit string   // zero value = ""
	var flagNonInit bool    // zero value = false

	fmt.Println("\n=== ZERO VALUES ===")
	fmt.Printf("float64 non init : %f\n", prixNonInit) // 0.000000
	fmt.Printf("string non init  : '%s'\n", nomNonInit) // ''
	fmt.Printf("bool non init    : %v\n", flagNonInit)   // false

	// PIEGE COURANT : oublier d'initialiser un prix -> calculs faux
	// En entretien on te demande souvent : "que vaut X si non initialise ?"

	// ============================================================
	// PARTIE 4 : CONSTANTES
	// ============================================================
	// Les constantes sont immuables. Utiles pour les parametres
	// fixes du systeme (taille des lots, limites de risque...).

	const TailleLot = 100      // un "lot" = 100 actions
	const LimitePosition = 50_000 // max 50 000 actions en position
	const NomMarche = "NASDAQ"

	fmt.Println("\n=== CONSTANTES ===")
	fmt.Println("Taille lot     :", TailleLot)
	fmt.Println("Limite position:", LimitePosition)
	fmt.Println("Marche         :", NomMarche)

	// NOTA : les constantes ne peuvent PAS changer.
	// Ce code ne compile pas : TailleLot = 200

	// ============================================================
	// PARTIE 5 : CONVERSIONS DE TYPES
	// ============================================================
	// Go est STRICTEMENT type. Il n'y a PAS de conversion implicite.
	// Il faut convertir manuellement.

	var nbLots int = 3
	var prixUnitaire float64 = 142.50

	// ERREUR : on ne peut pas multiplier int * float64 directement
	// valeur := nbLots * prixUnitaire  <-- NE COMPILE PAS

	// CORRECT : on convertit
	valeurPortefeuille := float64(nbLots) * TailleLot * prixUnitaire

	fmt.Println("\n=== CONVERSION DE TYPES ===")
	fmt.Printf("Valeur portefeuille: %.2f USD\n", valeurPortefeuille)

	// ============================================================
	// PARTIE 6 : fmt.Printf — les verbes de formatage
	// ============================================================
	// Tu verras ces verbes partout dans Go :
	//
	//   %d  : entier decimal
	//   %f  : flottant (%.2f = 2 decimales)
	//   %s  : string
	//   %v  : valeur generique (fonctionne avec tout)
	//   %T  : affiche le TYPE de la variable
	//   %p  : adresse memoire (pointeur)

	fmt.Println("\n=== VERBES fmt.Printf ===")
	fmt.Printf("Type de prix      : %T\n", prix)         // float64
	fmt.Printf("Type de quantite  : %T\n", quantite)     // int
	fmt.Printf("Type de ticker    : %T\n", ticker)       // string
	fmt.Printf("Prix formate      : %.2f USD\n", prix)   // 142.50 USD

	// ============================================================
	// RESUME — Ce que tu dois retenir
	// ============================================================
	// 1. var x int = 5    -> declaration explicite
	// 2. x := 5           -> declaration implicite (plus courant)
	// 3. const X = 5      -> constante immuable
	// 4. Zero value       -> chaque type a une valeur par defaut
	// 5. Pas de conversion implicite -> float64(monInt)
	// 6. En finance : float64 pour les prix (pas float32)

	fmt.Println("\n=== FIN LECON 01 ===")
	fmt.Println("Prochaine etape : lesson02_functions_errors.go")
}
