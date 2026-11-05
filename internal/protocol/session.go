package protocol

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

// SessionJoinURL is Mojang's session server endpoint the client calls to
// register itself as joining an online-mode server.
const SessionJoinURL = "https://sessionserver.mojang.com/session/minecraft/join"

// ServerHash computes Minecraft's non-standard SHA-1 hexdigest used as the
// serverId in a session join: SHA-1 over the server ID (empty since 1.7),
// the shared secret and the server's DER public key, rendered as a signed
// big-endian integer in base 16 — with a minus sign when negative. Verified
// against the documented vectors ("Notch", "jeb_", "simon") in tests.
func ServerHash(serverID string, sharedSecret, publicKey []byte) string {
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret)
	h.Write(publicKey)
	return mcHexDigest(h.Sum(nil))
}

// mcHexDigest renders a digest the way Minecraft does: as one large signed
// two's-complement integer in base 16, no zero padding.
func mcHexDigest(sum []byte) string {
	n := new(big.Int).SetBytes(sum)
	if len(sum) > 0 && sum[0]&0x80 != 0 {
		mod := new(big.Int).Lsh(big.NewInt(1), uint(len(sum)*8))
		n.Sub(n, mod)
	}
	return n.Text(16)
}

// SessionJoin tells Mojang this client, authenticated as accessToken, is
// about to join an online-mode server. The server verifies it with /hasJoined
// using the same serverHash. Success is HTTP 204 No Content.
func SessionJoin(ctx context.Context, client *http.Client, accessToken string, profileUUID [16]byte, serverHash string) error {
	if client == nil {
		client = http.DefaultClient
	}
	body, err := json.Marshal(map[string]string{
		"accessToken":     accessToken,
		"selectedProfile": UndashedUUID(profileUUID),
		"serverId":        serverHash,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SessionJoinURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusForbidden:
		var e struct {
			Error        string `json:"error"`
			ErrorMessage string `json:"errorMessage"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&e); err == nil && e.Error != "" {
			return fmt.Errorf("session join rejected: %s: %s", e.Error, e.ErrorMessage)
		}
		return fmt.Errorf("session join rejected: HTTP 403")
	case http.StatusTooManyRequests:
		return fmt.Errorf("session join rate limited (max 6 joins per 30s per account)")
	default:
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("session join: unexpected HTTP %d: %s", resp.StatusCode, b)
	}
}
