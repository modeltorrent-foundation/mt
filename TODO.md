# Remaining before public launch

Do not post to Reddit or boost the social copy until real models are seeding.

- [x] First commit is on GitHub (`modeltorrent-foundation/mt`)
- [x] Hide or fix the browser WebTorrent button (needs WSS/WebRTC seeder)
      Public WSS trackers (`wss://tracker.openwebtorrent.com`,
      `wss://tracker.webtorrent.dev`) are on generated magnets; infohashes
      unchanged. Browser path is HTTP webseeds (R2 + same-origin fixtures) with
      CORS, not `webtorrent-hybrid` on the Hetzner box. Button stays enabled
      when an HTTP webseed exists; otherwise disabled with copy-magnet + `mt get`.
- [x] Deploy `web/` (Pages or Cloudflare) — GitHub Pages workflow on
      `modeltorrent-foundation/mt` serves `web/` (after `make fixtures`).
      Live: https://modeltorrent.pages.dev/ (Cloudflare Pages) and
      https://modeltorrent-foundation.github.io/mt/ (GitHub Pages workflow).
      **modeltorrent.org is not owned; do not claim it.** Shim
      `DEFAULT_ENDPOINT` points at the live Pages URL (`HF_ENDPOINT` overrides).
      Domain registration remains a follow-up issue.
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
- [ ] User-facing README / release binaries (`go install` works after this push)
- [x] CI + 8–10 `good first issue` tickets
      GitHub Actions `CI` (`go test ./...` + pytest shim) with README badge.
      Issues #2–#11 (`good first issue`).
- [x] Second catalog mirror (Codeberg/GitLab)
      Script + docs: `scripts/mirror-to-codeberg.sh`, `docs/codeberg-mirror.md`.
      Creating the Codeberg repo needs a Codeberg account (signup documented).
- [ ] Reply in r/LocalLLaMA `1w6bkkh` (Cereal_Grapeist torrent comment), then a build-log post
