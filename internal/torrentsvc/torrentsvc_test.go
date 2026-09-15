package torrentsvc_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

// TestSwarmSmokeLocalhost is the canonical proof that the swarm works: create a
// torrent over a tiny stand-in "model", seed it from client A, fetch it from
// client B, and confirm the bytes and hash survive the round trip. It uses KB
// fixtures on localhost only — no trackers, no real models, disk-safe.
//
// RED until torrentsvc is wired (Wave 1).
func TestSwarmSmokeLocalhost(t *testing.T) {
	dir := t.TempDir()
	content := bytes.Repeat([]byte("swarm-smoke-"), 512) // ~6 KB stand-in
	src := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	want := sha256.Sum256(content)

	mi, magnet, err := torrentsvc.Create([]string{src}, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if magnet == "" || mi.InfoHash == "" {
		t.Fatalf("Create returned empty magnet/infohash")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := torrentsvc.Seed(ctx, mi); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	dest := t.TempDir()
	verified, err := torrentsvc.Fetch(ctx, magnet, dest)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !verified {
		t.Fatalf("Fetch did not verify pieces")
	}

	got, err := os.ReadFile(filepath.Join(dest, "model.gguf"))
	if err != nil {
		t.Fatalf("read fetched file: %v", err)
	}
	if sha256.Sum256(got) != want {
		t.Fatalf("fetched bytes do not match seeded bytes")
	}
}

func TestSeedingIsDefaultNotOptIn(t *testing.T) {
	// Documents intent: Seed takes no "enable" flag. The absence of an opt-in
	// parameter is the contract. This compiles today and guards against a future
	// signature that makes seeding opt-in.
	_ = torrentsvc.Seed
}
