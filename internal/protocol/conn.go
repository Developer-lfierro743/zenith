package protocol

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net"
)

// Conn is a Minecraft protocol connection: a stream plus the optional
// compression and encryption layers the login sequence switches on.
// The Conn owns its buffered reader and writer, so packet framing never
// over-reads into the next packet's bytes — the same bufio.Reader serves
// every read for the life of the connection.
type Conn struct {
	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer

	compressionThreshold int32 // negative = compression disabled
}

// NewConn wraps c in a plain (uncompressed, unencrypted) protocol connection.
func NewConn(c net.Conn) *Conn {
	return &Conn{
		conn:                 c,
		br:                   bufio.NewReader(c),
		bw:                   bufio.NewWriter(c),
		compressionThreshold: -1,
	}
}

// EnableCompression switches the connection to the post-compression packet
// format with the given threshold. Packets whose (ID+payload) length is at
// or above the threshold are zlib-compressed; smaller ones are sent raw with
// an uncompressed length of 0. A threshold of 0 compresses everything.
func (c *Conn) EnableCompression(threshold int32) {
	c.compressionThreshold = threshold
}

// CompressionThreshold reports the current threshold, or -1 if compression
// is disabled.
func (c *Conn) CompressionThreshold() int32 { return c.compressionThreshold }

// EnableEncryption switches the stream to AES-128/CFB8 with the given key
// (the 16-byte shared secret; the IV is the same bytes), applied to every
// wire byte from this point on — including length prefixes.
//
// Call this only after the last plaintext packet has been read and any
// buffered plaintext has been consumed: the vanilla flow enables encryption
// right after the Encryption Request/Response exchange, with nothing else
// in flight.
func (c *Conn) EnableEncryption(key []byte) error {
	if len(key) != 16 {
		return fmt.Errorf("protocol: encryption key must be 16 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	dec := &cfb8{b: block, decrypt: true}
	copy(dec.iv[:], key)
	enc := &cfb8{b: block, decrypt: false}
	copy(enc.iv[:], key)
	c.br = bufio.NewReader(&cipher.StreamReader{S: dec, R: c.conn})
	c.bw = bufio.NewWriter(&cipher.StreamWriter{S: enc, W: c.conn})
	return nil
}

// ReadPacket reads one packet, honoring the compression layer if enabled.
func (c *Conn) ReadPacket() (Packet, error) {
	outerLen, err := ReadVarInt(c.br)
	if err != nil {
		return Packet{}, err
	}
	if outerLen < 0 || outerLen > MaxPacketSize {
		return Packet{}, ErrPacketTooBig
	}
	buf := make([]byte, outerLen)
	if _, err := io.ReadFull(c.br, buf); err != nil {
		return Packet{}, err
	}
	var payload []byte
	if c.compressionThreshold < 0 {
		payload = buf
	} else {
		payload, err = decompress(buf)
		if err != nil {
			return Packet{}, err
		}
	}
	fields := bytes.NewReader(payload)
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

func decompress(buf []byte) ([]byte, error) {
	rd := bytes.NewReader(buf)
	dataLen, err := ReadVarInt(rd)
	if err != nil {
		return nil, err
	}
	body := buf[len(buf)-rd.Len():]
	if dataLen == 0 {
		return body, nil // below threshold: raw packet, uncompressed
	}
	if dataLen < 0 || dataLen > MaxPacketSize {
		return nil, ErrPacketTooBig
	}
	zr, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	payload := make([]byte, dataLen)
	if _, err := io.ReadFull(zr, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// WritePacket writes one packet, honoring the compression layer if enabled.
func (c *Conn) WritePacket(p Packet) error {
	body := AppendVarInt(nil, uint32(p.ID))
	body = append(body, p.Data...)

	if c.compressionThreshold < 0 {
		buf := AppendVarInt(nil, uint32(len(body)))
		buf = append(buf, body...)
		if _, err := c.bw.Write(buf); err != nil {
			return err
		}
		return c.bw.Flush()
	}

	var dataLen int32
	var payload []byte
	if c.compressionThreshold == 0 || int32(len(body)) >= c.compressionThreshold {
		var zb bytes.Buffer
		zw := zlib.NewWriter(&zb)
		if _, err := zw.Write(body); err != nil {
			return err
		}
		if err := zw.Close(); err != nil {
			return err
		}
		payload, dataLen = zb.Bytes(), int32(len(body))
	} else {
		payload, dataLen = body, 0
	}

	inner := AppendVarInt(nil, uint32(dataLen))
	inner = append(inner, payload...)
	out := AppendVarInt(nil, uint32(len(inner)))
	out = append(out, inner...)
	if _, err := c.bw.Write(out); err != nil {
		return err
	}
	return c.bw.Flush()
}
