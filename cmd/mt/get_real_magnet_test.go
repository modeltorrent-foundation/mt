package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/catalog"
)

// TestGetDemoTinyGGUFRealMagnet ensures the web catalog uses a real magnet and
// mt get can bootstrap the demo model via catalog-relative file webseeds.
func TestGetDemoTinyGGUFRealMagnet(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Skip(err)
	}
	catalogPath := filepath.Join(root, "web", "catalog.json")
	if _, err := os.Stat(catalogPath); err != nil {
		t.Skip("web catalog missing")
	}
	fixture := filepath.Join(root, "web", "fixtures", "demo", "tiny-gguf", "Q4_K_M.gguf")
	if _, err := os.Stat(fixture); err != nil {
		t.Skip("run make fixtures to generate web demo fixtures")
	}

	idx, _, err := catalog.LoadFile(catalogPath)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	m, ok := idx.Lookup("demo/tiny-gguf")
	if !ok {
		t.Fatal("demo/tiny-gguf not in catalog")
	}
	if strings.HasPrefix(m.Magnet, "fixture://") {
		t.Fatalf("magnet still fixture://, want real magnet")
	}
	if !strings.HasPrefix(m.Magnet, "magnet:") {
		t.Fatalf("magnet = %q want magnet: URI", m.Magnet)
	}

	t.Setenv("MT_CATALOG", catalogPath)
	dest := t.TempDir()
	if code := run([]string{"get", "demo/tiny-gguf", dest, "--no-seed"}); code != 0 {
		t.Fatalf("get exit = %d want 0", code)
	}
	got := filepath.Join(dest, "snapshots", "main", "Q4_K_M.gguf")
	if _, err := os.Stat(got); err != nil {
		t.Fatalf("missing downloaded file: %v", err)
	}
}
