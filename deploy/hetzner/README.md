# Always-on seeder deployment

A BitTorrent swarm dies when the last seed sleeps. A laptop is not a seeder — it
suspends, roams NATs, and drops off the DHT. This directory deploys the `mt`
seeder (and an optional WebTorrent-hybrid process) to a **durable, non-sleeping
host** as `systemd` services that auto-restart and survive reboot.

The directory name is historical; nothing here is Hetzner-specific. The target
host is supplied at runtime and is **never** committed (this repo is public).

## What it installs

- `/opt/model-torrent/bin/mt` — the cross-compiled CLI (linux/amd64).
- `/opt/model-torrent/seed/<slug>/<file>` — one directory per model, each holding
  exactly one model file (so the rebuilt infohash matches the published magnet).
- `/opt/model-torrent/env/<slug>.env` — per-instance `MT_LISTEN_PORT` (a distinct,
  fixed, reachable BitTorrent port per model).
- `/etc/systemd/system/mt-seed@.service` — templated unit; `Restart=always`,
  `WantedBy=multi-user.target` (survives reboot), memory-capped and niced so it
  never threatens the box's primary services.
- `/etc/systemd/system/mt-webtorrent-hybrid.service` — optional WebRTC/WSS seeder
  for browser WebTorrent (three small GGUFs only). See below.

The seeder runs the **online** client config: DHT on, announcing to the public
trackers in `internal/torrentsvc/trackers.go`. (The offline/test config disables
both; that path is only for the hermetic test suite.)

## Usage

```bash
# 1. Put the model files under the staging dir (default .work/dl):
#      .work/dl/<slug>/<file>.gguf
#    (e.g. via huggingface-cli download ... --local-dir .work/dl/<slug>)

# 2. Generate real magnets + signed manifests + catalog entries:
go run ./scripts/gen_real_models.go

# 3. Deploy TCP/UDP seeders to your durable host (ssh alias or user@host):
MT_SEED_HOST=my-vps ./deploy/hetzner/deploy_seeder.sh
#    Non-root remote user? add:  MT_SSH_SUDO=sudo

# 4. Deploy the WebRTC/WSS hybrid seeder (browser peers):
MT_SEED_HOST=my-vps ./deploy/hetzner/deploy_hybrid.sh
```

## Verify the swarm

From any machine with the published catalog:

```bash
MT_CATALOG=web/catalog.json ./bin/mt get Qwen/Qwen3-0.6B-GGUF /tmp/verify --no-seed
```

A non-zero exit means integrity was not established and nothing is written —
`mt get` fails closed and only writes/【seeds】 bytes whose SHA-256 matches the
signed manifest.

## Operating

```bash
systemctl status 'mt-seed@*'          # TCP/UDP BitTorrent seeders
journalctl -u 'mt-seed@qwen3-0.6b' -f # logs for one model
systemctl restart mt-seed@qwen3-0.6b  # bounce one seeder

systemctl status mt-webtorrent-hybrid  # WebRTC/WSS seeder
journalctl -u mt-webtorrent-hybrid -f
```

Live ports on the shared VPS: `42413` (Qwen3-0.6B), `42414` (SmolLM2-360M),
`42415` (Qwen2.5-0.5B), `42416` (Qwen3-8B). Hybrid uses ephemeral WebRTC UDP
plus public WSS trackers; it has no extra TCP listen port.

## Notes

- **This host is a BitTorrent peer, not the HTTP durability layer.** Leave the
  `mt-seed@*` units running; do not upgrade or replace them for webseeds.
  HTTP webseeds (BEP-19) are a public R2 bucket — see `deploy/r2-webseeds/`.
  Do not point webseeds at Hugging Face (SCOPE.md) or leak a personal IP into
  the public catalog.
- **WebTorrent-hybrid is on this box on purpose.** Browser WebTorrent speaks
  WebRTC/WSS, not TCP/UDP, so `mt-seed@*` never appears as a browser peer.
  `mt-webtorrent-hybrid` seeds the **three small Apache-2.0 GGUFs** from the
  existing `/opt/model-torrent/seed/<slug>` files and the catalog
  `publish.torrent` (same infohashes). It is niced (`Nice=10`),
  `CPUQuota=25%`, `MemoryHigh=384M`, `MemoryMax=768M`. **Qwen3-8B is not
  hybrid-seeded**: the file is 4.68GiB and this VPS has 3.7Gi RAM plus a
  LinkedIn Chrome session; putting 8B in webtorrent would OOM the tenant.
  8B remains R2 HTTP webseed + `mt-seed@qwen3-8b` (mmap, port 42416).
- **Disk:** after copying 8B (~4.68GiB) onto the VPS there was 8.6GiB free
  (floor was 5GiB). Keep an eye on `df -h`; the unit is memory-capped but not
  disk-capped.
- **Do not change the three small infohashes.** Hybrid adds WSS announce URLs
  at runtime; it never re-creates torrents from files.
