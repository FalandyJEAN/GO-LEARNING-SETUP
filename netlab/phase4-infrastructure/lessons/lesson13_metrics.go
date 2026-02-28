// Lesson 13 — Metrics: Counter, Gauge, Histogram, /healthz, /metrics (Prometheus-style)
// Run: go run phase4-infrastructure/lessons/lesson13_metrics.go
//
// Then check:
//   curl http://localhost:9090/healthz
//   curl http://localhost:9090/metrics
//
// This implements a simplified Prometheus text format (OpenMetrics).
// In production: use github.com/prometheus/client_golang
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// COUNTER — monotonically increasing (requests_total, errors_total)
// ═══════════════════════════════════════════════════════════════════════════

type Counter struct {
	v    atomic.Uint64
	name string
	help string
}

func (c *Counter) Inc()            { c.v.Add(1) }
func (c *Counter) Add(n uint64)   { c.v.Add(n) }
func (c *Counter) Value() uint64  { return c.v.Load() }

// ═══════════════════════════════════════════════════════════════════════════
// GAUGE — can go up or down (active_connections, queue_depth, cpu_usage)
// ═══════════════════════════════════════════════════════════════════════════

type Gauge struct {
	v    atomic.Int64
	name string
	help string
}

func (g *Gauge) Set(v int64) { g.v.Store(v) }
func (g *Gauge) Inc()        { g.v.Add(1) }
func (g *Gauge) Dec()        { g.v.Add(-1) }
func (g *Gauge) Value() int64 { return g.v.Load() }

// ═══════════════════════════════════════════════════════════════════════════
// HISTOGRAM — distribution of values (request_latency, order_size)
//
// Stores cumulative bucket counts: how many observations <= each bound.
// Enables percentile calculation (p50, p95, p99) in Grafana/Prometheus.
// ═══════════════════════════════════════════════════════════════════════════

type Histogram struct {
	mu      sync.Mutex
	name    string
	help    string
	buckets []float64 // upper bounds in seconds
	counts  []uint64  // counts[i] = observations <= buckets[i], counts[last] = +Inf
	sum     float64
	count   uint64
}

func NewHistogram(name, help string, buckets []float64) *Histogram {
	b := make([]float64, len(buckets))
	copy(b, buckets)
	sort.Float64s(b)
	return &Histogram{
		name:    name,
		help:    help,
		buckets: b,
		counts:  make([]uint64, len(b)+1), // +1 for +Inf bucket
	}
}

func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sum += v
	h.count++
	// Cumulative: each bucket counts observations ≤ its upper bound
	for i, bound := range h.buckets {
		if v <= bound {
			h.counts[i]++
		}
	}
	h.counts[len(h.buckets)]++ // +Inf always gets every observation
}

// Percentile estimates the p-th percentile (p in 0–100).
func (h *Histogram) Percentile(p float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	target := uint64(math.Ceil(float64(h.count) * p / 100))
	for i, cnt := range h.counts {
		if cnt >= target {
			if i < len(h.buckets) {
				return h.buckets[i]
			}
			return math.Inf(1)
		}
	}
	return math.Inf(1)
}

// ═══════════════════════════════════════════════════════════════════════════
// REGISTRY — collects all metrics and renders /metrics
// ═══════════════════════════════════════════════════════════════════════════

type Registry struct {
	mu         sync.RWMutex
	counters   []*Counter
	gauges     []*Gauge
	histograms []*Histogram
}

var defaultReg = &Registry{}

func NewCounter(name, help string) *Counter {
	c := &Counter{name: name, help: help}
	defaultReg.mu.Lock()
	defaultReg.counters = append(defaultReg.counters, c)
	defaultReg.mu.Unlock()
	return c
}

func NewGauge(name, help string) *Gauge {
	g := &Gauge{name: name, help: help}
	defaultReg.mu.Lock()
	defaultReg.gauges = append(defaultReg.gauges, g)
	defaultReg.mu.Unlock()
	return g
}

func RegisterHistogram(h *Histogram) *Histogram {
	defaultReg.mu.Lock()
	defaultReg.histograms = append(defaultReg.histograms, h)
	defaultReg.mu.Unlock()
	return h
}

