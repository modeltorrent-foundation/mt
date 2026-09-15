package torrentsvc

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// FetchFixture copies files from a fixture:// magnet or file:// webseed into dest,
// verifying SHA-256 against want when non-empty. Used for Wave 2 dev when the full
// BitTorrent stack is not wired yet.
//
// It fails closed: a SHA-256 mismatch is corruption, so it returns a descriptive
// error rather than a soft verified=false, and verified is true only when every
// declared hash matched.
func FetchFixture(ctx context.Context, magnet string, webseeds []string, files map[string]string, dest string) (verified bool, err error) {
	srcRoot, err := fixtureRoot(magnet, webseeds)
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return false, err
	}
	for rel := range files {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		src := filepath.Join(srcRoot, rel)
		dst := filepath.Join(dest, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return false, err
		}
		if err := copyFile(src, dst); err != nil {
			return false, fmt.Errorf("torrentsvc: fixture copy %q: %w", rel, err)
		}
	}
	if err := verifyManifestHashes(dest, files); err != nil {
		return false, err
	}
	return true, nil
}

func fixtureRoot(magnet string, webseeds []string) (string, error) {
	if strings.HasPrefix(magnet, "fixture://") {
		// fixture://Qwen/Qwen3-8B — caller must pass webseeds with file:// root
		for _, ws := range webseeds {
			if root, ok := fileURLPath(ws); ok {
				return root, nil
			}
		}
	}
	for _, ws := range webseeds {
		if root, ok := fileURLPath(ws); ok {
			return root, nil
		}
	}
	return "", fmt.Errorf("torrentsvc: no file:// fixture root in magnet/webseeds")
}

func fileURLPath(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "file" {
		return "", false
	}
	if u.Path != "" && u.Path != "/" {
		return u.Path, true
	}
	if u.Host != "" {
		return filepath.Join("/", u.Host), true
	}
	return "", false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
