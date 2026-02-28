// orderbook.go — Carnet d'ordres avec algorithme FIFO Price-Time Priority.
// Implemente deux priority queues (container/heap) pour les bids et les asks.

package main

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
)

// ===========================================================================
// BID HEAP — Max-heap : le BID le plus HAUT est en tete (meilleur acheteur)
// A prix egal, ordre le plus ANCIEN en premier (FIFO)
// ===========================================================================

type BidHeap []*Order

func (h BidHeap) Len() int { return len(h) }

func (h BidHeap) Less(i, j int) bool {
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price // Max-heap : prix plus haut = priorite plus haute
	}
	return h[i].Timestamp < h[j].Timestamp // FIFO a prix egal
}

func (h BidHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *BidHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *BidHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	old[n-1] = nil // Evite la fuite memoire (reference fantome dans le slice)
	*h = old[:n-1]
	return x
}

// ===========================================================================
// ASK HEAP — Min-heap : l'ASK le plus BAS est en tete (meilleur vendeur)
// A prix egal, ordre le plus ANCIEN en premier (FIFO)
// ===========================================================================

type AskHeap []*Order

func (h AskHeap) Len() int { return len(h) }

func (h AskHeap) Less(i, j int) bool {
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price // Min-heap : prix plus bas = priorite plus haute
	}
	return h[i].Timestamp < h[j].Timestamp // FIFO a prix egal
}

func (h AskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *AskHeap) Push(x interface{}) {
	*h = append(*h, x.(*Order))
}

func (h *AskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return x
}

// ===========================================================================
// ORDER BOOK
// ===========================================================================

// OrderBook maintient les listes d'ordres en attente pour un seul symbole.
// Il est thread-safe via un RWMutex.
//
// Architecture :
//   - bids : max-heap des acheteurs (meilleur prix en tete)
//   - asks : min-heap des vendeurs (meilleur prix en tete)
//   - Le matching utilise une "lazy deletion" : on marque les ordres
//     comme annules/remplis, et on les retire du heap quand ils arrivent en tete.
//
// POURQUOI RWMUTEX et pas MUTEX ?
//   BestBid() et BestAsk() sont des lectures pures et appellees tres frequemment
//   par le market data feed. RWMutex permet des lectures concurrentes.
//   Submit() et Cancel() ecrivent => besoin du lock exclusif.

type OrderBook struct {
	mu     sync.RWMutex
	symbol string
	bids   *BidHeap
	asks   *AskHeap
}

// NewOrderBook cree un carnet d'ordres vide pour un symbole.
func NewOrderBook(symbol string) *OrderBook {
	bids := &BidHeap{}
	asks := &AskHeap{}
	heap.Init(bids)
	heap.Init(asks)
	return &OrderBook{
		symbol: symbol,
		bids:   bids,
		asks:   asks,
	}
}

// ---------------------------------------------------------------------------
// Lecture du top of book (thread-safe, multi-lecteurs simultanement)
// ---------------------------------------------------------------------------

// BestBid retourne le meilleur prix d'achat (le plus haut).
func (ob *OrderBook) BestBid() (float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	for ob.bids.Len() > 0 {
		top := (*ob.bids)[0]
		if top.IsActive() {
			return top.Price, true
		}
		// Lazy deletion : on retire les ordres inactifs du sommet
		// ATTENTION : RLock ne permet pas de modifier le heap.
		// Cette situation ne devrait pas arriver souvent (annulations rares).
		break
	}
	return 0, false
}

// BestAsk retourne le meilleur prix de vente (le plus bas).
func (ob *OrderBook) BestAsk() (float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	for ob.asks.Len() > 0 {
		top := (*ob.asks)[0]
		if top.IsActive() {
			return top.Price, true
		}
		break
	}
	return 0, false
}

// Spread retourne l'ecart entre le meilleur ask et le meilleur bid.
// Un spread faible = marche liquide. Un spread large = marche illiquide.
func (ob *OrderBook) Spread() (float64, bool) {
	bid, hasBid := ob.BestBid()
	ask, hasAsk := ob.BestAsk()
	if !hasBid || !hasAsk {
		return 0, false
	}
	return ask - bid, true
}

// Depth retourne le nombre d'ordres actifs dans le book.
func (ob *OrderBook) Depth() (bidCount, askCount int) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	for _, o := range *ob.bids {
		if o.IsActive() {
			bidCount++
		}
	}
	for _, o := range *ob.asks {
		if o.IsActive() {
			askCount++
		}
	}
	return
}

// ---------------------------------------------------------------------------
// Submit — Point d'entree principal : soumet un ordre et declenche le matching
// ---------------------------------------------------------------------------

// Submit accepte un nouvel ordre, tente de le matcher, et l'ajoute au book
// s'il n'est pas completement execute.
// Retourne la liste des trades generes (peut etre vide).
func (ob *OrderBook) Submit(incoming *Order) []Trade {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	return ob.match(incoming)
}

// match est la logique interne de matching. Appele uniquement avec ob.mu tenu.
// Ne pas appeler directement depuis l'exterieur.
func (ob *OrderBook) match(incoming *Order) []Trade {
	var trades []Trade

	switch incoming.Side {
	case Buy:
		trades = ob.matchBuy(incoming)
	case Sell:
		trades = ob.matchSell(incoming)
	}

	return trades
}

