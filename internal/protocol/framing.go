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

// readVarIntN reads a VarInt from data and reports its encoded length.
func readVarIntN(data []byte) (int32, int, error) {
	var result uint32
	for i := 0; i < 5 && i < len(data); i++ {
		result |= uint32(data[i]&0x7f) << (7 * i)
		if data[i]&0x80 == 0 {
			return int32(result), i + 1, nil
		}
	}
	if len(data) >= 5 {
		return 0, 0, ErrVarIntTooBig
	}
	return 0, 0, io.ErrUnexpectedEOF
}

// fieldReader walks the fields of a packet payload left to right.
type fieldReader struct {
	data []byte
	off  int
}

func newFieldReader(data []byte) *fieldReader { return &fieldReader{data: data} }

func (r *fieldReader) varInt() (int32, error) {
	v, n, err := readVarIntN(r.data[r.off:])
	if err != nil {
		return 0, err
	}
	r.off += n
	return v, nil
}

// fixed returns the next n bytes.
func (r *fieldReader) fixed(n int) ([]byte, error) {
	if n < 0 || r.off+n > len(r.data) {
		return nil, io.ErrUnexpectedEOF
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b, nil
}

func (r *fieldReader) str() (string, error) {
	v, n, err := readVarIntN(r.data[r.off:])
	if err != nil {
		return "", err
	}
	if v < 0 || r.off+n+int(v) > len(r.data) {
		return "", fmt.Errorf("protocol: string length %d exceeds packet data", v)
	}
	out := string(r.data[r.off+n : r.off+n+int(v)])
	r.off += n + int(v)
	return out, nil
}

// byteSlice reads a length-prefixed array of bytes.
func (r *fieldReader) byteSlice() ([]byte, error) {
	v, n, err := readVarIntN(r.data[r.off:])
	if err != nil {
		return nil, err
	}
	if v < 0 || r.off+n+int(v) > len(r.data) {
		return nil, fmt.Errorf("protocol: byte array length %d exceeds packet data", v)
	}
	out := append([]byte(nil), r.data[r.off+n:r.off+n+int(v)]...)
	r.off += n + int(v)
	return out, nil
}

func (r *fieldReader) boolean() (bool, error) {
	b, err := r.fixed(1)
	if err != nil {
		return false, err
	}
	return b[0] != 0, nil
}

func (r *fieldReader) uuid() ([16]byte, error) {
	b, err := r.fixed(16)
	var u [16]byte
	copy(u[:], b)
	return u, err
}
