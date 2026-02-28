// Lesson 06 — TLS 1.3: self-signed certificate, tls.Listen, InsecureSkipVerify
// Run: go run phase2-protocols/lessons/lesson06_tls.go
//
// TLS in 30 seconds:
//   1. Client connects → Server sends certificate
//   2. Client verifies certificate against trusted CAs
//   3. Key exchange (ECDH in TLS 1.3) → session keys derived
//   4. All subsequent traffic is encrypted + authenticated (AEAD)
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"time"
)

// ─── CERTIFICATE GENERATION ────────────────────────────────────────────────

// generateSelfSignedCert creates an in-memory ECDSA self-signed certificate.
// Self-signed = the cert is signed by itself, not by a trusted CA.
// Use cases: development, internal services with custom CA, testing.
//
// Production alternatives:
//   - Let's Encrypt / ACME protocol (public services)
//   - cert-manager in Kubernetes
//   - Your organisation's internal PKI (HashiCorp Vault, AWS ACM)
func generateSelfSignedCert() (tls.Certificate, error) {
	// ECDSA P-256 is preferred over RSA for TLS:
	//   - Smaller key sizes (256-bit ECC ≈ 3072-bit RSA security)
	//   - Faster handshakes
	//   - Required by TLS 1.3 key exchange
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Netlab Demo Org"},
			CommonName:   "localhost",
		},
		// SAN (Subject Alternative Names) — browsers require these, not just CN
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost"},
		NotBefore:   time.Now().Add(-1 * time.Minute), // clock skew buffer
		NotAfter:    time.Now().Add(1 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// IsCA: false — this cert is for a server, not a CA
	}

	// Self-sign: parent == template, signingKey == private key
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

// ─── TLS SERVER ────────────────────────────────────────────────────────────

func tlsServer(addr string, cert tls.Certificate, ready chan<- struct{}) {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13, // enforce TLS 1.3 — reject older versions
		// TLS 1.0/1.1 are deprecated (RFC 8996). TLS 1.2 is still acceptable.
		// TLS 1.3 removes weak cipher suites entirely.
	}

	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		fmt.Println("[tls-server] error:", err)
		return
	}
	defer ln.Close()
	fmt.Println("[tls-server] listening on", addr, "(TLS 1.3 only)")
	close(ready)

	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	// Explicit Handshake() lets us inspect the TLS state before data flows
	tlsConn := conn.(*tls.Conn)
	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("[tls-server] handshake failed:", err)
		return
	}

	state := tlsConn.ConnectionState()
	fmt.Printf("[tls-server] TLS version : 0x%04x (%s)\n", state.Version, tlsVersionName(state.Version))
	fmt.Printf("[tls-server] cipher suite : %d\n", state.CipherSuite)
	fmt.Printf("[tls-server] server name  : %q\n", state.ServerName)

	// Echo mode
	io.Copy(conn, conn)
}

// ─── TLS CLIENT ────────────────────────────────────────────────────────────

func tlsClientInsecure(addr string) {
	fmt.Println("\n[tls-client] DEMO A: InsecureSkipVerify (BAD — dev only)")

	// InsecureSkipVerify: skips ALL certificate validation.
	// RISK: you're vulnerable to MITM — an attacker can intercept your traffic
	// by presenting any certificate. NEVER use in production.
	// WHY developers use it: lazy, or self-signed cert with no CA chain.
	// CORRECT solution: add the self-signed cert to a custom CertPool.
	config := &tls.Config{
		InsecureSkipVerify: true, // SECURITY WARNING: MITM-vulnerable
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", addr, config)
	if err != nil {
		fmt.Println("[tls-client] error:", err)
		return
	}
	defer conn.Close()

	state := conn.ConnectionState()
	fmt.Printf("[tls-client] connected, version=0x%04x (%s)\n",
		state.Version, tlsVersionName(state.Version))
	if len(state.PeerCertificates) > 0 {
		c := state.PeerCertificates[0]
		fmt.Printf("[tls-client] server cert CN=%s, expires=%s\n",
			c.Subject.CommonName, c.NotAfter.Format("2006-01-02 15:04"))
	}

	fmt.Fprintf(conn, "HELLO over TLS 1.3\n")
	buf := make([]byte, 64)
	n, _ := conn.Read(buf)
	fmt.Printf("[tls-client] echo: %q\n", buf[:n])
}

func tlsClientSecure(addr string, serverCert tls.Certificate) {
	fmt.Println("\n[tls-client] DEMO B: Proper CA pool (CORRECT)")

	// Extract the x509 cert from our tls.Certificate
	x509Cert, err := x509.ParseCertificate(serverCert.Certificate[0])
	if err != nil {
		fmt.Println("[tls-client] parse cert error:", err)
		return
	}

	// Build a custom CertPool that trusts only our self-signed cert
	pool := x509.NewCertPool()
	pool.AddCert(x509Cert)

	config := &tls.Config{
		RootCAs:    pool,
		ServerName: "localhost",
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", addr, config)
	if err != nil {
		fmt.Println("[tls-client] error:", err)
		return
	}
	defer conn.Close()
	fmt.Printf("[tls-client] connected with proper cert verification\n")
	conn.Close()
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "unknown"
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	cert, err := generateSelfSignedCert()
	if err != nil {
		fmt.Println("cert generation error:", err)
		return
	}
	fmt.Println("Self-signed ECDSA P-256 certificate generated (in memory)")

	addr := "127.0.0.1:9443"
	ready := make(chan struct{})
	go tlsServer(addr, cert, ready)
	<-ready

	tlsClientInsecure(addr)

	time.Sleep(100 * time.Millisecond)

	// Restart server for second demo
	ready2 := make(chan struct{})
	go tlsServer(addr, cert, ready2)
	<-ready2
	tlsClientSecure(addr, cert)

	time.Sleep(50 * time.Millisecond)
	fmt.Println("\nDone.")

	// KEY TAKEAWAYS:
	// 1. Always enforce MinVersion: tls.VersionTLS12 (minimum) or TLS13
	// 2. InsecureSkipVerify = dev only; always pinpoint the risk in a comment
	// 3. Custom CertPool is the correct way to trust self-signed certs
	// 4. tls.Conn.Handshake() lets you inspect state before transferring data
	// 5. ECDSA P-256 > RSA 2048 for modern TLS
}
