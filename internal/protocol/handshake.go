package protocol

// NextState values for the Handshake packet.
const (
	NextStatus int32 = 1
	NextLogin  int32 = 2
)

// Handshake is packet 0x00 in the Handshaking state, and the only packet
// there. It is always the first packet sent on a new connection.
type Handshake struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

// Encode returns the Handshake payload (the 0x00 packet ID is added by
// WritePacket).
func (h Handshake) Encode() []byte {
	var b []byte
	b = AppendVarInt(b, uint32(h.ProtocolVersion))
	b = AppendString(b, h.ServerAddress)
	b = append(b, byte(h.ServerPort>>8), byte(h.ServerPort)) // Unsigned Short, big-endian
	b = AppendVarInt(b, uint32(h.NextState))
	return b
}
