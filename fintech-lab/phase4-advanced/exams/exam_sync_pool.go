// exam_sync_pool.go — EXAMEN 4-A : "sync.Pool : Eliminer la Pression sur le GC"
// Simulateur de parsing de messages FIX Protocol.
// Ce code fonctionne correctement mais genere une pression GC excessive.
//
// Lancer    : go run phase4-advanced/exams/exam_sync_pool.go
// Benchmark : go test ./phase4-advanced/exams/ -bench=BenchmarkParseFIX -benchmem

package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// MarketDataMsg — Message de donnees de marche (format FIX simplifie)
// ---------------------------------------------------------------------------

// MarketDataMsg represente un message deparse.
// Alloue dans le heap a chaque message => pression GC.
type MarketDataMsg struct {
	Symbol    string
	BidPrice  float64
	AskPrice  float64
	BidSize   int64
	AskSize   int64
	Timestamp int64
	raw       []byte // BUG #4 : buffer realloue a chaque parse
}

// BUG #1 : Alloue TOUJOURS un nouveau struct.
// Avec 500K messages/sec, le GC doit collecter 500K objets/sec.
// Correction : utiliser sync.Pool pour recycler les MarketDataMsg.
func NewMarketDataMsg() *MarketDataMsg {
	return &MarketDataMsg{
		raw: make([]byte, 0, 256), // BUG #4 : nouvel alloc a chaque fois
	}
}

// Reset remet le message a zero pour reutilisation.
// BUG #2 : Cette methode existe mais n'est PAS appelee avant Put() dans le pool.
// Un message remis au pool sans reset exposera les donnees precedentes
// au prochain utilisateur.
func (m *MarketDataMsg) Reset() {
	m.Symbol = ""
	m.BidPrice = 0
	m.AskPrice = 0
	m.BidSize = 0
	m.AskSize = 0
	m.Timestamp = 0
	m.raw = m.raw[:0] // Garder la capacite, vider le contenu
}

// ---------------------------------------------------------------------------
// FIXParser — Parse les messages FIX Protocol
// ---------------------------------------------------------------------------

// Format FIX simplifie :
// "55=AAPL|132=189.45|133=189.55|134=100|135=200|52=1706000000"
// Tag 55  = Symbol
// Tag 132 = BidPrice
// Tag 133 = AskPrice
// Tag 134 = BidSize
// Tag 135 = AskSize
// Tag 52  = Timestamp

// FIXParser parse les messages FIX.
// BUG #1 : cree un nouveau MarketDataMsg a chaque parse (pas de pool).
type FIXParser struct{}

// Parse parse un message FIX et retourne un *MarketDataMsg.
// BUG #1 : Pas de sync.Pool. Allocation systematique.
// BUG #2 : Si on ajoutait un pool, il faudrait Reset() avant Put().
func (p *FIXParser) Parse(fixMsg string) (*MarketDataMsg, error) {
	msg := NewMarketDataMsg() // BUG #1 : alloue toujours

	fields := strings.Split(fixMsg, "|")
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}
		tag, value := parts[0], parts[1]

		switch tag {
		case "55":
			msg.Symbol = value
		case "132":
			price, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("bid price invalide: %s", value)
			}
			msg.BidPrice = price
		case "133":
			price, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("ask price invalide: %s", value)
			}
			msg.AskPrice = price
		case "134":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("bid size invalide: %s", value)
			}
			msg.BidSize = size
		case "135":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("ask size invalide: %s", value)
			}
			msg.AskSize = size
		case "52":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("timestamp invalide: %s", value)
			}
			msg.Timestamp = ts
		}
	}

	return msg, nil
}

// Release remet le message dans le pool apres utilisation.
// BUG #2 : Appele ici mais Reset() est oublie — les donnees sont corrompues.
// BUG #3 : Un pool n'est PAS un stockage persistant. Le GC peut le vider
//          a tout moment. Ne jamais stocker des donnees critiques dans un pool.
func (p *FIXParser) Release(msg *MarketDataMsg) {
	// BUG #2 : On ne reset PAS avant de remettre dans le pool.
	// Le prochain utilisateur verra les anciennes donnees.
	// msg.Reset() // <-- MANQUANT
	_ = msg // Simule un "put dans un pool inexistant pour cet exemple"
}

// ---------------------------------------------------------------------------
// Simulation de flux
// ---------------------------------------------------------------------------

// generateFIXMessages genere des messages FIX synthetiques.
func generateFIXMessages(n int) []string {
	symbols := []string{"AAPL", "MSFT", "TSLA", "GOOGL", "AMZN"}
	msgs := make([]string, n)
	for i := range msgs {
		sym := symbols[i%len(symbols)]
		bid := 189.50 + float64(i%100)*0.01
		ask := bid + 0.10
		msgs[i] = fmt.Sprintf("55=%s|132=%.2f|133=%.2f|134=100|135=200|52=%d",
			sym, bid, ask, time.Now().UnixNano()+int64(i))
	}
	return msgs
}

// processMessages traite un lot de messages FIX.
// C'est le hot path — chaque allocation ici a un cout.
func processMessages(messages []string) int {
	parser := &FIXParser{}
	processed := 0

	for _, rawMsg := range messages {
		msg, err := parser.Parse(rawMsg) // BUG #1 : alloc a chaque iteration
		if err != nil {
			continue
		}

		// Simuler le traitement du message
		_ = msg.BidPrice + msg.AskPrice // mid price
		_ = (msg.AskPrice - msg.BidPrice) // spread
		processed++

		// BUG #2 : Release appele mais sans Reset()
		parser.Release(msg)
	}

	return processed
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	fmt.Println("=== FIX Parser — Analyse des allocations ===")
	fmt.Println()

	// Statistiques GC avant
	var gcBefore runtime.MemStats
	runtime.ReadMemStats(&gcBefore)

	messages := generateFIXMessages(100_000)

	start := time.Now()
	processed := processMessages(messages)
	elapsed := time.Since(start)

	// Forcer un GC pour voir l'impact
	runtime.GC()

	// Statistiques GC apres
	var gcAfter runtime.MemStats
	runtime.ReadMemStats(&gcAfter)

	fmt.Printf("Messages traites  : %d\n", processed)
	fmt.Printf("Temps total       : %v\n", elapsed)
	fmt.Printf("Throughput        : %.0f msgs/sec\n", float64(processed)/elapsed.Seconds())
	fmt.Println()
	fmt.Printf("Allocations heap  : %d (total depuis le start)\n", gcAfter.TotalAlloc-gcBefore.TotalAlloc)
	fmt.Printf("GC cycles         : %d\n", gcAfter.NumGC-gcBefore.NumGC)
	fmt.Printf("GC pause total    : %v\n", time.Duration(gcAfter.PauseTotalNs-gcBefore.PauseTotalNs))
	fmt.Println()
	fmt.Println("Objectif apres correction :")
	fmt.Println("  - GC cycles : 0 ou minimal")
	fmt.Println("  - GC pause  : 0 ns")
	fmt.Println("  - Allocs    : minimales (idealement 0 dans le hot path)")
	fmt.Println()
	fmt.Println("Lance : go test -bench=BenchmarkParseFIX -benchmem")
	fmt.Println("Objectif : 0 allocs/op dans le benchmark")
}
