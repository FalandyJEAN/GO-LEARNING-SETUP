// Lesson 12 — Rate Limiter: token bucket + circuit breaker (closed/open/half-open)
// Run: go run phase4-infrastructure/lessons/lesson12_rate_limiter.go
//
// Used in: API gateways, trading risk controls, DDoS protection, service meshes.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// TOKEN BUCKET
// ═══════════════════════════════════════════════════════════════════════════
//
// Concept:
//   - Bucket holds up to `capacity` tokens
//   - Tokens refill at `rate` per second
//   - Each request consumes 1 token
//   - If bucket empty → request rejected (or queued)
//
// vs Leaky Bucket:
//   Token bucket  → allows controlled bursts (up to capacity)
//   Leaky bucket  → constant output rate, no bursts allowed
//
// Used by: AWS API Gateway, Nginx limit_req, Redis INCR + EXPIRE pattern

type TokenBucket struct {
	mu        sync.Mutex
	capacity  float64
	tokens    float64
	rate      float64 // tokens added per second
	lastCheck time.Time
}

func NewTokenBucket(capacity, ratePerSec float64) *TokenBucket {
	return &TokenBucket{
		capacity:  capacity,
		tokens:    capacity, // start full
		rate:      ratePerSec,
		lastCheck: time.Now(),
	}
}

// Allow returns true if a token is available (and consumes it).
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastCheck).Seconds()
	tb.lastCheck = now

	// Refill tokens based on elapsed time
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1.0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// ═══════════════════════════════════════════════════════════════════════════
// CIRCUIT BREAKER
// ═══════════════════════════════════════════════════════════════════════════
//
// State machine:
//
//   CLOSED ──(failures >= threshold)──► OPEN
//     ▲                                   │
//     │                          (timeout expires)
//     │                                   ▼
//   (successes >= threshold) ◄── HALF-OPEN
//
// CLOSED   : Normal operation. All requests pass through.
// OPEN     : Service is down. All requests fail fast (no network calls).
// HALF-OPEN: Recovery probe. Limited requests pass; if they succeed → CLOSED,
//            if they fail → back to OPEN.
//
// Used by: Netflix Hystrix, Resilience4j, Istio, service mesh sidecars.

type cbState int32

const (
	cbClosed   cbState = 0 // normal
	cbOpen     cbState = 1 // failing fast
	cbHalfOpen cbState = 2 // probing recovery
)

func (s cbState) String() string {
	switch s {
	case cbClosed:
		return "CLOSED"
	case cbOpen:
		return "OPEN"
	case cbHalfOpen:
		return "HALF-OPEN"
	}
	return "UNKNOWN"
}

type CircuitBreaker struct {
	st           atomic.Int32
	failures     atomic.Int32
	successes    atomic.Int32
	mu           sync.Mutex
	openedAt     time.Time // when the breaker last opened
	maxFail      int32
	successThresh int32
	resetTimeout time.Duration
}

func NewCircuitBreaker(maxFail, successThresh int32, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFail:      maxFail,
		successThresh: successThresh,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) State() cbState { return cbState(cb.st.Load()) }

// Allow returns true if the request should proceed.
func (cb *CircuitBreaker) Allow() bool {
	switch cb.State() {
	case cbClosed:
		return true

	case cbOpen:
		cb.mu.Lock()
		if time.Since(cb.openedAt) > cb.resetTimeout {
			cb.st.Store(int32(cbHalfOpen))
			cb.successes.Store(0)
			cb.mu.Unlock()
			fmt.Printf("  [CB] OPEN → HALF-OPEN (probing recovery)\n")
			return true
		}
		cb.mu.Unlock()
		return false

	case cbHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful downstream call.
func (cb *CircuitBreaker) RecordSuccess() {
	if cb.State() == cbHalfOpen {
		n := cb.successes.Add(1)
		if n >= cb.successThresh {
			cb.st.Store(int32(cbClosed))
			cb.failures.Store(0)
			fmt.Printf("  [CB] HALF-OPEN → CLOSED (service recovered after %d successes)\n", n)
		}
	} else {
		cb.failures.Store(0)
	}
}

// RecordFailure records a failed downstream call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	cb.openedAt = time.Now()
	cb.mu.Unlock()

	n := cb.failures.Add(1)
	if cb.State() != cbOpen && n >= cb.maxFail {
		cb.st.Store(int32(cbOpen))
		fmt.Printf("  [CB] CLOSED → OPEN (failures=%d, threshold=%d)\n", n, cb.maxFail)
	}
	if cb.State() == cbHalfOpen {
		cb.st.Store(int32(cbOpen))
		cb.mu.Lock()
		cb.openedAt = time.Now()
		cb.mu.Unlock()
		fmt.Printf("  [CB] HALF-OPEN → OPEN (probe failed)\n")
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// MAIN
// ═══════════════════════════════════════════════════════════════════════════

func main() {
	// ── Token Bucket Demo ─────────────────────────────────────────────
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║  Token Bucket (capacity=5, rate=2/s)   ║")
	fmt.Println("╚════════════════════════════════════════╝")

	tb := NewTokenBucket(5, 2)

	// Burst: drain the bucket
	for i := 1; i <= 7; i++ {
		allowed := tb.Allow()
		fmt.Printf("  request %d: allowed=%-5v  tokens_left=%.1f\n",
			i, allowed, tb.Tokens())
	}

	fmt.Println("  (sleeping 1 second — bucket refills at 2 tokens/sec)")
	time.Sleep(time.Second)
	fmt.Printf("  after 1s: allowed=%v  tokens=%.1f\n", tb.Allow(), tb.Tokens())

	// ── Circuit Breaker Demo ──────────────────────────────────────────
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║  Circuit Breaker (maxFail=3, threshold=2, to=500ms)  ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	cb := NewCircuitBreaker(3, 2, 500*time.Millisecond)

	// Phase 1: Failures trip the breaker
	fmt.Println("\nPhase 1: Simulating failures")
	for i := 1; i <= 5; i++ {
		if cb.Allow() {
			cb.RecordFailure()
			fmt.Printf("  call %d: ALLOWED → failed → state=%s\n", i, cb.State())
		} else {
			fmt.Printf("  call %d: BLOCKED (fast fail) → state=%s\n", i, cb.State())
		}
	}

	// Phase 2: Breaker is OPEN — wait for timeout
	fmt.Printf("\nPhase 2: Waiting %v for reset timeout...\n", 600*time.Millisecond)
	time.Sleep(600 * time.Millisecond)

	// Phase 3: HALF-OPEN — first request probes; successes close the breaker
	fmt.Println("\nPhase 3: Recovery probes (HALF-OPEN)")
	for i := 1; i <= 4; i++ {
		if cb.Allow() {
			cb.RecordSuccess()
			fmt.Printf("  probe %d: ALLOWED → success → state=%s\n", i, cb.State())
		} else {
			fmt.Printf("  probe %d: BLOCKED → state=%s\n", i, cb.State())
		}
	}

	// KEY TAKEAWAYS:
	// 1. Token bucket: allows bursts up to capacity, then enforces rate
	// 2. Leaky bucket: no bursts, constant drain rate — use for strict SLA
	// 3. Circuit breaker: fail-fast pattern protects upstream services
	// 4. Three states: CLOSED (normal) → OPEN (fail-fast) → HALF-OPEN (probe)
	// 5. Use atomic.Int32 for state to avoid locking on the hot path
}
