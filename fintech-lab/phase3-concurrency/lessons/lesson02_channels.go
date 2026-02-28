// ================================================================================
// LECON 08 — Channels
// ================================================================================
// COMMENT EXECUTER :  go run phase3-concurrency/lessons/lesson02_channels.go
// OBJECTIF         :  Communication typee et safe entre goroutines
// PROCHAINE LECON  :  lesson03_mutex.go
// ================================================================================
// Citation Go : "Don't communicate by sharing memory;
//                share memory by communicating."
// ================================================================================

package main

import (
	"fmt"
	"time"
)

// ============================================================
// TYPES POUR LES EXEMPLES
// ============================================================

type MarketTick struct {
	Symbol string
	Bid    float64
	Ask    float64
}

type Signal struct {
	Symbol string
	Action string // "BUY" ou "SELL"
	Size   int
}

func main() {

	// ============================================================
	// PARTIE 1 : CHANNEL DE BASE
	// ============================================================
	// Un channel est un "tuyau" qui permet a des goroutines
	// de s'envoyer des valeurs de facon synchronisee.
	//
	// Syntaxe : make(chan Type)
	// Envoyer : chan <- valeur   (bloque jusqu'a ce que qq lit)
	// Recevoir : valeur := <-chan (bloque jusqu'a ce que qq envoie)

	fmt.Println("=== CHANNEL SIMPLE ===")

	ch := make(chan int) // channel non-buffere (synchrone)

	// Goroutine qui envoie
	go func() {
		fmt.Println("  Goroutine : envoi de 42")
		ch <- 42 // BLOQUE jusqu'a ce que main recoive
		fmt.Println("  Goroutine : envoi confirme")
	}()

	// Main qui recoit
	valeur := <-ch // BLOQUE jusqu'a ce que la goroutine envoie
	fmt.Println("  Main : recu", valeur)

	// ============================================================
	// PARTIE 2 : CHANNEL BUFFERE
	// ============================================================
	// make(chan Type, capacite) cree un channel avec buffer.
	// L'envoi ne bloque pas si le buffer n'est pas plein.
	// La reception bloque si le buffer est vide.
	// Tres utilise en HFT pour absorber les pics de trafic.

	fmt.Println("\n=== CHANNEL BUFFERE ===")

	chBuf := make(chan MarketTick, 10) // buffer de 10 ticks

	// On peut envoyer sans goroutine (ne bloque pas si buffer non plein)
	chBuf <- MarketTick{Symbol: "AAPL", Bid: 142.48, Ask: 142.50}
	chBuf <- MarketTick{Symbol: "MSFT", Bid: 384.98, Ask: 385.00}

	fmt.Printf("Buffer: %d/%d elements\n", len(chBuf), cap(chBuf))

	// Recevoir les ticks
	tick1 := <-chBuf
	tick2 := <-chBuf
	fmt.Printf("Tick 1 : %s Bid=%.2f Ask=%.2f\n", tick1.Symbol, tick1.Bid, tick1.Ask)
	fmt.Printf("Tick 2 : %s Bid=%.2f Ask=%.2f\n", tick2.Symbol, tick2.Bid, tick2.Ask)

	// ============================================================
	// PARTIE 3 : FERMER UN CHANNEL
	// ============================================================
	// close(ch) signale qu'aucune autre valeur ne sera envoyee.
	// IMPORTANT :
	//   - Seul l'expediteur doit fermer le channel
	//   - Envoyer dans un channel ferme = PANIC
	//   - Recevoir d'un channel ferme et vide retourne zero value

	fmt.Println("\n=== FERMETURE DE CHANNEL + for range ===")

	chPrix := make(chan float64, 5)

	// Producteur : envoie puis ferme
	go func() {
		prix := []float64{142.50, 142.51, 142.49, 142.52, 142.48}
		for _, p := range prix {
			chPrix <- p
		}
		close(chPrix) // ferme quand tout est envoye
	}()

	// Consommateur : for range lit jusqu'a fermeture
	for prix := range chPrix { // s'arrete automatiquement quand closed
		fmt.Printf("  Prix recu : %.2f\n", prix)
	}
	fmt.Println("Channel ferme, boucle terminee")

	// ============================================================
	// PARTIE 4 : SELECT — attendre plusieurs channels
	// ============================================================
	// select est comme un switch pour les channels.
	// Il attend que L'UN DES cas soit pret, puis l'execute.
	// Si plusieurs sont prets -> choix aleatoire (equitable).

	fmt.Println("\n=== SELECT ===")

	chApple := make(chan float64, 1)
	chMSFT := make(chan float64, 1)
	chTimeout := time.After(100 * time.Millisecond)

	// Simuler des mises a jour de prix asynchrones
	go func() {
		time.Sleep(10 * time.Millisecond)
		chApple <- 142.55
	}()
	go func() {
		time.Sleep(20 * time.Millisecond)
		chMSFT <- 385.10
	}()

	// Attendre les mises a jour avec timeout
	recu := 0
	for recu < 2 {
		select {
		case px := <-chApple:
			fmt.Printf("  AAPL mis a jour : %.2f\n", px)
			recu++
		case px := <-chMSFT:
			fmt.Printf("  MSFT mis a jour : %.2f\n", px)
			recu++
		case <-chTimeout:
			fmt.Println("  Timeout ! Pas de mise a jour dans les 100ms")
			recu = 2 // sortir de la boucle
		}
	}

	// ============================================================
	// PARTIE 5 : PATTERN PIPELINE
	// ============================================================
	// En HFT, les donnees traversent des "pipelines" de traitement.
	// Chaque etape est une goroutine, les channels les connectent.
	//
	//   [MarketData] -> ch1 -> [Filtrer] -> ch2 -> [Signal] -> ch3 -> [Executer]

	fmt.Println("\n=== PIPELINE : market data -> signal -> execution ===")

	// Etape 1 : generer des ticks
	genererTicks := func() <-chan MarketTick {
		out := make(chan MarketTick, 5)
		go func() {
			ticks := []MarketTick{
				{Symbol: "AAPL", Bid: 142.48, Ask: 142.52},
				{Symbol: "AAPL", Bid: 142.55, Ask: 142.57}, // hausse -> signal BUY ?
				{Symbol: "AAPL", Bid: 141.90, Ask: 141.93}, // baisse -> signal SELL ?
			}
			for _, t := range ticks {
				out <- t
			}
			close(out)
		}()
		return out
	}

	// Etape 2 : analyser et generer des signaux
	genererSignaux := func(in <-chan MarketTick) <-chan Signal {
		out := make(chan Signal, 5)
		go func() {
			defer close(out)
			var dernierMid float64
			for tick := range in {
				mid := (tick.Bid + tick.Ask) / 2
				if dernierMid > 0 {
					if mid > dernierMid*1.001 { // +0.1%
						out <- Signal{Symbol: tick.Symbol, Action: "BUY", Size: 100}
					} else if mid < dernierMid*0.999 { // -0.1%
						out <- Signal{Symbol: tick.Symbol, Action: "SELL", Size: 100}
					}
				}
				dernierMid = mid
			}
		}()
		return out
	}

	// Etape 3 : executer les signaux
	executer := func(in <-chan Signal) {
		for sig := range in {
			fmt.Printf("  EXECUTION : %s %s x%d\n", sig.Action, sig.Symbol, sig.Size)
		}
	}

	// Connecter le pipeline
	ticks := genererTicks()
	signaux := genererSignaux(ticks)
	executer(signaux)

	// ============================================================
	// PARTIE 6 : CHANNELS DIRECTIONNELS
	// ============================================================
	// Pour la securite, on peut typer un channel en lecture seule
	// ou ecriture seule dans les signatures de fonctions.
	//
	//   chan<- Type  : ecriture seulement (send-only)
	//   <-chan Type  : lecture seulement (receive-only)

	fmt.Println("\n=== CHANNELS DIRECTIONNELS ===")

	chBidi := make(chan string, 3)

	// producteur ne peut qu'ecrire
	producteur(chBidi, []string{"ordre_1", "ordre_2", "ordre_3"})

	// consommateur ne peut que lire
	consommateur(chBidi)

	// ============================================================
	// RESUME
	// ============================================================
	// 1. make(chan T)      <- non-buffere : sync complet
	// 2. make(chan T, n)   <- buffere n : envoi non-bloquant si place dispo
	// 3. ch <- val         <- envoyer (peut bloquer)
	// 4. val := <-ch       <- recevoir (peut bloquer)
	// 5. close(ch)         <- signaler fin, seul l'expediteur ferme
	// 6. for v := range ch <- lit jusqu'a fermeture
	// 7. select            <- attendre plusieurs channels + timeout
	// 8. <-chan T          <- receive-only, chan<- T <- send-only

	fmt.Println("\n=== FIN LECON 08 ===")
	fmt.Println("Prochaine etape : lesson03_mutex.go")
}

func producteur(out chan<- string, ordres []string) {
	for _, o := range ordres {
		out <- o
	}
	close(out)
}

func consommateur(in <-chan string) {
	for o := range in {
		fmt.Println("  Traitement:", o)
	}
}
