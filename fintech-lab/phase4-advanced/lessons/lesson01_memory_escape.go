// ================================================================================
// LECON 10 — Memoire, Escape Analysis et le Garbage Collector
// ================================================================================
// COMMENT EXECUTER :  go run phase4-advanced/lessons/lesson01_memory_escape.go
// POUR VOIR L'ESCAPE :  go build -gcflags="-m" phase4-advanced/lessons/lesson01_memory_escape.go
// OBJECTIF         :  Comprendre comment Go gere la memoire (crucial en HFT)
// PROCHAINE LECON  :  lesson02_sync_pool.go
// ================================================================================

package main

import "fmt"

// ============================================================
// PARTIE 1 : STACK vs HEAP
// ============================================================
// En Go (et dans la plupart des langages), la memoire est divisee en 2 zones :
//
// STACK (pile) :
//   - Tres rapide (pas de GC, allocation/liberation automatique)
//   - Variables locales qui ne "s'echappent" pas de leur fonction
//   - Taille limitee
//
// HEAP (tas) :
//   - Plus lent (gere par le Garbage Collector)
//   - Variables qui doivent survivre au-dela de leur fonction
//   - Taille dynamique
//
// En HFT, on veut MINIMISER les allocations sur le heap pour
// reduire la pression sur le GC et les latences impredictibles.

// ============================================================
// PARTIE 2 : ESCAPE ANALYSIS
// ============================================================
// Le compilateur Go determine automatiquement si une variable
// doit aller sur le stack ou le heap.
// C'est l'"escape analysis".
//
// Une variable "s'echappe" vers le heap si :
//   - Elle est retournee par reference (pointeur)
//   - Elle est stockee dans une structure qui vit plus longtemps
//   - Elle est passee a une interface (boxing)
//   - Sa taille n'est pas connue a la compilation

// CAS 1 : reste sur le stack
// (utiliser go build -gcflags="-m" pour le voir)
func creerPrixLocal() float64 {
	prix := 142.50 // reste sur le stack -> pas de GC
	return prix    // on retourne la VALEUR, pas un pointeur
}

// CAS 2 : s'echappe vers le heap
func creerPrixHeap() *float64 {
	prix := 142.50
	return &prix // le pointeur est retourne -> prix s'echappe vers le heap
}

// CAS 3 : struct sur le stack
func creerOrdreLocal() struct{ ID int; Prix float64 } {
	return struct{ ID int; Prix float64 }{ID: 1, Prix: 142.50}
}

// CAS 4 : struct sur le heap (parce qu'on retourne un pointeur)
type OrdreSimple struct {
	ID   int
	Prix float64
}

func creerOrdreHeap() *OrdreSimple {
	return &OrdreSimple{ID: 1, Prix: 142.50} // s'echappe vers le heap
}

// ============================================================
// PARTIE 3 : IMPACT SUR LES PERFORMANCES
// ============================================================
// Chaque allocation heap :
//   1. Couts CPU supplementaires (le GC doit la tracer)
//   2. Peut causer des pauses GC (stop-the-world, meme si tres courtes)
//   3. Augmente la pression memoire
//
// En HFT, une pause GC de 100 microseconds = catastrophe.
// Solution : reutil-iser les objets au lieu d'en creer de nouveaux.
// C'est l'objet de lesson02_sync_pool.go

// ============================================================
// PARTIE 4 : INTERFACES ET ESCAPE
// ============================================================
// Passer une valeur concrete dans une interface cause TOUJOURS
// une allocation heap (boxing). C'est un pieges courant.

type Valorisable interface {
	Valeur() float64
}

type Action struct {
	Prix float64
}

func (a Action) Valeur() float64 { return a.Prix }

func afficherValeur(v Valorisable) {
	// v est une interface -> boxing -> allocation heap !
	fmt.Printf("Valeur : %.2f\n", v.Valeur())
}

// ============================================================
// PARTIE 5 : STRATEGIES POUR REDUIRE LES ALLOCATIONS
// ============================================================

