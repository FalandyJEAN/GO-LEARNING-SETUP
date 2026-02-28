// Lesson 05 — DNS, CIDR, IP math, network interfaces
// Run: go run phase1-fundamentals/lessons/lesson05_dns_ip.go
//
// Concepts used daily in networking code:
//   - DNS lookups (forward + reverse)
//   - CIDR notation and subnet math
//   - IP containment checks
//   - Enumerating local network interfaces
package main

import (
	"fmt"
	"net"
)

// ─── DNS RESOLUTION ────────────────────────────────────────────────────────

func dnsDemo() {
	fmt.Println("╔══════════════════════════════╗")
	fmt.Println("║  DNS Resolution               ║")
	fmt.Println("╚══════════════════════════════╝")

	// Forward lookup: hostname → IPs
	hosts := []string{"google.com", "cloudflare.com", "1.1.1.1"}
	for _, host := range hosts {
		addrs, err := net.LookupHost(host)
		if err != nil {
			fmt.Printf("  LookupHost(%q): ERROR %v\n", host, err)
			continue
		}
		fmt.Printf("  LookupHost(%-18q) = %v\n", host, addrs)
	}

	// Reverse lookup: IP → hostnames
	fmt.Println()
	for _, ip := range []string{"1.1.1.1", "8.8.8.8"} {
		names, err := net.LookupAddr(ip)
		if err != nil {
			fmt.Printf("  LookupAddr(%s): ERROR %v\n", ip, err)
			continue
		}
		fmt.Printf("  LookupAddr(%-15s) = %v\n", ip, names)
	}

	// MX records — useful for email infrastructure
	fmt.Println()
	mxs, err := net.LookupMX("gmail.com")
	if err == nil {
		fmt.Printf("  MX gmail.com: %s (pref %d)\n", mxs[0].Host, mxs[0].Pref)
	}
}

// ─── CIDR MATH ─────────────────────────────────────────────────────────────

func cidrDemo() {
	fmt.Println("\n╔══════════════════════════════╗")
	fmt.Println("║  CIDR / Subnet Math           ║")
	fmt.Println("╚══════════════════════════════╝")

	networks := []string{
		"10.0.0.0/24",  // class-A private, 254 hosts
		"192.168.1.0/30", // point-to-point link, 2 hosts
		"172.16.0.0/12",  // class-B private range
		"2001:db8::/32",  // IPv6 documentation prefix
	}

	for _, cidr := range networks {
		ip, network, err := net.ParseCIDR(cidr)
		if err != nil {
			fmt.Println("  ParseCIDR error:", err)
			continue
		}
		ones, bits := network.Mask.Size()
		usable := (1 << (bits - ones)) - 2
		if bits == 128 { // IPv6
			fmt.Printf("  %-20s  network=%-20s  prefix=/%d (IPv6)\n",
				ip, network.IP, ones)
		} else {
			fmt.Printf("  %-20s  network=%-20s  mask=%-16s  hosts=%d\n",
				ip, network.IP, net.IP(network.Mask), usable)
		}
	}

	// Subnet containment — used in firewall rules, routing decisions
	fmt.Println()
	_, internal, _ := net.ParseCIDR("10.0.0.0/8")
	probeIPs := []string{"10.1.2.3", "10.255.0.1", "192.168.0.1", "172.217.0.0"}
	for _, ipStr := range probeIPs {
		ip := net.ParseIP(ipStr)
		fmt.Printf("  %-16s  in 10.0.0.0/8 ? %v\n", ipStr, internal.Contains(ip))
	}
}

// ─── IP ARITHMETIC ─────────────────────────────────────────────────────────

func ipMathDemo() {
	fmt.Println("\n╔══════════════════════════════╗")
	fmt.Println("║  IP Arithmetic                ║")
	fmt.Println("╚══════════════════════════════╝")

	// Increment an IP — useful for CIDR scanning / DHCP range allocation
	ip := net.ParseIP("192.168.1.0").To4()
	fmt.Printf("  Start: %s\n", ip)
	for i := 0; i < 5; i++ {
		incrementIP(ip)
		fmt.Printf("  +%d    : %s\n", i+1, ip)
	}

	// Parse a raw IPv4 address from 4 bytes (common in binary protocol parsing)
	raw := []byte{172, 16, 42, 10}
	parsed := net.IP(raw)
	fmt.Printf("\n  Raw bytes %v → IP: %s\n", raw, parsed)

	// IP version check
	for _, s := range []string{"192.168.1.1", "::1", "2001:db8::1"} {
		ip := net.ParseIP(s)
		version := "IPv4"
		if ip.To4() == nil {
			version = "IPv6"
		}
		fmt.Printf("  %-20s  → %s\n", s, version)
	}
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

// ─── NETWORK INTERFACES ────────────────────────────────────────────────────

func interfaceDemo() {
	fmt.Println("\n╔══════════════════════════════╗")
	fmt.Println("║  Local Network Interfaces     ║")
	fmt.Println("╚══════════════════════════════╝")

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("  Interfaces() error:", err)
		return
	}

	for _, iface := range ifaces {
		fmt.Printf("  %-15s  MTU=%-6d  flags=%s\n",
			iface.Name, iface.MTU, iface.Flags)

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			fmt.Printf("    addr: %s\n", addr)
		}

		if iface.HardwareAddr != nil {
			fmt.Printf("    mac:  %s\n", iface.HardwareAddr)
		}
	}
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	dnsDemo()
	cidrDemo()
	ipMathDemo()
	interfaceDemo()

	// KEY TAKEAWAYS:
	// 1. net.LookupHost    → forward DNS (name → IPs)
	// 2. net.LookupAddr    → reverse DNS (IP → names)
	// 3. net.ParseCIDR     → parses CIDR, returns host IP + network
	// 4. network.Contains  → checks if IP is within a subnet
	// 5. net.Interfaces()  → lists NICs, their addresses, flags, MACs
	// 6. net.IP.To4()      → nil means IPv6; non-nil means IPv4
}
