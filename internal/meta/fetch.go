package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VersionMeta is the slice of Mojang's per-version JSON that Zenith needs.
type VersionMeta struct {
	ID              string `json:"id"`
	ProtocolVersion int32  `json:"protocolVersion"`
	ReleaseTime     string `json:"releaseTime"`
}

// FetchVersion resolves a version id (e.g. "26.3") through piston-meta to
// its version JSON and returns the parsed metadata. piston-meta is public
// metadata; this never downloads game files.
func FetchVersion(ctx context.Context, client *http.Client, id string) (VersionMeta, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ManifestURL, nil)
	if err != nil {
		return VersionMeta{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return VersionMeta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return VersionMeta{}, fmt.Errorf("meta: manifest returned %s", resp.Status)
	}
	var manifest VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return VersionMeta{}, err
	}
	for _, v := range manifest.Versions {
		if v.ID == id {
			return fetchVersionJSON(ctx, client, v.URL)
		}
	}
	return VersionMeta{}, fmt.Errorf("meta: version %q not in manifest", id)
}

func fetchVersionJSON(ctx context.Context, client *http.Client, url string) (VersionMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return VersionMeta{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return VersionMeta{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return VersionMeta{}, fmt.Errorf("meta: version JSON returned %s", resp.Status)
	}
	var vm VersionMeta
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return VersionMeta{}, err
	}
	return vm, nil
}
