# Remaining before public launch

Do not post to Reddit or boost the social copy until real models are seeding.

- [x] First commit is on GitHub (`modeltorrent-foundation/mt`)
- [x] Hide or fix the browser WebTorrent button (needs WSS/WebRTC seeder)
      Public WSS trackers (`wss://tracker.openwebtorrent.com`,
      `wss://tracker.webtorrent.dev`) are on generated magnets; infohashes
      unchanged. Browser path is HTTP webseeds (R2 + same-origin fixtures) plus
      a memory-capped `mt-webtorrent-hybrid` unit on Hetzner for the three small
      GGUFs (not 8B). Button stays enabled when an HTTP webseed exists;
      otherwise disabled with copy-magnet + `mt get`.
- [x] Deploy `web/` (Pages or Cloudflare) — GitHub Pages workflow on
      `modeltorrent-foundation/mt` serves `web/` (after `make fixtures`).
      Live: https://modeltorrent.org/ (Cloudflare Registrar + Pages custom
      hostname; HTTPS is live — Google Trust Services WE1 cert, HTTP 301 to
      HTTPS), https://modeltorrent.pages.dev/ (Cloudflare Pages), and
      https://modeltorrent-foundation.github.io/mt/ (GitHub Pages workflow).
      Shim `DEFAULT_ENDPOINT` is `https://modeltorrent.org` (`HF_ENDPOINT` overrides).
      Qwen3-8B is on the live Pages catalog (`wrangler pages deploy web`).
- [x] 2–3 real Apache-2.0 GGUFs with magnets + a 24/7 seeder
      (Qwen3-0.6B-Q8_0, SmolLM2-360M-Instruct-Q8_0, Qwen2.5-0.5B-Instruct-Q4_K_M;
      signed manifests + hybrid v1/v2 magnets with public trackers in
      `web/catalog.json`; always-on `systemd` seeder deployed via
      `deploy/hetzner/` — DHT + trackers, auto-restart, survives reboot.
      Cross-internet swarm fetch verified end-to-end, checksum passed.
      Merged to `main` via PR #1.)
      - [x] HTTP webseeds (BEP-19): public `mt-webseeds` bucket at
            `https://pub-60d277f41b154e4f826375a17187b003.r2.dev`.
            Three Apache-2.0 GGUFs uploaded; HEAD 200 + matching
            Content-Length. Regenerated magnets keep identical `btih`/`btmh`
            and add `&ws=`. Hetzner `mt-seed@*` units left as a bonus peer.
            CORS rules in `deploy/r2-webseeds/cors.json` so the browser can
            Range-GET those objects.
      Wave 4: `mt pack popular` now seeds the four live Apache-2.0 catalog
      GGUFs (3 small + Qwen3-8B) from `web/catalog.json`, not `demo/tiny-*`
      fixtures or the testdata `Qwen/Qwen3-8B` stub. The 8B file is also on
      the Hetzner disk (`mt-seed@qwen3-8b`, port 42416) because free space
      stayed ≥5GiB after the copy. Hybrid WebRTC seeds the three small GGUFs
      only (`mt-webtorrent-hybrid`).
- [ ] User-facing README / release binaries (`go install` works after this push)
- [x] CI + 8–10 `good first issue` tickets
      GitHub Actions `CI` (`go test ./...` + pytest shim) with README badge.
      Issues #2–#11 (`good first issue`).
- [x] Second catalog mirror (Codeberg)
      Live: https://codeberg.org/modeltorrent-foundation/mt (`curl -sI` 200).
      Local remote `codeberg`. Re-run `./scripts/mirror-to-codeberg.sh` after
      GitHub pushes. Docs: `docs/codeberg-mirror.md`.
- [ ] Reply in r/LocalLLaMA `1w6bkkh` (Cereal_Grapeist torrent comment), then a build-log post