// Strategie 1 : Pre-allouer les slices avec make
// BAD : append cree de nouvelles allocations
func collecterPrixBad(n int) []float64 {
	var result []float64 // allocation initiale zero
	for i := 0; i < n; i++ {
		result = append(result, float64(i)*0.01+100.0) // peut reallouer !
	}
	return result
}

// GOOD : capacite connue a l'avance
func collecterPrixGood(n int) []float64 {
	result := make([]float64, 0, n) // alloue d'un coup, pas de reallocation
	for i := 0; i < n; i++ {
		result = append(result, float64(i)*0.01+100.0)
	}
	return result
}

// Strategie 2 : Passer un buffer en parametre au lieu de retourner
// Permet a l'appelant de reutiliser le buffer

func remplirPrix(buf []float64, debut float64) {
	for i := range buf {
		buf[i] = debut + float64(i)*0.01
	}
}

// ============================================================
// PARTIE 6 : LE GARBAGE COLLECTOR EN GO
// ============================================================
// Go utilise un GC concurrent tri-colore (mark-and-sweep).
// Caracteristiques en Go moderne :
//   - Pauses "stop-the-world" < 1ms en general
//   - Tourne en parallele avec le programme
//   - Declenchement controlable via GOGC
//
// Variables d'environnement importantes :
//   GOGC=100  (defaut) : GC se declenche quand le heap double
//   GOGC=200  : GC moins frequent (moins de pauses, plus de memoire)
//   GOGC=off  : desactiver le GC (DANGEREUX, uniquement pour bench)
//   GOMEMLIMIT : limite la taille totale de la memoire heap

func main() {

	fmt.Println("=== ESCAPE ANALYSIS ===")
	fmt.Println("(Lance 'go build -gcflags=\"-m\" ...' pour voir les escapes)")

	valeurStack := creerPrixLocal() // sur le stack
	valeurHeap := creerPrixHeap()   // sur le heap

	fmt.Printf("Stack value : %.2f\n", valeurStack)
	fmt.Printf("Heap value  : %.2f\n", *valeurHeap)

	fmt.Println("\n=== INTERFACE BOXING ===")
	a := Action{Prix: 142.50}
	afficherValeur(a) // a est boxe dans l'interface -> alloc heap

	fmt.Println("\n=== PRE-ALLOCATION ===")
	n := 1000

	// Version sans pre-allocation (plusieurs reallocations possibles)
	prixBad := collecterPrixBad(n)
	fmt.Printf("prixBad  : %d elements, premier=%.2f\n", len(prixBad), prixBad[0])

	// Version avec pre-allocation (une seule allocation)
	prixGood := collecterPrixGood(n)
	fmt.Printf("prixGood : %d elements, premier=%.2f\n", len(prixGood), prixGood[0])

	fmt.Println("\n=== REUTILISATION DE BUFFER ===")
	// Cree une fois, reutilise plusieurs fois
	buf := make([]float64, 100)

	remplirPrix(buf, 142.00) // premiere utilisation
	fmt.Printf("Buffer[0]=%.2f Buffer[99]=%.2f\n", buf[0], buf[99])

	remplirPrix(buf, 385.00) // reutilisation du meme buffer
	fmt.Printf("Buffer[0]=%.2f Buffer[99]=%.2f\n", buf[0], buf[99])

	// ============================================================
	// RESUME
	// ============================================================
	// 1. Stack = rapide, pas de GC. Heap = GC, pauses potentielles.
	// 2. Retourner un pointeur = escape vers le heap
	// 3. Passer dans une interface = boxing = escape vers le heap
	// 4. go build -gcflags="-m" pour voir les escapes
	// 5. Pre-allouer les slices : make([]T, 0, capacite)
	// 6. Reutiliser les buffers au lieu d'en creer de nouveaux
	// 7. sync.Pool = reutiliser des objets heap (voir lecon suivante)
	// 8. GOGC et GOMEMLIMIT pour controler le GC en production

	fmt.Println("\n=== FIN LECON 10 ===")
	fmt.Println("Prochaine etape : lesson02_sync_pool.go")
}
