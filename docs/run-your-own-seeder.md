# Run your own seeder

Model Torrent stays alive because people seed. You do not need our VPS, a
Cloudflare account, or any special permission. A laptop that stays awake, a
homelab box, or a cheap VPS all work.

## `mt get` already seeds

After a successful download, `mt get` **keeps seeding** until you stop it
(`Ctrl+C`). That is the default. `--no-seed` exists for one-shot fetches
(CI, a quick checksum check) and is not the usual path.

```bash
go install github.com/modeltorrent-foundation/mt/cmd/mt@latest

# Download a catalog model and remain in the swarm
MT_CATALOG=web/catalog.json mt get Qwen/Qwen3-0.6B-GGUF ./downloads/qwen3-0.6b
```

Bytes are written only after per-file SHA-256 matches the signed manifest.
Failed integrity checks do not seed.

## Seed a directory you already have

If the files are already on disk (Hugging Face snapshot layout, or a flat
folder of the torrent's files):

```bash
mt seed ./downloads/qwen3-0.6b
# or the snapshot dir directly:
mt seed ./downloads/qwen3-0.6b/snapshots/main
```

`mt seed` matches the files against the local catalog and announces the same
infohash as the published magnet. Identical bytes join the same swarm.

## Always-on (optional)

A laptop that sleeps is a fair-weather seeder. For a box that stays up, the
templated systemd unit in [`deploy/hetzner/`](../deploy/hetzner/) is
host-agnostic despite the directory name — point `MT_SEED_HOST` at any
SSH-able Linux machine. That is a convenience, not a requirement. One
`mt seed` process on a machine that does not suspend is enough to keep a
thin swarm from going dark.

## What you are not signing up for

- You are **not** the catalog. Magnets and signed manifests live in git.
- You are **not** a trust root. Clients verify SHA-256 regardless of who
  seeded or which HTTP webseed answered.
- You do **not** need to copy our deployment. Extra independent seeders are
  the point.

See [CONTRIBUTING.md](../CONTRIBUTING.md) if you also want to add a model.
