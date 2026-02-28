// Lesson 03 — HTTP Client: timeout, context, GET/POST JSON, retry
// Run: go run phase1-fundamentals/lessons/lesson03_http_client.go
//
// CARDINAL RULE: NEVER use http.DefaultClient in production.
// It has no timeout — one slow server can hang your entire program.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// ─── CLIENT SETUP ──────────────────────────────────────────────────────────

// newHTTPClient creates a properly configured client.
// All three timeouts matter for different failure scenarios:
//   - Timeout          : wall-clock limit for the entire request
//   - IdleConnTimeout  : how long to keep idle keep-alive connections
//   - MaxIdleConns     : connection pool size
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// ─── GET WITH CONTEXT ──────────────────────────────────────────────────────

// getJSON fetches a URL and decodes the JSON response.
// context.WithTimeout adds a per-request deadline ON TOP of client.Timeout.
// Always defer cancel() — even if the context times out, cancel frees resources.
func getJSON(client *http.Client, url string, out any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // IMPORTANT: prevents goroutine leak in net/http internals

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// ─── POST JSON ─────────────────────────────────────────────────────────────

// postJSON marshals payload to JSON, POSTs it, and decodes the response.
func postJSON(client *http.Client, url string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", resp.StatusCode, raw)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// ─── RETRY ─────────────────────────────────────────────────────────────────

// retryGet retries a GET up to maxAttempts using exponential backoff.
// Pattern used by: Kubernetes client-go, AWS SDK, Terraform providers.
func retryGet(client *http.Client, url string, maxAttempts int, out any) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<attempt) * 100 * time.Millisecond // 200ms, 400ms, 800ms…
			fmt.Printf("[retry] attempt %d, backing off %v\n", attempt+1, wait)
			time.Sleep(wait)
		}
		if err := getJSON(client, url, out); err == nil {
			return nil
		} else {
			lastErr = err
			fmt.Printf("[retry] error: %v\n", err)
		}
	}
	return fmt.Errorf("all %d attempts failed, last: %w", maxAttempts, lastErr)
}

// ─── DEMO SERVER ───────────────────────────────────────────────────────────

type Order struct {
	Symbol   string  `json:"symbol"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Side     string  `json:"side"`
}

func newDemoServer() *httptest.Server {
	mux := http.NewServeMux()

	orders := []Order{
		{Symbol: "AAPL", Quantity: 100, Price: 150.00, Side: "BUY"},
		{Symbol: "GOOG", Quantity: 50, Price: 2800.00, Side: "SELL"},
	}

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(orders)

		case http.MethodPost:
			var o Order
			json.NewDecoder(r.Body).Decode(&o)
			orders = append(orders, o)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "accepted",
				"id":     len(orders),
				"symbol": o.Symbol,
			})

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // simulate slow endpoint → will trigger timeout
		w.Write([]byte(`{"status":"too late"}`))
	})

	return httptest.NewServer(mux)
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	server := newDemoServer()
	defer server.Close()

	client := newHTTPClient()

	// ── Demo 1: GET ──────────────────────────────────────────────────────
	fmt.Println("=== GET /orders ===")
	var orders []Order
	if err := getJSON(client, server.URL+"/orders", &orders); err != nil {
		fmt.Println("GET error:", err)
	} else {
		for _, o := range orders {
			fmt.Printf("  %s %s %d @ %.2f\n", o.Side, o.Symbol, o.Quantity, o.Price)
		}
	}

	// ── Demo 2: POST JSON ────────────────────────────────────────────────
	fmt.Println("\n=== POST /orders ===")
	newOrder := Order{Symbol: "MSFT", Quantity: 200, Price: 300.00, Side: "BUY"}
	var ack map[string]any
	if err := postJSON(client, server.URL+"/orders", newOrder, &ack); err != nil {
		fmt.Println("POST error:", err)
	} else {
		fmt.Printf("  Acknowledged: %v\n", ack)
	}

	// ── Demo 3: Context timeout ──────────────────────────────────────────
	fmt.Println("\n=== Context timeout (slow endpoint) ===")
	var dummy map[string]any
	err := getJSON(client, server.URL+"/slow", &dummy)
	fmt.Printf("  Expected timeout: %v\n", err)

	// ── Demo 4: Retry on 404 ─────────────────────────────────────────────
	fmt.Println("\n=== Retry (3 attempts on 404) ===")
	err = retryGet(client, server.URL+"/nonexistent", 3, &dummy)
	fmt.Printf("  Final error (expected): %v\n", err)

	// KEY TAKEAWAYS:
	// 1. Never use http.DefaultClient — always set Timeout
	// 2. http.NewRequestWithContext + defer cancel() prevents goroutine leaks
	// 3. Always defer resp.Body.Close() to return connection to pool
	// 4. Check resp.StatusCode — no error doesn't mean success (404 is not an error)
	// 5. Retry with exponential backoff: standard practice for distributed systems
}
