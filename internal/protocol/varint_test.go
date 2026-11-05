package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestVarInt(t *testing.T) {
	cases := map[uint32][]byte{
		0:          {0x00},
		1:          {0x01},
		127:        {0x7f},
		128:        {0x80, 0x01},
		2097151:    {0xff, 0xff, 0x7f},
		2147483647: {0xff, 0xff, 0xff, 0xff, 0x07},
	}
	for v, want := range cases {
		if got := AppendVarInt(nil, v); !bytes.Equal(got, want) {
			t.Errorf("AppendVarInt(%d) = % x, want % x", v, got, want)
		}
		if got := VarIntLen(v); got != len(want) {
			t.Errorf("VarIntLen(%d) = %d, want %d", v, got, len(want))
		}
		got, err := ReadVarInt(bytes.NewReader(want))
		if err != nil {
			t.Errorf("ReadVarInt(% x): %v", want, err)
			continue
		}
		if got != int32(v) {
			t.Errorf("ReadVarInt(% x) = %d, want %d", want, got, v)
		}
	}
}

func TestReadVarIntSigned(t *testing.T) {
	got, err := ReadVarInt(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff, 0x0f}))
	if err != nil || got != -1 {
		t.Errorf("ReadVarInt(-1) = %d, %v; want -1, nil", got, err)
	}
}

func TestReadVarIntTooBig(t *testing.T) {
	_, err := ReadVarInt(bytes.NewReader([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x01}))
	if !errors.Is(err, ErrVarIntTooBig) {
		t.Errorf("err = %v, want ErrVarIntTooBig", err)
	}
}

func TestVarIntNegative(t *testing.T) {
	if got := VarIntLen(uint32(0xffffffff)); got != 5 {
		t.Errorf("VarIntLen(0xffffffff) = %d, want 5", got)
	}
}

func TestPacketRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("hello")
	if err := WritePacket(&buf, 0x2a, payload); err != nil {
		t.Fatal(err)
	}
	p, err := ReadPacket(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != 0x2a || !bytes.Equal(p.Data, payload) {
		t.Errorf("got id=%d data=%q, want id=0x2a data=%q", p.ID, p.Data, payload)
	}
}
