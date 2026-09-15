package torrentsvc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	infohash_v2 "github.com/anacrolix/torrent/types/infohash-v2"
)

var (
	seedMu      sync.Mutex
	activeSeeds = map[string]*seedSession{}
)

type seedSession struct {
	client *torrent.Client
	cancel context.CancelFunc
}

func offlineClientConfig(dataDir string) *torrent.ClientConfig {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenHost = torrent.LoopbackListenHost
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.DisableTrackers = true
	cfg.DisablePEX = true
	cfg.NoDefaultPortForwarding = true
	cfg.DropMutuallyCompletePeers = false
	cfg.Seed = true
	cfg.DataDir = dataDir
	return cfg
}

func onlineClientConfig(dataDir string) *torrent.ClientConfig {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.NoDefaultPortForwarding = true
	cfg.Seed = true
	cfg.DataDir = dataDir
	return cfg
}

func buildMetaInfo(files []string, webseeds []string) (metainfo.MetaInfo, *metainfo.Info, error) {
	if len(files) == 0 {
		return metainfo.MetaInfo{}, nil, fmt.Errorf("torrentsvc: Create: no files")
	}
	root, err := torrentRoot(files)
	if err != nil {
		return metainfo.MetaInfo{}, nil, err
	}
	mi := metainfo.MetaInfo{
		AnnounceList: nil,
		UrlList:      webseeds,
	}
	mi.SetDefaults()
	info := metainfo.Info{PieceLength: 256 * 1024}
	if err := info.BuildFromFilePath(root); err != nil {
		return metainfo.MetaInfo{}, nil, fmt.Errorf("torrentsvc: build metainfo: %w", err)
	}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		return metainfo.MetaInfo{}, nil, fmt.Errorf("torrentsvc: marshal info: %w", err)
	}
	return mi, &info, nil
}

func torrentRoot(files []string) (string, error) {
	if len(files) == 1 {
		return files[0], nil
	}
	abs := make([]string, len(files))
	for i, f := range files {
		a, err := filepath.Abs(f)
		if err != nil {
			return "", err
		}
		abs[i] = a
	}
	dir := filepath.Dir(abs[0])
	for _, f := range abs[1:] {
		if filepath.Dir(f) != dir {
			return "", fmt.Errorf("torrentsvc: files must share a directory")
		}
	}
	return dir, nil
}

func wrapMetainfo(mi metainfo.MetaInfo, info *metainfo.Info, files []string) Metainfo {
	v2Hash := infohash_v2.HashBytes(mi.InfoBytes)
	raw, _ := bencode.Marshal(mi)
	return Metainfo{
		InfoHash: v2HexString(v2Hash),
		Name:     info.BestName(),
		Files:    files,
		Webseeds: mi.UrlList,
		Raw:      raw,
	}
}

func magnetString(mi *metainfo.MetaInfo) (string, error) {
	m, err := mi.MagnetV2()
	if err != nil {
		return "", err
	}
	// BEP 52 infohash is SHA-256 of the info dict even for v1-encoded infos.
	m.V2InfoHash.Set(infohash_v2.HashBytes(mi.InfoBytes))
	return m.String(), nil
}

func createBitTorrent(files []string, webseeds []string) (Metainfo, string, error) {
	absFiles := make([]string, len(files))
	for i, f := range files {
		a, err := filepath.Abs(f)
		if err != nil {
			return Metainfo{}, "", err
		}
		absFiles[i] = a
	}
	mi, info, err := buildMetaInfo(absFiles, webseeds)
	if err != nil {
		return Metainfo{}, "", err
	}
	magnet, err := magnetString(&mi)
	if err != nil {
		return Metainfo{}, "", err
	}
	return wrapMetainfo(mi, info, absFiles), magnet, nil
}

func seedDataDir(mi Metainfo) string {
	if len(mi.Files) == 0 {
		return ""
	}
	if len(mi.Files) == 1 {
		return filepath.Dir(mi.Files[0])
	}
	dir := filepath.Dir(mi.Files[0])
	for _, f := range mi.Files[1:] {
		if filepath.Dir(f) != dir {
			return dir
		}
	}
	return dir
}

