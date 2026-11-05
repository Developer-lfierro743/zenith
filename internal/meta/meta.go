// Package meta reads Mojang's piston-meta endpoints: the version manifest,
// per-version JSON (downloads, libraries, asset index) and asset indexes.
// Zenith uses these only to locate files for a signed-in owner, and always
// verifies SHA1 pins before use.
package meta

const (
	// ManifestURL is piston-meta's version_manifest_v2.
	ManifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

	// ResourcesBaseURL hosts asset objects by their first two hash chars.
	ResourcesBaseURL = "https://resources.download.minecraft.net"
)

// VersionManifest is the top of version_manifest_v2.json.
type VersionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		SHA1        string `json:"sha1"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}
