//go:build ignore

// append_wss_trackers adds WebTorrent WSS tracker params (and demo HTTP
// webseeds) to existing catalog magnets and manifests without rebuilding
// torrents. Infohashes (btih/btmh) are left unchanged. Signed real-model
// manifests are re-signed with .work/publisher.key when present.
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

	"github.com/modeltorrent-foundation/mt/internal/manifest"
	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

const r2PublicBase = "https://pub-60d277f41b154e4f826375a17187b003.r2.dev"

func main() {
	webRoot, err := filepath.Abs("web")
	if err != nil {
		fatal(err)
	}
	priv := loadPublisherKey(filepath.Join(".work", "publisher.key"))

	catalogPath := filepath.Join(webRoot, "catalog.json")
	cb, err := os.ReadFile(catalogPath)
	if err != nil {
		fatal(err)
	}
	var cat map[string]any
	if err := json.Unmarshal(cb, &cat); err != nil {
		fatal(err)
	}
	models, _ := cat["models"].([]any)
	for i, raw := range models {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id, _ := m["modelId"].(string)
		magnet, _ := m["magnet"].(string)
		if magnet == "" {
			continue
		}
		btih0, btmh0 := torrentsvc.MagnetInfoHashes(magnet)
		magnet = torrentsvc.AppendMagnetTrackers(magnet, torrentsvc.WebTorrentTrackers)
		extraWS := demoHTTPWebseeds(id)
		magnet = torrentsvc.AppendMagnetWebseeds(magnet, extraWS)
		btih1, btmh1 := torrentsvc.MagnetInfoHashes(magnet)
		if btih0 != btih1 || btmh0 != btmh1 {
			fatal("infohash changed for", id)
		}
		m["magnet"] = magnet
		models[i] = m

		manifestRel, _ := m["manifest"].(string)
		if manifestRel == "" {
			continue
		}
		if err := updateManifest(filepath.Join(webRoot, manifestRel), magnet, extraWS, priv); err != nil {
			fatal(id, err)
		}
		fmt.Printf("updated %s\n", id)
	}
	cat["models"] = models
	out, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(catalogPath, append(out, '\n'), 0o644); err != nil {
		fatal(err)
	}
}

func demoHTTPWebseeds(modelID string) []string {
	switch modelID {
	case "demo/tiny-gguf":
		return []string{r2PublicBase + "/demo/tiny-gguf/Q4_K_M.gguf"}
	case "demo/tiny-safetensors":
		return []string{r2PublicBase + "/demo/tiny-safetensors/"}
	case "demo/tiny-bundle":
		return []string{r2PublicBase + "/demo/tiny-bundle/"}
	default:
		return nil
	}
}

func updateManifest(path, magnet string, extraWS []string, priv ed25519.PrivateKey) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	m, err := manifest.Parse(b)
	if err != nil {
		return err
	}
	m.Magnet = magnet
	m.Webseeds = mergeWebseeds(m.Webseeds, extraWS)
	if m.Publisher.Signature != "" {
		if len(priv) == 0 {
			return fmt.Errorf("signed manifest %s needs .work/publisher.key to re-sign", path)
		}
		m.Publisher.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(priv, m.Canonical()))
		if err := m.Verify(priv.Public().(ed25519.PublicKey)); err != nil {
			return err
		}
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func mergeWebseeds(existing, extra []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range append(append([]string{}, existing...), extra...) {
		if w == "" || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

func loadPublisherKey(path string) ed25519.PrivateKey {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(b)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil
	}
	return ed25519.NewKeyFromSeed(seed)
}

func fatal(parts ...any) {
	fmt.Fprintln(os.Stderr, parts...)
	os.Exit(1)
}
