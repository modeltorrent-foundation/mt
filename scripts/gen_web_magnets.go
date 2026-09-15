//go:build ignore

// gen_web_magnets builds real BitTorrent magnets for web demo models and
// updates web/catalog.json plus per-model manifests. Run after fixtures exist:
//
//	make fixtures && go run ./scripts/gen_web_magnets.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

type demoModel struct {
	modelID  string
	files    []string // paths relative to fixture root
	fixture  string   // under web/fixtures/
	manifest string
	license  string
	webseeds []string
}

var demos = []demoModel{
	{
		modelID:  "demo/tiny-gguf",
		files:    []string{"Q4_K_M.gguf"},
		fixture:  "demo/tiny-gguf",
		manifest: "models/demo/tiny-gguf/manifest.json",
		license:  "MIT",
		webseeds: []string{"file://fixtures/demo/tiny-gguf"},
	},
	{
		modelID:  "demo/tiny-safetensors",
		files:    []string{"model.safetensors", "tokenizer.json"},
		fixture:  "demo/tiny-safetensors",
		manifest: "models/demo/tiny-safetensors/manifest.json",
		license:  "Apache-2.0",
		webseeds: []string{"file://fixtures/demo/tiny-safetensors"},
	},
	{
		modelID:  "demo/tiny-bundle",
		files:    []string{"model.safetensors", "tokenizer.json", "Q4_K_M.gguf"},
		fixture:  "demo/tiny-bundle",
		manifest: "models/demo/tiny-bundle/manifest.json",
		license:  "Apache-2.0",
		webseeds: []string{"file://fixtures/demo/tiny-bundle"},
	},
}

func main() {
	webRoot, err := filepath.Abs("web")
	if err != nil {
		fatal("abs web:", err)
	}
	fixSrc, err := filepath.Abs(filepath.Join("testdata", "fixtures"))
	if err != nil {
		fatal("abs fixtures:", err)
	}

	type catalogEntry struct {
		ModelID   string         `json:"modelId"`
		License   map[string]any `json:"license"`
		Magnet    string         `json:"magnet"`
		SizeBytes int64          `json:"sizeBytes"`
		Manifest  string         `json:"manifest"`
	}

	var entries []catalogEntry

	for _, d := range demos {
		fixDir := filepath.Join(webRoot, "fixtures", d.fixture)
		if err := os.MkdirAll(fixDir, 0o755); err != nil {
			fatal("mkdir:", err)
		}
		var absFiles []string
		var total int64
		for _, name := range d.files {
			src := filepath.Join(fixSrc, name)
			dst := filepath.Join(fixDir, name)
			if err := copyFile(src, dst); err != nil {
				fatal("copy", name, err)
			}
			st, err := os.Stat(dst)
			if err != nil {
				fatal("stat", dst, err)
			}
			total += st.Size()
			absFiles = append(absFiles, dst)
		}

		mi, magnet, err := torrentsvc.Create(absFiles, nil)
		if err != nil {
			fatal("Create", d.modelID, err)
		}
		if magnet == "" || !hasPrefix(magnet, "magnet:") {
			fatal("Create", d.modelID, "empty or invalid magnet")
		}

		torrentPath := filepath.Join(webRoot, filepath.Dir(d.manifest), "publish.torrent")
		if err := os.WriteFile(torrentPath, mi.Raw, 0o644); err != nil {
			fatal("write torrent:", err)
		}

		manifestPath := filepath.Join(webRoot, d.manifest)
		mb, err := os.ReadFile(manifestPath)
		if err != nil {
			fatal("read manifest:", err)
		}
		var doc map[string]any
		if err := json.Unmarshal(mb, &doc); err != nil {
			fatal("decode manifest:", err)
		}
		doc["magnet"] = magnet
		doc["webseeds"] = d.webseeds
		out, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			fatal("marshal manifest:", err)
		}
		if err := os.WriteFile(manifestPath, append(out, '\n'), 0o644); err != nil {
			fatal("write manifest:", err)
		}

		entries = append(entries, catalogEntry{
			ModelID:   d.modelID,
			License:   map[string]any{"spdx": d.license, "redistributable": true},
			Magnet:    magnet,
			SizeBytes: total,
			Manifest:  d.manifest,
		})
		fmt.Printf("demo %s\n  magnet: %s\n  infohash: %s\n", d.modelID, magnet, mi.InfoHash)
	}

	catalogPath := filepath.Join(webRoot, "catalog.json")
	cb, err := os.ReadFile(catalogPath)
	if err != nil {
		fatal("read catalog:", err)
	}
	var catalog map[string]any
	if err := json.Unmarshal(cb, &catalog); err != nil {
		fatal("decode catalog:", err)
	}

	models, _ := catalog["models"].([]any)
	byID := map[string]catalogEntry{}
	for _, e := range entries {
		byID[e.ModelID] = e
	}
	for i, raw := range models {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := m["modelId"].(string)
		if upd, ok := byID[id]; ok {
			m["magnet"] = upd.Magnet
			m["sizeBytes"] = upd.SizeBytes
			models[i] = m
			delete(byID, id)
		}
	}
	for _, e := range byID {
		models = append(models, map[string]any{
			"modelId":   e.ModelID,
			"license":   e.License,
			"magnet":    e.Magnet,
			"sizeBytes": e.SizeBytes,
			"manifest":  e.Manifest,
		})
	}
	catalog["models"] = models
	catalog["generatedAt"] = time.Now().UTC().Format(time.RFC3339)

	out, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		fatal("marshal catalog:", err)
	}
	if err := os.WriteFile(catalogPath, append(out, '\n'), 0o644); err != nil {
		fatal("write catalog:", err)
	}
	fmt.Println("updated", catalogPath)
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func fatal(parts ...any) {
	fmt.Fprintln(os.Stderr, parts...)
	os.Exit(1)
}
