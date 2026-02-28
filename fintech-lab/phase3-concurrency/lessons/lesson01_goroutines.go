// ================================================================================
// LECON 07 — Goroutines
// ================================================================================
// COMMENT EXECUTER :  go run phase3-concurrency/lessons/lesson01_goroutines.go
// OBJECTIF         :  Comprendre la concurrence legere de Go
// PROCHAINE LECON  :  lesson02_channels.go
// ================================================================================
// PREREQUIS : Avoir fait les lecons 01 a 06 (Phase 1)
// ================================================================================

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================
// PARTIE 1 : QU'EST-CE QU'UNE GOROUTINE ?
// ============================================================
// Une goroutine est un "fil d'execution leger" gere par Go.
// Elle est bien plus legere qu'un thread OS :
//   - Thread OS   : ~1MB de stack, cher a creer
//   - Goroutine   : ~2KB de stack, peut en avoir des millions
//
// En HFT, on utilise les goroutines pour :
//   - Traiter les flux de marche en parallele
//   - Gerer les connexions reseau simultanement
//   - Effectuer des calculs de risque en arriere-plan
//
// Syntaxe : go maFonction()

func main() {

	fmt.Printf("Nombre de CPUs disponibles : %d\n", runtime.NumCPU())

	// ============================================================
	// PARTIE 2 : LANCER UNE GOROUTINE
	// ============================================================

	fmt.Println("\n=== GOROUTINE SIMPLE ===")

	// Sans goroutine : execution sequentielle
	fmt.Println("Main : avant la goroutine")

	// go func() { ... }() = goroutine anonyme
	// S'execute en PARALLELE avec le main
	go func() {
		fmt.Println("Goroutine : je m'execute en parallele !")
	}()

	// PROBLEME : si main se termine, TOUTES les goroutines sont tuees.
	// Sans synchronisation, la goroutine ci-dessus pourrait ne jamais s'afficher.
	time.Sleep(10 * time.Millisecond) // mauvaise pratique, voir sync.WaitGroup
	fmt.Println("Main : apres la goroutine")

	// ============================================================
	// PARTIE 3 : sync.WaitGroup — attendre les goroutines
	// ============================================================
	// WaitGroup est le bon outil pour attendre que des goroutines finissent.
	//
	// wg.Add(n)  : signale que n goroutines vont demarrer
	// wg.Done()  : une goroutine signale qu'elle est terminee
	// wg.Wait()  : bloque jusqu'a ce que toutes soient terminees

	fmt.Println("\n=== sync.WaitGroup ===")

	var wg sync.WaitGroup

	// Simuler le traitement parallele de 5 flux de marche
	symboles := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}

	for _, sym := range symboles {
		wg.Add(1) // on va lancer 1 goroutine de plus

		sym := sym // IMPORTANT : capturer la variable dans la closure !
		// (voir explication ci-dessous)

		go func() {
			defer wg.Done() // s'execute quand la goroutine se termine

			// Simuler un traitement variable
			prix := obtenirPrix(sym)
			fmt.Printf("  [%s] Prix : %.2f\n", sym, prix)
		}()
	}

	wg.Wait() // attend que toutes les goroutines aient appele Done()
	fmt.Println("Tous les prix recus !")

	// ============================================================
	// PARTIE 4 : PIEGE CLASSIQUE — CAPTURE DE VARIABLE DE BOUCLE
	// ============================================================
	// C'est un des bugs les plus courants avec les goroutines.
	// Si tu passes une variable de boucle dans une closure sans
	// la "capturer", toutes les goroutines verront la DERNIERE valeur.

	fmt.Println("\n=== PIEGE CAPTURE DE VARIABLE ===")

	fmt.Println("--- BUGGE (toutes verront la derniere valeur) ---")
	var wgBug sync.WaitGroup
	for i := 0; i < 3; i++ {
		wgBug.Add(1)
		go func() {
			defer wgBug.Done()
			fmt.Println("  valeur i (bug) :", i) // affiche probablement 3,3,3
		}()
	}
	wgBug.Wait()

	fmt.Println("--- CORRECT (chaque goroutine a sa propre copie) ---")
	var wgOk sync.WaitGroup
	for i := 0; i < 3; i++ {
		wgOk.Add(1)
		i := i // copie locale ! chaque iteration cree un nouveau i
		go func() {
			defer wgOk.Done()
			fmt.Println("  valeur i (ok) :", i) // affiche 0, 1, 2 (ordre variable)
		}()
	}
	wgOk.Wait()

	// ============================================================
	// PARTIE 5 : GOROUTINES ET DATA RACES
	// ============================================================
	// Si deux goroutines lisent/ecrivent la MEME variable sans
	// synchronisation -> DATA RACE -> comportement indefini !
	// C'est le sujet de exam_data_race.go en Phase 3.

	fmt.Println("\n=== DATA RACE DEMO ===")

	// VERSION BUGGEE : acces concurrent sans synchronisation
	compteur := 0
	var wgRace sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wgRace.Add(1)
		go func() {
			defer wgRace.Done()
			compteur++ // DATA RACE ! lecture + incrementation + ecriture non atomique
		}()
	}
	wgRace.Wait()

	// Le resultat est IMPREVISIBLE (probablement != 1000)
	fmt.Printf("Compteur (avec race) : %d (attendu: 1000)\n", compteur)
	fmt.Println("  -> Pour corriger : utiliser sync.Mutex ou sync/atomic")
	fmt.Println("  -> Pour detecter : go run -race ...")

	// ============================================================
	// PARTIE 6 : GOROUTINES EN PRODUCTION HFT
	// ============================================================
	// En HFT, architecture typique :
	//
	//   goroutine 1 : ecoute le reseau (market data feed)
	//   goroutine 2 : traite les ordres
	//   goroutine 3 : calcule le risque en temps reel
	//   goroutine 4 : envoie les ordres au marche
	//   goroutine 5 : log et monitoring
	//
	// Elles communiquent via des CHANNELS (lecon suivante)

	fmt.Println("\n=== GOROUTINES RUNTIME ===")
	fmt.Printf("Goroutines actives : %d\n", runtime.NumGoroutine())

	// ============================================================
	// RESUME
	// ============================================================
	// 1. go func() { ... }()         <- lancer une goroutine
	// 2. go maFonction(args)          <- goroutine sur une fonction existante
	// 3. sync.WaitGroup               <- attendre la fin des goroutines
	//      wg.Add(1)                  <- avant de lancer la goroutine
	//      defer wg.Done()            <- au debut de la goroutine
	//      wg.Wait()                  <- dans main pour attendre
	// 4. TOUJOURS capturer les variables de boucle : i := i
	// 5. Sans synchronisation = data race = bugs impredictibles
	// 6. go run -race ./... pour detecter les races

	fmt.Println("\n=== FIN LECON 07 ===")
	fmt.Println("Prochaine etape : lesson02_channels.go")
}

// Simule la recuperation d'un prix (avec delai variable)
func obtenirPrix(symbol string) float64 {
	prix := map[string]float64{
		"AAPL":  142.50,
		"MSFT":  385.00,
		"GOOGL": 141.80,
		"AMZN":  178.25,
		"TSLA":  248.50,
	}
	// Simuler un delai reseau variable
	time.Sleep(time.Duration(len(symbol)) * time.Millisecond)
	return prix[symbol]
}
