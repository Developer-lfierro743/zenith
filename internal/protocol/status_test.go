package protocol

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// fakeStatusServer implements the server side of the status flow with
// canned data, so the client is exercised end-to-end without a real
// Minecraft server (and without any Mojang files) in CI.
func fakeStatusServer(t *testing.T) net.Addr {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveFakeStatus(conn)
		}
	}()
	return ln.Addr()
}

func serveFakeStatus(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	hs, err := ReadPacket(conn)
	if err != nil || hs.ID != 0x00 || !validStatusHandshake(hs.Data) {
		return
	}
	req, err := ReadPacket(conn)
	if err != nil || req.ID != PacketStatusRequest {
		return
	}
	status := `{"version":{"name":"fake-26.3","protocol":9999},` +
		`"players":{"max":20,"online":3},` +
		`"description":{"text":"fake server ✓"}}`
	if err := WritePacket(conn, PacketStatusResponse, AppendString(nil, status)); err != nil {
		return
	}
	ping, err := ReadPacket(conn)
	if err != nil || ping.ID != PacketPingRequest || len(ping.Data) < 8 {
		return
	}
	_ = WritePacket(conn, PacketPongResponse, ping.Data)
}

// validStatusHandshake decodes the handshake fields and reports whether it
// targets the status state — a server-side round-trip check of Encode.
func validStatusHandshake(data []byte) bool {
	rd := bytes.NewReader(data)
	if _, err := ReadVarInt(rd); err != nil { // protocol version
		return false
	}
	rem := data[len(data)-rd.Len():]
	_, used, err := ReadString(rem)
	if err != nil {
		return false
	}
	rest := rem[used:]
	if len(rest) < 3 {
		return false
	}
	rd.Reset(rest[2:])
	next, err := ReadVarInt(rd)
	return err == nil && next == NextStatus
}

func TestPingStatusFlow(t *testing.T) {
	addr := fakeStatusServer(t)
	conn, err := net.DialTimeout("tcp", addr.String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	hs := Handshake{ProtocolVersion: 9999, ServerAddress: "127.0.0.1", ServerPort: 25565, NextState: NextStatus}
	resp, rtt, err := PingStatus(conn, hs, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Version.Name != "fake-26.3" || resp.Version.Protocol != 9999 {
		t.Errorf("version = %+v", resp.Version)
	}
	if resp.Players.Online != 3 || resp.Players.Max != 20 {
		t.Errorf("players = %+v", resp.Players)
	}
	if got := resp.MotdText(); got != "fake server ✓" {
		t.Errorf("MotdText() = %q, want %q", got, "fake server ✓")
	}
	if rtt < 0 || rtt > 2*time.Second {
		t.Errorf("rtt = %v", rtt)
	}
}
