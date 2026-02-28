// Lesson 10 — Network Inventory: walk CIDR, concurrent probe, JSON output
// Run: go run phase3-automation/lessons/lesson10_inventory.go
//
// This simulates a network discovery tool — used in:
//   - Infrastructure as Code (Terraform, Ansible dynamic inventory)
//   - Cisco DevNet network automation
//   - SRE on-call tooling (which hosts are up?)
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

// ─── HOST INFO ──────────────────────────────────────────────────────────────

type HostInfo struct {
	IP         string   `json:"ip"`
	Alive      bool     `json:"alive"`
	OpenPorts  []int    `json:"open_ports,omitempty"`
	LatencyMs  float64  `json:"latency_ms,omitempty"`
	Hostname   string   `json:"hostname,omitempty"`
	ProbeError string   `json:"error,omitempty"`
}

// ─── PROBE ─────────────────────────────────────────────────────────────────

// probePorts checks which well-known ports are open on a host.
// We use TCP connect — no ICMP ping (requires raw sockets / root privileges).
func probePorts(ip string, ports []int, timeout time.Duration) HostInfo {
	info := HostInfo{IP: ip}

	for _, port := range ports {
		addr := fmt.Sprintf("%s:%d", ip, port)
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err == nil {
			conn.Close()
			info.Alive = true
			info.OpenPorts = append(info.OpenPorts, port)
			if info.LatencyMs == 0 {
				ms := float64(time.Since(start).Microseconds()) / 1000.0
				info.LatencyMs = roundFloat(ms, 2)
			}
		}
	}

	// Reverse DNS — optional, may time out on large scans
	if info.Alive {
		names, err := net.LookupAddr(ip)
		if err == nil && len(names) > 0 {
			info.Hostname = names[0]
		}
	}

	return info
}

// ─── CIDR EXPANSION ────────────────────────────────────────────────────────

// expandCIDR returns all usable host IPs in a network (excluding network address and broadcast).
func expandCIDR(cidr string) ([]string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	ones, bits := network.Mask.Size()
	if bits == 128 {
		return nil, fmt.Errorf("IPv6 CIDR expansion not supported in this demo")
	}

	// /31 and /32 are special cases (RFC 3021)
	hostBits := bits - ones
	if hostBits == 0 {
		return []string{network.IP.String()}, nil // /32 = single host
	}
	if hostBits == 1 {
		a := cloneIP(network.IP)
		b := cloneIP(network.IP)
		b[len(b)-1]++
		return []string{a.String(), b.String()}, nil // /31 = 2 hosts
	}

	var ips []string
	for ip := cloneIP(network.IP); network.Contains(ip); incrementIP(ip) {
		ipStr := ip.String()
		// Skip network address (first) and broadcast (last)
		if ipStr == network.IP.String() {
			continue
		}
		if isBroadcast(ip, network) {
			continue
		}
		ips = append(ips, ipStr)
	}
	return ips, nil
}

func cloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func isBroadcast(ip net.IP, n *net.IPNet) bool {
	broadcast := make(net.IP, len(n.IP))
	for i := range broadcast {
		broadcast[i] = n.IP[i] | ^n.Mask[i]
	}
	return ip.Equal(broadcast)
}

// ─── INVENTORY SCAN ────────────────────────────────────────────────────────

func scanInventory(cidr string, probePorts_ []int, maxWorkers int, timeout time.Duration) []HostInfo {
	ips, err := expandCIDR(cidr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "expandCIDR error: %v\n", err)
		return nil
	}

	fmt.Printf("  Probing %d hosts in %s...\n", len(ips), cidr)

	ipCh := make(chan string, len(ips))
	for _, ip := range ips {
		ipCh <- ip
	}
	close(ipCh)

	workers := maxWorkers
	if len(ips) < workers {
		workers = len(ips)
	}

	resultCh := make(chan HostInfo, len(ips))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipCh {
				resultCh <- probePorts(ip, probePorts_, timeout)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []HostInfo
	for info := range resultCh {
		results = append(results, info)
	}

	// Sort by IP for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return ipToUint32(results[i].IP) < ipToUint32(results[j].IP)
	})
	return results
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	// Small CIDR for a quick demo (change to your subnet in real use)
	// 127.0.0.0/29 = 127.0.0.1 through 127.0.0.6
	cidr := "127.0.0.0/29"
	ports := []int{22, 80, 443, 3306, 5432, 6379, 8080}
	timeout := 100 * time.Millisecond

	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  Network Inventory Scan           ║\n")
	fmt.Printf("╚══════════════════════════════════╝\n")
	fmt.Printf("CIDR     : %s\n", cidr)
	fmt.Printf("Ports    : %v\n", ports)
	fmt.Printf("Timeout  : %v per host\n", timeout)
	fmt.Printf("Workers  : 10\n\n")

	start := time.Now()
	inventory := scanInventory(cidr, ports, 10, timeout)
	elapsed := time.Since(start)

	// ── JSON output ───────────────────────────────────────────────────
	fmt.Println("=== JSON Inventory ===")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(inventory)

	// ── Summary ────────────────────────────────────────────────────────
	alive := 0
	for _, h := range inventory {
		if h.Alive {
			alive++
		}
	}
	fmt.Printf("\nScan complete in %v\n", elapsed)
	fmt.Printf("Hosts alive  : %d / %d\n", alive, len(inventory))

	// KEY TAKEAWAYS:
	// 1. CIDR expansion: iterate from network IP, skip network addr & broadcast
	// 2. TCP connect instead of ICMP ping: no root/admin privileges needed
	// 3. Worker pool: bounded concurrency prevents connection exhaustion
	// 4. JSON output: standard format for Ansible dynamic inventory, Terraform, CI
	// 5. Sort results for idempotent output (important for diffs in automation)
}

// ─── HELPERS ───────────────────────────────────────────────────────────────

func roundFloat(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}

func ipToUint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
