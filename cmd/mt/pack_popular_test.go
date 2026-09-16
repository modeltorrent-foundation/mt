package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/catalog"
)

func TestPackPopularSeedsRealCatalogGGUFs(t *testing.T) {
	want := []string{
		"Qwen/Qwen3-0.6B-GGUF",
		"HuggingFaceTB/SmolLM2-360M-Instruct-GGUF",
		"Qwen/Qwen2.5-0.5B-Instruct-GGUF",
		"Qwen/Qwen3-8B-GGUF",
	}
	if len(packPopular) != len(want) {
		t.Fatalf("packPopular len=%d want %d", len(packPopular), len(want))
	}
	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	_, file, err := catalog.LoadFile(filepath.Join(root, "web", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]catalog.Entry, len(file.Models))
	for _, e := range file.Models {
		byID[e.ModelID] = e
	}
	for i, e := range packPopular {
		if e.modelID != want[i] {
			t.Errorf("packPopular[%d]=%s want %s", i, e.modelID, want[i])
		}
		if e.catalogPath != "web/catalog.json" {
			t.Errorf("%s catalogPath=%s want web/catalog.json", e.modelID, e.catalogPath)
		}
		if strings.HasPrefix(e.modelID, "demo/") {
			t.Errorf("popular pack must not include demo fixtures: %s", e.modelID)
		}
		got, ok := byID[e.modelID]
		if !ok {
			t.Errorf("missing from web catalog: %s", e.modelID)
			continue
		}
		if got.License.SPDX != "Apache-2.0" || !got.License.Redistributable {
			t.Errorf("%s license=%+v want Apache-2.0 redistributable", e.modelID, got.License)
		}
	}
}
