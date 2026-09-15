package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/catalog"
	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

// TestGetFailsClosedOnManifestHashMismatch is the integrity regression guard:
// the same fixture download must succeed when the manifest hash matches and must
// fail — writing nothing and seeding nothing — when it does not. A corrupt
// download that still lands in the HF layout defeats the core guarantee of
// PROTOCOL.md §1, so this asserts the destination is untouched, not merely that
// a warning was printed.
func TestGetFailsClosedOnManifestHashMismatch(t *testing.T) {
	content := bytes.Repeat([]byte("integrity-"), 512)
	sum := sha256.Sum256(content)
	goodSHA := hex.EncodeToString(sum[:])

	tampered := append([]byte(nil), content...)
	tampered[0] ^= 0xff
	tamperedSum := sha256.Sum256(tampered)
	wrongSHA := hex.EncodeToString(tamperedSum[:])

	t.Run("matching hash writes the layout", func(t *testing.T) {
		catalogPath := writeFixtureCatalog(t, content, goodSHA)
		dest := filepath.Join(t.TempDir(), "out")

		if code := doGet(fixtureModelID, dest, getOptions{catalogPath: catalogPath, noSeed: true}); code != 0 {
			t.Fatalf("doGet exit = %d want 0", code)
		}
		got := filepath.Join(dest, "snapshots", "main", fixtureFileName)
		if _, err := os.Stat(got); err != nil {
			t.Fatalf("verified download missing from layout: %v", err)
		}
	})

	t.Run("mismatched hash writes nothing and does not seed", func(t *testing.T) {
		catalogPath := writeFixtureCatalog(t, content, wrongSHA)
		dest := filepath.Join(t.TempDir(), "out")

		// Seeding is left enabled (noSeed false) so that reaching the seed path
		// at all would be visible as a staging dir under dest.
		if code := doGet(fixtureModelID, dest, getOptions{catalogPath: catalogPath}); code == 0 {
			t.Fatal("doGet exit = 0, want non-zero for a manifest hash mismatch")
		}
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			entries, _ := os.ReadDir(dest)
			t.Fatalf("dest %s exists after failed verification (entries: %v)", dest, entries)
		}
	})
}

const (
	fixtureModelID  = "demo/tiny-gguf"
	fixtureFileName = "Q4_K_M.gguf"
)

// writeFixtureCatalog builds a self-contained fixture catalog in a temp dir whose
// single model declares wantSHA for its only file, and returns the catalog.json
// path. The bytes on disk are always content, so passing a wantSHA that does not
// hash content simulates a tampered manifest.
func writeFixtureCatalog(t *testing.T, content []byte, wantSHA string) string {
	t.Helper()
	root := t.TempDir()

	fixtureDir := filepath.Join(root, "fixtures", "demo", "tiny-gguf")
	if err := os.MkdirAll(fixtureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, fixtureFileName), content, 0o644); err != nil {
		t.Fatal(err)
	}

	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ModelID:       fixtureModelID,
		Files:         []manifest.File{{Path: fixtureFileName, Size: int64(len(content)), SHA256: wantSHA}},
		License:       manifest.License{SPDX: "MIT", Redistributable: true},
		Magnet:        "fixture://" + fixtureModelID,
		Webseeds:      []string{"file://fixtures/demo/tiny-gguf"},
		CreatedAt:     "2026-09-15T00:00:00Z",
	}
	manifestRel := filepath.Join("models", "demo", "tiny-gguf", "manifest.json")
	writeJSON(t, filepath.Join(root, manifestRel), m)

	catalogPath := filepath.Join(root, "catalog.json")
	writeJSON(t, catalogPath, catalog.File{
		SchemaVersion: manifest.SchemaVersion,
		GeneratedAt:   m.CreatedAt,
		Models: []catalog.Entry{{
			ModelID:   m.ModelID,
			License:   m.License,
			Magnet:    m.Magnet,
			SizeBytes: int64(len(content)),
			Manifest:  filepath.ToSlash(manifestRel),
		}},
	})
	return catalogPath
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}
