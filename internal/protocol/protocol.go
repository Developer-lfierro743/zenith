// Package protocol implements Minecraft's network protocol for the pinned
// version: framing, VarInt/VarLong, compression, and the handshake, status,
// login, configuration and play states. First goal (M1): join a real
// vanilla/Fabric server as a client.
package protocol

import (
	"errors"
	"io"
)

// ErrVarIntTooBig is returned when a VarInt exceeds 5 bytes.
var ErrVarIntTooBig = errors.New("protocol: varint is too big")

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

// ReadVarInt reads a Minecraft VarInt from r and returns it as int32
// (VarInts are signed 32-bit in Minecraft). It fails with ErrVarIntTooBig
// after 5 continuation bytes. It reads exactly one byte at a time, so it
// never consumes past the end of the VarInt — safe on connection streams.
func ReadVarInt(r io.Reader) (int32, error) {
	var result uint32
	var buf [1]byte
	for i := 0; i < 5; i++ {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		result |= uint32(buf[0]&0x7f) << (7 * i)
		if buf[0]&0x80 == 0 {
			return int32(result), nil
		}
	}
	return 0, ErrVarIntTooBig
}
