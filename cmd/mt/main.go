// Command mt is the Model Torrent CLI. Seeding is on by default — get leaves
// the torrent seeding unless --no-seed is passed.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modeltorrent-foundation/mt/internal/catalog"
	"github.com/modeltorrent-foundation/mt/internal/hflayout"
	"github.com/modeltorrent-foundation/mt/internal/manifest"
	"github.com/modeltorrent-foundation/mt/internal/torrentsvc"
)

const usage = `mt — Model Torrent

Usage:
  mt get <org/name> [dest]   Download a model over BitTorrent and keep seeding it
  mt seed <magnet|path>      Seed a model you already have
  mt pack <name>             Build/seed a bundle of popular models (seed-pack)
  mt catalog [--port N]      Serve the dev catalog index (default :8765)

Seeding is ON by default after get. Pass --no-seed to stop after download.

Pack bundles:
  dev       Seed the testdata fixture bundle (Qwen/Qwen3-8B)
  popular   Seed bundled small web demo models plus the Qwen fixture; hook for
            curated real GGUFs later (currently offline-sized stand-ins only)

Environment:
  MT_CATALOG    Path to catalog.json (default: search upward for web/catalog.json)
  MT_ROOT       Repo root for fixture paths`

// run returns a process exit code so it is testable.
func run(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "get":
		return cmdGet(args[1:])
	case "seed":
		return cmdSeed(args[1:])
	case "pack":
		return cmdPack(args[1:])
	case "catalog":
		return cmdCatalog(args[1:])
	case "-h", "--help", "help":
		fmt.Println(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "mt: unknown command %q\n\n%s\n", args[0], usage)
		return 2
	}
}

func main() {
	os.Exit(run(os.Args[1:]))
}

type getOptions struct {
	noSeed      bool
	catalogPath string
	blockSeed   bool
}

func cmdGet(args []string) int {
	opts := getOptions{blockSeed: true}
	var rest []string
	for _, a := range args {
		if a == "--no-seed" {
			opts.noSeed = true
			continue
		}
		rest = append(rest, a)
	}
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "mt get: requires <org/name> [dest]")
		return 2
	}
	modelID := rest[0]
	dest := defaultDest(modelID)
	if len(rest) >= 2 {
		dest = rest[1]
	}
	return doGet(modelID, dest, opts)
}

