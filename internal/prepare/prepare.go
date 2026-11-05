// Package prepare builds Zenith's game resources from the user's own copy of
// Minecraft: Java Edition 26.3 — the same trick Horizon's horizon-prepare
// uses. It verifies every input against the SHA1 pins in research-lock.json,
// runs vanilla's own data generator on a JDK to export registries, and
// extracts assets. Everything lands in a gitignored .local/ directory;
// nothing is committed, downloaded or shared.
package prepare

import "fmt"

// Options configures a preparation run.
type Options struct {
	// MinecraftDir is the user's .minecraft (official launcher layout).
	MinecraftDir string

	// ServerJar is the path to the user's downloaded 26.3 server jar.
	ServerJar string

	// JavaHome points at a JDK 25 runtime for the vanilla data generator.
	JavaHome string

	// OutputDir is where prepared resources land (default .local/).
	OutputDir string
}

// Run prepares resources. TODO(M3).
func Run(opts Options) error {
	return fmt.Errorf("prepare: not implemented yet (roadmap M3); opts=%+v", opts)
}
