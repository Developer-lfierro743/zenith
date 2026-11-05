package protocol

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"net"
	"testing"
	"time"
)

// startServer accepts one connection, runs serve on it, and returns a wait
// func that reports the serve result (with a timeout).
func startServer(t *testing.T, serve func(net.Conn) error) (net.Addr, func() error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	done := make(chan error, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer c.Close()
		done <- serve(c)
	}()
	wait := func() error {
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			return errors.New("server goroutine timed out")
		}
	}
	return ln.Addr(), wait
}

func dial(t *testing.T, addr net.Addr) *Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", addr.String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return NewConn(c)
}

func TestOfflineLoginWithCompression(t *testing.T) {
	const threshold = int32(256)
	addr, wait := startServer(t, func(c net.Conn) error {
		sc := NewConn(c)
		start, err := sc.ReadPacket()
		if err != nil {
			return err
		}
		if start.ID != LoginSBLoginStart {
			return errors.New("expected login start")
		}
		// Set Compression is the transition packet: it goes out in the OLD
		// framing; the compressed format applies to everything after it.
		if err := sc.WritePacket(Packet{ID: LoginCBSetCompression, Data: AppendVarInt(nil, uint32(threshold))}); err != nil {
			return err
		}
		sc.EnableCompression(threshold)
		// Big payload forces real zlib compression on the wire.
		success := EncodeLoginSuccess(OfflinePlayerUUID("Steve"), "Steve", [16]byte{9})
		success = append(success, bytes.Repeat([]byte("A"), 1024)...)
		if err := sc.WritePacket(Packet{ID: LoginCBLoginSuccess, Data: success}); err != nil {
			return err
		}
		ack, err := sc.ReadPacket()
		if err != nil {
			return err
		}
		if ack.ID != LoginSBAcknowledged {
			return errors.New("expected login acknowledged")
		}
		return nil
	})

	cc := dial(t, addr)
	if err := cc.WritePacket(Packet{ID: LoginSBLoginStart, Data: LoginStart{Name: "Steve", UUID: OfflinePlayerUUID("Steve")}.Encode()}); err != nil {
		t.Fatal(err)
	}
	pkt, err := cc.ReadPacket()
	if err != nil {
		t.Fatal(err)
	}
	if pkt.ID != LoginCBSetCompression {
		t.Fatalf("expected set compression, got 0x%02x", pkt.ID)
	}
	thr, err := ReadSetCompression(pkt.Data)
	if err != nil || thr != threshold {
		t.Fatalf("threshold = %d, %v", thr, err)
	}
	cc.EnableCompression(thr)
	pkt, err = cc.ReadPacket()
	if err != nil {
		t.Fatal(err)
	}
	if pkt.ID != LoginCBLoginSuccess {
		t.Fatalf("expected login success, got 0x%02x", pkt.ID)
	}
	ls, err := ReadLoginSuccess(pkt.Data)
	if err != nil {
		t.Fatal(err)
	}
	if ls.Name != "Steve" || ls.UUID != OfflinePlayerUUID("Steve") {
		t.Errorf("profile = %+v", ls)
	}
	if err := cc.WritePacket(Packet{ID: LoginSBAcknowledged}); err != nil {
		t.Fatal(err)
	}
	if err := wait(); err != nil {
		t.Fatal(err)
	}
}

