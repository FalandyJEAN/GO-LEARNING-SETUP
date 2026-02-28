// gateway.go — Couche de validation et de routage des ordres entrants.
// Le Gateway est la premiere ligne de defense avant le matching engine.

package main

import (
	"fmt"
)

// ---------------------------------------------------------------------------
// Erreurs metier specifiques
// ---------------------------------------------------------------------------

// ValidationError encapsule une erreur de validation avec le contexte.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

// ---------------------------------------------------------------------------
// Gateway — Valide et route les ordres
// ---------------------------------------------------------------------------

// Gateway est le point d'entree du systeme.
// Il valide les ordres, puis les soumet au bon OrderBook.
type Gateway struct {
	books map[string]*OrderBook // symbol -> OrderBook
	log   *TradeLog
}

// NewGateway cree un Gateway avec les symboles pre-enregistres.
func NewGateway(symbols []string, log *TradeLog) *Gateway {
	books := make(map[string]*OrderBook, len(symbols))
	for _, s := range symbols {
		books[s] = NewOrderBook(s)
	}
	return &Gateway{
		books: books,
		log:   log,
	}
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// validateOrder verifie qu'un ordre est conforme aux regles metier.
// Cette fonction est le SEUL endroit ou la validation est effectuee.
// Pattern : retourner une erreur explicite, jamais un bool silencieux.
func validateOrder(o *Order) error {
	if o == nil {
		return &ValidationError{Field: "order", Message: "ordre nil"}
	}

	if o.Symbol == "" {
		return &ValidationError{Field: "symbol", Message: "symbole vide"}
	}

	if o.Side != Buy && o.Side != Sell {
		return &ValidationError{Field: "side", Message: fmt.Sprintf("cote invalide: %q", o.Side)}
	}

	if o.Type != Limit && o.Type != Market && o.Type != IOC {
		return &ValidationError{Field: "type", Message: fmt.Sprintf("type invalide: %q", o.Type)}
	}

	if o.Quantity <= 0 {
		return &ValidationError{Field: "quantity", Message: fmt.Sprintf("quantite doit etre > 0, recu: %d", o.Quantity)}
	}

	// Les Market orders n'ont pas de prix limite
	if o.Type == Limit && o.Price <= 0 {
		return &ValidationError{Field: "price", Message: fmt.Sprintf("prix limite doit etre > 0, recu: %.2f", o.Price)}
	}

	// Garde-fou contre les prix aberrants (circuit breaker simplifie)
	if o.Type == Limit && o.Price > 1_000_000 {
		return &ValidationError{Field: "price", Message: fmt.Sprintf("prix anormalement eleve: %.2f", o.Price)}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Submit — Validation + Routing + Matching
// ---------------------------------------------------------------------------

// Submit valide un ordre, le route vers le bon book, et retourne les trades.
// C'est la methode principale appelee par les clients.
func (gw *Gateway) Submit(o *Order) ([]Trade, error) {
	// Etape 1 : Validation
	if err := validateOrder(o); err != nil {
		o.Status = StatusRejected
		return nil, fmt.Errorf("ordre #%d rejete: %w", o.ID, err)
	}

	// Etape 2 : Routing vers le bon OrderBook
	book, exists := gw.books[o.Symbol]
	if !exists {
		o.Status = StatusRejected
		return nil, fmt.Errorf("ordre #%d rejete: symbole %q non supporte", o.ID, o.Symbol)
	}

	// Etape 3 : Matching
	trades := book.Submit(o)

	// Etape 4 : Logging des trades
	if gw.log != nil {
		gw.log.AddAll(trades)
	}

	return trades, nil
}

// Cancel annule un ordre dans le book correspondant.
func (gw *Gateway) Cancel(symbol string, orderID uint64) error {
	book, exists := gw.books[symbol]
	if !exists {
		return fmt.Errorf("symbole %q non supporte", symbol)
	}
	if !book.Cancel(orderID) {
		return fmt.Errorf("ordre #%d non trouve ou deja inactif", orderID)
	}
	return nil
}

// Book retourne le OrderBook d'un symbole (pour affichage/monitoring).
func (gw *Gateway) Book(symbol string) (*OrderBook, bool) {
	b, ok := gw.books[symbol]
	return b, ok
}
