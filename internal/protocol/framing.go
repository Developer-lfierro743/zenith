package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// MaxPacketSize is the largest packet Zenith will read uncompressed:
// vanilla's default maximum is 2 MiB, plus a little headroom.
const MaxPacketSize = 2<<20 + 64

// ErrPacketTooBig is returned when a packet exceeds MaxPacketSize.
var ErrPacketTooBig = errors.New("protocol: packet exceeds MaxPacketSize")

// Packet is one length-prefixed protocol packet in uncompressed framing:
// VarInt length, then VarInt packet ID, then the ID's payload.
type Packet struct {
	ID   int32
	Data []byte
}

// ReadPacket reads one packet from r: a VarInt length followed by a VarInt
// packet ID and exactly that many payload bytes. It never reads past the
// packet boundary, so consecutive packets on one stream decode correctly.
func ReadPacket(r io.Reader) (Packet, error) {
	length, err := ReadVarInt(r)
	if err != nil {
		return Packet{}, err
	}
	if length < 0 || int(length) > MaxPacketSize {
		return Packet{}, ErrPacketTooBig
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Packet{}, err
	}
	fields := bytes.NewReader(buf)
	id, err := ReadVarInt(fields)
	if err != nil {
		return Packet{}, err
	}
	data, err := io.ReadAll(fields)
	if err != nil {
		return Packet{}, err
	}
	return Packet{ID: id, Data: data}, nil
}

// WritePacket writes one packet: VarInt length of (ID+payload), VarInt ID,
// then the payload.
func WritePacket(w io.Writer, id int32, data []byte) error {
	total := VarIntLen(uint32(id)) + len(data)
	buf := AppendVarInt(nil, uint32(total))
	buf = AppendVarInt(buf, uint32(id))
	buf = append(buf, data...)
	_, err := w.Write(buf)
	return err
}

// AppendString appends s as a Minecraft String: a VarInt UTF-8 byte length
// followed by the UTF-8 bytes.
func AppendString(dst []byte, s string) []byte {
	dst = AppendVarInt(dst, uint32(len(s)))
	return append(dst, s...)
}

// ReadString reads a Minecraft String from data and returns it with the
// number of bytes consumed.
func ReadString(data []byte) (string, int, error) {
	br := bytes.NewReader(data)
	n, err := ReadVarInt(br)
	if err != nil {
		return "", 0, err
	}
	if n < 0 || n > int32(br.Len()) {
		return "", 0, fmt.Errorf("protocol: string length %d exceeds packet data", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(br, buf); err != nil {
		return "", 0, err
	}
	return string(buf), len(data) - br.Len(), nil
}
