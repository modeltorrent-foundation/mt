// Package blob is the content-addressing layer: SHA-256 file hashing and the
// on-disk layout mapping. Identity is derived from bytes (PROTOCOL.md §1), so
// identical file bytes always hash equal ("hash the file, not the folder").
package blob

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// HashFile returns the lowercase hex SHA-256 and byte size of the file at path.
// Two files with identical bytes MUST return the same hash (quant rule).
func HashFile(path string) (sha256hex string, size int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// Layout maps a modelId and a file's relative path to its on-disk location,
// rooted at Root. Used by the CLI cache and by hflayout.
type Layout struct {
	Root string
}

// Path returns the absolute on-disk path for one file of a model.
func (l Layout) Path(modelID, file string) string {
	return filepath.Join(l.Root, modelID, file)
}