func doGet(modelID, dest string, opts getOptions) int {
	catalogPath, err := resolveCatalogPath(opts.catalogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt get: %v\n", err)
		return 1
	}
	idx, _, err := catalog.LoadFile(catalogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt get: catalog: %v\n", err)
		return 1
	}
	m, ok := idx.Lookup(modelID)
	if !ok {
		fmt.Fprintf(os.Stderr, "mt get: model %q not in catalog\n", modelID)
		return 1
	}

	downloadDir, err := os.MkdirTemp("", "mt-get-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt get: temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(downloadDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	catalogRoot := filepath.Dir(catalogPath)
	fileHashes := manifestFileHashes(m)
	verified, err := torrentsvc.FetchWithManifest(ctx, m.Magnet, resolveWebseeds(catalogRoot, m.Webseeds), fileHashes, downloadDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt get: fetch: %v\n", err)
		if err == torrentsvc.ErrNotImplemented {
			fmt.Fprintln(os.Stderr, "mt get: BitTorrent fetch not wired yet; use fixture catalog for dev")
		}
		fmt.Fprintf(os.Stderr, "mt get: integrity not established for %s; nothing written to %s, not seeding\n", modelID, dest)
		return 1
	}
	// Fail closed: unverified bytes are never written to the HF layout or seeded.
	if !verified {
		fmt.Fprintf(os.Stderr, "mt get: checksum verification failed for %s; nothing written to %s, not seeding\n", modelID, dest)
		return 1
	}

	blobs := make(map[string]string, len(m.Files))
	for _, f := range m.Files {
		blobs[f.Path] = filepath.Join(downloadDir, f.Path)
	}
	if err := hflayout.Write(dest, m, blobs); err != nil {
		fmt.Fprintf(os.Stderr, "mt get: layout: %v\n", err)
		return 1
	}

	fmt.Printf("wrote Hugging Face layout to %s\n", dest)
	if filepath.Base(dest) == "main" || strings.Contains(dest, "snapshots") {
		fmt.Println(dest)
	} else {
		fmt.Println(filepath.Join(dest, "snapshots", "main"))
	}

	if opts.noSeed {
		fmt.Println("seeding disabled (--no-seed)")
		return 0
	}

	seedDir := filepath.Join(dest, "snapshots", "main")
	if err := startSeed(context.Background(), m, seedDir); err != nil {
		fmt.Fprintf(os.Stderr, "mt get: seed: %v (download succeeded; seeding deferred)\n", err)
		return 0
	}
	if strings.HasPrefix(m.Magnet, "fixture://") {
		fmt.Println("seed registered (fixture dev mode)")
		return 0
	}
	if opts.blockSeed {
		fmt.Println("seeding (press Ctrl+C to stop)")
		select {}
	}
	return 0
}

func cmdSeed(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "mt seed: requires <magnet|path>")
		return 2
	}
	target := args[0]
	ctx := context.Background()

	if strings.HasPrefix(target, "fixture://") || strings.HasPrefix(target, "magnet:") {
		catalogPath, err := resolveCatalogPath("")
		if err == nil {
			if idx, _, err := catalog.LoadFile(catalogPath); err == nil {
				var m manifest.Manifest
				var ok bool
				if strings.HasPrefix(target, "fixture://") {
					modelID := strings.TrimPrefix(target, "fixture://")
					m, ok = idx.Lookup(modelID)
				} else {
					m, ok = idx.LookupByMagnet(target)
				}
				if ok {
					if strings.HasPrefix(m.Magnet, "fixture://") {
						if err := startSeed(ctx, m, ""); err != nil {
							fmt.Fprintf(os.Stderr, "mt seed: %v\n", err)
							return 1
						}
						fmt.Printf("seed registered for %s (fixture dev mode)\n", m.ModelID)
						return 0
					}
					fmt.Fprintf(os.Stderr, "mt seed: catalog entry %q requires a local HF-layout path to seed real magnets\n", m.ModelID)
					return 1
				}
			}
		}
	}

	if st, err := os.Stat(target); err == nil && st.IsDir() {
		dataDir, err := resolveSeedDataDir(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mt seed: %v\n", err)
			return 1
		}
		webseeds := webseedsForSeedPath(target)
		if err := seedFromPath(ctx, dataDir, webseeds); err != nil {
			fmt.Fprintf(os.Stderr, "mt seed: %v\n", err)
			return 1
		}
		fmt.Println("seeding (press Ctrl+C to stop)")
		select {}
	}

	fmt.Fprintf(os.Stderr, "mt seed: %q: not found or not a directory\n", target)
	return 1
}

type packEntry struct {
	modelID     string
	catalogPath string
}

var packDev = []packEntry{
	{modelID: "Qwen/Qwen3-8B", catalogPath: "testdata/catalog.json"},
}

var packPopular = []packEntry{
	{modelID: "demo/tiny-gguf", catalogPath: "web/catalog.json"},
	{modelID: "demo/tiny-safetensors", catalogPath: "web/catalog.json"},
	{modelID: "demo/tiny-bundle", catalogPath: "web/catalog.json"},
	{modelID: "Qwen/Qwen3-8B", catalogPath: "testdata/catalog.json"},
}

func cmdPack(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "mt pack: requires <name>  (try: mt pack dev | mt pack popular)")
		return 2
	}
	name := args[0]
	switch name {
	case "dev", "fixtures", "test":
		return runPack("dev", packDev)
	case "popular":
		return runPack("popular", packPopular)
	default:
		fmt.Fprintf(os.Stderr, "mt pack: unknown bundle %q\n", name)
		fmt.Fprintln(os.Stderr, "Available: dev (fixture bundle), popular (bundled small web demos + Qwen fixture)")
		return 2
	}
}

