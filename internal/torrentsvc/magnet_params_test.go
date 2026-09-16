package torrentsvc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

func TestCreateWithTrackersDoesNotChangeInfohash(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(src, []byte("infohash-stability-fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plain, magPlain, err := torrentsvc.Create([]string{src}, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	with, magWith, err := torrentsvc.CreateWithTrackers(
		[]string{src},
		[]string{"https://example.test/model.gguf"},
		torrentsvc.AnnounceTrackers(),
	)
	if err != nil {
		t.Fatalf("CreateWithTrackers: %v", err)
	}
	if plain.InfoHash != with.InfoHash {
		t.Fatalf("infohash changed: %s vs %s", plain.InfoHash, with.InfoHash)
	}
	btihA, btmhA := torrentsvc.MagnetInfoHashes(magPlain)
	btihB, btmhB := torrentsvc.MagnetInfoHashes(magWith)
	if btihA == "" || btmhA == "" {
		t.Fatalf("plain magnet missing hashes: %s", magPlain)
	}
	if btihA != btihB || btmhA != btmhB {
		t.Fatalf("xt hashes changed:\n  %s\n  %s", magPlain, magWith)
	}
	for _, tr := range torrentsvc.WebTorrentTrackers {
		if !strings.Contains(magWith, "wss") {
			t.Fatalf("web magnet missing WSS trackers: %s", magWith)
		}
		_ = tr
	}
	if !strings.Contains(magWith, "tracker.openwebtorrent.com") {
		t.Fatalf("missing openwebtorrent WSS tracker: %s", magWith)
	}
	if !strings.Contains(magWith, "tracker.webtorrent.dev") {
		t.Fatalf("missing webtorrent.dev WSS tracker: %s", magWith)
	}
}

func TestAppendMagnetTrackersPreservesInfohashes(t *testing.T) {
	base := "magnet:?xt=urn:btih:23bbb1ba4c30d3c844e753182fabedd528983511&xt=urn:btmh:12204bd4f72e4e957686f0bc48df4eeb128c892f75159c136333f9b9e1d4dbc92661&dn=Qwen3-0.6B-Q8_0.gguf"
	got := torrentsvc.AppendMagnetTrackers(base, torrentsvc.WebTorrentTrackers)
	got = torrentsvc.AppendMagnetTrackers(got, torrentsvc.WebTorrentTrackers) // idempotent
	btihA, btmhA := torrentsvc.MagnetInfoHashes(base)
	btihB, btmhB := torrentsvc.MagnetInfoHashes(got)
	if btihA != btihB || btmhA != btmhB {
		t.Fatalf("append changed xt: %s -> %s", base, got)
	}
	if strings.Count(got, "tracker.openwebtorrent.com") != 1 {
		t.Fatalf("expected one openwebtorrent tr, got %s", got)
	}
}
