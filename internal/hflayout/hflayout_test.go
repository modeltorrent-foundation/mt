package hflayout_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/hflayout"
	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

// The layout must be exactly what huggingface_hub expects so that a plain
// snapshot_download / from_pretrained resolves without code changes.
func TestWriteProducesHFLayout(t *testing.T) {
	dir := t.TempDir()
	blobDir := t.TempDir()

	// stand-in downloaded blob
	blobPath := filepath.Join(blobDir, "model.safetensors")
	if err := os.WriteFile(blobPath, []byte("weights"), 0o644); err != nil {
		t.Fatalf("write blob: %v", err)
	}

	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ModelID:       "Qwen/Qwen3-8B",
		Files:         []manifest.File{{Path: "model.safetensors", Size: 7, SHA256: "deadbeef"}},
		License:       manifest.License{SPDX: "Apache-2.0", Redistributable: true},
	}

	if err := hflayout.Write(dir, m, map[string]string{"model.safetensors": blobPath}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// blob addressed by sha256
	if _, err := os.Stat(filepath.Join(dir, "blobs", "deadbeef")); err != nil {
		t.Errorf("expected blobs/deadbeef: %v", err)
	}
	// snapshot file resolves to the original bytes
	snap := filepath.Join(dir, "snapshots", "main", "model.safetensors")
	got, err := os.ReadFile(snap)
	if err != nil {
		t.Fatalf("read snapshot file: %v", err)
	}
	if string(got) != "weights" {
		t.Errorf("snapshot content = %q want %q", got, "weights")
	}
}
