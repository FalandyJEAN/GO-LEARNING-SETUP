// Lesson 04 — HTTP Server: ServeMux, JSON responses, middleware (logging + auth)
// Run: go run phase1-fundamentals/lessons/lesson04_http_server.go
//
// Then test with curl:
//   curl http://localhost:8080/healthz
//   curl -H "Authorization: Bearer secret-token-123" http://localhost:8080/orders
//   curl -s -X POST -H "Authorization: Bearer secret-token-123" \
//        -H "Content-Type: application/json" \
//        -d '{"symbol":"TSLA","side":"BUY","qty":10,"price":250.0}' \
//        http://localhost:8080/orders
//   curl -H "Authorization: Bearer wrong-token" http://localhost:8080/orders
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ─── HELPERS ───────────────────────────────────────────────────────────────

// writeJSON sets Content-Type, writes status, encodes v to JSON.
// Centralizing this avoids forgetting headers or writing them in wrong order.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("writeJSON encode error:", err)
	}
}

// ─── MIDDLEWARE ────────────────────────────────────────────────────────────

// loggingMiddleware wraps a handler to log method, path, and duration.
// Middleware signature: func(http.Handler) http.Handler — composable.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %-6s %-30s %v\n",
			time.Now().Format("15:04:05"), r.Method, r.URL.Path, time.Since(start))
	})
}

// authMiddleware rejects requests without a valid Bearer token.
// In production: validate JWTs or call an auth service.
func authMiddleware(validToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "missing Authorization header",
			})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token != validToken {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid token",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ─── ORDER STORE ────────────────────────────────────────────────────────────

type Order struct {
	ID     int     `json:"id"`
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Qty    int     `json:"qty"`
	Price  float64 `json:"price"`
}

// OrderStore is an in-memory store with a mutex for concurrent access.
type OrderStore struct {
	mu     sync.RWMutex
	orders []Order
	nextID int
}

func NewOrderStore() *OrderStore {
	s := &OrderStore{nextID: 1}
	s.orders = []Order{
		{ID: 1, Symbol: "AAPL", Side: "BUY", Qty: 100, Price: 150.0},
		{ID: 2, Symbol: "GOOG", Side: "SELL", Qty: 50, Price: 2800.0},
	}
	s.nextID = 3
	return s
}

func (s *OrderStore) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		orders := make([]Order, len(s.orders))
		copy(orders, s.orders)
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, orders)

	case http.MethodPost:
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if o.Symbol == "" || o.Qty <= 0 || o.Price <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order fields"})
			return
		}
		s.mu.Lock()
		o.ID = s.nextID
		s.nextID++
		s.orders = append(s.orders, o)
		s.mu.Unlock()
		writeJSON(w, http.StatusCreated, o)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	store := NewOrderStore()
	mux := http.NewServeMux()

	// Public endpoint — no auth required
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Protected endpoint — requires valid Bearer token
	mux.Handle("/orders",
		authMiddleware("secret-token-123",
			http.HandlerFunc(store.handleOrders),
		),
	)

	// Apply logging middleware to everything
	handler := loggingMiddleware(mux)

	// Production-grade server config: never use http.ListenAndServe directly
	// without timeouts — a slow client can hold connections open forever.
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("HTTP server running on :8080 (Ctrl+C to stop)")
	fmt.Println()
	fmt.Println("  curl http://localhost:8080/healthz")
	fmt.Println("  curl -H 'Authorization: Bearer secret-token-123' http://localhost:8080/orders")
	fmt.Println()

	log.Fatal(srv.ListenAndServe())

	// KEY TAKEAWAYS:
	// 1. http.ServeMux routes requests by path prefix
	// 2. Middleware pattern: func(http.Handler) http.Handler — stack them with nesting
	// 3. Always set ReadTimeout/WriteTimeout/IdleTimeout on http.Server
	// 4. Use sync.RWMutex for concurrent store access (RLock for reads, Lock for writes)
	// 5. Separate routing (mux) from middleware from business logic (store)
}
