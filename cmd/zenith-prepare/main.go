// Command zenith-prepare builds Zenith's game resources from the user's own
// copy of Minecraft: Java Edition 26.3 — the same trick Horizon's
// horizon-prepare uses. Every input is verified against the SHA1 pins in
// research-lock.json, vanilla's own data generator runs on a JDK to export
// registries, and assets are extracted locally.
//
// Prepared resources land in gitignored .local/ and are never committed,
// downloaded or shared.
//
// TODO(M3).
package main

import (
	"flag"
	"fmt"

	"github.com/Developer-lfierro743/zenith/internal/version"
)

func main() {
	minecraftDir := flag.String("minecraft-dir", "", "path to the user's .minecraft (official launcher layout)")
	serverJar := flag.String("server-jar", "", "path to the user's downloaded 26.3 server jar")
	javaHome := flag.String("java-home", "", "JDK 25 runtime for vanilla's data generator")
	flag.Parse()

	fmt.Printf("zenith-prepare (pre-alpha) — target: Minecraft Java Edition %s\n", version.Pinned)
	fmt.Printf("minecraft-dir=%q server-jar=%q java-home=%q\n", *minecraftDir, *serverJar, *javaHome)
	fmt.Println("not implemented yet — see docs/ROADMAP.md (M3)")
}
