// Package protocol implements Minecraft's network protocol for the pinned
// version: framing, VarInt/VarLong, compression, and the handshake, status,
// login, configuration and play states. First goal (M1): join a real
// vanilla/Fabric server as a client.
package protocol

// AppendVarInt appends v in Minecraft's VarInt encoding to dst and returns
// the extended slice. VarInts are LEB128: 7 bits per byte, little-endian,
// with the high bit marking continuation; at most 5 bytes.
func AppendVarInt(dst []byte, v uint32) []byte {
	for {
		if v < 0x80 {
			return append(dst, byte(v))
		}
		dst = append(dst, byte(v)|0x80)
		v >>= 7
	}
}

// VarIntLen returns the encoded length of v.
func VarIntLen(v uint32) int {
	n := 1
	for v >= 0x80 {
		v >>= 7
		n++
	}
	return n
}
