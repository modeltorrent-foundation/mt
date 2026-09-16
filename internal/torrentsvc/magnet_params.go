package torrentsvc

import (
	"net/url"
	"strings"
)

// MagnetInfoHashes returns the btih and btmh xt values from a magnet URI
// (lowercase, query-decoded). Missing hashes are empty strings.
func MagnetInfoHashes(magnet string) (btih, btmh string) {
	u, err := url.Parse(magnet)
	if err != nil {
		return "", ""
	}
	for _, xt := range u.Query()["xt"] {
		low := strings.ToLower(xt)
		switch {
		case strings.HasPrefix(low, "urn:btih:"):
			btih = strings.ToLower(xt[len("urn:btih:"):])
		case strings.HasPrefix(low, "urn:btmh:"):
			btmh = strings.ToLower(xt[len("urn:btmh:"):])
		}
	}
	return btih, btmh
}

// AppendMagnetTrackers adds tr= params for any tracker not already present.
// The xt= infohashes are left byte-for-byte intact (trackers are not identity).
func AppendMagnetTrackers(magnet string, trackers []string) string {
	return appendMagnetQuery(magnet, "tr", trackers)
}

// AppendMagnetWebseeds adds ws= params for any webseed not already present.
func AppendMagnetWebseeds(magnet string, webseeds []string) string {
	return appendMagnetQuery(magnet, "ws", webseeds)
}

func appendMagnetQuery(magnet, key string, values []string) string {
	if magnet == "" || len(values) == 0 {
		return magnet
	}
	existing := magnetQueryValues(magnet, key)
	out := magnet
	for _, v := range values {
		if v == "" || containsFold(existing, v) {
			continue
		}
		if !strings.Contains(out, "?") {
			out += "?"
		} else if !strings.HasSuffix(out, "?") && !strings.HasSuffix(out, "&") {
			out += "&"
		}
		out += key + "=" + url.QueryEscape(v)
		existing = append(existing, v)
	}
	return out
}

func magnetQueryValues(magnet, key string) []string {
	u, err := url.Parse(magnet)
	if err != nil {
		return nil
	}
	return u.Query()[key]
}

func containsFold(values []string, want string) bool {
	for _, v := range values {
		if strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}
