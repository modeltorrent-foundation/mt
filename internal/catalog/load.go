package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

// Entry is one row in catalog.json (api/catalog.md).
type Entry struct {
	ModelID   string           `json:"modelId"`
	License   manifest.License `json:"license"`
	Magnet    string           `json:"magnet"`
	SizeBytes int64            `json:"sizeBytes"`
	Manifest  string           `json:"manifest"` // path relative to catalog root, e.g. /models/org/name/manifest.json
}

// File is the top-level catalog.json envelope.
type File struct {
	SchemaVersion int     `json:"schemaVersion"`
	GeneratedAt   string  `json:"generatedAt"`
	Models        []Entry `json:"models"`
}

// LoadFile reads catalog.json from path and builds an Index. Manifest files are
// resolved relative to the directory containing catalog.json.
func LoadFile(catalogPath string) (*Index, *File, error) {
	b, err := os.ReadFile(catalogPath)
	if err != nil {
		return nil, nil, fmt.Errorf("catalog: read %s: %w", catalogPath, err)
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, nil, fmt.Errorf("catalog: decode: %w", err)
	}
	root := filepath.Dir(catalogPath)
	idx := NewIndex()
	for _, e := range f.Models {
		m, err := loadManifest(root, e.Manifest)
		if err != nil {
			return nil, nil, fmt.Errorf("catalog: %s: %w", e.ModelID, err)
		}
		if m.Magnet == "" {
			m.Magnet = e.Magnet
		}
		if err := idx.Add(m); err != nil {
			return nil, nil, fmt.Errorf("catalog: add %s: %w", e.ModelID, err)
		}
	}
	return idx, &f, nil
}

func loadManifest(catalogRoot, manifestPath string) (manifest.Manifest, error) {
	p := manifestPath
	if strings.HasPrefix(p, "/") {
		p = strings.TrimPrefix(p, "/")
	}
	p = filepath.Join(catalogRoot, filepath.FromSlash(p))
	b, err := os.ReadFile(p)
	if err != nil {
		return manifest.Manifest{}, err
	}
	return manifest.Parse(b)
}

// FindCatalog walks upward from start looking for testdata/catalog.json or
// catalog.json in the current directory.
func FindCatalog(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		for _, name := range []string{
			filepath.Join("testdata", "catalog.json"),
			"catalog.json",
		} {
			p := filepath.Join(dir, name)
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("catalog: no catalog.json found from %s", start)
}
