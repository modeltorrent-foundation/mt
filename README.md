# Model Torrent

Open-weight models, peer to peer. Hugging Face catalog UX. A hub nobody can buy.

Read **[SCOPE.md](./SCOPE.md)** for the system design, demand evidence, and 90-day wedge.

Repo: [github.com/modeltorrent-foundation/mt](https://github.com/modeltorrent-foundation/mt)

```bash
go install github.com/modeltorrent-foundation/mt/cmd/mt@latest
```

## Quickstart

```bash
export GOTOOLCHAIN=auto

# Build the CLI
make build

# Generate tiny stand-in fixtures, dev catalog, and web demo magnets
make fixtures

# Download a demo model (seeding stays on unless --no-seed)
MT_CATALOG=web/catalog.json ./bin/mt get demo/tiny-gguf ./downloads/demo/tiny-gguf

# Seed a model you already have on disk (HF layout or snapshots/main)
./bin/mt seed ./downloads/demo/tiny-gguf/snapshots/main

# Seed the bundled small-model pack (web demos + Qwen fixture)
./bin/mt pack popular

# Dev fixture bundle only (testdata catalog)
./bin/mt pack dev

# Serve the web catalog UI
cd web && python3 -m http.server 8080
# open http://localhost:8080
```

Seeding is on by default after `mt get`. Pass `--no-seed` to stop after download.
You do not need our VPS — see [Run your own seeder](./docs/run-your-own-seeder.md).

## License

- **Software** (`cmd/`, `internal/`, `shim/`, `web/`, scripts): [AGPL-3.0-or-later](./LICENSE)
- **Catalog metadata** (`catalog.json`, manifests, model cards): [CC0-1.0](./CATALOG-LICENSE)
- **Model weights**: each publisher's license (SPDX on every manifest)

See [GOVERNANCE.md](./GOVERNANCE.md) for the full licensing model.
