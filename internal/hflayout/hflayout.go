// Package hflayout writes a Hugging Face-style cache layout so existing tools
// (huggingface_hub, transformers, llama.cpp) find files at the paths they
// already expect. This is the compatibility wedge: no ecosystem rewrite.
//
// Target layout (mirrors huggingface_hub's snapshot cache):
//
//	<dest>/blobs/<sha256>                      # content-addressed blob
//	<dest>/snapshots/<revision>/<file path>    # symlink -> ../../blobs/<sha256>
//	<dest>/refs/main                           # revision id
package hflayout

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

const revision = "main"

// Write materializes a Hugging Face-style snapshot for manifest m under dest.
// blobs maps each manifest file path to the local path of its downloaded bytes.
func Write(dest string, m manifest.Manifest, blobs map[string]string) error {
	if err := os.MkdirAll(filepath.Join(dest, "blobs"), 0o755); err != nil {
		return fmt.Errorf("hflayout: mkdir blobs: %w", err)
	}
	snapRoot := filepath.Join(dest, "snapshots", revision)
	if err := os.MkdirAll(snapRoot, 0o755); err != nil {
		return fmt.Errorf("hflayout: mkdir snapshots: %w", err)
	}

	for _, f := range m.Files {
		src, ok := blobs[f.Path]
		if !ok {
			return fmt.Errorf("hflayout: missing blob for %q", f.Path)
		}

		blobDest := filepath.Join(dest, "blobs", f.SHA256)
		if err := copyFile(src, blobDest); err != nil {
			return fmt.Errorf("hflayout: copy blob %q: %w", f.Path, err)
		}

		snapPath := filepath.Join(snapRoot, f.Path)
		if err := os.MkdirAll(filepath.Dir(snapPath), 0o755); err != nil {
			return fmt.Errorf("hflayout: mkdir snapshot dir: %w", err)
		}

		relBlob, err := filepath.Rel(filepath.Dir(snapPath), blobDest)
		if err != nil {
			return fmt.Errorf("hflayout: rel blob path: %w", err)
		}
		_ = os.Remove(snapPath)
		if err := os.Symlink(relBlob, snapPath); err != nil {
			if err := copyFile(blobDest, snapPath); err != nil {
				return fmt.Errorf("hflayout: link/copy snapshot %q: %w", f.Path, err)
			}
		}
	}

	refPath := filepath.Join(dest, "refs", "main")
	if err := os.MkdirAll(filepath.Dir(refPath), 0o755); err != nil {
		return fmt.Errorf("hflayout: mkdir refs: %w", err)
	}
	if err := os.WriteFile(refPath, []byte(revision), 0o644); err != nil {
		return fmt.Errorf("hflayout: write refs/main: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
