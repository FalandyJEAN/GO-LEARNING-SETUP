// order.go — Types fondamentaux du systeme de matching d'ordres.
// Toutes les structures de base que le reste du projet utilise.

package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// Enumerations (Go n'a pas d'enum natif : on utilise des types string/int)
// ---------------------------------------------------------------------------

// Side represente le cote d'un ordre : achat ou vente.
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// OrderType definit le comportement d'execution de l'ordre.
type OrderType string

const (
	Limit  OrderType = "LIMIT"  // Execute seulement au prix fixe ou mieux
	Market OrderType = "MARKET" // Execute immediatement au meilleur prix dispo
	IOC    OrderType = "IOC"    // Immediate or Cancel : execute ce qui peut l'etre, annule le reste
)

// OrderStatus suit le cycle de vie d'un ordre.
type OrderStatus string

const (
	StatusOpen      OrderStatus = "OPEN"
	StatusPartial   OrderStatus = "PARTIAL"   // Partiellement execute
	StatusFilled    OrderStatus = "FILLED"    // Completement execute
	StatusCancelled OrderStatus = "CANCELLED"
	StatusRejected  OrderStatus = "REJECTED"
)

// ---------------------------------------------------------------------------
// Sequence d'IDs : atomic pour la thread-safety sans mutex
// ---------------------------------------------------------------------------

var globalOrderSeq uint64

func nextOrderID() uint64 {
	return atomic.AddUint64(&globalOrderSeq, 1)
}

// ---------------------------------------------------------------------------
// Order — La structure centrale de tout le systeme.
// ---------------------------------------------------------------------------
//
// NOTE SUR LES TYPES :
//   - ID      : uint64 au lieu de string (plus rapide a comparer, pas d'allocation)
//   - Price   : float64 (simplifie pour ce labo — en prod: int64 en ticks/centimes)
//   - Quantity: int64 (jamais de valeur negative en production)
//   - Timestamp: int64 Unix nanoseconds (plus leger que time.Time pour le hot path)

type Order struct {
	ID        uint64
	Symbol    string
	Side      Side
	Type      OrderType
	Status    OrderStatus
	Price     float64 // 0 pour les Market orders
	Quantity  int64
	Filled    int64 // Quantite deja executee
	Timestamp int64 // Unix nanoseconds — pour la priorite FIFO
}

// NewLimitOrder cree un nouvel ordre a cours limite.
func NewLimitOrder(symbol string, side Side, price float64, qty int64) *Order {
	return &Order{
		ID:        nextOrderID(),
		Symbol:    symbol,
		Side:      side,
		Type:      Limit,
		Status:    StatusOpen,
		Price:     price,
		Quantity:  qty,
		Timestamp: time.Now().UnixNano(),
	}
}

// NewMarketOrder cree un ordre au marche (execute immediatement).
func NewMarketOrder(symbol string, side Side, qty int64) *Order {
	return &Order{
		ID:        nextOrderID(),
		Symbol:    symbol,
		Side:      side,
		Type:      Market,
		Status:    StatusOpen,
		Price:     0,
		Quantity:  qty,
		Timestamp: time.Now().UnixNano(),
	}
}

// Remaining retourne la quantite restante a executer.
func (o *Order) Remaining() int64 {
	return o.Quantity - o.Filled
}

// IsFilled indique si l'ordre est completement execute.
func (o *Order) IsFilled() bool {
	return o.Filled >= o.Quantity
}

// IsActive indique si l'ordre peut encore etre execute.
func (o *Order) IsActive() bool {
	return o.Status == StatusOpen || o.Status == StatusPartial
}

// String implemente fmt.Stringer pour un affichage lisible.
func (o *Order) String() string {
	return fmt.Sprintf("[#%d] %s %s %s x%d/%d @ $%.2f (%s)",
		o.ID, o.Side, o.Type, o.Symbol,
		o.Filled, o.Quantity, o.Price, o.Status)
}

// Reset remet un ordre a zero pour reutilisation via sync.Pool (Phase 4).
// IMPORTANT : ne jamais utiliser un Order apres Reset() sans le reinitialiser.
func (o *Order) Reset() {
	o.ID = 0
	o.Symbol = ""
	o.Side = ""
	o.Type = ""
	o.Status = ""
	o.Price = 0
	o.Quantity = 0
	o.Filled = 0
	o.Timestamp = 0
}
