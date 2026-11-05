// Package auth implements Microsoft-account authentication for Zenith:
// OAuth device flow → Xbox Live (XBL) → XSTS → Minecraft services token,
// plus the mcstore entitlement check that proves the account owns Java
// Edition. Ownership is required; there is no offline mode.
package auth

import "fmt"

// Session is an authenticated Minecraft session.
type Session struct {
	// Gamertag is the player's Xbox gamertag.
	Gamertag string

	// MCAccessToken is the Minecraft services JWT ("MinecraftAccessToken").
	MCAccessToken string

	// ExpiresAt is a Unix timestamp after which MCAccessToken is invalid.
	ExpiresAt int64

	// OwnsJavaEdition reports the mcstore entitlement check result.
	OwnsJavaEdition bool
}

// DeviceFlow begins the Microsoft OAuth device-code flow and returns the
// verification URL plus the user code for the player to enter in a browser.
//
// TODO(M2).
func DeviceFlow() (verifyURL, userCode string, err error) {
	return "", "", fmt.Errorf("auth: not implemented yet (roadmap M2)")
}
