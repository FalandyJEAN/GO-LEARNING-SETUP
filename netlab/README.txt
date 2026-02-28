NETLAB — Go Network Programming Lab
====================================

Module  : netlab
Go      : 1.22
Dep     : golang.org/x/crypto v0.28.0 (SSH uniquement)

STRUCTURE
---------
netlab/
  START_HERE.txt                   <- Commence ici
  README.txt                       <- Ce fichier
  go.mod

  phase1-fundamentals/
    lessons/
      lesson01_tcp.go              net.Listen, Dial, goroutine/conn, echo
      lesson02_udp.go              PacketConn, datagrams, syslog sender
      lesson03_http_client.go      Client+timeout, context, GET/POST JSON, retry
      lesson04_http_server.go      ServeMux, JSON, middleware logging+auth
      lesson05_dns_ip.go           LookupHost, CIDR, IP math, net.Interfaces
    exams/
      exam01_instructions.txt
      exam01_port_scanner.go       [4 bugs : timeout, goroutine limit, race, channel]

  phase2-protocols/
    lessons/
      lesson06_tls.go              Self-signed cert, tls.Listen, InsecureSkipVerify
      lesson07_binary_protocol.go  encoding/binary, frame [len:4|type:1|payload:N]
    exams/
      exam02_instructions.txt
      exam02_http_proxy.go         [4 bugs : body, hop-by-hop, err, WriteHeader]

  phase3-automation/
    lessons/
      lesson08_ssh_client.go       x/crypto/ssh, session, exec, stream
      lesson09_scanner.go          Worker pool, semaphore, port scan
      lesson10_inventory.go        CIDR walk, concurrent probe, JSON output
    exams/
      exam03_instructions.txt
      exam03_ssh_automation.go     [4 bugs : session leak, host key, timeout, stdout]

  phase4-infrastructure/
    lessons/
      lesson11_load_balancer.go    Round-robin atomic, ReverseProxy, health check
      lesson12_rate_limiter.go     Token bucket, circuit breaker 3 states
      lesson13_metrics.go          Counter/Gauge/Histogram, /healthz, /metrics
    exams/
      exam04_instructions.txt
      exam04_mini_lb.go            [4 bugs : atomic, nil panic, leak, body lu 2x]
    interview_questions.txt        30 questions OSI/TCP/TLS/BGP/k8s/eBPF
    notes_networking.txt           Aide-memoire protocoles reseau

CIBLES D'ENTRETIEN
------------------
Cisco DevNet Associate/Professional
Cloudflare Network Engineer
HashiCorp Infrastructure Engineer
Google/Meta SRE
Arista Networks Software Engineer
