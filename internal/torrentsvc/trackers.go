package torrentsvc

import "github.com/anacrolix/torrent/metainfo"

// DefaultTrackers is the set of well-known public BitTorrent trackers the real
// (online) seed and fetch paths announce to, in addition to the DHT. Trackers
// are NOT part of the manifest/magnet identity (the infohash is), so adding them
// does not change the PROTOCOL wire format — it only improves peer discovery so
// a real swarm can form (SCOPE.md §6: "seeding default"; a 0-peer catalog is a
// graveyard).
var DefaultTrackers = []string{
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://open.demonii.com:1337/announce",
	"udp://tracker.openbittorrent.com:6969/announce",
	"udp://exodus.desync.com:6969/announce",
	"udp://tracker.torrent.eu.org:451/announce",
	"udp://open.stealth.si:80/announce",
}

// WebTorrentTrackers are public WebSocket trackers used by browser WebTorrent.
// They are outside the info dict (same as UDP trackers): adding them does not
// change btih/btmh. Classic TCP/UDP seeders ignore wss:// announce URLs.
var WebTorrentTrackers = []string{
	"wss://tracker.openwebtorrent.com",
	"wss://tracker.webtorrent.dev",
}

// AnnounceTrackers is the UDP public set plus WebTorrent WSS trackers. Use this
// when generating magnets that browsers might consume. Seeders that only speak
// TCP/UDP still use DefaultTrackers when rebuilding metainfo from disk.
func AnnounceTrackers() []string {
	out := make([]string, 0, len(DefaultTrackers)+len(WebTorrentTrackers))
	out = append(out, DefaultTrackers...)
	out = append(out, WebTorrentTrackers...)
	return out
}

// announceListFrom turns a flat tracker list into a BEP-12 announce-list where
// every tracker is its own tier (all tiers tried).
func announceListFrom(trackers []string) metainfo.AnnounceList {
	if len(trackers) == 0 {
		return nil
	}
	al := make(metainfo.AnnounceList, 0, len(trackers))
	for _, t := range trackers {
		al = append(al, []string{t})
	}
	return al
}

// CreateWithTrackers is Create plus a BEP-12 announce-list. The returned magnet
// carries the trackers as "tr" params (via MagnetV2), and the raw metainfo
// announces to them when seeded. The infohash is identical to Create over the
// same bytes — trackers live outside the info dict.
func CreateWithTrackers(files, webseeds, trackers []string) (Metainfo, string, error) {
	return createBitTorrentWithTrackers(files, webseeds, trackers)
}
