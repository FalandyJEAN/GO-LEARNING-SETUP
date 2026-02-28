// Lesson 07 — Binary Protocol: encoding/binary, custom frame format
// Run: go run phase2-protocols/lessons/lesson07_binary_protocol.go
//
// Binary protocols are used when performance or bandwidth matters:
//   - FIX/FAST (trading), ITCH/OUCH (Nasdaq), BFD, BGP, OSPF, DNS wire format
//   - gRPC (Protobuf), Cap'n Proto, FlatBuffers
//
// Frame format designed in this lesson:
//   ┌──────────────┬──────────┬─────────────────┐
//   │ Length (4B)  │ Type (1B)│ Payload (N bytes)│
//   │ big-endian   │          │                  │
//   └──────────────┴──────────┴──────────────────┘
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// ─── MESSAGE TYPES ─────────────────────────────────────────────────────────

const (
	TypeHeartbeat uint8 = 0x01
	TypeOrderNew  uint8 = 0x02
	TypeOrderFill uint8 = 0x03
	TypeOrderCancel uint8 = 0x04
)

func typeName(t uint8) string {
	switch t {
	case TypeHeartbeat:
		return "HEARTBEAT"
	case TypeOrderNew:
		return "ORDER_NEW"
	case TypeOrderFill:
		return "ORDER_FILL"
	case TypeOrderCancel:
		return "ORDER_CANCEL"
	default:
		return fmt.Sprintf("UNKNOWN(0x%02x)", t)
	}
}

// ─── GENERIC FRAME ─────────────────────────────────────────────────────────

type Frame struct {
	Type    uint8
	Payload []byte
}

// EncodeFrame serializes a Frame to []byte.
// Always big-endian for network byte order (network standard since RFC 1700).
func EncodeFrame(f Frame) []byte {
	length := uint32(len(f.Payload))
	buf := make([]byte, 4+1+length)
	binary.BigEndian.PutUint32(buf[0:4], length) // length field: 4 bytes
	buf[4] = f.Type                              // type field:   1 byte
	copy(buf[5:], f.Payload)                     // payload
	return buf
}

// DecodeFrame reads exactly one frame from r.
// io.ReadFull ensures we get all bytes even if the read is fragmented (TCP).
func DecodeFrame(r io.Reader) (Frame, error) {
	// Read header: 4 bytes length + 1 byte type
	var header [5]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return Frame{}, fmt.Errorf("read header: %w", err)
	}
	length := binary.BigEndian.Uint32(header[0:4])
	msgType := header[4]

	// Guard against malicious or corrupt length field
	if length > 1<<20 { // 1 MB max payload
		return Frame{}, fmt.Errorf("payload too large: %d bytes", length)
	}

	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return Frame{}, fmt.Errorf("read payload: %w", err)
		}
	}
	return Frame{Type: msgType, Payload: payload}, nil
}

// ─── ORDER MESSAGE (fixed-width struct for cache efficiency) ───────────────

// OrderMsg uses fixed-size fields so the entire struct can be encoded
// with a single binary.Write — no dynamic allocation.
// This pattern is used in Nasdaq ITCH, CME MDP, and FIX binary protocols.
type OrderMsg struct {
	Symbol   [8]byte  // ASCII, null-padded
	OrderID  uint64   // 8 bytes
	Side     uint8    // 0=BUY 1=SELL
	Quantity uint32   // 4 bytes
	Price    uint64   // price in basis points (e.g. 15025 = $150.25)
	_        [3]byte  // padding to align to 8-byte boundary
}

func newOrderMsg(symbol, side string, qty uint32, price float64) OrderMsg {
	var msg OrderMsg
	copy(msg.Symbol[:], symbol)
	msg.OrderID = nextOrderID()
	if side == "SELL" {
		msg.Side = 1
	}
	msg.Quantity = qty
	msg.Price = uint64(math.Round(price * 100)) // store as cents
	return msg
}

var orderIDCounter uint64

func nextOrderID() uint64 {
	orderIDCounter++
	return orderIDCounter
}

func encodeOrder(msg OrderMsg) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeOrder(data []byte) (OrderMsg, error) {
	var msg OrderMsg
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (o OrderMsg) String() string {
	sym := string(bytes.TrimRight(o.Symbol[:], "\x00"))
	side := "BUY"
	if o.Side == 1 {
		side = "SELL"
	}
	return fmt.Sprintf("OrderID=%d %s %s qty=%d price=%.2f",
		o.OrderID, side, sym, o.Quantity, float64(o.Price)/100)
}

// ─── MAIN ──────────────────────────────────────────────────────────────────

func main() {
	// ── Demo 1: Encode/decode a stream of frames ──────────────────────
	fmt.Println("=== Frame Encode / Decode ===")

	frames := []Frame{
		{Type: TypeHeartbeat, Payload: []byte("seq=1")},
		{Type: TypeOrderCancel, Payload: []byte("ORD-0042")},
	}

	var stream bytes.Buffer
	for _, f := range frames {
		encoded := EncodeFrame(f)
		stream.Write(encoded)
		fmt.Printf("  Encoded [type=%-12s len=%d]: %x\n",
			typeName(f.Type), len(f.Payload), encoded)
	}

	fmt.Printf("\n  Total stream: %d bytes\n\n", stream.Len())

	reader := bytes.NewReader(stream.Bytes())
	for {
		f, err := DecodeFrame(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("  decode error:", err)
			break
		}
		fmt.Printf("  Decoded: type=%-12s payload=%q\n", typeName(f.Type), f.Payload)
	}

	// ── Demo 2: Fixed-width order struct ─────────────────────────────
	fmt.Println("\n=== Binary Order Encoding ===")

	orders := []OrderMsg{
		newOrderMsg("AAPL", "BUY", 1000, 150.25),
		newOrderMsg("GOOG", "SELL", 200, 2850.00),
		newOrderMsg("MSFT", "BUY", 500, 299.99),
	}

	for _, orig := range orders {
		data, err := encodeOrder(orig)
		if err != nil {
			fmt.Println("  encode error:", err)
			continue
		}
		decoded, err := decodeOrder(data)
		if err != nil {
			fmt.Println("  decode error:", err)
			continue
		}
		fmt.Printf("  %-40s  wire=%d bytes\n", decoded, len(data))
	}

	// ── Demo 3: Order wrapped in a Frame ─────────────────────────────
	fmt.Println("\n=== Order wrapped in Frame ===")

	orderData, _ := encodeOrder(orders[0])
	frame := Frame{Type: TypeOrderNew, Payload: orderData}
	wireBytes := EncodeFrame(frame)

	fmt.Printf("  Header  : %d bytes (4B len + 1B type)\n", 5)
	fmt.Printf("  Payload : %d bytes (fixed OrderMsg struct)\n", len(orderData))
	fmt.Printf("  Total   : %d bytes\n", len(wireBytes))
	fmt.Printf("  Header  : %x\n", wireBytes[:5])

	// KEY TAKEAWAYS:
	// 1. Always big-endian for network protocols (host byte order varies)
	// 2. io.ReadFull is critical — TCP may fragment your writes
	// 3. Fixed-size structs: cache-friendly, zero allocation on encode
	// 4. Price as integer (basis points/cents): avoids floating-point rounding
	// 5. Always validate length field before allocating payload buffer
}
