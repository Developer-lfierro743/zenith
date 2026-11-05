// Package version pins the vanilla version Zenith reimplements.
//
// The pin exists so every subsystem (protocol, prepare, rendering, harnesses)
// agrees on one target. Moving the pin is a deliberate, repo-wide event.
package version

const (
	// Pinned is the Minecraft: Java Edition version Zenith targets.
	Pinned = "26.3"
)
