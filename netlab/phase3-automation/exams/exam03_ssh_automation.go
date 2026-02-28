// EXAM 03 — SSH Automation (contient 4 bugs intentionnels)
// Run: go run phase3-automation/exams/exam03_ssh_automation.go
//
// Lis exam03_instructions.txt avant de commencer.
// Necessite: go mod tidy (golang.org/x/crypto)
//
// Un serveur SSH de test (mockSSHServer) est integre — pas besoin d'infra externe.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

// ─── SSH CLIENT (BUGGY) ────────────────────────────────────────────────────

// connectToHost establishes an SSH connection to host:port.
func connectToHost(host, port, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// BUG #2: InsecureIgnoreHostKey used with no security comment.
		// In production this is a critical MITM vulnerability.
		// Fix: add a // SECURITY WARNING comment, or use ssh.FixedHostKey / knownhosts.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// BUG #3: Timeout field is missing.
		// If the remote host is unreachable, ssh.Dial blocks forever.
		// Fix: add  Timeout: 10 * time.Second
	}

	addr := net.JoinHostPort(host, port)
	return ssh.Dial("tcp", addr, config)
}

// runCommand executes a command on the SSH server and returns its output.
func runCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	// BUG #1: session.Close() is never called.
	// Each session consumes server-side resources. After enough commands,
	// the SSH server refuses new sessions with "too many open sessions".
	// Fix: add  defer session.Close()  here.

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("run %q: %w", cmd, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// streamOutput runs a command and "streams" the output via a buffer.
func streamOutput(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	var buf bytes.Buffer
	// BUG #4: session.Stdout is never assigned to &buf.
	// session.Run() executes the command, but the output goes nowhere.
	// buf stays empty, and the function returns "" even on success.
	// Fix: add  session.Stdout = &buf  before session.Run(cmd).
	session.Stderr = &buf // stderr is captured, but stdout is NOT

	if err := session.Run(cmd); err != nil {
		return buf.String(), fmt.Errorf("run: %w", err)
	}
	return buf.String(), nil
}

// ─── AUTOMATION SCRIPT ─────────────────────────────────────────────────────

func runAutomation(host, port string) {
	client, err := connectToHost(host, port, "admin", "password123")
	if err != nil {
		fmt.Println("Connect error:", err)
		return
	}
	defer client.Close()

	commands := []string{
		"show version",
		"show interfaces",
		"show ip route",
	}

	fmt.Println("=== Batch Command Execution ===")
	for _, cmd := range commands {
		out, err := runCommand(client, cmd)
		if err != nil {
			fmt.Printf("  [%s] ERROR: %v\n", cmd, err)
			continue
		}
		fmt.Printf("  [%s] output=%q\n", cmd, out)
	}

	fmt.Println("\n=== Stream Output Demo ===")
	out, err := streamOutput(client, "show run")
	if err != nil {
		fmt.Println("  streamOutput error:", err)
	} else {
		fmt.Printf("  [show run] captured %d bytes: %q\n", len(out), out)
		fmt.Println("  (should NOT be empty after BUG #4 fix)")
	}
}

// ─── MOCK SSH SERVER (test fixture) ────────────────────────────────────────

// startMockSSHServer starts a minimal SSH server that responds to commands.
// This lets the exam run without a real SSH server.
func startMockSSHServer(addr string) (string, error) {
	// Generate an ephemeral host key
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		return "", err
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) == "password123" {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid password")
		},
	}
	config.AddHostKey(signer)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleSSHConn(conn, config)
		}
	}()

	return ln.Addr().String(), nil
}

func handleSSHConn(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()
	sconn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sconn.Close()
	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		ch, requests, err := newChan.Accept()
		if err != nil {
			continue
		}
		go handleSSHSession(ch, requests)
	}
}

func handleSSHSession(ch ssh.Channel, requests <-chan *ssh.Request) {
	defer ch.Close()
	for req := range requests {
		if req.Type == "exec" {
			// Parse command from payload: [4-byte length][command bytes]
			if len(req.Payload) < 4 {
				if req.WantReply {
					req.Reply(false, nil)
				}
				continue
			}
			cmdLen := int(req.Payload[0])<<24 | int(req.Payload[1])<<16 |
				int(req.Payload[2])<<8 | int(req.Payload[3])
			if len(req.Payload) < 4+cmdLen {
				if req.WantReply {
					req.Reply(false, nil)
				}
				continue
			}
			cmd := string(req.Payload[4 : 4+cmdLen])

			if req.WantReply {
				req.Reply(true, nil)
			}

			// Send response to stdout
			response := fmt.Sprintf("MOCK OUTPUT for: %s\n", cmd)
			io.WriteString(ch, response)

			// Send exit status 0
			ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
			return
		}
		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	// Start the embedded mock SSH server
	serverAddr, err := startMockSSHServer("127.0.0.1:0")
	if err != nil {
		fmt.Println("Failed to start mock SSH server:", err)
		return
	}
	fmt.Printf("Mock SSH server running at %s\n\n", serverAddr)

	// Parse host:port
	host, port, _ := net.SplitHostPort(serverAddr)
	runAutomation(host, port)
}