func runPack(label string, entries []packEntry) int {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt pack %s: %v\n", label, err)
		return 1
	}
	fmt.Printf("seed-pack (%s): downloading and seeding bundled models:\n", label)
	var seeded int
	for _, e := range entries {
		fmt.Printf("  - %s\n", e.modelID)
		catalogPath := e.catalogPath
		if !filepath.IsAbs(catalogPath) {
			catalogPath = filepath.Join(root, catalogPath)
		}
		dest := filepath.Join(os.TempDir(), "mt-pack-"+strings.ReplaceAll(e.modelID, "/", "_"))
		code := doGet(e.modelID, dest, getOptions{
			catalogPath: catalogPath,
			blockSeed:   false,
		})
		if code != 0 {
			fmt.Fprintf(os.Stderr, "mt pack: get %s failed (exit %d)\n", e.modelID, code)
			continue
		}
		seeded++
	}
	if seeded == 0 {
		fmt.Fprintf(os.Stderr, "mt pack %s: no models seeded\n", label)
		return 1
	}
	fmt.Printf("seed-pack (%s): seeding %d model(s) (press Ctrl+C to stop)\n", label, seeded)
	select {}
}

func cmdCatalog(args []string) int {
	fs := flag.NewFlagSet("catalog", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	port := fs.String("port", "8765", "HTTP port to serve catalog static files")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	catalogPath, err := resolveCatalogPath("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mt catalog: %v\n", err)
		return 1
	}
	root := filepath.Dir(catalogPath)
	addr := ":" + *port
	fmt.Printf("serving catalog from %s at http://localhost%s/catalog.json\n", root, addr)
	fmt.Println("press Ctrl+C to stop")
	handler := http.FileServer(http.Dir(root))
	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "mt catalog: %v\n", err)
		return 1
	}
	return 0
}

func resolveCatalogPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if p := os.Getenv("MT_CATALOG"); p != "" {
		return p, nil
	}
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if root := os.Getenv("MT_ROOT"); root != "" {
		start = root
	}
	return catalog.FindCatalog(start)
}

func defaultDest(modelID string) string {
	safe := strings.ReplaceAll(modelID, "/", string(os.PathSeparator))
	return filepath.Join(".", "downloads", safe)
}

// resolveWebseeds turns catalog-relative file:// paths (file://fixtures/…)
// into absolute file URLs so manifests stay portable in git.
func resolveWebseeds(catalogRoot string, seeds []string) []string {
	if len(seeds) == 0 {
		return seeds
	}
	root, err := filepath.Abs(catalogRoot)
	if err != nil {
		root = catalogRoot
	}
	out := make([]string, len(seeds))
	for i, ws := range seeds {
		if strings.HasPrefix(ws, "file://fixtures/") {
			rel := strings.TrimPrefix(ws, "file://")
			abs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				abs = filepath.Join(root, filepath.FromSlash(rel))
			}
			out[i] = "file://" + abs
			continue
		}
		out[i] = ws
	}
	return out
}

func manifestFileHashes(m manifest.Manifest) map[string]string {
	out := make(map[string]string, len(m.Files))
	for _, f := range m.Files {
		out[f.Path] = f.SHA256
	}
	return out
}

func startSeed(ctx context.Context, m manifest.Manifest, dataDir string) error {
	if strings.HasPrefix(m.Magnet, "fixture://") {
		mi := torrentsvc.Metainfo{
			Name:     m.ModelID,
			Webseeds: m.Webseeds,
		}
		return torrentsvc.Seed(ctx, mi)
	}
	if dataDir == "" {
		return fmt.Errorf("seed: missing download directory for %s", m.ModelID)
	}
	return seedFromPath(ctx, dataDir, m.Webseeds)
}

func seedFromPath(ctx context.Context, dataDir string, webseeds []string) error {
	seedDir, err := prepareSeedDir(dataDir)
	if err != nil {
		return err
	}
	mi, err := torrentsvc.MetainfoForDownload(seedDir, webseeds)
	if err != nil {
		return err
	}
	return torrentsvc.Seed(ctx, mi)
}

// prepareSeedDir returns a directory whose files have real sizes for BitTorrent.
// Hugging Face layouts store snapshot paths as symlinks into blobs/; anacrolix
// mis-reads symlink metadata, so we hard-link blob bytes into .mt-seed/<name>.
// Multi-file torrent names must match the published fixture folder (e.g.
// tiny-safetensors), not snapshots/main, or the infohash will differ.
func prepareSeedDir(dataDir string) (string, error) {
	if hfRoot := findHFRoot(dataDir); hfRoot != "" {
		name := torrentNameForDir(dataDir)
		return ensureSeedStaging(hfRoot, name)
	}
	return dataDir, nil
}

