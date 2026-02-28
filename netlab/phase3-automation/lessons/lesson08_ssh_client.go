// Lesson 08 — SSH Client: golang.org/x/crypto/ssh
// Run: SSH_USER=admin SSH_PASS=secret go run phase3-automation/lessons/lesson08_ssh_client.go
//
// SETUP: go mod tidy   (downloads golang.org/x/crypto)
//
// To test locally:
//   Linux/Mac: openssh-server must be running (sudo systemctl start ssh)
//   Windows  : Enable "OpenSSH Server" in Optional Features
//   Docker   : docker run -d -p 2222:22 linuxserver/openssh-server
//
// Set SSH_HOST / SSH_PORT env vars to override defaults.
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ─── CONNECTION ────────────────────────────────────────────────────────────

type SSHConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Timeout  time.Duration
}

// connect establishes an authenticated SSH connection.
//
// HOST KEY SECURITY:
//   InsecureIgnoreHostKey() skips host verification — vulnerable to MITM.
//   Production options:
//     a) ssh.FixedHostKey(pubKey)          — pin a known host public key
//     b) knownhosts.New("~/.ssh/known_hosts") — use system known_hosts file
func connect(cfg SSHConfig) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
			// For key-based auth:
			// ssh.PublicKeys(signer),  where signer comes from ssh.ParsePrivateKey
		},
		// SECURITY WARNING: InsecureIgnoreHostKey() is dev/demo only.
		// In production: use knownhosts.New or ssh.FixedHostKey.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.Timeout,
	}

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	return ssh.Dial("tcp", addr, config)
}

// ─── COMMAND EXECUTION ─────────────────────────────────────────────────────

// runCommand executes a command and returns combined stdout+stderr.
// Each ssh.Session is single-use: one command per session.
func runCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close() // IMPORTANT: always close to release server resources

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("run %q: %w", cmd, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// streamCommand executes a command and streams output to w in real time.
// Useful for long-running commands: log tailing, packet captures, etc.
func streamCommand(client *ssh.Client, cmd string, w io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	session.Stdout = w
	session.Stderr = w

	return session.Run(cmd)
}

// ─── BATCH EXECUTION ───────────────────────────────────────────────────────

type CommandResult struct {
	Command string
	Output  string
	Error   error
}

// runBatch runs all commands sequentially and collects results.
// Each command gets its own session (ssh.Session is single-use).
func runBatch(client *ssh.Client, commands []string) []CommandResult {
	results := make([]CommandResult, 0, len(commands))
	for _, cmd := range commands {
		out, err := runCommand(client, cmd)
		results = append(results, CommandResult{Command: cmd, Output: out, Error: err})
	}
	return results
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	cfg := SSHConfig{
		Host:     getEnv("SSH_HOST", "127.0.0.1"),
		Port:     getEnv("SSH_PORT", "22"),
		User:     os.Getenv("SSH_USER"),
		Password: os.Getenv("SSH_PASS"),
		Timeout:  5 * time.Second,
	}

	if cfg.User == "" {
		printUsage()
		return
	}

	fmt.Printf("Connecting to %s@%s:%s...\n", cfg.User, cfg.Host, cfg.Port)

	client, err := connect(cfg)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		fmt.Println("Make sure SSH server is running and credentials are correct.")
		return
	}
	defer client.Close()

	fmt.Println("Connected!\n")

	// ── Batch: system info ────────────────────────────────────────────
	fmt.Println("=== System Information ===")
	sysCommands := []string{
		"hostname",
		"uname -a",
		"uptime",
		"nproc",
		"free -h",
	}
	for _, r := range runBatch(client, sysCommands) {
		if r.Error != nil {
			fmt.Printf("  $ %-20s  ERROR: %v\n", r.Command, r.Error)
		} else {
			fmt.Printf("  $ %-20s  → %s\n", r.Command, firstLine(r.Output))
		}
	}

	// ── Batch: network state ──────────────────────────────────────────
	fmt.Println("\n=== Network State ===")
	netCommands := []string{
		"ip addr show | grep 'inet ' | head -5",
		"ip route | head -5",
		"ss -tlnp | head -10",
	}
	for _, r := range runBatch(client, netCommands) {
		if r.Error != nil {
			fmt.Printf("  ERROR: %v\n", r.Error)
		} else {
			fmt.Printf("  $ %s\n%s\n", r.Command, indent(r.Output, "    "))
		}
	}

	// ── Streaming ─────────────────────────────────────────────────────
	fmt.Println("\n=== Streaming: df -h ===")
	if err := streamCommand(client, "df -h", os.Stdout); err != nil {
		fmt.Println("Stream error:", err)
	}

	// KEY TAKEAWAYS:
	// 1. ssh.ClientConfig.HostKeyCallback — NEVER InsecureIgnoreHostKey in prod
	// 2. Each session is single-use: one command, then session.Close()
	// 3. defer session.Close() prevents SSH session leaks on the server
	// 4. CombinedOutput() for simple commands; session.Stdout = w for streaming
	// 5. Use net.JoinHostPort for IPv6-safe host:port concatenation
}

func printUsage() {
	fmt.Println("SSH_USER and SSH_PASS environment variables required.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  SSH_USER=admin SSH_PASS=secret go run phase3-automation/lessons/lesson08_ssh_client.go")
	fmt.Println("  SSH_HOST=10.0.0.1 SSH_PORT=22 SSH_USER=admin SSH_PASS=secret go run ...")
	fmt.Println()
	fmt.Println("--- Code walkthrough (no server needed to read) ---")
	fmt.Println("1. ssh.ClientConfig{User, Auth, HostKeyCallback, Timeout}")
	fmt.Println("2. ssh.Dial(\"tcp\", host:port, config)  →  *ssh.Client")
	fmt.Println("3. client.NewSession()                 →  *ssh.Session  (single use)")
	fmt.Println("4. session.CombinedOutput(cmd)         →  stdout+stderr as []byte")
	fmt.Println("5. session.Stdout = w + session.Run()  →  streaming output")
	fmt.Println("6. defer session.Close()               →  mandatory resource cleanup")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return s[:idx]
	}
	return s
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
