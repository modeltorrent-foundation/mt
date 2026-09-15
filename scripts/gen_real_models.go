//go:build ignore

// gen_real_models builds real BitTorrent magnets + signed manifests for the
// first wave of real, redistributable GGUF models and merges them into
// web/catalog.json. It follows the same shape as gen_web_magnets.go /
// gen_catalog.go — it does NOT hand-hack the protocol wire format.
//
// Inputs (already downloaded, e.g. via huggingface-cli) are read from a staging
// directory (default .work/dl). Outputs:
//
//	web/models/<modelId>/manifest.json    signed (Ed25519), SPDX-gated
//	web/models/<modelId>/publish.torrent  raw metainfo (DHT + public trackers)
//	web/catalog.json                      merged catalog entries
//	web/keys/<keyId>.pub                  publisher public key (committed)
//	.work/publisher.key                   publisher private seed (gitignored)
//
// Usage:
//
//	go run ./scripts/gen_real_models.go
//
// Env:
//
//	MT_STAGING        staging dir with downloaded GGUFs (default .work/dl)
//	MT_WEBSEED_BASE   if set, per-model HTTP webseed base, e.g.
//	                  https://mirror.example.org  ->  <base>/<slug>/<file>
package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/blob"
	"github.com/modeltorrent-foundation/mt/internal/manifest"
	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

const publisherKeyID = "ed25519:modeltorrent-foundation-2026"

type realModel struct {
	modelID string // catalog id, e.g. "Qwen/Qwen3-0.6B-GGUF"
	slug    string // staging subdir under MT_STAGING
	file    string // gguf filename (single-file torrent)
	license string // SPDX id (must be on the allowlist)
}

// The first wave: small, ungated, Apache-2.0 GGUFs (SCOPE.md §"Small first").
var realModels = []realModel{
	{modelID: "Qwen/Qwen3-0.6B-GGUF", slug: "qwen3-0.6b", file: "Qwen3-0.6B-Q8_0.gguf", license: "Apache-2.0"},
	{modelID: "HuggingFaceTB/SmolLM2-360M-Instruct-GGUF", slug: "smollm2-360m", file: "smollm2-360m-instruct-q8_0.gguf", license: "Apache-2.0"},
	{modelID: "Qwen/Qwen2.5-0.5B-Instruct-GGUF", slug: "qwen2.5-0.5b", file: "qwen2.5-0.5b-instruct-q4_k_m.gguf", license: "Apache-2.0"},
}

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fatal("getwd:", err)
	}
	staging := os.Getenv("MT_STAGING")
	if staging == "" {
		staging = filepath.Join(repoRoot, ".work", "dl")
	}
	webseedBase := strings.TrimRight(os.Getenv("MT_WEBSEED_BASE"), "/")
	webRoot := filepath.Join(repoRoot, "web")

	priv := loadOrCreateKey(filepath.Join(repoRoot, ".work", "publisher.key"), webRoot)

	type catalogEntry struct {
		ModelID   string           `json:"modelId"`
		License   manifest.License `json:"license"`
		Magnet    string           `json:"magnet"`
		SizeBytes int64            `json:"sizeBytes"`
		Manifest  string           `json:"manifest"`
	}
	var entries []catalogEntry

	for _, rm := range realModels {
		if !manifest.LicenseAllowed(rm.license) {
			fatal("license not on allowlist:", rm.modelID, rm.license)
		}
		src := filepath.Join(staging, rm.slug, rm.file)
		hash, size, err := blob.HashFile(src)
		if err != nil {
			fatal("hash", src, err)
		}

		var webseeds []string
		if webseedBase != "" {
			webseeds = []string{fmt.Sprintf("%s/%s/%s", webseedBase, slugPath(rm.modelID), rm.file)}
		}

		mi, magnet, err := torrentsvc.CreateWithTrackers([]string{src}, webseeds, torrentsvc.DefaultTrackers)
		if err != nil {
			fatal("create torrent", rm.modelID, err)
		}
		if !strings.HasPrefix(magnet, "magnet:") {
			fatal("bad magnet for", rm.modelID)
		}

		modelDir := filepath.Join(webRoot, "models", filepath.FromSlash(rm.modelID))
		if err := os.MkdirAll(modelDir, 0o755); err != nil {
			fatal("mkdir", modelDir, err)
		}
		if err := os.WriteFile(filepath.Join(modelDir, "publish.torrent"), mi.Raw, 0o644); err != nil {
			fatal("write torrent", err)
		}

		m := manifest.Manifest{
			SchemaVersion: manifest.SchemaVersion,
			ModelID:       rm.modelID,
			Files:         []manifest.File{{Path: rm.file, Size: size, SHA256: hash}},
			License:       manifest.License{SPDX: rm.license, Redistributable: true},
			Publisher:     manifest.Publisher{KeyID: publisherKeyID},
			Magnet:        magnet,
			Webseeds:      webseeds,
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		}
		sig := ed25519.Sign(priv, m.Canonical())
		m.Publisher.Signature = base64.StdEncoding.EncodeToString(sig)

		// Sanity: the manifest we just signed must parse and verify.
		mb, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			fatal("marshal manifest", err)
		}
		if _, err := manifest.Parse(mb); err != nil {
			fatal("manifest self-check parse", rm.modelID, err)
		}
		if err := m.Verify(priv.Public().(ed25519.PublicKey)); err != nil {
			fatal("manifest self-check verify", rm.modelID, err)
		}
		manifestPath := filepath.Join(modelDir, "manifest.json")
		if err := os.WriteFile(manifestPath, append(mb, '\n'), 0o644); err != nil {
			fatal("write manifest", err)
		}

		rel := "models/" + slugPath(rm.modelID) + "/manifest.json"
		entries = append(entries, catalogEntry{
			ModelID:   rm.modelID,
			License:   m.License,
			Magnet:    magnet,
			SizeBytes: size,
			Manifest:  rel,
		})
		fmt.Printf("real %s\n  size:     %d\n  sha256:   %s\n  infohash: %s\n  magnet:   %s\n",
			rm.modelID, size, hash, mi.InfoHash, magnet)
	}

	// Merge into web/catalog.json, preserving existing (demo) entries.
	catalogPath := filepath.Join(webRoot, "catalog.json")
	cb, err := os.ReadFile(catalogPath)
	if err != nil {
		fatal("read catalog", err)
	}
	var cat map[string]any
	if err := json.Unmarshal(cb, &cat); err != nil {
		fatal("decode catalog", err)
	}
	models, _ := cat["models"].([]any)
	byID := map[string]catalogEntry{}
	for _, e := range entries {
		byID[e.ModelID] = e
	}
	// Update existing rows in place.
	for i, raw := range models {
		mm, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := mm["modelId"].(string)
		if upd, ok := byID[id]; ok {
			mm["magnet"] = upd.Magnet
			mm["sizeBytes"] = upd.SizeBytes
			mm["manifest"] = upd.Manifest
			mm["license"] = map[string]any{"spdx": upd.License.SPDX, "redistributable": upd.License.Redistributable}
			models[i] = mm
			delete(byID, id)
		}
	}
	// Append new rows.
	for _, e := range entries {
		if _, still := byID[e.ModelID]; !still {
			continue
		}
		models = append(models, map[string]any{
			"modelId":   e.ModelID,
			"license":   map[string]any{"spdx": e.License.SPDX, "redistributable": e.License.Redistributable},
			"magnet":    e.Magnet,
			"sizeBytes": e.SizeBytes,
			"manifest":  e.Manifest,
		})
		delete(byID, e.ModelID)
	}
	cat["models"] = models
	cat["generatedAt"] = time.Now().UTC().Format(time.RFC3339)

	out, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		fatal("marshal catalog", err)
	}
	if err := os.WriteFile(catalogPath, append(out, '\n'), 0o644); err != nil {
		fatal("write catalog", err)
	}
	fmt.Println("updated", catalogPath)
}