func torrentNameForDir(dataDir string) string {
	if m, ok := manifestForDir(dataDir); ok {
		return filepath.Base(m.ModelID)
	}
	return filepath.Base(dataDir)
}

func findHFRoot(dataDir string) string {
	if filepath.Base(dataDir) != "main" || filepath.Base(filepath.Dir(dataDir)) != "snapshots" {
		return ""
	}
	hfRoot := filepath.Dir(filepath.Dir(dataDir))
	if st, err := os.Stat(filepath.Join(hfRoot, "blobs")); err == nil && st.IsDir() {
		return hfRoot
	}
	return ""
}

func ensureSeedStaging(hfRoot, torrentName string) (string, error) {
	snapMain := filepath.Join(hfRoot, "snapshots", "main")
	staging := filepath.Join(hfRoot, ".mt-seed", torrentName)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return "", err
	}
	for _, name := range []string{".torrent.db", ".torrent.db-shm", ".torrent.db-wal"} {
		_ = os.Remove(filepath.Join(staging, name))
	}
	entries, err := os.ReadDir(snapMain)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		src := filepath.Join(snapMain, e.Name())
		dst := filepath.Join(staging, e.Name())
		target, err := filepath.EvalSymlinks(src)
		if err != nil {
			return "", fmt.Errorf("seed staging: %s: %w", e.Name(), err)
		}
		_ = os.Remove(dst)
		if err := os.Link(target, dst); err != nil {
			if err := copySeedFile(target, dst); err != nil {
				return "", fmt.Errorf("seed staging: %s: %w", e.Name(), err)
			}
		}
	}
	return staging, nil
}

func copySeedFile(src, dst string) error {
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
	_, err = io.Copy(out, in)
	return err
}

func resolveSeedDataDir(path string) (string, error) {
	path = filepath.Clean(path)
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("not a directory: %s", path)
	}
	if filepath.Base(path) == "main" && filepath.Base(filepath.Dir(path)) == "snapshots" {
		return path, nil
	}
	snapMain := filepath.Join(path, "snapshots", "main")
	if st, err := os.Stat(snapMain); err == nil && st.IsDir() {
		return snapMain, nil
	}
	return path, nil
}

func webseedsForSeedPath(path string) []string {
	if m, ok := manifestForDir(path); ok {
		root, err := findRepoRoot()
		if err != nil {
			return nil
		}
		for _, rel := range []string{"web/catalog.json", "testdata/catalog.json"} {
			catalogPath := filepath.Join(root, rel)
			idx, _, err := catalog.LoadFile(catalogPath)
			if err != nil {
				continue
			}
			if _, ok := idx.Lookup(m.ModelID); ok {
				return resolveWebseeds(filepath.Dir(catalogPath), m.Webseeds)
			}
		}
	}
	return nil
}

func manifestForDir(path string) (manifest.Manifest, bool) {
	seedDir, err := resolveSeedDataDir(path)
	if err != nil {
		return manifest.Manifest{}, false
	}
	root, err := findRepoRoot()
	if err != nil {
		return manifest.Manifest{}, false
	}
	var best manifest.Manifest
	bestFiles := 0
	for _, rel := range []string{"web/catalog.json", "testdata/catalog.json"} {
		catalogPath := filepath.Join(root, rel)
		idx, _, err := catalog.LoadFile(catalogPath)
		if err != nil {
			continue
		}
		for _, m := range idx.All() {
			if modelFilesMatch(seedDir, m) && len(m.Files) > bestFiles {
				best = m
				bestFiles = len(m.Files)
			}
		}
	}
	return best, bestFiles > 0
}

func modelFilesMatch(dataDir string, m manifest.Manifest) bool {
	for _, f := range m.Files {
		if _, err := os.Stat(filepath.Join(dataDir, f.Path)); err != nil {
			return false
		}
	}
	return len(m.Files) > 0
}

func findRepoRoot() (string, error) {
	if root := os.Getenv("MT_ROOT"); root != "" {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found from %s", dir)
		}
		dir = parent
	}
}
