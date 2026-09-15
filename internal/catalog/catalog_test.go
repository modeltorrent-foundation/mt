package catalog_test

import (
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/catalog"
	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

func mf(id, spdx string) manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ModelID:       id,
		Files:         []manifest.File{{Path: "model.safetensors", Size: 1, SHA256: "aa"}},
		License:       manifest.License{SPDX: spdx, Redistributable: true},
	}
}

func TestAddSearchHealth(t *testing.T) {
	idx := catalog.NewIndex()
	for _, m := range []manifest.Manifest{
		mf("Qwen/Qwen3-8B", "Apache-2.0"),
		mf("mistralai/Mistral-7B", "Apache-2.0"),
		mf("bigscience/bloom", "OpenRAIL-permissive"),
	} {
		if err := idx.Add(m); err != nil {
			t.Fatalf("Add %s: %v", m.ModelID, err)
		}
	}

	got := idx.Search("Qwen")
	if len(got) != 1 || got[0].ModelID != "Qwen/Qwen3-8B" {
		t.Fatalf("Search(Qwen) = %v, want single Qwen match", got)
	}

	h, err := idx.Health("Qwen/Qwen3-8B")
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if h.ModelID != "Qwen/Qwen3-8B" {
		t.Errorf("Health.ModelID = %q", h.ModelID)
	}
}

func TestAddRejectsNonRedistributable(t *testing.T) {
	idx := catalog.NewIndex()
	bad := mf("meta-llama/leaked", "LicenseRef-Proprietary")
	bad.License.Redistributable = false
	if err := idx.Add(bad); err == nil {
		t.Fatalf("Add must reject a non-redistributable manifest")
	}
}

func TestHealthUnknownModel(t *testing.T) {
	idx := catalog.NewIndex()
	if _, err := idx.Health("nope/nope"); err == nil {
		t.Fatalf("Health of unknown model must error")
	}
}
