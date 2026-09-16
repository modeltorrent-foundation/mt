# Always-on seeder deployment

A BitTorrent swarm dies when the last seed sleeps. A laptop is not a seeder — it
suspends, roams NATs, and drops off the DHT. This directory deploys the `mt`
seeder to a **durable, non-sleeping host** (a VPS, a homelab box that stays up, a
NAS that runs containers — anything with a stable uplink) as `systemd` services
that auto-restart and survive reboot.

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

# 3. Deploy to your durable host (ssh alias or user@host):
MT_SEED_HOST=my-vps ./deploy/hetzner/deploy_seeder.sh
#    Non-root remote user? add:  MT_SSH_SUDO=sudo
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
systemctl status 'mt-seed@*'          # health
journalctl -u 'mt-seed@qwen3-0.6b' -f # logs for one model
systemctl restart mt-seed@qwen3-0.6b  # bounce one seeder
```

## Notes

- **This host is a BitTorrent peer, not the HTTP durability layer.** Leave the
  `mt-seed@*` units running; do not upgrade or replace them for webseeds.
  HTTP webseeds (BEP-19) are a public R2 bucket — see `deploy/r2-webseeds/`.
  Do not point webseeds at Hugging Face (SCOPE.md) or leak a personal IP into
  the public catalog.
- **Disk hygiene**: model files are small on purpose. Keep an eye on the target's
  free space; the unit is memory-capped but not disk-capped.
