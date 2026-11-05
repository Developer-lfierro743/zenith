package protocol

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/Developer-lfierro743/zenith/internal/version"
)

// Packet IDs in the Login state, pinned to Java Edition 26.3 (protocol 777;
// source: minecraft.wiki "Java Edition protocol/Packets"). Kept together so
// a version bump is a reviewable single-file change.
const (
	LoginCBDisconnect     int32 = 0x00
	LoginCBEncryptionReq  int32 = 0x01
	LoginCBLoginSuccess   int32 = 0x02
	LoginCBSetCompression int32 = 0x03
	LoginCBPluginRequest  int32 = 0x04
	LoginCBCookieRequest  int32 = 0x05

	LoginSBLoginStart     int32 = 0x00
	LoginSBEncryptionResp int32 = 0x01
	LoginSBPluginResp     int32 = 0x02
	LoginSBAcknowledged   int32 = 0x03
	LoginSBCookieResp     int32 = 0x04
)

// TargetProtocolNote documents where the login packet IDs above come from.
const TargetProtocolNote = "Java Edition " + version.Pinned + " (protocol 777)"

// LoginStart is serverbound packet 0x00 and always opens the login state.
type LoginStart struct {
	Name string
	UUID [16]byte
}

// Encode returns the Login Start payload.
func (l LoginStart) Encode() []byte {
	b := AppendString(nil, l.Name)
	return append(b, l.UUID[:]...)
}

// EncryptionRequest is clientbound packet 0x01. When ShouldAuthenticate is
// true the client must authenticate with Mojang's session server before
// answering (online mode); when false the exchange still enables encryption
// but no session-server call is made (offline mode with encryption).
type EncryptionRequest struct {
	ServerID           string
	PublicKey          []byte // SubjectPublicKeyInfo (X.509 SPKI) DER
	VerifyToken        []byte
	ShouldAuthenticate bool
}

// ReadEncryptionRequest parses the Encryption Request payload.
func ReadEncryptionRequest(data []byte) (EncryptionRequest, error) {
	r := newFieldReader(data)
	var req EncryptionRequest
	var err error
	if req.ServerID, err = r.str(); err != nil {
		return EncryptionRequest{}, err
	}
	if req.PublicKey, err = r.byteSlice(); err != nil {
		return EncryptionRequest{}, err
	}
	if req.VerifyToken, err = r.byteSlice(); err != nil {
		return EncryptionRequest{}, err
	}
	if req.ShouldAuthenticate, err = r.boolean(); err != nil {
		return EncryptionRequest{}, err
	}
	return req, nil
}

// EncryptionResponse is serverbound packet 0x01: the shared secret and the
// server's verify token, both RSA-encrypted (PKCS#1 v1.5) with the public
// key from the Encryption Request.
type EncryptionResponse struct {
	SharedSecret []byte
	VerifyToken  []byte
}

// Encode encrypts both fields with pub and returns the payload.
func (e EncryptionResponse) Encode(pub *rsa.PublicKey) ([]byte, error) {
	encSecret, err := rsa.EncryptPKCS1v15(rand.Reader, pub, e.SharedSecret)
	if err != nil {
		return nil, err
	}
	encToken, err := rsa.EncryptPKCS1v15(rand.Reader, pub, e.VerifyToken)
	if err != nil {
		return nil, err
	}
	b := AppendVarInt(nil, uint32(len(encSecret)))
	b = append(b, encSecret...)
	b = AppendVarInt(b, uint32(len(encToken)))
	b = append(b, encToken...)
	return b, nil
}

// ReadEncryptionResponse parses the payload server-side (used by tests and
// future tooling): two length-prefixed byte arrays.
func ReadEncryptionResponse(data []byte) (secret, verifyToken []byte, err error) {
	r := newFieldReader(data)
	if secret, err = r.byteSlice(); err != nil {
		return nil, nil, err
	}
	if verifyToken, err = r.byteSlice(); err != nil {
		return nil, nil, err
	}
	return secret, verifyToken, nil
}

// LoginSuccess is clientbound packet 0x02: the authenticated profile plus a
// session UUID. Any trailing fields (older versions carried profile
// properties here) are ignored.
type LoginSuccess struct {
	UUID [16]byte
	Name string
}

// ReadLoginSuccess parses the Login Success payload.
func ReadLoginSuccess(data []byte) (LoginSuccess, error) {
	r := newFieldReader(data)
	var ls LoginSuccess
	var err error
	if ls.UUID, err = r.uuid(); err != nil {
		return LoginSuccess{}, err
	}
	if ls.Name, err = r.str(); err != nil {
		return LoginSuccess{}, err
	}
	return ls, nil
}

// EncodeLoginSuccess builds a Login Success payload (used by tests and the
// future fake/real server tooling).
func EncodeLoginSuccess(uuid [16]byte, name string, sessionUUID [16]byte) []byte {
	b := append([]byte(nil), uuid[:]...)
	b = AppendString(b, name)
	return append(b, sessionUUID[:]...)
}

// ReadSetCompression parses clientbound packet 0x03.
func ReadSetCompression(data []byte) (int32, error) {
	r := newFieldReader(data)
	return r.varInt()
}

// LoginDisconnect is clientbound packet 0x00: the server refused the login.
type LoginDisconnect struct {
	Reason string // raw JSON text component
}

// ReadLoginDisconnect parses the Disconnect (login) payload.
func ReadLoginDisconnect(data []byte) (LoginDisconnect, error) {
	r := newFieldReader(data)
	reason, err := r.str()
	if err != nil {
		return LoginDisconnect{}, err
	}
	return LoginDisconnect{Reason: reason}, nil
}

// PlainReason extracts readable text from the JSON component reason.
func (l LoginDisconnect) PlainReason() string {
	var raw any
	if err := json.Unmarshal([]byte(l.Reason), &raw); err != nil {
		return l.Reason
	}
	switch v := raw.(type) {
	case string:
		return v
	case map[string]any:
		var sb strings.Builder
		walkComponent(&sb, v)
		return sb.String()
	default:
		return l.Reason
	}
}

// OfflinePlayerUUID derives the vanilla offline-mode UUID: UUID v3 (MD5) of
// "OfflinePlayer:<name>" with RFC 4122 version/variant bits set.
func OfflinePlayerUUID(name string) [16]byte {
	h := md5.Sum([]byte("OfflinePlayer:" + name))
	h[6] = (h[6] & 0x0f) | 0x30 // version 3
	h[8] = (h[8] & 0x3f) | 0x80 // IETF variant
	return h
}

// UndashedUUID formats u as the 32-char lowercase hex form Mojang APIs use.
func UndashedUUID(u [16]byte) string {
	return hex.EncodeToString(u[:])
}
