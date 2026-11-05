package protocol

import (
	"bytes"
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
	}
}
