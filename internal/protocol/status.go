package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// Packet IDs in the Status state.
const (
	PacketStatusRequest  int32 = 0x00
	PacketStatusResponse int32 = 0x00
	PacketPingRequest    int32 = 0x01
	PacketPongResponse   int32 = 0x01
)

// StatusResponse is the JSON body of the Status Response packet (the
// Server List Ping response).
type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int32  `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int32 `json:"max"`
		Online int32 `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description json.RawMessage `json:"description"`
	Favicon     string          `json:"favicon"`
}

// MotdText extracts plain text from Description, which is a chat component
// (either a bare string or a {"text", "extra": [...]} object).
func (s StatusResponse) MotdText() string {
	var raw any
	if err := json.Unmarshal(s.Description, &raw); err != nil {
		return string(s.Description)
	}
	switch v := raw.(type) {
	case string:
		return v
	case map[string]any:
		var sb strings.Builder
		walkComponent(&sb, v)
		return sb.String()
	default:
		return string(s.Description)
	}
}

func walkComponent(sb *strings.Builder, m map[string]any) {
	if t, ok := m["text"].(string); ok {
		sb.WriteString(t)
	}
	if extra, ok := m["extra"].([]any); ok {
		for _, e := range extra {
			if em, ok := e.(map[string]any); ok {
				walkComponent(sb, em)
			}
		}
	}
}

// PingStatus performs the full Server List Ping flow over conn:
// Handshake → Status Request → Status Response → Ping Request → Pong
// Response. It returns the parsed response and the ping round-trip time
// (Ping Request to Pong Response). The whole flow must finish within
// timeout.
func PingStatus(conn net.Conn, hs Handshake, timeout time.Duration) (StatusResponse, time.Duration, error) {
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return StatusResponse{}, 0, err
	}
	if err := WritePacket(conn, 0x00, hs.Encode()); err != nil {
		return StatusResponse{}, 0, fmt.Errorf("write handshake: %w", err)
	}
	if err := WritePacket(conn, PacketStatusRequest, nil); err != nil {
		return StatusResponse{}, 0, fmt.Errorf("write status request: %w", err)
	}
	respPacket, err := ReadPacket(conn)
	if err != nil {
		return StatusResponse{}, 0, fmt.Errorf("read status response: %w", err)
	}
	if respPacket.ID != PacketStatusResponse {
		return StatusResponse{}, 0, fmt.Errorf("protocol: expected status response (0x00), got 0x%02x", respPacket.ID)
	}
	body, _, err := ReadString(respPacket.Data)
	if err != nil {
		return StatusResponse{}, 0, fmt.Errorf("read status JSON string: %w", err)
	}
	var resp StatusResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return StatusResponse{}, 0, fmt.Errorf("decode status JSON: %w", err)
	}
	payload := time.Now().UnixNano()
	var pl [8]byte
	binary.BigEndian.PutUint64(pl[:], uint64(payload))
	if err := WritePacket(conn, PacketPingRequest, pl[:]); err != nil {
		return StatusResponse{}, 0, fmt.Errorf("write ping: %w", err)
	}
	sent := time.Now()
	pong, err := ReadPacket(conn)
	if err != nil {
		return StatusResponse{}, 0, fmt.Errorf("read pong: %w", err)
	}
	if pong.ID != PacketPongResponse || len(pong.Data) < 8 {
		return StatusResponse{}, 0, fmt.Errorf("protocol: expected pong (0x01) with 8-byte payload, got 0x%02x len=%d", pong.ID, len(pong.Data))
	}
	return resp, time.Since(sent), nil
}