func seedBitTorrent(ctx context.Context, mi Metainfo, dataDir string, offline bool) error {
	if len(mi.Raw) == 0 {
		return fmt.Errorf("torrentsvc: Seed: missing metainfo bytes")
	}
	key := mi.InfoHash
	if key == "" {
		var parsed metainfo.MetaInfo
		if err := bencode.Unmarshal(mi.Raw, &parsed); err == nil {
			h := infohash_v2.HashBytes(parsed.InfoBytes)
			key = v2HexString(h)
		}
	}
	if key == "" {
		return fmt.Errorf("torrentsvc: Seed: missing infohash")
	}

	seedMu.Lock()
	if _, ok := activeSeeds[key]; ok {
		seedMu.Unlock()
		return nil
	}
	seedMu.Unlock()

	if dataDir == "" {
		dataDir = seedDataDir(mi)
	}
	if dataDir == "" {
		return fmt.Errorf("torrentsvc: Seed: cannot locate data directory")
	}

	cfg := onlineClientConfig(dataDir)
	if offline {
		cfg = offlineClientConfig(dataDir)
	}

	seedCtx, cancel := context.WithCancel(context.Background())
	cl, err := torrent.NewClient(cfg)
	if err != nil {
		cancel()
		return fmt.Errorf("torrentsvc: Seed client: %w", err)
	}

	var parsed metainfo.MetaInfo
	if err := bencode.Unmarshal(mi.Raw, &parsed); err != nil {
		cl.Close()
		cancel()
		return fmt.Errorf("torrentsvc: Seed parse metainfo: %w", err)
	}
	tor, err := cl.AddTorrent(&parsed)
	if err != nil {
		cl.Close()
		cancel()
		return fmt.Errorf("torrentsvc: Seed add torrent: %w", err)
	}
	if err := tor.VerifyData(); err != nil {
		select {
		case <-tor.Complete().On():
		case <-ctx.Done():
			cl.Close()
			cancel()
			return ctx.Err()
		}
	}

	seedMu.Lock()
	activeSeeds[key] = &seedSession{client: cl, cancel: cancel}
	seedMu.Unlock()

	go func() {
		select {
		case <-seedCtx.Done():
		case <-ctx.Done():
		}
		cl.Close()
		seedMu.Lock()
		delete(activeSeeds, key)
		seedMu.Unlock()
		cancel()
	}()

	return nil
}

func fetchBitTorrent(ctx context.Context, magnet, dest string, offline bool) (bool, error) {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return false, err
	}
	cfg := onlineClientConfig(dest)
	if offline {
		cfg = offlineClientConfig(dest)
	}
	cfg.DefaultStorage = storage.NewFile(dest)
	cl, err := torrent.NewClient(cfg)
	if err != nil {
		return false, fmt.Errorf("torrentsvc: Fetch client: %w", err)
	}
	defer cl.Close()

	spec, err := torrent.TorrentSpecFromMagnetUri(magnet)
	if err != nil {
		return false, fmt.Errorf("torrentsvc: Fetch magnet: %w", err)
	}
	t, ok, err := cl.AddTorrentSpec(spec)
	if err != nil {
		return false, fmt.Errorf("torrentsvc: Fetch add: %w", err)
	}
	if !ok {
		return false, fmt.Errorf("torrentsvc: Fetch: duplicate torrent")
	}

	seedMu.Lock()
	for _, sess := range activeSeeds {
		t.AddClientPeer(sess.client)
	}
	seedMu.Unlock()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-t.GotInfo():
	}

	t.DownloadAll()
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-t.Complete().On():
	}

	if err := t.VerifyData(); err != nil {
		return false, fmt.Errorf("torrentsvc: Fetch verify: %w", err)
	}
	return true, nil
}

// MetainfoForDownload rebuilds metainfo from on-disk files so Seed can serve them.
func MetainfoForDownload(dataDir string, webseeds []string) (Metainfo, error) {
	var paths []string
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return Metainfo{}, err
	}
	if len(paths) == 0 {
		return Metainfo{}, fmt.Errorf("torrentsvc: no files in %s", dataDir)
	}
	mi, _, err := createBitTorrent(paths, webseeds)
	return mi, err
}

func isOfflineTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

func v2HexString(h infohash_v2.T) string {
	return fmt.Sprintf("%x", h[:])
}
