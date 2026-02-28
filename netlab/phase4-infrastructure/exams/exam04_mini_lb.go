// EXAM 04 — Mini Load Balancer (contient 4 bugs intentionnels)
// Run:  go run  phase4-infrastructure/exams/exam04_mini_lb.go
// Race: go run -race phase4-infrastructure/exams/exam04_mini_lb.go
//
// Lis exam04_instructions.txt avant de commencer.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// ─── LOAD BALANCER (BUGGY) ─────────────────────────────────────────────────

type Backend struct {
	URL     *url.URL
	healthy atomic.Bool
	proxy   *httputil.ReverseProxy
}

func newBackend(rawURL string) *Backend {
	u, _ := url.Parse(rawURL)
	b := &Backend{URL: u}
	b.healthy.Store(true)
	b.proxy = httputil.NewSingleHostReverseProxy(u)
	return b
}

type LoadBalancer struct {
	backends []*Backend
	// BUG #1: counter is a plain int, not atomic.
	// When multiple goroutines call ServeHTTP concurrently, they race on counter.
	// go run -race will detect: "DATA RACE on lb.counter"
	// Fix: replace with  counter atomic.Uint64  and use  lb.counter.Add(1)
	counter int
}

func newLoadBalancer(urls []string) *LoadBalancer {
	lb := &LoadBalancer{}
	for _, u := range urls {
		lb.backends = append(lb.backends, newBackend(u))
	}
	return lb
}

// nextBackend picks the next healthy backend using round-robin.
func (lb *LoadBalancer) nextBackend() *Backend {
	// BUG #2: if lb.backends is empty, len(lb.backends) == 0
	// and the modulo  lb.counter % 0  causes a panic: integer divide by zero.
	// Fix: add  if len(lb.backends) == 0 { return nil }  before this line.

	// BUG #1 (continued): lb.counter++ is not atomic — race condition.
	lb.counter++
	idx := lb.counter % len(lb.backends)
	b := lb.backends[idx]
	if b.healthy.Load() {
		return b
	}
	return nil
}

// StartHealthCheck starts a background goroutine that probes backends.
func (lb *LoadBalancer) StartHealthCheck(interval time.Duration) {
	// BUG #3: this goroutine runs forever — there is no stop channel.
	// Every call to StartHealthCheck() leaks a goroutine.
	// Fix: add a stopCh chan struct{} field to LoadBalancer,
	//      and use select { case <-ticker.C: ... case <-lb.stopCh: return }
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C { // BUG: no stop mechanism
			lb.probeAll()
		}
	}()
}

func (lb *LoadBalancer) probeAll() {
	client := &http.Client{Timeout: 1 * time.Second}
	for _, b := range lb.backends {
		backend := b
		go func() {
			resp, err := client.Get(backend.URL.String() + "/healthz")
			if err != nil || resp.StatusCode != 200 {
				backend.healthy.Store(false)
			} else {
				resp.Body.Close()
				backend.healthy.Store(true)
			}
		}()
	}
}

// ServeHTTP handles incoming requests and forwards them to a backend.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// BUG #4: io.ReadAll reads req.Body completely.
	// req.Body is an io.ReadCloser — once read, it returns EOF.
	// newBackend's ReverseProxy will forward an empty body to the upstream.
	// Fix: save bodyBytes, then restore: req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	bodyBytes, _ := io.ReadAll(req.Body)
	log.Printf("[lb] %s %s  body=%d bytes", req.Method, req.URL.Path, len(bodyBytes))
	// req.Body is now exhausted — NOT restored

	backend := lb.nextBackend()
	if backend == nil {
		http.Error(w, "503 no healthy backends", http.StatusServiceUnavailable)
		return
	}

	log.Printf("[lb] → %s", backend.URL.Host)
	backend.proxy.ServeHTTP(w, req)
}

// ─── DEMO ──────────────────────────────────────────────────────────────────

func main() {
	// Start 2 backend servers
	backends := make([]*httptest.Server, 2)
	for i := range backends {
		idx := i + 1
		backends[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/healthz":
				fmt.Fprintf(w, `{"status":"ok","backend":%d}`, idx)
			default:
				body, _ := io.ReadAll(r.Body)
				fmt.Fprintf(w, "backend%d received body: %q", idx, string(body))
			}
		}))
		defer backends[i].Close()
		log.Printf("[backend%d] at %s", idx, backends[i].URL)
	}

	lb := newLoadBalancer([]string{backends[0].URL, backends[1].URL})
	lb.StartHealthCheck(2 * time.Second)

	lbServer := httptest.NewServer(lb)
	defer lbServer.Close()
	log.Printf("[proxy] at %s", lbServer.URL)

	client := &http.Client{Timeout: 3 * time.Second}

	// Test 1: Round-robin GET (BUG #1 — race if concurrent)
	fmt.Println("\n--- Test 1: GET round-robin ---")
	for i := 1; i <= 4; i++ {
		resp, _ := client.Get(lbServer.URL + "/")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("  request %d: %s\n", i, body)
	}

	// Test 2: POST with body (BUG #4 — body should NOT be empty)
	fmt.Println("\n--- Test 2: POST body forwarding ---")
	orderJSON := `{"symbol":"AAPL","qty":100}`
	resp, _ := client.Post(lbServer.URL+"/orders", "application/json",
		strings.NewReader(orderJSON))
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("  backend response: %s\n", body)
	fmt.Println("  (body should NOT be empty after BUG #4 fix)")

	// Test 3: Empty backends (BUG #2 — should return 503, not panic)
	fmt.Println("\n--- Test 3: Empty backends (BUG #2 panic test) ---")
	emptyLB := newLoadBalancer([]string{}) // 0 backends
	emptyServer := httptest.NewServer(emptyLB)
	defer emptyServer.Close()
	resp2, err := client.Get(emptyServer.URL + "/")
	if err != nil {
		fmt.Println("  ERROR (panic likely):", err)
	} else {
		body2, _ := io.ReadAll(resp2.Body)
		resp2.Body.Close()
		fmt.Printf("  status=%d body=%s\n", resp2.StatusCode, body2)
	}
}
