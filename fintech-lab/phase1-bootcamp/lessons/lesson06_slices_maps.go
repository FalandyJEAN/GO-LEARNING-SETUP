// ================================================================================
// LECON 06 — Slices et Maps
// ================================================================================
// COMMENT EXECUTER :  go run phase1-bootcamp/lessons/lesson06_slices_maps.go
// OBJECTIF         :  Maitriser les collections Go (la base de tout algorithme)
// APRES CETTE LECON : Tu peux attaquer exam01, exam02, exam03 !
// ================================================================================

package main

import (
	"fmt"
	"sort"
)

func main() {

	// ============================================================
	// PARTIE 1 : SLICES — la liste dynamique de Go
	// ============================================================
	// Une slice est une vue sur un tableau. Elle est DYNAMIQUE
	// (contrairement aux arrays de taille fixe).
	//
	// Les arrays en Go : [5]int{1,2,3,4,5} <- taille fixe, rare en pratique
	// Les slices en Go : []int{1,2,3,4,5}  <- taille variable, tres utilise

	fmt.Println("=== SLICES : CREATION ===")

	// Methode 1 : literal (quand on connait les valeurs)
	prix := []float64{142.50, 385.00, 3200.00, 178.25}
	fmt.Println("Prix :", prix)

	// Methode 2 : make(type, longueur, capacite)
	// make alloue la memoire. longueur = nb elements actuels,
	// capacite = memoire reservee (optionnel, optimisation)
	ordreIDs := make([]int64, 0, 100) // vide mais pret pour 100 elements
	fmt.Printf("ordreIDs : len=%d cap=%d\n", len(ordreIDs), cap(ordreIDs))

	// Methode 3 : var (nil slice, meme comportement que make(,0))
	var trades []string // nil mais append fonctionne quand meme
	fmt.Printf("trades est nil : %v\n", trades == nil)

	// ============================================================
	// PARTIE 2 : OPERATIONS DE BASE SUR LES SLICES
	// ============================================================

	fmt.Println("\n=== SLICES : OPERATIONS ===")

	// append : ajouter des elements
	ordreIDs = append(ordreIDs, 1001, 1002, 1003)
	ordreIDs = append(ordreIDs, 1004) // un seul element aussi
	fmt.Println("IDs :", ordreIDs)

	// len : nombre d'elements courant
	// cap : capacite memoire allouee
	fmt.Printf("len=%d cap=%d\n", len(ordreIDs), cap(ordreIDs))

	// acces par index (commence a 0)
	fmt.Println("Premier ID :", ordreIDs[0])
	fmt.Println("Dernier ID :", ordreIDs[len(ordreIDs)-1])

	// PIEGE : index hors bornes = PANIC
	// fmt.Println(ordreIDs[99]) <- panic: index out of range

	// ============================================================
	// PARTIE 3 : ITERATION (for range)
	// ============================================================

	fmt.Println("\n=== SLICES : ITERATION ===")

	// range retourne (index, valeur)
	for i, p := range prix {
		fmt.Printf("  prix[%d] = %.2f\n", i, p)
	}

	// Si on ne veut que les valeurs (ignorer l'index) :
	total := 0.0
	for _, p := range prix {
		total += p
	}
	fmt.Printf("Somme des prix : %.2f\n", total)

	// Si on ne veut que les index :
	for i := range prix {
		fmt.Printf("  index %d\n", i)
	}

	// ============================================================
	// PARTIE 4 : SLICING (sous-slices)
	// ============================================================

	fmt.Println("\n=== SLICING ===")
	// s[debut:fin] -> elements de debut (inclus) a fin (exclu)
	// Mnemonique : comme Python

	donnees := []float64{10.0, 20.0, 30.0, 40.0, 50.0}
	fmt.Println("Original       :", donnees)
	fmt.Println("donnees[1:3]   :", donnees[1:3]) // [20 30]
	fmt.Println("donnees[:2]    :", donnees[:2])   // [10 20]
	fmt.Println("donnees[3:]    :", donnees[3:])   // [40 50]

	// ATTENTION : les slices partagent la memoire !
	// Modifier une sous-slice modifie l'original.
	sousDonnees := donnees[1:3]
	sousDonnees[0] = 999.0
	fmt.Println("Apres sousDonnees[0]=999 :", donnees) // donnees[1] est aussi 999 !

	// Pour eviter ce partage, utiliser copy :
	copie := make([]float64, len(donnees))
	copy(copie, donnees) // copie complete et independante

	// ============================================================
	// PARTIE 5 : TRIER UNE SLICE
	// ============================================================

	fmt.Println("\n=== TRI ===")
	bids := []float64{142.48, 142.45, 142.50, 142.47, 142.49}
	fmt.Println("Avant tri :", bids)
	sort.Float64s(bids) // tri croissant
	fmt.Println("Apres tri (croissant) :", bids)

	// Tri decroissant (ordre du carnet : meilleur bid en premier)
	sort.Sort(sort.Reverse(sort.Float64Slice(bids)))
	fmt.Println("Apres tri (decroissant) :", bids)

	// ============================================================
	// PARTIE 6 : MAPS — dictionnaire cle->valeur
	// ============================================================
	// map[typeCle]typeValeur
	// TOUJOURS initialiser avec make (ou literal) avant d'utiliser !

	fmt.Println("\n=== MAPS : CREATION ===")

	// Methode 1 : make
	prixParSymbol := make(map[string]float64)

	// Ajout/modification : map[cle] = valeur
	prixParSymbol["AAPL"] = 142.50
	prixParSymbol["MSFT"] = 385.00
	prixParSymbol["GOOGL"] = 141.80

	fmt.Println("Prix AAPL :", prixParSymbol["AAPL"])
	fmt.Println("Prix MSFT :", prixParSymbol["MSFT"])

	// Methode 2 : literal
	positions := map[string]int{
		"AAPL":  1000,
		"MSFT":  500,
		"GOOGL": 200,
	}
	fmt.Println("\nPositions :", positions)

	// ============================================================
	// PARTIE 7 : OPERATIONS SUR LES MAPS
	// ============================================================

	fmt.Println("\n=== MAPS : OPERATIONS ===")

	// Acces et verification d'existence (two-value form)
	// TOUJOURS utiliser cette forme pour verifier si la cle existe
	px, existe := prixParSymbol["AAPL"]
	if existe {
		fmt.Printf("AAPL existe : %.2f\n", px)
	}

	// Si la cle n'existe pas -> zero value (pas d'erreur, pas de panic)
	px, existe = prixParSymbol["TSLA"]
	if !existe {
		fmt.Printf("TSLA n'existe pas (zero value: %.2f)\n", px)
	}

	// PIEGE : sans la verification, on obtient zero value silencieusement
	// Ce n'est PAS une erreur en Go -> source de bugs !
	prixTSLA := prixParSymbol["TSLA"]
	fmt.Printf("TSLA prix (piege) : %.2f\n", prixTSLA) // 0 silencieux !

	// Supprimer une cle
	delete(prixParSymbol, "GOOGL")
	fmt.Println("Apres delete GOOGL:", prixParSymbol)

	// Nombre d'elements
	fmt.Println("Nombre de symboles:", len(prixParSymbol))

	// ============================================================
	// PARTIE 8 : ITERER SUR UNE MAP
	// ============================================================

	fmt.Println("\n=== MAPS : ITERATION ===")
	// ATTENTION : l'ordre d'iteration des maps est ALEATOIRE en Go
	// Ne jamais supposer un ordre particulier !
	for symbol, pos := range positions {
		fmt.Printf("  %s : %d actions\n", symbol, pos)
	}

	// ============================================================
	// PARTIE 9 : PIEGE — nil map
	// ============================================================

	fmt.Println("\n=== PIEGE NIL MAP ===")

	var mapNulle map[string]float64 // nil map

	// LIRE depuis une nil map : ok (retourne zero value)
	val := mapNulle["AAPL"]
	fmt.Printf("Lecture nil map : %.2f (ok)\n", val)

	// ECRIRE dans une nil map : PANIC !
	// mapNulle["AAPL"] = 100.0 <- panic: assignment to entry in nil map

	// TOUJOURS initialiser avant d'ecrire :
	mapNulle = make(map[string]float64)
	mapNulle["AAPL"] = 142.50
	fmt.Printf("Apres make + ecriture : %.2f (ok)\n", mapNulle["AAPL"])

	// ============================================================
	// PARTIE 10 : CAS PRATIQUE — Carnet d'ordres simplifie
	// ============================================================

	fmt.Println("\n=== CAS PRATIQUE : CARNET D'ORDRES ===")

	// On organise les ordres par prix (bid side du carnet)
	// map[prix][]ordreID : plusieurs ordres peuvent avoir le meme prix
	carneBids := make(map[float64][]int64)

	// Ajouter des ordres
	ajouterOrdre := func(prix float64, id int64) {
		carneBids[prix] = append(carneBids[prix], id)
	}

	ajouterOrdre(142.50, 1001)
	ajouterOrdre(142.48, 1002)
	ajouterOrdre(142.50, 1003) // meme prix que 1001
	ajouterOrdre(142.49, 1004)

	// Extraire les prix et les trier
	prixBids := make([]float64, 0, len(carneBids))
	for p := range carneBids {
		prixBids = append(prixBids, p)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(prixBids)))

	fmt.Println("Carnet Bid (prix decroissant) :")
	for _, p := range prixBids {
		fmt.Printf("  %.2f -> ordres %v\n", p, carneBids[p])
	}

	// ============================================================
	// RESUME — Ce que tu dois retenir
	// ============================================================
	// SLICES :
	//   []Type{}                   <- literal
	//   make([]Type, len, cap)     <- allocation
	//   append(s, elem)            <- ajouter
	//   s[i]                       <- acces
	//   s[debut:fin]               <- sous-slice (attention: memoire partagee)
	//   for i, v := range s { }   <- iteration
	//   len(s) / cap(s)            <- longueur / capacite
	//
	// MAPS :
	//   map[K]V{}                  <- literal
	//   make(map[K]V)              <- toujours initialiser avant ecriture !
	//   m[cle] = val               <- ecrire
	//   val, ok := m[cle]          <- lire + verifier existence
	//   delete(m, cle)             <- supprimer
	//   for k, v := range m { }   <- iteration (ordre aleatoire !)

	fmt.Println("\n=== FIN LECON 06 ===")
	fmt.Println("Tu peux maintenant attaquer les examens !")
	fmt.Println("Ordre recommande :")
	fmt.Println("  1. exam01_syntax.go       (utilise lecons 01-04)")
	fmt.Println("  2. exam02_interfaces.go   (utilise lecon 05)")
	fmt.Println("  3. exam03_slices_maps.go  (utilise lecon 06)")
}
