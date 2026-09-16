# Model Torrent

[![CI](https://github.com/modeltorrent-foundation/mt/actions/workflows/ci.yml/badge.svg)](https://github.com/modeltorrent-foundation/mt/actions/workflows/ci.yml)

BitTorrent meets Hugging Face. Open-weight models, content-addressed. A hub nobody can buy.

**Catalog** — [modeltorrent.org](https://modeltorrent.org/)
**Source** — [github.com/modeltorrent-foundation/mt](https://github.com/modeltorrent-foundation/mt) · [codeberg.org/modeltorrent-foundation/mt](https://codeberg.org/modeltorrent-foundation/mt)

```mermaid
flowchart LR
  catalog --> magnet
  magnet --> swarm
  magnet --> R2[R2 webseed]
```

```bash
GOTOOLCHAIN=auto go install github.com/modeltorrent-foundation/mt/cmd/mt@latest
```

```bash
mt get Qwen/Qwen2.5-0.5B-Instruct-GGUF
```

A real catalog model. Seeding stays on after download (`--no-seed` for a one-shot). You do not need our VPS — [run your own seeder](./docs/run-your-own-seeder.md).

Weights verify against a signed manifest. The HTTP webseed is an accelerator, never a trust root.

Software is [AGPL-3.0-or-later](./LICENSE). Catalog metadata is [CC0-1.0](./CATALOG-LICENSE). Each weight keeps its publisher SPDX.

Bootstrap domain: [modeltorrent.org](https://modeltorrent.org/). Mirrors: [modeltorrent.pages.dev](https://modeltorrent.pages.dev/), [modeltorrent-foundation.github.io/mt](https://modeltorrent-foundation.github.io/mt/). Catalog git mirror: [codeberg.org/modeltorrent-foundation/mt](https://codeberg.org/modeltorrent-foundation/mt) ([docs/codeberg-mirror.md](./docs/codeberg-mirror.md)). Design: [SCOPE.md](./SCOPE.md). Governance: [GOVERNANCE.md](./GOVERNANCE.md).

Wave 4 (real SPDX weights) is live: three small Apache-2.0 GGUFs (Qwen3-0.6B Q8_0, SmolLM2-360M Instruct Q8_0, Qwen2.5-0.5B Instruct Q4_K_M) seed 24/7 from the Hetzner VPS over BitTorrent, and Qwen3-8B Q4_K_M (~4.68GiB) is on that same disk as an `mt-seed` peer (there was ~14GiB free before the copy and 8.6GiB after, above the 5GiB floor). HTTP for every catalog GGUF is the R2 webseed only — never Hugging Face. Browsers can also hit a memory-capped WebTorrent-hybrid process for the three small GGUFs (not 8B: WebRTC would OOM the 3.7Gi shared box). `mt pack popular` seeds those same four catalog GGUFs from `web/catalog.json`; it no longer seeds `demo/tiny-*` 64KB fixtures or the testdata `Qwen/Qwen3-8B` stub. Infohashes for the original three small torrents are unchanged. PROTOCOL.md stays frozen.

## Dev

Tiny fixtures, a local catalog, and web demo magnets:

```bash
export GOTOOLCHAIN=auto
make fixtures
make build
MT_CATALOG=web/catalog.json ./bin/mt get demo/tiny-gguf ./downloads/demo/tiny-gguf
cd web && python3 -m http.server 8080
```
