//go:build ignore

// gen_catalog writes testdata/catalog.json and a signed manifest for the
// Qwen/Qwen3-8B fixture model. Run after make fixtures:
//
//	go run ./scripts/gen_fixtures.go && go run ./scripts/gen_catalog.go
package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/blob"
)

const modelID = "Qwen/Qwen3-8B"

var fixtureNames = []string{
	"model.safetensors",
	"tokenizer.json",
	"Q4_K_M.gguf",
}

type fileEntry struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type manifestDoc struct {
	SchemaVersion int         `json:"schemaVersion"`
	ModelID       string      `json:"modelId"`
	Files         []fileEntry `json:"files"`
	License       struct {
		SPDX            string `json:"spdx"`
		Redistributable bool   `json:"redistributable"`
	} `json:"license"`
	Publisher struct {
		KeyID     string `json:"keyId"`
		Signature string `json:"signature"`
	} `json:"publisher"`
	Magnet    string   `json:"magnet"`
	Webseeds  []string `json:"webseeds"`
	CreatedAt string   `json:"createdAt"`
}

func main() {
	fixDir, err := filepath.Abs(filepath.Join("testdata", "fixtures"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "abs fixtures:", err)
		os.Exit(1)
	}
	var files []fileEntry
	var total int64
	for _, name := range fixtureNames {
		p := filepath.Join(fixDir, name)
		hash, size, err := blob.HashFile(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "hash:", p, err)
			os.Exit(1)
		}
		files = append(files, fileEntry{Path: name, Size: size, SHA256: hash})
		total += size
	}

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "keygen:", err)
		os.Exit(1)
	}
	_ = pub

	m := manifestDoc{
		SchemaVersion: 1,
		ModelID:       modelID,
		Files:         files,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Magnet:        "fixture://" + modelID,
		Webseeds:      []string{"file://" + fixDir}, // absolute path; three-slash form on Unix
	}
	m.License.SPDX = "Apache-2.0"
	m.License.Redistributable = true
	m.Publisher.KeyID = "ed25519:dev-fixture"

	canonical, err := json.Marshal(&struct {
		SchemaVersion int         `json:"schemaVersion"`
		ModelID       string      `json:"modelId"`
		Files         []fileEntry `json:"files"`
		License       any         `json:"license"`
		Publisher     struct {
			KeyID     string `json:"keyId"`
			Signature string `json:"signature"`
		} `json:"publisher"`
		Magnet    string   `json:"magnet"`
		Webseeds  []string `json:"webseeds"`
		CreatedAt string   `json:"createdAt"`
	}{
		SchemaVersion: m.SchemaVersion,
		ModelID:       m.ModelID,
		Files:         m.Files,
		License:       m.License,
		Publisher: struct {
			KeyID     string `json:"keyId"`
			Signature string `json:"signature"`
		}{KeyID: m.Publisher.KeyID, Signature: ""},
		Magnet:    m.Magnet,
		Webseeds:  m.Webseeds,
		CreatedAt: m.CreatedAt,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "canonical:", err)
		os.Exit(1)
	}
	sig := ed25519.Sign(priv, canonical)
	m.Publisher.Signature = base64.StdEncoding.EncodeToString(sig)

	manifestDir := filepath.Join("testdata", "models", "Qwen", "Qwen3-8B")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}
	manifestPath := filepath.Join(manifestDir, "manifest.json")
	mb, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal manifest:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(manifestPath, mb, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write manifest:", err)
		os.Exit(1)
	}

	catalog := map[string]any{
		"schemaVersion": 1,
		"generatedAt":   time.Now().UTC().Format(time.RFC3339),
		"models": []map[string]any{
			{
				"modelId":   modelID,
				"license":   map[string]any{"spdx": "Apache-2.0", "redistributable": true},
				"magnet":    m.Magnet,
				"sizeBytes": total,
				"manifest":  "/models/Qwen/Qwen3-8B/manifest.json",
			},
		},
	}
	cb, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal catalog:", err)
		os.Exit(1)
	}
	catalogPath := filepath.Join("testdata", "catalog.json")
	if err := os.WriteFile(catalogPath, cb, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write catalog:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", catalogPath)
	fmt.Println("wrote", manifestPath)
}
