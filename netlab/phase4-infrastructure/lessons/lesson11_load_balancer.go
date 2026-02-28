// Lesson 11 — Load Balancer: round-robin atomic, httputil.ReverseProxy, health checks
// Run: go run phase4-infrastructure/lessons/lesson11_load_balancer.go
//
// Test with curl (while the program is running):
//   for i in {1..9}; do curl -s http://localhost:9100/; done
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// ─── BACKEND ───────────────────────────────────────────────────────────────

// Backend represents one upstream server.
type Backend struct {
	URL     *url.URL
	healthy atomic.Bool
	proxy   *httputil.ReverseProxy
}

func newBackend(rawURL string) *Backend {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("invalid backend URL %q: %v", rawURL, err)
	}
	b := &Backend{URL: u}
	b.healthy.Store(true)

	// httputil.ReverseProxy handles:
	//   - Forwarding the request to the upstream URL
	//   - Copying response headers and body back to the client
	//   - Setting X-Forwarded-For, X-Forwarded-Host, etc.
	b.proxy = httputil.NewSingleHostReverseProxy(u)

	// Custom error handler — the default logs but returns 502 with no body
	b.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[lb] backend %s error: %v", u.Host, err)
		b.healthy.Store(false)
		http.Error(w, "backend unavailable", http.StatusBadGateway)
	}

	return b
}

// ─── LOAD BALANCER ─────────────────────────────────────────────────────────

// LoadBalancer distributes requests across backends using round-robin.
type LoadBalancer struct {
	backends []*Backend
	counter  atomic.Uint64 // atomic: safe to increment from multiple goroutines
	stopCh   chan struct{}
}

func newLoadBalancer(urls []string) *LoadBalancer {
	lb := &LoadBalancer{stopCh: make(chan struct{})}
	for _, u := range urls {
		lb.backends = append(lb.backends, newBackend(u))
	}
	return lb
}

// nextHealthy picks the next healthy backend using round-robin.
// atomic.Uint64.Add is lock-free — critical for high-throughput LBs.
func (lb *LoadBalancer) nextHealthy() *Backend {
	n := len(lb.backends)
	if n == 0 {
		return nil
	}
	// Try up to N times to skip unhealthy backends
	for i := 0; i < n; i++ {
		idx := lb.counter.Add(1) % uint64(n)
		b := lb.backends[idx]
		if b.healthy.Load() {
			return b
		}
	}
	return nil // all backends down
}

// ServeHTTP implements http.Handler.
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b := lb.nextHealthy()
	if b == nil {
		http.Error(w, "503 Service Unavailable — all backends down", http.StatusServiceUnavailable)
		return
	}
	log.Printf("[lb] %s %s → %s", r.Method, r.URL.Path, b.URL.Host)
	b.proxy.ServeHTTP(w, r)
}

// ─── HEALTH CHECKER ────────────────────────────────────────────────────────

// startHealthCheck probes /healthz on each backend periodically.
// A stop channel lets us cleanly shut down the goroutine.
func (lb *LoadBalancer) startHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				lb.probeAll()
			case <-lb.stopCh:
				log.Println("[health] checker stopped")
				return
			}
		}
	}()
}

func (lb *LoadBalancer) Stop() {
	close(lb.stopCh)
}

func (lb *LoadBalancer) probeAll() {
	client := &http.Client{Timeout: 2 * time.Second}
	for _, b := range lb.backends {
		backend := b
		go func() {
			url := backend.URL.String() + "/healthz"
			resp, err := client.Get(url)
			wasHealthy := backend.healthy.Load()

			if err != nil || resp.StatusCode != http.StatusOK {
				backend.healthy.Store(false)
				if wasHealthy {
					log.Printf("[health] DOWN: %s", backend.URL.Host)
				}
				if resp != nil {
					resp.Body.Close()
				}
			} else {
				resp.Body.Close()
				backend.healthy.Store(true)
				if !wasHealthy {
					log.Printf("[health] UP: %s", backend.URL.Host)
				}
			}
		}()
	}
}

// ─── DEMO BACKENDS ─────────────────────────────────────────────────────────

func startDemoBackends(ports []string) {
	for i, port := range ports {
		idx := i + 1
		p := port
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello from backend %d (port %s)\n", idx, p)
		})
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"status":"ok","backend":%d}`, idx)
		})
		srv := &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[backend%d] error: %v", idx, err)
			}
		}()
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	backendPorts := []string{"9101", "9102", "9103"}
	startDemoBackends(backendPorts)
	time.Sleep(50 * time.Millisecond) // wait for backends to bind

	backendURLs := []string{
		"http://127.0.0.1:9101",
		"http://127.0.0.1:9102",
		"http://127.0.0.1:9103",
	}

	lb := newLoadBalancer(backendURLs)
	lb.startHealthCheck(5 * time.Second)
	defer lb.Stop()

	lbServer := &http.Server{
		Addr:    ":9100",
		Handler: lb,
	}
	go lbServer.ListenAndServe()
	time.Sleep(50 * time.Millisecond)

	fmt.Println("Load balancer running on :9100")
	fmt.Println("Backends: :9101, :9102, :9103")
	fmt.Println()

	// Demo: 9 requests → should distribute 3 per backend
	fmt.Println("=== Sending 9 requests (round-robin) ===")
	client := &http.Client{Timeout: 3 * time.Second}
	for i := 1; i <= 9; i++ {
		resp, err := client.Get("http://127.0.0.1:9100/")
		if err != nil {
			fmt.Printf("  request %d: error %v\n", i, err)
			continue
		}
		buf := make([]byte, 128)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()
		fmt.Printf("  request %d → %s", i, buf[:n])
	}

	fmt.Println()
	fmt.Println("Server continues running. Press Ctrl+C to stop.")
	fmt.Println("Test: for i in {1..9}; do curl -s http://localhost:9100/; done")

	// KEY TAKEAWAYS:
	// 1. atomic.Uint64 for round-robin counter: lock-free, thread-safe
	// 2. httputil.ReverseProxy: handles all proxy boilerplate (headers, body, etc.)
	// 3. Health check goroutine: must have a stop channel to avoid goroutine leak
	// 4. nextHealthy() skips down backends by retrying N times
	// 5. Always check len(backends) > 0 before indexing — empty slice → panic

	select {} // block forever
}
