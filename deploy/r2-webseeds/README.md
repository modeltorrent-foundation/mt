# R2 HTTP webseeds (BEP-19)

BitTorrent webseeds (BEP-19 `url-list`) let a plain HTTP mirror serve the exact
model bytes alongside the P2P swarm. A leecher can pull from the webseed when no
peers are online, then verify every piece against the signed manifest — so the
mirror is an accelerator, never a trust root.

This directory documents wiring a **public Cloudflare R2 bucket** in as the
webseed host for the first-wave models.

## Status

- Torrents, signed manifests, and the always-on Hetzner swarm are live (see
  `deploy/hetzner/`). The published magnets already work P2P.
- Webseeds are **not yet in the catalog**: they require a *public* bucket with a
  public domain (`r2.dev` dev URL or a custom domain). The only R2 credentials
  currently on this machine are scoped to a **private** research bucket
  (`claude-intercept-research`) — model weights must **not** go there.
- Everything else is ready: the 3 real GGUFs are staged locally and the
  generator reproduces byte-identical infohashes when `MT_WEBSEED_BASE` is set.

## Why webseeds don't change the torrent

`ws=` webseed hints live in the metainfo `url-list` and in the magnet query
string — they are **outside the info dict**, so the infohash is unchanged.
Verified locally: regenerating all three models with `MT_WEBSEED_BASE` set
produced **identical** `btih` (v1) and `btmh` (v2) infohashes; the only delta was
an additive `&ws=<url>` parameter on each magnet. Existing peers, the Hetzner
seeder, and any already-shared magnet stay fully compatible.

## Webseed URL layout

The generator (`scripts/gen_real_models.go`) resolves each single-file torrent to:

```
$MT_WEBSEED_BASE/<modelId>/<file>
```

So the public bucket must serve these exact object keys (path = `modelId/file`):

| Object key (under the bucket root)                                             | Bytes       |
|-------------------------------------------------------------------------------|-------------|
| `Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf`                                    | 639446688   |
| `HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/smollm2-360m-instruct-q8_0.gguf`     | 386404992   |
| `Qwen/Qwen2.5-0.5B-Instruct-GGUF/qwen2.5-0.5b-instruct-q4_k_m.gguf`            | 491400032   |

`MT_WEBSEED_BASE` is the bucket's public base URL with **no trailing slash**,
e.g. `https://pub-<hash>.r2.dev` or `https://models.<yourdomain>`.

## One-time provisioning (needs the account owner)

Pick either path. Both need R2 permissions this machine's tokens don't have.

### A. Dashboard (fastest)

1. R2 → **Create bucket** → name e.g. `mt-webseeds` (Standard, auto region).
2. Bucket → **Settings** → **Public access**: either enable the **r2.dev dev
   subdomain** (gives `https://pub-<hash>.r2.dev`) or **Connect a custom domain**
   (e.g. `models.example.org`, needs the zone on Cloudflare).
3. R2 → **Manage API Tokens** → **Create API token** → Object Read & Write,
   scoped to `mt-webseeds`. Save the Access Key ID + Secret.

### B. wrangler (CLI)

The wrangler OAuth token on this box lacks R2 scope (R2 API returns
`Authentication error [code: 10000]`). Refresh it first:

```bash
wrangler login          # re-consent; include R2 (Workers R2 Storage) scope
wrangler r2 bucket create mt-webseeds
wrangler r2 bucket dev-url enable mt-webseeds   # -> https://pub-<hash>.r2.dev
```

## Upload + wire in (once the bucket + creds exist)

```bash
# 1) Upload the exact seeded bytes (keys must match the layout table above).
#    Using an S3 token scoped to the new bucket (NOT the private research one):
export AWS_ACCESS_KEY_ID=...        # new bucket token
export AWS_SECRET_ACCESS_KEY=...
export AWS_ENDPOINT_URL=https://<accountid>.r2.cloudflarestorage.com
B=mt-webseeds
aws s3 cp .work/dl/qwen3-0.6b/Qwen3-0.6B-Q8_0.gguf \
  s3://$B/Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf
aws s3 cp .work/dl/smollm2-360m/smollm2-360m-instruct-q8_0.gguf \
  s3://$B/HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/smollm2-360m-instruct-q8_0.gguf
aws s3 cp .work/dl/qwen2.5-0.5b/qwen2.5-0.5b-instruct-q4_k_m.gguf \
  s3://$B/Qwen/Qwen2.5-0.5B-Instruct-GGUF/qwen2.5-0.5b-instruct-q4_k_m.gguf

# 2) Regenerate catalog + signed manifests with the public base URL.
#    (.work/publisher.key is reused, so magnets/infohashes stay identical.)
MT_WEBSEED_BASE=https://pub-<hash>.r2.dev go run ./scripts/gen_real_models.go

# 3) Verify each webseed serves the right bytes (200 + Content-Length).
for u in \
  "$MT_WEBSEED_BASE/Qwen/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf" \
  "$MT_WEBSEED_BASE/HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/smollm2-360m-instruct-q8_0.gguf" \
  "$MT_WEBSEED_BASE/Qwen/Qwen2.5-0.5B-Instruct-GGUF/qwen2.5-0.5b-instruct-q4_k_m.gguf" ; do
  curl -sI "$u" | grep -iE 'HTTP/|content-length'
done

# 4) Optional: full webseed-only fetch (no peers) verifies bytes end-to-end.
MT_CATALOG=web/catalog.json ./bin/mt get Qwen/Qwen3-0.6B-GGUF /tmp/ws-verify --no-seed

# 5) Commit the regenerated web/catalog.json + web/models/**/manifest.json.
#    Only the public R2 domain is committed — never tokens or the private bucket.
```

## Guardrails

- Public model weights go in a **dedicated public** bucket only — never the
  private `claude-intercept-research` research bucket.
- Do **not** webseed from Hugging Face (SCOPE.md) and never commit R2 tokens,
  S3 secrets, or personal IPs. The public R2 domain itself is fine to commit.
- `PROTOCOL.md` is frozen: webseeds are `url-list`/magnet metadata only and must
  not alter the wire format or the infohash.
