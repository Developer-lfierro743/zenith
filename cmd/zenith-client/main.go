// Command zenith-client is the playable Go client for Minecraft Java Edition
// 26.3: protocol, rendering (wgpu), input and client-side physics.
//
// Today (M1): `zenith-client ping HOST [PORT]` performs a Server List Ping
// against any server and prints what it sees.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Developer-lfierro743/zenith/internal/meta"
	"github.com/Developer-lfierro743/zenith/internal/protocol"
	"github.com/Developer-lfierro743/zenith/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "ping":
		ping(os.Args[2:])
	case "version":
		fmt.Printf("zenith-client — target: Minecraft Java Edition %s\n", version.Pinned)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  zenith-client ping [-protocol N] [-timeout D] HOST [PORT]   server list ping
  zenith-client version`)
}

func ping(args []string) {
	fs := flag.NewFlagSet("ping", flag.ExitOnError)
	proto := fs.Int("protocol", 0, "handshake protocol version (0 = resolve from piston-meta)")
	timeout := fs.Duration("timeout", 5*time.Second, "network timeout")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		usage()
		os.Exit(2)
	}
	host := fs.Arg(0)
	port := 25565
	if fs.NArg() > 1 {
		p, err := strconv.Atoi(fs.Arg(1))
		if err != nil {
			fatal(err)
		}
		port = p
	}
	pv := int32(*proto)
	if pv == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		vm, err := meta.FetchVersion(ctx, nil, version.Pinned)
		if err != nil {
			fatal(fmt.Errorf("resolve protocol version from piston-meta: %w (use -protocol to pin it)", err))
		}
		pv = vm.ProtocolVersion
	}
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, *timeout)
	if err != nil {
		fatal(err)
	}
	defer conn.Close()
	hs := protocol.Handshake{
		ProtocolVersion: pv,
		ServerAddress:   host,
		ServerPort:      uint16(port),
		NextState:       protocol.NextStatus,
	}
	resp, rtt, err := protocol.PingStatus(conn, hs, *timeout)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("%s — %s (protocol %d) — %d/%d players — %v\n",
		addr, resp.Version.Name, resp.Version.Protocol, resp.Players.Online, resp.Players.Max, rtt.Round(time.Millisecond))
	if motd := resp.MotdText(); motd != "" {
		fmt.Printf("MOTD: %s\n", motd)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "zenith-client:", err)
	os.Exit(1)
}
