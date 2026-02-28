// trade.go — Resultat d'une execution : le Trade.
// Un Trade est cree chaque fois que le matching engine trouve une contrepartie.

package main

import (
	"fmt"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// Sequence d'IDs pour les trades
// ---------------------------------------------------------------------------

var globalTradeSeq uint64

func nextTradeID() uint64 {
	return atomic.AddUint64(&globalTradeSeq, 1)
}

// ---------------------------------------------------------------------------
// Trade — Enregistrement d'une execution.
// ---------------------------------------------------------------------------
//
// Quand BUY #12 (x100 @ $189.50) rencontre SELL #7 (x100 @ $189.00) :
//   => Trade {
//        BuyOrderID  : 12,
//        SellOrderID : 7,
//        Price       : 189.00,  <- Prix du vendeur (ordre passif)
//        Quantity    : 100,
//      }
//
// REGLE PRIX : c'est l'ordre "passif" (deja dans le book) qui fixe le prix.
//              L'ordre "agressif" (nouveau) accepte ce prix.

type Trade struct {
	ID          uint64
	Symbol      string
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64 // Prix d'execution = prix de l'ordre passif
	Quantity    int64   // Quantite executee (peut etre partielle)
	Timestamp   int64   // Unix nanoseconds
}

// newTrade cree un Trade entre un ordre d'achat et un ordre de vente.
// Cette fonction est appelee uniquement par le matching engine.
func newTrade(symbol string, buyID, sellID uint64, price float64, qty int64, ts int64) Trade {
	return Trade{
		ID:          nextTradeID(),
		Symbol:      symbol,
		BuyOrderID:  buyID,
		SellOrderID: sellID,
		Price:       price,
		Quantity:    qty,
		Timestamp:   ts,
	}
}

// Notional retourne la valeur notionnelle du trade (price * quantity).
// Critique pour le calcul du P&L et des commissions.
func (t Trade) Notional() float64 {
	return t.Price * float64(t.Quantity)
}

// String implemente fmt.Stringer.
func (t Trade) String() string {
	return fmt.Sprintf("TRADE[#%d] %s: BUY#%d vs SELL#%d | x%d @ $%.2f (notional: $%.2f)",
		t.ID, t.Symbol,
		t.BuyOrderID, t.SellOrderID,
		t.Quantity, t.Price,
		t.Notional())
}

// ---------------------------------------------------------------------------
// TradeLog — Historique des trades executes.
// ---------------------------------------------------------------------------

// TradeLog accumule les trades pour analyse post-session.
// En production, cela serait remplace par un publisher vers Kafka/Solace.
type TradeLog struct {
	trades []Trade
}

func NewTradeLog() *TradeLog {
	return &TradeLog{
		trades: make([]Trade, 0, 1024), // pre-alloue 1024 slots
	}
}

// Add ajoute un trade au log.
func (tl *TradeLog) Add(t Trade) {
	tl.trades = append(tl.trades, t)
}

// AddAll ajoute plusieurs trades.
func (tl *TradeLog) AddAll(trades []Trade) {
	tl.trades = append(tl.trades, trades...)
}

// Count retourne le nombre de trades.
func (tl *TradeLog) Count() int {
	return len(tl.trades)
}

// TotalVolume retourne le volume total execute (somme des quantites).
func (tl *TradeLog) TotalVolume() int64 {
	var total int64
	for _, t := range tl.trades {
		total += t.Quantity
	}
	return total
}

// TotalNotional retourne la valeur totale executee.
func (tl *TradeLog) TotalNotional() float64 {
	var total float64
	for _, t := range tl.trades {
		total += t.Notional()
	}
	return total
}

// VWAP retourne le prix moyen pondere par le volume (Volume Weighted Average Price).
// Metrique cle en trading pour evaluer la qualite d'execution.
func (tl *TradeLog) VWAP() float64 {
	vol := tl.TotalVolume()
	if vol == 0 {
		return 0
	}
	return tl.TotalNotional() / float64(vol)
}

// PrintSummary affiche un resume de la session de trading.
func (tl *TradeLog) PrintSummary() {
	fmt.Println("=== TRADE LOG SUMMARY ===")
	fmt.Printf("  Trades executes : %d\n", tl.Count())
	fmt.Printf("  Volume total    : %d actions\n", tl.TotalVolume())
	fmt.Printf("  Notional total  : $%.2f\n", tl.TotalNotional())
	fmt.Printf("  VWAP            : $%.4f\n", tl.VWAP())
}