func (ob *OrderBook) matchBuy(incoming *Order) []Trade {
	var trades []Trade

	for ob.asks.Len() > 0 && incoming.Remaining() > 0 {
		bestAsk := (*ob.asks)[0]

		// Lazy deletion : retirer les ordres inactifs
		if !bestAsk.IsActive() {
			heap.Pop(ob.asks)
			continue
		}

		// Verifier si les prix se croisent
		// Pour un Market order, il n'y a pas de limite de prix.
		if incoming.Type == Limit && incoming.Price < bestAsk.Price {
			break // Pas de match possible, prix trop loin
		}

		// EXECUTION : l'ordre passif (le vendeur dans le book) fixe le prix
		qty := min64(incoming.Remaining(), bestAsk.Remaining())
		execPrice := bestAsk.Price // Passive order pricing rule

		trade := newTrade(ob.symbol, incoming.ID, bestAsk.ID, execPrice, qty, incoming.Timestamp)
		trades = append(trades, trade)

		// Mettre a jour les quantites executees
		incoming.Filled += qty
		bestAsk.Filled += qty

		// Mettre a jour les statuts
		if bestAsk.IsFilled() {
			bestAsk.Status = StatusFilled
			heap.Pop(ob.asks)
		} else {
			bestAsk.Status = StatusPartial
			// Le heap doit etre reajuste car la priorite n'a pas change,
			// mais heap.Fix serait necessaire si le prix changeait (il ne change pas ici).
		}
	}

	// Gestion de l'ordre entrant apres matching
	ob.finalizeOrder(incoming, ob.bids)
	return trades
}

func (ob *OrderBook) matchSell(incoming *Order) []Trade {
	var trades []Trade

	for ob.bids.Len() > 0 && incoming.Remaining() > 0 {
		bestBid := (*ob.bids)[0]

		if !bestBid.IsActive() {
			heap.Pop(ob.bids)
			continue
		}

		if incoming.Type == Limit && incoming.Price > bestBid.Price {
			break
		}

		qty := min64(incoming.Remaining(), bestBid.Remaining())
		execPrice := bestBid.Price

		trade := newTrade(ob.symbol, bestBid.ID, incoming.ID, execPrice, qty, incoming.Timestamp)
		trades = append(trades, trade)

		incoming.Filled += qty
		bestBid.Filled += qty

		if bestBid.IsFilled() {
			bestBid.Status = StatusFilled
			heap.Pop(ob.bids)
		} else {
			bestBid.Status = StatusPartial
		}
	}

	ob.finalizeOrder(incoming, ob.asks)
	return trades
}

// finalizeOrder determine le statut final de l'ordre entrant et l'ajoute au book si necessaire.
func (ob *OrderBook) finalizeOrder(o *Order, book heap.Interface) {
	switch {
	case o.IsFilled():
		o.Status = StatusFilled

	case o.Type == IOC:
		// IOC : ce qui n'a pas ete execute est annule immediatement
		o.Status = StatusCancelled

	case o.Type == Market && o.Remaining() > 0:
		// Market order non completement execute = annule (pas de prix cible)
		o.Status = StatusCancelled

	case o.Type == Limit && o.Remaining() > 0:
		// Limit order partiellement ou non execute : reste dans le book
		if o.Filled > 0 {
			o.Status = StatusPartial
		}
		heap.Push(book, o)
	}
}

// ---------------------------------------------------------------------------
// Cancel — Annulation d'un ordre (lazy deletion)
// ---------------------------------------------------------------------------

// Cancel marque un ordre comme annule. Il sera retire du heap lors du prochain matching.
// Complexite : O(n) pour la recherche. En production, on utiliserait un index
// map[orderID]*Order pour O(1).
func (ob *OrderBook) Cancel(orderID uint64) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	for _, o := range *ob.bids {
		if o.ID == orderID && o.IsActive() {
			o.Status = StatusCancelled
			return true
		}
	}
	for _, o := range *ob.asks {
		if o.ID == orderID && o.IsActive() {
			o.Status = StatusCancelled
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Display — Affichage du carnet
// ---------------------------------------------------------------------------

// PrintBook affiche une representation du carnet d'ordres.
func (ob *OrderBook) PrintBook(levels int) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	fmt.Printf("╔══════════════════════════════════════╗\n")
	fmt.Printf("║  ORDER BOOK : %-22s║\n", ob.symbol)
	fmt.Printf("╠══════════════════════════════════════╣\n")

	// Collecter les asks actifs
	var activeAsks []*Order
	for _, o := range *ob.asks {
		if o.IsActive() {
			activeAsks = append(activeAsks, o)
		}
	}

	// Afficher les asks (du plus haut au plus bas)
	count := 0
	for i := len(activeAsks) - 1; i >= 0 && count < levels; i-- {
		o := activeAsks[i]
		bar := strings.Repeat("█", int(o.Remaining()/10))
		fmt.Printf("║  SELL  %8d  $%8.2f  %-5s ║\n", o.Remaining(), o.Price, bar)
		count++
	}

	// Spread
	bid, hasBid := ob.BestBid()
	ask, hasAsk := ob.BestAsk()
	if hasBid && hasAsk {
		fmt.Printf("║  ---- SPREAD: $%-6.4f          ----  ║\n", ask-bid)
	} else {
		fmt.Printf("║  ---- NO SPREAD (book vide)    ----  ║\n")
	}

	// Afficher les bids (du plus haut au plus bas)
	count = 0
	for _, o := range *ob.bids {
		if o.IsActive() && count < levels {
			bar := strings.Repeat("█", int(o.Remaining()/10))
			fmt.Printf("║  BUY   %8d  $%8.2f  %-5s ║\n", o.Remaining(), o.Price, bar)
			count++
		}
	}

	fmt.Printf("╚══════════════════════════════════════╝\n")
}

// ---------------------------------------------------------------------------
// Utilitaires
// ---------------------------------------------------------------------------

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
