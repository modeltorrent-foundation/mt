# Remaining before public launch

Do not post to Reddit or boost the social copy until real models are seeding.

- [x] First commit is on GitHub (`modeltorrent-foundation/mt`)
- [ ] Hide or fix the browser WebTorrent button (needs WSS/WebRTC seeder)
- [ ] Deploy `web/` (Pages or Cloudflare) and register a domain
- [x] 2–3 real Apache-2.0 GGUFs with magnets + a 24/7 seeder
      (Qwen3-0.6B-Q8_0, SmolLM2-360M-Instruct-Q8_0, Qwen2.5-0.5B-Instruct-Q4_K_M;
      signed manifests + hybrid v1/v2 magnets with public trackers in
      `web/catalog.json`; always-on `systemd` seeder deployed via
      `deploy/hetzner/` — DHT + trackers, auto-restart, survives reboot.
      Cross-internet swarm fetch verified end-to-end, checksum passed.)
      - [ ] HTTP webseeds (BEP-19): deferred until a neutral public mirror/domain
            exists — must not webseed from Hugging Face (SCOPE) or leak a personal IP.
- [ ] User-facing README / release binaries (`go install` works after this push)
- [ ] CI + 8–10 `good first issue` tickets
- [ ] Second catalog mirror (Codeberg/GitLab)
- [ ] Reply in r/LocalLLaMA `1w6bkkh` (Cereal_Grapeist torrent comment), then a build-log post
