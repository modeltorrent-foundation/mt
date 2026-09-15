package torrentsvc

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/modeltorrent-foundation/mt/internal/blob"
)

// verifyManifestHashes re-hashes every manifest file under dest and compares it
// to the SHA-256 claimed by the manifest. This is deliberately independent of
// BitTorrent piece verification: pieces only prove the payload matches the
// infohash it was fetched with, while the manifest hashes are the signed
// integrity claim (PROTOCOL.md §1).
//
// It fails closed — a missing, unreadable, or mismatched file is an error, and a
// nil return means every file with a declared hash is present and matching.
// Entries with an empty want hash carry no claim and are skipped.
func verifyManifestHashes(dest string, files map[string]string) error {
	mismatched := make([]string, 0, len(files))
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		want := files[rel]
		if want == "" {
			continue
		}
		got, _, err := blob.HashFile(filepath.Join(dest, rel))
		if err != nil {
			return fmt.Errorf("torrentsvc: verify %q: %w", rel, err)
		}
		if got != want {
			mismatched = append(mismatched, fmt.Sprintf("%s (want %s, got %s)", rel, want, got))
		}
	}
	if len(mismatched) > 0 {
		return fmt.Errorf("torrentsvc: manifest sha256 mismatch: %s", strings.Join(mismatched, "; "))
	}
	return nil
}
