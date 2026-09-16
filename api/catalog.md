# Catalog read API (v1)

All read endpoints are **anonymous** — no account, token, or phone number
(PROTOCOL.md §6). Everything here is static JSON that can be served from git, a
domain, or a content-addressed store. There is no write API in v1: publishing is
done by adding a signed manifest to the catalog repo.

Base is any mirror root (e.g. `https://modeltorrent-foundation.github.io/mt/`, a raw git host, or an
IPFS gateway). Identical bytes must serve from every mirror.

## `GET /catalog.json`

The full index. Small enough to ship as one file in Wave 0; paginated later.

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-09-15T00:00:00Z",
  "models": [
    {
      "modelId": "Qwen/Qwen3-8B",
      "license": { "spdx": "Apache-2.0", "redistributable": true },
      "magnet": "magnet:?xt=urn:btmh:...",
      "sizeBytes": 131072,
      "manifest": "/models/Qwen/Qwen3-8B/manifest.json"
    }
  ]
}
```

## `GET /models/{org}/{name}/manifest.json`

One model's full signed manifest (see PROTOCOL.md §2). Consumers MUST verify the
publisher signature and the per-file SHA-256 before trusting bytes.

## `GET /models/{org}/{name}/health.json`

Optional, best-effort swarm liveness. Absent or stale health MUST NOT block a
download — a magnet plus webseeds is enough.

```json
{
  "modelId": "Qwen/Qwen3-8B",
  "seeders": 12,
  "peers": 3,
  "lastWebseedOk": "2026-09-15T00:00:00Z",
  "checksumOk": true
}
```

## Consumers

- **Web UI** (`web/`) fetches `catalog.json` for search and `health.json` per
  model page (seeds/peers/checksum).
- **Python shim** (`shim/`) maps a Hugging Face `repo_id` to `manifest.json`,
  then hands the magnet/webseeds to `mt` for the actual transfer.
