// Lesson 01 — TCP: net.Listen, net.Dial, goroutine-per-connection, echo server
// Run: go run phase1-fundamentals/lessons/lesson01_tcp.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// ─── SERVER ────────────────────────────────────────────────────────────────

// echoServer listens on addr and spawns one goroutine per accepted connection.
// This is the canonical Go pattern: net.Listen → Accept loop → go handleConn.
func echoServer(addr string, ready chan<- struct{}) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("[server] Listen error:", err)
		return
	}
	defer ln.Close()
	fmt.Println("[server] listening on", addr)
	close(ready) // signal: server is up

	for {
		conn, err := ln.Accept()
		if err != nil {
			return // listener was closed
		}
		go handleConn(conn) // one goroutine per connection — cheap in Go
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr()
	fmt.Println("[server] new connection from", remote)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(conn, "ECHO: %s\n", line)
	}
	fmt.Println("[server] connection closed:", remote)
}

// ─── CLIENT ────────────────────────────────────────────────────────────────

// echoClient dials addr, sends messages, and prints echoed responses.
// net.DialTimeout is always preferred over net.Dial in production
// to avoid blocking forever on unreachable hosts.
func echoClient(addr string, messages []string) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		fmt.Println("[client] connect error:", err)
		return
	}
	defer conn.Close()
	fmt.Println("[client] connected to", addr)

	reader := bufio.NewReader(conn)
	for _, msg := range messages {
		fmt.Fprintf(conn, "%s\n", msg)
		resp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("[client] read error:", err)
			break
		}
		fmt.Printf("[client] sent=%-20q  got=%q\n", msg, resp)
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	addr := "127.0.0.1:9001"
	ready := make(chan struct{})

	go echoServer(addr, ready)
	<-ready // wait until server is listening before dialing

	// Simulate an order gateway sending messages
	messages := []string{
		"NEW_ORDER BUY 100 AAPL@150.00",
		"NEW_ORDER SELL 50 GOOG@2800.00",
		"CANCEL_ORDER ORD-42",
	}
	echoClient(addr, messages)

	// KEY TAKEAWAYS:
	// 1. net.Listen("tcp", addr)   → creates a listening socket
	// 2. ln.Accept()               → blocks until a client connects
	// 3. go handleConn(conn)       → goroutines are cheap; one per conn is idiomatic
	// 4. net.DialTimeout           → always set a timeout; net.Dial has none
	// 5. defer conn.Close()        → critical to avoid file descriptor leaks
	fmt.Println("\nDone.")
}
