package protocol

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"
)

// Ground-truth AES-128/CFB8 vector generated with:
//
//	openssl enc -aes-128-cfb8 \
//	  -K 000102030405060708090a0b0c0d0e0f \
//	  -iv f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff -nosalt
//
// plaintext: "Zenith CFB8 test vector!!" (25 bytes)
var cfb8Vector = struct{ key, iv, pt, ct string }{
	key: "000102030405060708090a0b0c0d0e0f",
	iv:  "f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff",
	pt:  "5a656e6974682043464238207465737420766563746f722121",
	ct:  "3c34e250cf0acb28bc3f7c7fdd126ddb89f892d827aea72012",
}

func TestCFB8AgainstOpenSSL(t *testing.T) {
	key, err := hex.DecodeString(cfb8Vector.key)
	if err != nil {
		t.Fatal(err)
	}
	iv, err := hex.DecodeString(cfb8Vector.iv)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := hex.DecodeString(cfb8Vector.pt)
	if err != nil {
		t.Fatal(err)
	}
	wantCT, err := hex.DecodeString(cfb8Vector.ct)
	if err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	enc := &cfb8{b: block}
	copy(enc.iv[:], iv)
	gotCT := make([]byte, len(pt))
	enc.XORKeyStream(gotCT, pt)
	if !bytes.Equal(gotCT, wantCT) {
		t.Errorf("encrypt = %x, want %x", gotCT, wantCT)
	}

	dec := &cfb8{b: block, decrypt: true}
	copy(dec.iv[:], iv)
	gotPT := make([]byte, len(wantCT))
	dec.XORKeyStream(gotPT, wantCT)
	if !bytes.Equal(gotPT, pt) {
		t.Errorf("decrypt = %x, want %x", gotPT, pt)
	}
}

// TestCFB8InPlace checks the aliasing case cipher.StreamReader relies on
// (dst == src).
func TestCFB8InPlace(t *testing.T) {
	key, _ := hex.DecodeString(cfb8Vector.key)
	iv, _ := hex.DecodeString(cfb8Vector.iv)
	pt, _ := hex.DecodeString(cfb8Vector.pt)
	wantCT, _ := hex.DecodeString(cfb8Vector.ct)
	block, _ := aes.NewCipher(key)

	enc := &cfb8{b: block}
	copy(enc.iv[:], iv)
	buf := append([]byte(nil), pt...)
	enc.XORKeyStream(buf, buf)
	if !bytes.Equal(buf, wantCT) {
		t.Fatalf("in-place encrypt = %x, want %x", buf, wantCT)
	}

	dec := &cfb8{b: block, decrypt: true}
	copy(dec.iv[:], iv)
	dec.XORKeyStream(buf, buf)
	if !bytes.Equal(buf, pt) {
		t.Fatalf("in-place decrypt = %x, want %x", buf, pt)
	}
}
