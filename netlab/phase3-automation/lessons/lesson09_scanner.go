// Lesson 09 — Concurrent Port Scanner: worker pool, semaphore, channel fan-in
// Run: go run phase3-automation/lessons/lesson09_scanner.go
//
// This lesson shows the CORRECT way to do concurrent port scanning —
// the patterns that fix the bugs in exam01.
package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// ─── RESULT TYPE ───────────────────────────────────────────────────────────

type ScanResult struct {
	Port    int
	Open    bool
	Latency time.Duration
	Service string
}

// ─── SCAN FUNCTION ─────────────────────────────────────────────────────────

// scanPort probes a single TCP port.
// Always uses DialTimeout — net.Dial with no timeout blocks forever on
// firewalled hosts ("stealth" mode drops packets without responding).
func scanPort(host string, port int, timeout time.Duration) ScanResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return ScanResult{Port: port, Open: false}
	}
	latency := time.Since(start)
	conn.Close()
	return ScanResult{
		Port:    port,
		Open:    true,
		Latency: latency,
		Service: wellKnownService(port),
	}
}

// ─── PATTERN 1: SEMAPHORE (chan struct{}) ──────────────────────────────────

// scanWithSemaphore uses a buffered channel as a semaphore to cap concurrency.
//
// Semaphore pattern:
//   sem := make(chan struct{}, N)  // N = max concurrent goroutines
//   sem <- struct{}{}              // acquire (blocks when N goroutines active)
//   defer func() { <-sem }()      // release
//
// This is lightweight — no sync.Pool, no worker pool boilerplate.
func scanWithSemaphore(host string, ports []int, maxConcurrent int, timeout time.Duration) []ScanResult {
	sem := make(chan struct{}, maxConcurrent)
	results := make(chan ScanResult, len(ports))

	var wg sync.WaitGroup
	for _, p := range ports {
		wg.Add(1)
		port := p
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // acquire slot
			defer func() { <-sem }() // release slot
			results <- scanPort(host, port, timeout)
		}()
	}

	// Close results channel only after all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []ScanResult
	for r := range results {
		all = append(all, r)
	}
	return all
}

// ─── PATTERN 2: WORKER POOL ────────────────────────────────────────────────

// scanWithWorkerPool spawns exactly N worker goroutines that drain a ports channel.
// More structured than semaphore — good when workers need initialization/teardown.
func scanWithWorkerPool(host string, ports []int, workers int, timeout time.Duration) []ScanResult {
	portCh := make(chan int, len(ports))
	for _, p := range ports {
		portCh <- p
	}
	close(portCh)

	results := make(chan ScanResult, len(ports))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portCh { // range over closed channel drains it
				results <- scanPort(host, port, timeout)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []ScanResult
	for r := range results {
		all = append(all, r)
	}
	return all
}

// ─── HELPER ────────────────────────────────────────────────────────────────

func portRange(start, end int) []int {
	ports := make([]int, end-start+1)
	for i := range ports {
		ports[i] = start + i
	}
	return ports
}

func sortResults(results []ScanResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})
}

func wellKnownService(port int) string {
	m := map[int]string{
		21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp",
		53: "dns", 80: "http", 110: "pop3", 143: "imap",
		443: "https", 465: "smtps", 587: "submission",
		3306: "mysql", 5432: "postgres", 6379: "redis",
		8080: "http-alt", 8443: "https-alt", 9200: "elasticsearch",
	}
	if s, ok := m[port]; ok {
		return s
	}
	return "unknown"
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	host := "127.0.0.1"
	ports := portRange(1, 1024)
	timeout := 150 * time.Millisecond

	// ── Pattern 1: Semaphore ──────────────────────────────────────────
	fmt.Println("=== Semaphore pattern (max 200 concurrent) ===")
	t0 := time.Now()
	results1 := scanWithSemaphore(host, ports, 200, timeout)
	elapsed1 := time.Since(t0)
	sortResults(results1)
	printSummary(host, results1, elapsed1)

	// ── Pattern 2: Worker pool ─────────────────────────────────────────
	fmt.Println("\n=== Worker pool pattern (100 workers) ===")
	t0 = time.Now()
	results2 := scanWithWorkerPool(host, ports, 100, timeout)
	elapsed2 := time.Since(t0)
	sortResults(results2)
	printSummary(host, results2, elapsed2)

	// KEY TAKEAWAYS:
	// 1. net.DialTimeout: always use; never net.Dial for network code
	// 2. Semaphore (chan struct{}): simplest concurrency limit — 2 lines
	// 3. Worker pool: N goroutines draining a channel — predictable resource usage
	// 4. Channel fan-in: goroutines → chan ScanResult → collector
	// 5. close(resultsCh) ONLY after wg.Wait() in a separate goroutine
}

func printSummary(host string, results []ScanResult, elapsed time.Duration) {
	var open []ScanResult
	for _, r := range results {
		if r.Open {
			open = append(open, r)
		}
	}
	fmt.Printf("  Scanned %d ports on %s in %v\n", len(results), host, elapsed)
	fmt.Printf("  Open: %d\n", len(open))
	for _, r := range open {
		fmt.Printf("    %5d/tcp  %-20s  latency=%v\n", r.Port, r.Service, r.Latency.Round(time.Microsecond))
	}
}
