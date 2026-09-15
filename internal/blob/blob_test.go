package blob_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/blob"
)

func write(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

func TestHashFileMatchesStdlib(t *testing.T) {
	dir := t.TempDir()
	data := []byte("model-torrent fixture bytes\n")
	p := write(t, dir, "model.safetensors", data)

	sum := sha256.Sum256(data)
	want := hex.EncodeToString(sum[:])

	got, size, err := blob.HashFile(p)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if got != want {
		t.Errorf("sha256 = %q want %q", got, want)
	}
	if size != int64(len(data)) {
		t.Errorf("size = %d want %d", size, len(data))
	}
}

// Quant rule: identical bytes at different paths share one identity.
func TestIdenticalBytesSameHash(t *testing.T) {
	dir := t.TempDir()
	data := []byte("identical gguf quant bytes")
	a := write(t, dir, "repoA_Q4_K_M.gguf", data)
	b := write(t, dir, "repoB_Q4_K_M.gguf", data)

	ha, _, err := blob.HashFile(a)
	if err != nil {
		t.Fatalf("HashFile a: %v", err)
	}
	hb, _, err := blob.HashFile(b)
	if err != nil {
		t.Fatalf("HashFile b: %v", err)
	}
	if ha != hb {
		t.Errorf("identical bytes hashed differently: %q vs %q", ha, hb)
	}
}
