// Package torrentsvc is the BitTorrent layer: create a torrent over a fileset,
// seed it (on by default), and fetch+verify from a magnet.
//
// Wave 0 is dependency-free and defines its own Metainfo so the red wall runs
// offline. Wave 1 backs these bodies with github.com/anacrolix/torrent v1.60.0
// (WebSeeds/BEP-19, BitTorrent v2, DHT, WebTorrent). The signatures here are the
// contract Wave 1 must satisfy; do not change them to fit the library.
package torrentsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrNotImplemented marks a Wave 0 stub awaiting a later wave.
var ErrNotImplemented = errors.New("torrentsvc: not implemented")

// Metainfo is a minimal, library-agnostic torrent descriptor. Wave 1 may embed
// or wrap anacrolix's metainfo, but this shape is what the rest of the system
// depends on.
type Metainfo struct {
	InfoHash string   // BitTorrent v2 infohash, lowercase hex
	Name     string   // display name (usually modelId)
	Files    []string // file paths in the torrent, relative to the content root
	Webseeds []string // BEP-19 url-list fallbacks
	Raw      []byte   // bencoded .torrent bytes (Wave 1)
}

// Create builds a BitTorrent v2 torrent over the given files and returns the
// metainfo plus a magnet URI. Identical file bytes MUST produce a stable
// per-file identity (quant rule, PROTOCOL.md §1).
func Create(files []string, webseeds []string) (Metainfo, string, error) {
	return createBitTorrent(files, webseeds)
}

// Seed announces and serves a torrent. Seeding is ON by default: callers do not
// opt in, they may only opt out. This is the product (auto-seed), not a feature.
func Seed(ctx context.Context, mi Metainfo) error {
	if len(mi.Raw) == 0 {
		// Fixture/dev entries without metainfo bytes are no-ops.
		if len(mi.Webseeds) > 0 && hasFileWebseed(mi.Webseeds) {
			return nil
		}
		return ErrNotImplemented
	}
	return seedBitTorrent(ctx, mi, seedDataDir(mi), isOfflineTest())
}

// Fetch downloads the content addressed by magnet into dest, verifying every
// piece. It returns verified=true only when all files match their hashes.
func Fetch(ctx context.Context, magnet, dest string) (verified bool, err error) {
	if strings.HasPrefix(magnet, "fixture://") {
		return false, fmt.Errorf("torrentsvc: Fetch(fixture) requires FetchWithManifest")
	}
	return fetchBitTorrent(ctx, magnet, dest, isOfflineTest())
}

// FetchWithManifest is the catalog-aware fetch entry point. It tries the full
// BitTorrent path first; on ErrNotImplemented it falls back to fixture copy.
//
// It fails closed: verified is true only when the transport succeeded AND every
// file declared in files matches its manifest SHA-256 on disk. Piece
// verification alone is not sufficient — it proves only that the bytes match the
// infohash they were fetched with, not that they match the signed manifest.
func FetchWithManifest(ctx context.Context, magnet string, webseeds []string, files map[string]string, dest string) (verified bool, err error) {
	if strings.HasPrefix(magnet, "fixture://") || hasFileWebseed(webseeds) {
		return FetchFixture(ctx, magnet, webseeds, files, dest)
	}
	verified, err = Fetch(ctx, magnet, dest)
	if err != nil {
		return false, err
	}
	if !verified {
		return false, fmt.Errorf("torrentsvc: piece verification failed for %q", magnet)
	}
	if err := verifyManifestHashes(dest, files); err != nil {
		return false, err
	}
	return true, nil
}

func hasFileWebseed(webseeds []string) bool {
	for _, ws := range webseeds {
		if strings.HasPrefix(ws, "file://") {
			return true
		}
	}
	return false
}
