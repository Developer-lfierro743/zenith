// Command zenith-launcher installs, updates and starts Zenith, and prepares
// game resources from the user's own copy of Minecraft: Java Edition 26.3.
//
// TODO(M2): Microsoft OAuth device flow, entitlement check, piston-meta
// manifest, install management, self-update.
package main

import (
	"fmt"

	"github.com/Developer-lfierro743/zenith/internal/version"
)

func main() {
	fmt.Printf("zenith-launcher (pre-alpha) — target: Minecraft Java Edition %s\n", version.Pinned)
	fmt.Println("not implemented yet — see docs/ROADMAP.md (M2)")
}
