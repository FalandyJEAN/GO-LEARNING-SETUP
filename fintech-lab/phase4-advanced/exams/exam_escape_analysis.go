// exam_escape_analysis.go — EXAMEN 4-B : "Escape Analysis : Stack vs Heap"
// Ce code genere des allocations evitables dans le hot path.
//
// Analyser : go build -gcflags="-m" ./phase4-advanced/exams/exam_escape_analysis.go 2>&1
// Benchmark : go test ./phase4-advanced/exams/ -bench=BenchmarkHotPath -benchmem

package main

import (
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// Types du hot path
// ---------------------------------------------------------------------------

type Order struct {
	ID    uint64
	Price float64
	Qty   int64
}

type Trade struct {
	BuyID   uint64
	SellID  uint64
	Price   float64
	Qty     int64
	Ts      int64
}

// ---------------------------------------------------------------------------
// BUG #1 : Retourner *Trade => echappement vers heap garanti
// ---------------------------------------------------------------------------

// newTrade alloue un Trade et retourne son adresse.
// BUG #1 : &Trade{} dans newTrade() "escape to heap" car l'adresse
// sort de la fonction. Le GC devra collecter chaque Trade apres usage.
//
// go build -gcflags="-m" montrera :
//   "newTrade &Trade{} escapes to heap"
//
// Correction : retourner Trade par valeur (pas de pointeur),
// ou accepter un *Trade pre-alloue en parametre.
func newTrade(buyID, sellID uint64, price float64, qty int64) *Trade {
	return &Trade{ // BUG #1 : escape vers heap
		BuyID:  buyID,
		SellID: sellID,
		Price:  price,
		Qty:    qty,
		Ts:     time.Now().UnixNano(),
	}
}

// ---------------------------------------------------------------------------
// BUG #2 : Passer a interface{} => boxing => heap
// ---------------------------------------------------------------------------

// logTrade log un trade. Utilise fmt.Println qui prend des interface{}.
// BUG #2 : Les arguments de fmt.Println echappent vers le heap car
// fmt.Println accepte ...interface{} => boxing de chaque argument.
//
// go build -gcflags="-m" montrera :
//   "price escapes to heap" (boxing dans interface{})
//
// Correction dans le hot path : utiliser un io.Writer buferise + format manuel,
// ou un logger zero-allocation comme zerolog ou zap.
func logTrade(price float64, qty int64) {
	fmt.Println("Trade:", price, "x", qty) // BUG #2 : price et qty echappent
}

// ---------------------------------------------------------------------------
// BUG #3 : Variable capturee par closure => heap
// ---------------------------------------------------------------------------

// scheduleTrade planifie l'execution differee d'un trade via une goroutine.
// BUG #3 : La variable 'trade' est capturee par la closure.
// Elle doit donc vivre sur le heap (sa duree de vie depasse la fonction).
//
// go build -gcflags="-m" montrera :
//   "trade escapes to heap"
//
// Correction : passer 'trade' en parametre de la closure, pas par capture.
func scheduleTrade(buy, sell *Order) func() {
	trade := Trade{ // BUG #3 : sera capture par la closure => heap
		BuyID:  buy.ID,
		SellID: sell.ID,
		Price:  sell.Price,
		Qty:    min64(buy.Qty, sell.Qty),
	}

	return func() {
		// 'trade' est capturee par reference => elle doit etre sur le heap
		_ = trade.Price * float64(trade.Qty)
	}
}

// ---------------------------------------------------------------------------
// BUG #4 : Slice de taille inconnue => heap
// ---------------------------------------------------------------------------

// matchBatch tente de matcher des ordres dans un batch.
// BUG #4 : 'results' est une slice dont la capacite est inconnue
// a la compilation => alloue sur le heap.
//
// Pour un nombre fixe de resultats, un tableau est preferable (stack).
func matchBatch(orders []*Order) []Trade {
	results := make([]Trade, 0) // BUG #4 : taille 0, capacite inconnue => heap
	// Correction : make([]Trade, 0, len(orders)/2) ou [MaxTrades]Trade{}

	for i := 0; i+1 < len(orders); i += 2 {
		buy := orders[i]
		sell := orders[i+1]
		if buy.Price >= sell.Price {
			t := Trade{
				BuyID:  buy.ID,
				SellID: sell.ID,
				Price:  sell.Price,
				Qty:    min64(buy.Qty, sell.Qty),
			}
			results = append(results, t)
		}
	}
	return results
}

// ---------------------------------------------------------------------------
// Hot path — Ce que le compilateur doit optimiser
// ---------------------------------------------------------------------------

// calcMidPrice calcule le prix median. Doit rester sur la stack.
// Cette fonction est un bon exemple de ce qu'on VEUT : pas d'allocation.
// Le compilateur devrait l'inliner dans l'appelant.
func calcMidPrice(bid, ask float64) float64 {
	return (bid + ask) / 2.0
}

// hotPath simule le matching engine en mode haute performance.
// Objectif : 0 allocation dans cette fonction.
func hotPath(buy, sell *Order) {
	// BUG #1 en action : newTrade alloue sur le heap
	trade := newTrade(buy.ID, sell.ID, sell.Price, min64(buy.Qty, sell.Qty))

	// BUG #2 en action : logTrade cause du boxing
	logTrade(trade.Price, trade.Qty)

	// Calcul sur la stack (pas d'escape si calcMidPrice est inline)
	mid := calcMidPrice(buy.Price, sell.Price)
	_ = mid
}

// ---------------------------------------------------------------------------
// Main + affichage du guide d'analyse
// ---------------------------------------------------------------------------

func main() {
	fmt.Println("=== Escape Analysis Lab ===")
	fmt.Println()
	fmt.Println("Commande d'analyse :")
	fmt.Println("  go build -gcflags=\"-m\" ./phase4-advanced/exams/exam_escape_analysis.go 2>&1")
	fmt.Println()
	fmt.Println("Cherche les lignes contenant 'escapes to heap' dans l'output.")
	fmt.Println("Chaque echappement dans le hot path = latence GC en production.")
	fmt.Println()

	// Demonstration du hot path
	buy  := &Order{ID: 1, Price: 190.00, Qty: 100}
	sell := &Order{ID: 2, Price: 189.50, Qty: 100}

	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		hotPath(buy, sell)
	}
	elapsed := time.Since(start)

	fmt.Printf("1M iterations du hot path : %v\n", elapsed)
	fmt.Printf("Moyenne par iteration     : %v\n", elapsed/1_000_000)
	fmt.Println()
	fmt.Println("Apres correction, l'allocation doit disparaitre.")
	fmt.Println("La latence moyenne devrait baisser de 50-90%.")
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