// Render produces Prometheus text format output.
func (r *Registry) Render() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder

	for _, c := range r.counters {
		fmt.Fprintf(&sb, "# HELP %s %s\n", c.name, c.help)
		fmt.Fprintf(&sb, "# TYPE %s counter\n", c.name)
		fmt.Fprintf(&sb, "%s %d\n\n", c.name, c.Value())
	}

	for _, g := range r.gauges {
		fmt.Fprintf(&sb, "# HELP %s %s\n", g.name, g.help)
		fmt.Fprintf(&sb, "# TYPE %s gauge\n", g.name)
		fmt.Fprintf(&sb, "%s %d\n\n", g.name, g.Value())
	}

	for _, h := range r.histograms {
		h.mu.Lock()
		fmt.Fprintf(&sb, "# HELP %s %s\n", h.name, h.help)
		fmt.Fprintf(&sb, "# TYPE %s histogram\n", h.name)
		for i, bound := range h.buckets {
			fmt.Fprintf(&sb, "%s_bucket{le=\"%.4f\"} %d\n", h.name, bound, h.counts[i])
		}
		fmt.Fprintf(&sb, "%s_bucket{le=\"+Inf\"} %d\n", h.name, h.counts[len(h.buckets)])
		fmt.Fprintf(&sb, "%s_sum %.4f\n", h.name, h.sum)
		fmt.Fprintf(&sb, "%s_count %d\n\n", h.name, h.count)
		h.mu.Unlock()
	}

	return sb.String()
}

// ═══════════════════════════════════════════════════════════════════════════
// APPLICATION METRICS
// ═══════════════════════════════════════════════════════════════════════════

var (
	httpRequestsTotal = NewCounter("http_requests_total", "Total HTTP requests received")
	httpErrorsTotal   = NewCounter("http_errors_total", "Total HTTP requests that returned 5xx")
	activeConns       = NewGauge("active_connections", "Number of currently active HTTP connections")
	orderQueueDepth   = NewGauge("order_queue_depth", "Current depth of the order processing queue")
	requestLatency    = RegisterHistogram(NewHistogram(
		"http_request_duration_seconds",
		"HTTP request latency distribution",
		[]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
	))
)

// ─── INSTRUMENTED HANDLER ──────────────────────────────────────────────────

// instrumentedHandler wraps a handler to record metrics.
func instrumentedHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		activeConns.Inc()
		defer activeConns.Dec()

		httpRequestsTotal.Inc()
		next.ServeHTTP(w, r)

		latency := time.Since(start).Seconds()
		requestLatency.Observe(latency)
	})
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()

	// /healthz — standard Kubernetes liveness/readiness probe endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// /metrics — Prometheus scrape endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, defaultReg.Render())
	})

	// /orders — simulated business endpoint
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// Simulate varying latency (1ms–100ms)
		time.Sleep(time.Duration(rand.IntN(100)) * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"orders":[],"queue_depth":%d}`, orderQueueDepth.Value())
	})

	// Simulate background activity
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			// Simulate order queue fluctuating
			orderQueueDepth.Set(int64(rand.IntN(50)))
		}
	}()

	// Self-test: generate some traffic to populate metrics
	go func() {
		time.Sleep(100 * time.Millisecond)
		client := &http.Client{Timeout: 2 * time.Second}
		endpoints := []string{"/orders", "/orders", "/healthz", "/orders", "/healthz"}
		for _, ep := range endpoints {
			resp, err := client.Get("http://127.0.0.1:9090" + ep)
			if err == nil {
				resp.Body.Close()
			}
		}
		fmt.Println("Self-test requests sent. Check /metrics for results.")
		fmt.Println()
		fmt.Println("Metrics snapshot:")
		fmt.Printf("  requests_total : %d\n", httpRequestsTotal.Value())
		fmt.Printf("  active_conns   : %d\n", activeConns.Value())
		fmt.Printf("  latency p50    : %.3fs\n", requestLatency.Percentile(50))
		fmt.Printf("  latency p95    : %.3fs\n", requestLatency.Percentile(95))
		fmt.Printf("  latency p99    : %.3fs\n", requestLatency.Percentile(99))
	}()

	srv := &http.Server{
		Addr:    ":9090",
		Handler: instrumentedHandler(mux),
	}

	fmt.Println("Metrics server on :9090")
	fmt.Println("  curl http://localhost:9090/healthz")
	fmt.Println("  curl http://localhost:9090/metrics")
	fmt.Println("  curl http://localhost:9090/orders")
	fmt.Println()

	log.Fatal(srv.ListenAndServe())

	// KEY TAKEAWAYS:
	// 1. Counter    : monotonic, never decreases (requests_total, bytes_sent)
	// 2. Gauge      : current value, can go up/down (active_conns, queue_depth)
	// 3. Histogram  : latency distribution → enables p50/p95/p99 in dashboards
	// 4. /healthz   : Kubernetes liveness probe (200 = alive, 503 = not ready)
	// 5. /metrics   : Prometheus scrape endpoint — plain text, parseable by Grafana
}
