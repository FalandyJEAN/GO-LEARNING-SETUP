// EXAM 02 — HTTP Reverse Proxy (contient 4 bugs intentionnels)
// Run: go run phase2-protocols/exams/exam02_http_proxy.go
//
// Lis exam02_instructions.txt avant de commencer.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// ─── PROXY IMPLEMENTATION ──────────────────────────────────────────────────

// copyHeaders copies all headers from src to dst.
func copyHeaders(dst, src http.Header) {
	// BUG #2: copies ALL headers, including hop-by-hop headers that must NOT
	// be forwarded by a proxy (RFC 7230 Section 6.1).
	// Hop-by-hop: Connection, Keep-Alive, Transfer-Encoding, Upgrade,
	//             Proxy-Authorization, TE, Trailers
	// Fix: filter these out before copying.
	for key, values := range src {
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}

// buildUpstreamRequest creates the outgoing request to the backend.
func buildUpstreamRequest(r *http.Request, backendURL string) (*http.Request, error) {
	target := backendURL + r.URL.RequestURI()

	// BUG #3: the error from http.NewRequest is silently ignored with _.
	// If backendURL is malformed, outReq is nil and copyHeaders panics.
	// Fix: return the error: outReq, err := http.NewRequest(...); if err != nil { return nil, err }
	outReq, _ := http.NewRequest(r.Method, target, r.Body)
	copyHeaders(outReq.Header, r.Header)
	outReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	return outReq, nil
}

// writeResponse writes the backend response back to the client.
func writeResponse(w http.ResponseWriter, resp *http.Response) {
	copyHeaders(w.Header(), resp.Header)

	body, _ := io.ReadAll(resp.Body)

	// BUG #4: w.Write() is called BEFORE w.WriteHeader().
	// The first call to w.Write() triggers an implicit WriteHeader(200).
	// The subsequent w.WriteHeader(resp.StatusCode) is a no-op and logs:
	//   "http: superfluous response.WriteHeader call"
	// Fix: swap the two lines — call w.WriteHeader(resp.StatusCode) first.
	w.Write(body)                  // BUG: implicitly sends 200 status
	w.WriteHeader(resp.StatusCode) // BUG: too late, status already sent
}

// logAndProxy logs the request body and proxies the request to the backend.
func logAndProxy(w http.ResponseWriter, r *http.Request, backendURL string) {
	// BUG #1: io.ReadAll consumes r.Body entirely.
	// r.Body is an io.ReadCloser — once read, it is empty.
	// buildUpstreamRequest then passes the empty r.Body to the upstream request.
	// POST requests arrive at the backend with a zero-length body.
	//
	// Fix:
	//   bodyBytes, _ := io.ReadAll(r.Body)
	//   r.Body.Close()
	//   log.Printf(...)
	//   r.Body = io.NopCloser(bytes.NewReader(bodyBytes))  // restore body
	bodyBytes, _ := io.ReadAll(r.Body)
	log.Printf("[proxy] %s %s  body_logged=%d bytes", r.Method, r.URL.Path, len(bodyBytes))
	// r.Body is now empty — NOT restored before buildUpstreamRequest

	client := &http.Client{Timeout: 10 * time.Second}

	outReq, err := buildUpstreamRequest(r, backendURL)
	if err != nil {
		http.Error(w, "proxy build error: "+err.Error(), http.StatusBadGateway)
		return
	}

	resp, err := client.Do(outReq)
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	writeResponse(w, resp)
}

// ─── DEMO ──────────────────────────────────────────────────────────────────

func main() {
	// Backend server — simulates a real microservice
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/orders":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"orders":[{"id":1,"symbol":"AAPL"}]}`)

		case r.Method == http.MethodPost && r.URL.Path == "/orders":
			received, _ := io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"status":"accepted","received":"%s"}`, string(received))

		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()
	log.Printf("[backend] running at %s", backend.URL)

	// Proxy server
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logAndProxy(w, r, backend.URL)
	}))
	defer proxy.Close()
	log.Printf("[proxy]   running at %s", proxy.URL)

	client := &http.Client{Timeout: 5 * time.Second}

	// ── Test 1: GET /orders ───────────────────────────────────────────────
	fmt.Println("\n--- Test 1: GET /orders ---")
	resp, _ := client.Get(proxy.URL + "/orders")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("Status: %d | Body: %s\n", resp.StatusCode, body)

	// ── Test 2: POST /orders (BUG #1 — body should NOT be empty) ─────────
	fmt.Println("\n--- Test 2: POST /orders (watch for empty received body) ---")
	orderJSON := `{"symbol":"MSFT","qty":100,"price":300.0}`
	resp2, _ := client.Post(proxy.URL+"/orders", "application/json",
		strings.NewReader(orderJSON))
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	fmt.Printf("Status: %d | Body: %s\n", resp2.StatusCode, body2)
	fmt.Println("Expected: received field should contain the JSON, not empty string")

	// ── Test 3: 404 status code (BUG #4 — status should be 404, not 200) ─
	fmt.Println("\n--- Test 3: 404 path (watch for wrong status code) ---")
	resp3, _ := client.Get(proxy.URL + "/nonexistent")
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	fmt.Printf("Status: %d (expected 404) | Body: %s\n", resp3.StatusCode, body3)
}