func TestEncryptedLoginFlow(t *testing.T) {
	addr, wait := startServer(t, func(c net.Conn) error {
		sc := NewConn(c)
		if _, err := sc.ReadPacket(); err != nil { // login start
			return err
		}
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			return err
		}
		der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return err
		}
		token := []byte{0xde, 0xad, 0xbe, 0xef}
		var req bytes.Buffer
		req.Write(AppendString(nil, "")) // server ID: empty since 1.7
		req.Write(AppendVarInt(nil, uint32(len(der))))
		req.Write(der)
		req.Write(AppendVarInt(nil, uint32(len(token))))
		req.Write(token)
		req.WriteByte(0) // should authenticate: false (offline mode w/ encryption)
		if err := sc.WritePacket(Packet{ID: LoginCBEncryptionReq, Data: req.Bytes()}); err != nil {
			return err
		}
		resp, err := sc.ReadPacket()
		if err != nil {
			return err
		}
		if resp.ID != LoginSBEncryptionResp {
			return errors.New("expected encryption response")
		}
		encSecret, encToken, err := ReadEncryptionResponse(resp.Data)
		if err != nil {
			return err
		}
		plainToken, err := rsa.DecryptPKCS1v15(rand.Reader, key, encToken)
		if err != nil {
			return err
		}
		if !bytes.Equal(plainToken, token) {
			return errors.New("verify token mismatch")
		}
		plainSecret, err := rsa.DecryptPKCS1v15(rand.Reader, key, encSecret)
		if err != nil {
			return err
		}
		if len(plainSecret) != 16 {
			return errors.New("shared secret must be 16 bytes")
		}
		if err := sc.EnableEncryption(plainSecret); err != nil {
			return err
		}
		if err := sc.WritePacket(Packet{ID: LoginCBSetCompression, Data: AppendVarInt(nil, 256)}); err != nil {
			return err
		}
		sc.EnableCompression(256) // applies from the next packet on
		return sc.WritePacket(Packet{ID: LoginCBLoginSuccess, Data: EncodeLoginSuccess(OfflinePlayerUUID("Alex"), "Alex", [16]byte{7})})
	})

	cc := dial(t, addr)
	if err := cc.WritePacket(Packet{ID: LoginSBLoginStart, Data: LoginStart{Name: "Alex", UUID: OfflinePlayerUUID("Alex")}.Encode()}); err != nil {
		t.Fatal(err)
	}
	pkt, err := cc.ReadPacket()
	if err != nil {
		t.Fatal(err)
	}
	if pkt.ID != LoginCBEncryptionReq {
		t.Fatalf("expected encryption request, got 0x%02x", pkt.ID)
	}
	req, err := ReadEncryptionRequest(pkt.Data)
	if err != nil {
		t.Fatal(err)
	}
	if req.ShouldAuthenticate {
		t.Fatal("expected offline mode (should authenticate = false)")
	}
	pubAny, err := x509.ParsePKIXPublicKey(req.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("public key type %T", pubAny)
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	respData, err := EncryptionResponse{SharedSecret: secret, VerifyToken: req.VerifyToken}.Encode(pub)
	if err != nil {
		t.Fatal(err)
	}
	if err := cc.WritePacket(Packet{ID: LoginSBEncryptionResp, Data: respData}); err != nil {
		t.Fatal(err)
	}
	if err := cc.EnableEncryption(secret); err != nil {
		t.Fatal(err)
	}
	pkt, err = cc.ReadPacket() // set compression (now encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if pkt.ID != LoginCBSetCompression {
		t.Fatalf("expected set compression, got 0x%02x", pkt.ID)
	}
	thr, err := ReadSetCompression(pkt.Data)
	if err != nil {
		t.Fatal(err)
	}
	cc.EnableCompression(thr)
	pkt, err = cc.ReadPacket() // login success (encrypted + compressed)
	if err != nil {
		t.Fatal(err)
	}
	if pkt.ID != LoginCBLoginSuccess {
		t.Fatalf("expected login success, got 0x%02x", pkt.ID)
	}
	ls, err := ReadLoginSuccess(pkt.Data)
	if err != nil {
		t.Fatal(err)
	}
	if ls.Name != "Alex" {
		t.Errorf("name = %q", ls.Name)
	}
	if err := wait(); err != nil {
		t.Fatal(err)
	}
}

func TestServerHashVectors(t *testing.T) {
	// Documented Minecraft vectors (minecraft.wiki protocol/Encryption):
	// SHA-1 rendered as a signed big-endian integer in base 16.
	cases := map[string]string{
		"Notch": "4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48",
		"jeb_":  "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1",
		"simon": "88e16a1019277b15d58faf0541e11910eb756f6",
	}
	for in, want := range cases {
		if got := ServerHash(in, nil, nil); got != want {
			t.Errorf("ServerHash(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOfflinePlayerUUID(t *testing.T) {
	a := OfflinePlayerUUID("Notch")
	if b := OfflinePlayerUUID("Notch"); a != b {
		t.Fatal("not deterministic")
	}
	if a[6]>>4 != 3 {
		t.Errorf("version nibble = %x, want 3", a[6]>>4)
	}
	if a[8]>>6 != 2 {
		t.Errorf("variant bits = %x, want 2", a[8]>>6)
	}
	if got := UndashedUUID(a); len(got) != 32 {
		t.Errorf("UndashedUUID len = %d", len(got))
	}
}
