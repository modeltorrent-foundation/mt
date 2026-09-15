package torrentsvc_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

// TestFetchWithManifestChecksManifestHashes covers the real BitTorrent path
// (real magnet, no file:// webseed). Piece verification only proves the payload
// matches the infohash it was fetched with, so FetchWithManifest must also check
// the manifest SHA-256s and must never report verified for a mismatch.
func TestFetchWithManifestChecksManifestHashes(t *testing.T) {
	dir := t.TempDir()
	content := bytes.Repeat([]byte("manifest-verify-"), 256) // ~4 KB stand-in
	src := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	sum := sha256.Sum256(content)
	goodSHA := hex.EncodeToString(sum[:])

	mi, magnet, err := torrentsvc.Create([]string{src}, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := torrentsvc.Seed(ctx, mi); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	t.Run("matching hash verifies", func(t *testing.T) {
		verified, err := torrentsvc.FetchWithManifest(ctx, magnet, nil,
			map[string]string{"model.gguf": goodSHA}, t.TempDir())
		if err != nil {
			t.Fatalf("FetchWithManifest: %v", err)
		}
		if !verified {
			t.Fatal("verified = false for a matching manifest hash")
		}
	})

	t.Run("mismatched hash fails closed", func(t *testing.T) {
		wrongSHA := hex.EncodeToString(bytes.Repeat([]byte{0xab}, sha256.Size))
		verified, err := torrentsvc.FetchWithManifest(ctx, magnet, nil,
			map[string]string{"model.gguf": wrongSHA}, t.TempDir())
		if verified {
			t.Fatal("verified = true for a manifest hash mismatch")
		}
		if err == nil {
			t.Fatal("err = nil for a manifest hash mismatch, want a descriptive error")
		}
	})

	t.Run("file missing from download fails closed", func(t *testing.T) {
		verified, err := torrentsvc.FetchWithManifest(ctx, magnet, nil,
			map[string]string{"model.gguf": goodSHA, "tokenizer.json": goodSHA}, t.TempDir())
		if verified {
			t.Fatal("verified = true for a manifest file absent from the download")
		}
		if err == nil {
			t.Fatal("err = nil for a missing manifest file, want a descriptive error")
		}
	})
}

// TestFetchFixtureFailsClosedOnMismatch pins the fixture path: a hash mismatch is
// corruption and must surface as an error, not a soft verified=false.
func TestFetchFixtureFailsClosedOnMismatch(t *testing.T) {
	srcRoot := t.TempDir()
	content := bytes.Repeat([]byte("fixture-"), 128)
	if err := os.WriteFile(filepath.Join(srcRoot, "model.gguf"), content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	sum := sha256.Sum256(content)
	goodSHA := hex.EncodeToString(sum[:])
	webseeds := []string{"file://" + srcRoot}

	ctx := context.Background()

	verified, err := torrentsvc.FetchFixture(ctx, "fixture://demo/x", webseeds,
		map[string]string{"model.gguf": goodSHA}, t.TempDir())
	if err != nil || !verified {
		t.Fatalf("valid fixture: verified = %v, err = %v; want true, nil", verified, err)
	}

	wrongSHA := hex.EncodeToString(bytes.Repeat([]byte{0x11}, sha256.Size))
	verified, err = torrentsvc.FetchFixture(ctx, "fixture://demo/x", webseeds,
		map[string]string{"model.gguf": wrongSHA}, t.TempDir())
	if verified {
		t.Fatal("verified = true for a corrupt fixture")
	}
	if err == nil {
		t.Fatal("err = nil for a corrupt fixture, want a descriptive error")
	}
}
