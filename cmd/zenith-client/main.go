// Command zenith-client is the playable Go client for Minecraft Java Edition
// 26.3: protocol, rendering (wgpu), input and client-side physics.
//
// TODO(M1): protocol — join a real vanilla/Fabric server.
// TODO(M4): rendering — chunk meshes, vanilla lighting.
// TODO(M5): playable — physics, interaction, HUD.
package main

import (
	"fmt"

	"github.com/Developer-lfierro743/zenith/internal/version"
)

func main() {
	fmt.Printf("zenith-client (pre-alpha) — target: Minecraft Java Edition %s\n", version.Pinned)
	fmt.Println("not implemented yet — see docs/ROADMAP.md (M1)")
}
