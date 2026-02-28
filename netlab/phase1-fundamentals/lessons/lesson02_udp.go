// Lesson 02 — UDP: PacketConn, datagrams, syslog sender
// Run: go run phase1-fundamentals/lessons/lesson02_udp.go
//
// UDP vs TCP:
//   TCP  = reliable, ordered, connection-oriented  (HTTP, SSH, FTP)
//   UDP  = unreliable, unordered, connectionless   (DNS, syslog, video streaming, QUIC)
package main

import (
	"fmt"
	"net"
	"time"
)

// ─── UDP SERVER ────────────────────────────────────────────────────────────

// udpServer listens on addr and prints incoming datagrams.
// net.ListenPacket returns a net.PacketConn — no concept of "connection".
// Each Read gives you the sender's address and the datagram payload.
func udpServer(addr string, ready chan<- struct{}, count int) {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("[udp-server] error:", err)
		return
	}
	defer pc.Close()
	fmt.Println("[udp-server] listening on", addr)
	close(ready)

	buf := make([]byte, 2048)
	for i := 0; i < count; i++ {
		n, from, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		fmt.Printf("[udp-server] from %-20s : %s\n", from, buf[:n])
	}
}

// ─── SYSLOG SENDER ─────────────────────────────────────────────────────────

// sendSyslog sends a simplified RFC 5424 syslog message over UDP.
// Real syslog:  <priority>VERSION TIMESTAMP HOST APP PID MSGID - MSG
// Simplified:   <priority>MSG  (enough for demos and log collectors)
//
// Priority = Facility * 8 + Severity
//   Facility 16 = local0  → 16*8 = 128
//   Severity  6 = info    → 128 + 6 = 134
func sendSyslog(addr, appName, message string) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Println("[syslog] error:", err)
		return
	}
	defer conn.Close()

	payload := fmt.Sprintf("<134>%s: %s", appName, message)
	if _, err := fmt.Fprint(conn, payload); err != nil {
		fmt.Println("[syslog] write error:", err)
		return
	}
	fmt.Printf("[syslog] sent: %s\n", payload)
}

// ─── UDP ECHO (client side) ─────────────────────────────────────────────────

// udpEcho shows how to use WriteTo/ReadFrom directly (connectionless style).
func udpEcho(serverAddr string) {
	// net.Dial("udp", ...) works too — it sets a default remote address
	// so you can use Write/Read instead of WriteTo/ReadFrom.
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	msg := []byte("PING")
	conn.SetDeadline(time.Now().Add(200 * time.Millisecond))
	conn.Write(msg)

	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	if n > 0 {
		fmt.Printf("[udp-echo] got: %q\n", buf[:n])
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	addr := "127.0.0.1:9514" // 514 is standard syslog, 9514 for non-root demo

	// Network events that a router/switch might log
	events := []struct{ app, msg string }{
		{"router01", "Interface GigabitEthernet0/0 changed state to up"},
		{"switch01", "MAC address table full on VLAN 100"},
		{"router01", "BGP session established with peer 10.0.0.2 AS 65001"},
		{"firewall", "DENY TCP 192.168.1.5:54321 -> 8.8.8.8:53"},
		{"router01", "OSPF neighbor 192.168.0.1 Full adjacency established"},
	}

	ready := make(chan struct{})
	go udpServer(addr, ready, len(events))
	<-ready

	time.Sleep(10 * time.Millisecond)
	for _, e := range events {
		sendSyslog(addr, e.app, e.msg)
		time.Sleep(5 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	// KEY TAKEAWAYS:
	// 1. net.ListenPacket("udp", addr) → connectionless socket
	// 2. pc.ReadFrom(buf)              → returns (n, remoteAddr, err) — no persistent conn
	// 3. net.Dial("udp", addr)         → convenience: WriteTo/ReadFrom with default peer
	// 4. UDP datagrams may be lost, reordered, or duplicated — design accordingly
	// 5. Syslog (RFC 5424) uses UDP/514 or TCP/601 for reliable delivery
	fmt.Println("\nDone.")
}
