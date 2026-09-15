// Package catalog is the in-memory index over manifests plus swarm-health
// reporting. The catalog is served as static JSON (PROTOCOL.md §6); this package
// is the query surface behind the web UI and the Python shim.
package catalog

import (
	"errors"
	"strings"

	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

// ErrNotFound is returned by Health for an unknown modelId.
var ErrNotFound = errors.New("catalog: model not found")

// SwarmHealth is the per-model liveness shown on a model page: seeds, peers,
// whether the last webseed fetch worked, and whether checksums verified.
type SwarmHealth struct {
	ModelID       string `json:"modelId"`
	Seeders       int    `json:"seeders"`
	Peers         int    `json:"peers"`
	LastWebseedOK string `json:"lastWebseedOk"` // RFC3339, empty if never
	ChecksumOK    bool   `json:"checksumOk"`
}

// Index holds manifests keyed by modelId and answers searches.
type Index struct {
	byID map[string]manifest.Manifest
}

// NewIndex returns an empty, usable index.
func NewIndex() *Index {
	return &Index{byID: make(map[string]manifest.Manifest)}
}

// Add ingests a manifest. It MUST reject manifests that fail the license gate
// (PROTOCOL.md §3) before indexing.
func (i *Index) Add(m manifest.Manifest) error {
	if !manifest.LicenseAllowed(m.License.SPDX) || !m.License.Redistributable {
		return errors.New("catalog: license not redistributable")
	}
	i.byID[m.ModelID] = m
	return nil
}

// Search returns manifests matching a free-text query over modelId and files.
// Ranking is local and forkable by design (PROTOCOL.md §7).
func (i *Index) Search(q string) []manifest.Manifest {
	if q == "" {
		return nil
	}
	var out []manifest.Manifest
	for id, m := range i.byID {
		if strings.Contains(id, q) {
			out = append(out, m)
		}
	}
	return out
}

// Lookup returns a manifest by modelId.
func (i *Index) Lookup(modelID string) (manifest.Manifest, bool) {
	m, ok := i.byID[modelID]
	return m, ok
}

// LookupByMagnet returns a manifest whose magnet URI matches exactly.
func (i *Index) LookupByMagnet(magnet string) (manifest.Manifest, bool) {
	for _, m := range i.byID {
		if m.Magnet == magnet {
			return m, true
		}
	}
	return manifest.Manifest{}, false
}

// All returns every indexed manifest.
func (i *Index) All() []manifest.Manifest {
	out := make([]manifest.Manifest, 0, len(i.byID))
	for _, m := range i.byID {
		out = append(out, m)
	}
	return out
}

// Health reports swarm liveness for one model.
func (i *Index) Health(modelID string) (SwarmHealth, error) {
	if _, ok := i.byID[modelID]; !ok {
		return SwarmHealth{}, ErrNotFound
	}
	return SwarmHealth{ModelID: modelID}, nil
}
