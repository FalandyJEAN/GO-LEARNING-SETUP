// EXAM 01 — Port Scanner (contient 4 bugs intentionnels)
// Run: go run phase1-fundamentals/exams/exam01_port_scanner.go
// Race: go run -race phase1-fundamentals/exams/exam01_port_scanner.go
//
// Lis exam01_instructions.txt avant de commencer.
package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

type ScanResult struct {
	Port int
	Open bool
}

// scanPort checks whether a single TCP port is open.
func scanPort(host string, port int) ScanResult {
	addr := fmt.Sprintf("%s:%d", host, port)

	// BUG #1: net.Dial has NO timeout.
	// On a firewall that silently drops packets, this blocks forever.
	// Fix: replace with net.DialTimeout(...)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return ScanResult{Port: port, Open: false}
	}
	conn.Close()
	return ScanResult{Port: port, Open: true}
}

// scanRange scans a range of TCP ports concurrently.
func scanRange(host string, start, end int) []ScanResult {
	// BUG #3 (shared slice + BUG #2 goroutines): results is written from
	// multiple goroutines without synchronization → race condition.
	var results []ScanResult
	var mu sync.Mutex // mu is declared but only used in the fix

	// BUG #4: close(resultsCh) will be called BEFORE goroutines finish,
	// causing a panic when they try to send. Combine with WaitGroup fix.
	resultsCh := make(chan ScanResult, end-start+1)

	var wg sync.WaitGroup

	for port := start; port <= end; port++ {
		wg.Add(1)
		p := port

		// BUG #2: one goroutine per port — no limit.
		// Scanning ports 1-65535 = 65535 simultaneous goroutines → OOM.
		// Fix: use a semaphore: sem := make(chan struct{}, 200)
		go func() {
			defer wg.Done()
			r := scanPort(host, p)

			// BUG #3: append to shared slice from multiple goroutines.
			// The race detector will catch this immediately.
			// Fix: send to resultsCh channel instead, then collect below.
			mu.Lock()             // mu is here only to suppress compile error; logic still racy
			results = append(results, r) // BUG: should use resultsCh <- r
			mu.Unlock()
		}()
	}

	// BUG #4: close(resultsCh) is called here, BEFORE wg.Wait().
	// Any goroutine that hasn't finished yet will panic on send.
	// Fix: wrap in  go func() { wg.Wait(); close(resultsCh) }()
	close(resultsCh)
	wg.Wait()

	// Drain channel (currently empty because goroutines wrote to results slice)
	for r := range resultsCh {
		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})
	return results
}

func main() {
	host := "127.0.0.1"
	startPort := 1
	endPort := 1024

	fmt.Printf("Scanning %s ports %d-%d...\n", host, startPort, endPort)
	t0 := time.Now()
	results := scanRange(host, startPort, endPort)
	fmt.Printf("Done in %v\n\n", time.Since(t0))

	var open []ScanResult
	for _, r := range results {
		if r.Open {
			open = append(open, r)
		}
	}

	fmt.Printf("Open ports: %d\n", len(open))
	for _, r := range open {
		fmt.Printf("  %d/tcp  open\n", r.Port)
	}
}