// slugPath keeps modelId path-separated for URLs/paths (uses forward slashes).
func slugPath(modelID string) string {
	return strings.ReplaceAll(modelID, string(os.PathSeparator), "/")
}

// loadOrCreateKey reuses a persisted Ed25519 seed (so re-runs are reproducible)
// or mints one, writing the public key into web/keys for the catalog.
func loadOrCreateKey(privPath, webRoot string) ed25519.PrivateKey {
	if b, err := os.ReadFile(privPath); err == nil {
		seed, err := hex.DecodeString(strings.TrimSpace(string(b)))
		if err == nil && len(seed) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(seed)
		}
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		fatal("keygen", err)
	}
	if err := os.MkdirAll(filepath.Dir(privPath), 0o755); err != nil {
		fatal("mkdir key dir", err)
	}
	if err := os.WriteFile(privPath, []byte(hex.EncodeToString(priv.Seed())), 0o600); err != nil {
		fatal("write priv key", err)
	}
	keysDir := filepath.Join(webRoot, "keys")
	if err := os.MkdirAll(keysDir, 0o755); err != nil {
		fatal("mkdir keys", err)
	}
	pubName := strings.ReplaceAll(publisherKeyID, ":", "_") + ".pub"
	pubDoc := map[string]any{
		"keyId":     publisherKeyID,
		"algorithm": "ed25519",
		"publicKey": base64.StdEncoding.EncodeToString(pub),
	}
	pb, _ := json.MarshalIndent(pubDoc, "", "  ")
	if err := os.WriteFile(filepath.Join(keysDir, pubName), append(pb, '\n'), 0o644); err != nil {
		fatal("write pub key", err)
	}
	return priv
}

func fatal(parts ...any) {
	fmt.Fprintln(os.Stderr, parts...)
	os.Exit(1)
}
