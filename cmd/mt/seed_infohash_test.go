package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	infohash_v2 "github.com/anacrolix/torrent/types/infohash-v2"
	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

func TestSeedStagingMatchesPublishedTorrent(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Skip(err)
	}
	fix := filepath.Join(root, "web/fixtures/demo/tiny-safetensors")
	stagingParent := t.TempDir()
	hfRoot := filepath.Join(stagingParent, "model")
	snapMain := filepath.Join(hfRoot, "snapshots", "main")
	blobDir := filepath.Join(hfRoot, "blobs")
	if err := os.MkdirAll(snapMain, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"model.safetensors", "tokenizer.json"} {
		src := filepath.Join(fix, name)
		blob := filepath.Join(blobDir, name+"-blob")
		if err := os.Link(src, blob); err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(snapMain, name)
		if err := os.Symlink(filepath.Join("..", "..", "blobs", name+"-blob"), dst); err != nil {
			t.Fatal(err)
		}
	}
	seedDir, err := prepareSeedDir(snapMain)
	if err != nil {
		t.Fatal(err)
	}
	miStaging, err := torrentsvc.MetainfoForDownload(seedDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	pubPath := filepath.Join(root, "web/models/demo/tiny-safetensors/publish.torrent")
	pubHash, err := infoHashFromTorrent(pubPath)
	if err != nil {
		t.Fatal(err)
	}
	if miStaging.InfoHash != pubHash {
		t.Fatalf("staging infohash %s != publish.torrent %s (seedDir=%s)", miStaging.InfoHash, pubHash, seedDir)
	}
}

func infoHashFromTorrent(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var mi metainfo.MetaInfo
	if err := bencode.Unmarshal(b, &mi); err != nil {
		return "", err
	}
	return infohashHex(infohash_v2.HashBytes(mi.InfoBytes)), nil
}

func infohashHex(h infohash_v2.T) string {
	return hexString(h[:])
}

func hexString(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}
