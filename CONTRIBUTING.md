# Contributing to Model Torrent

Model Torrent is preservation infrastructure for **legally redistributable
open-weight models**. Two things follow from that, and they shape every
contribution:

1. **Only redistributable weights.** The license gate is enforced in code, not
   by reviewer mood. Leaked, closed, and non-redistributable checkpoints are
   refused. This is what lets universities, archives, and ordinary people seed
   without thinking twice.
2. **Seeding is on by default.** A torrent index whose listings show zero seeds
   is a graveyard, and several already exist. Every contribution path here ends
   with "and it stays seeded."

Start here for context: [SCOPE.md](./SCOPE.md) (what we are building and why),
[PROTOCOL.md](./PROTOCOL.md) (the frozen v1 wire contract), and
[GOVERNANCE.md](./GOVERNANCE.md) (who decides what, and what nobody gets to
decide). Tone matters too: we write and build like an archive, not like a warez
scene — see [GOVERNANCE.md §13](./GOVERNANCE.md#13-name-and-framing) for why
that costs us nothing and buys us mirror partners.

### Good first issues (90-day wedge)

The labeled tickets track [SCOPE.md's 90-day wedge](./SCOPE.md#90-day-wedge).
Pick one; PROTOCOL.md stays frozen.

| Issue | Wedge slice |
|---|---|
| [#2](https://github.com/modeltorrent-foundation/mt/issues/2) | Optional WebTorrent-hybrid seeder (not on the shared Hetzner box) |
| [#3](https://github.com/modeltorrent-foundation/mt/issues/3) | LM Studio tracker / seed-after-download |
| [#4](https://github.com/modeltorrent-foundation/mt/issues/4) | More redistributable GGUFs in `mt pack popular` |
| [#5](https://github.com/modeltorrent-foundation/mt/issues/5) | Codeberg/GitLab catalog git mirror (second remote) |
| [#6](https://github.com/modeltorrent-foundation/mt/issues/6) | Bootstrap domain pointer (`modeltorrent.org`) |
| [#7](https://github.com/modeltorrent-foundation/mt/issues/7) | Counsel + 501(c)3 papers for GOVERNANCE.md |
| [#8](https://github.com/modeltorrent-foundation/mt/issues/8) | Seed-pack poll: which models belong in `mt pack popular` |
| [#9](https://github.com/modeltorrent-foundation/mt/issues/9) | Browser catalog UX (large GGUFs, save, CDN pin) |
| [#10](https://github.com/modeltorrent-foundation/mt/issues/10) | `mt keygen` / `mt manifest sign` / `mt publish` |
| [#11](https://github.com/modeltorrent-foundation/mt/issues/11) | Keep GitHub Pages and Cloudflare Pages catalogs in lockstep |

Legal specifics in this document are engineering practice, **not legal advice**.
Anything touching license interpretation or the allowlist needs counsel.

---

## Part 1 — Adding a model to the catalog

### Step 0. Prerequisites

```bash
export GOTOOLCHAIN=auto
make build        # ./bin/mt
make test         # Go + Python shim tests
```

### Step 1. Pass the SPDX license gate

Check this **before** you hash 40GB of weights. `internal/manifest` rejects a
manifest whose SPDX id is off the allowlist or whose `redistributable` flag is
false, and `internal/catalog` refuses to index it
([PROTOCOL.md §3](./PROTOCOL.md)).

**Allowed today** — the Wave 0 allowlist, `RedistributableLicenses` in
[`internal/manifest/manifest.go`](./internal/manifest/manifest.go):

```
Apache-2.0   MIT   BSD-2-Clause   BSD-3-Clause
CC-BY-4.0    CC-BY-SA-4.0   CC0-1.0   OpenRAIL-permissive
```

**Rejected, no exceptions:**

- leaked or otherwise unauthorized checkpoints;
- closed / proprietary weights, whatever the source;
- anything behind a click-through, gate, or access request — if it needs
  permission, it is not this network;
- non-commercial or field-of-use restricted terms (`CC-BY-NC-*`, restricted RAIL
  variants);
- weights whose license you cannot name. "Probably fine" is not an SPDX id.

**Rejected by default, opt-in possible:** Llama Community, Gemma, and similar
"open weights" licenses that restrict redistribution or use. Two paths, both
slower than a pull request:

- **Publisher opt-in** — the rights holder signs the manifest themselves with
  their own key, explicitly authorizing redistribution. Narrower, safer,
  preferred.
- **Allowlist change** — written license analysis, counsel review, public
  comment, maintainer vote
  ([GOVERNANCE.md §5](./GOVERNANCE.md#5-licensing-model)). Never a silent commit
  widening the list.

Declare the license honestly. A wrong SPDX id is grounds for removal
([GOVERNANCE.md §8](./GOVERNANCE.md#8-catalog-neutrality-and-fork-rights)) and
is worse than not submitting.

### Step 2. Choose files — hash files, not folders

The named failure mode for model torrents is **quant fragmentation**: one
torrent per quant folder means every flavor gets its own thin swarm, and popular
files end up with three seeders each. The rule from
[PROTOCOL.md §1](./PROTOCOL.md):

> Identical file bytes MUST yield the same per-file SHA-256, so a single swarm
> serves both.

In practice:

- **Never** wrap weights in a `.tar`, `.zip`, or any archive. Archiving changes
  the bytes, so two publishers of the same `Q4_K_M.gguf` produce two unrelated
  swarms.
- Do not re-quantize, re-pack, or "optimize" someone else's GGUF before
  publishing. Byte-identical to the upstream file is the whole point.
- Do not rename to something clever. Path is metadata; the hash is identity, but
  matching upstream paths keeps Hugging Face-layout consumers working.
- Include the **base safetensors**, not only the popular quants. Quants can be
  regenerated from base weights; base weights cannot be regenerated from quants
  ([GOVERNANCE.md §9](./GOVERNANCE.md#9-mirrors-and-preservation-no-kill-switch)).
- Include the small files people actually need: `tokenizer.json`, config, the
  model card.

Hash each file with lowercase hex SHA-256 and record its exact byte size.
`blob.HashFile` in [`internal/blob/blob.go`](./internal/blob/blob.go) is the
reference implementation; `sha256sum` produces the same digest.

```bash
sha256sum model.safetensors tokenizer.json Q4_K_M.gguf
stat -c '%n %s' model.safetensors tokenizer.json Q4_K_M.gguf
```

### Step 3. Magnetize

Create a **BitTorrent v2** torrent over the file set — v2 because its Merkle
tree is SHA-256 based, so per-file identity matches the manifest. Keep the files
as separate entries in one torrent; the v2 file tree is what lets identical
files be recognized across torrents.

Programmatically, `torrentsvc.Create(files, webseeds)` returns the metainfo and
the magnet:

```go
mi, magnet, err := torrentsvc.Create(paths, webseeds)
```

Add **webseeds** (BEP-19 `url-list`) if you have any fast HTTP mirror of the same
files. Webseeds are how a swarm survives having zero peers. They can be hosts we
do not control or trust — including an existing model hub — because per-file
SHA-256 makes tampering detectable ([PROTOCOL.md §5](./PROTOCOL.md)). They are
never identity and never the catalog: **losing every webseed must not break the
listing.**

> **Tooling gap, stated honestly:** there is no `mt publish`, `mt keygen`, or
> `mt manifest sign` subcommand yet — `mt` currently exposes `get`, `seed`,
> `pack`, and `catalog`. Until the publishing commands land, use the Go packages
> directly and follow [`scripts/gen_catalog.go`](./scripts/gen_catalog.go) as the
> working end-to-end example of hashing, canonicalizing, and signing. Patches
> that close this gap are high-value.

### Step 4. Write and sign the manifest

The manifest is the signed, portable description of the model
([PROTOCOL.md §2](./PROTOCOL.md)). It lives at
`models/{org}/{name}/manifest.json` relative to the catalog root.

```json
{
  "schemaVersion": 1,
  "modelId": "Qwen/Qwen3-8B",
  "files": [
    { "path": "model.safetensors", "size": 65536, "sha256": "<hex>" },
    { "path": "tokenizer.json",    "size": 65536, "sha256": "<hex>" },
    { "path": "Q4_K_M.gguf",       "size": 65536, "sha256": "<hex>" }
  ],
  "license": { "spdx": "Apache-2.0", "redistributable": true },
  "publisher": { "keyId": "ed25519:<hex-or-fingerprint>", "signature": "<base64>" },
  "magnet": "magnet:?xt=urn:btmh:...",
  "webseeds": ["https://mirror.example.org/Qwen/Qwen3-8B/"],
  "createdAt": "2026-09-15T00:00:00Z"
}
```

`modelId` is a human label (`org/name`), **not** identity — collisions are
resolved by content hash, not by first-writer-wins
([PROTOCOL.md §1](./PROTOCOL.md)).

**Signing** ([PROTOCOL.md §4](./PROTOCOL.md)) — the math is fixed; do not
improvise a variant:

1. Set `publisher.signature` to the empty string.
2. Serialize deterministically: stable field order, no insignificant whitespace.
   This is `Manifest.Canonical()`.
3. `publisher.signature = base64(ed25519_sign(privkey, canonicalBytes))`.

Verification recomputes the canonical bytes with an empty signature field and
checks against the public key named by `publisher.keyId`. Verify your own
manifest before submitting — `manifest.Parse` then `Manifest.Verify(pub)`.

**Keys, not accounts.** Publishers are Ed25519 keys; there is no user database
([GOVERNANCE.md §6](./GOVERNANCE.md#6-publisher-identity)). If your key is not
yet in the catalog keyring, submit it in the same pull request with a
human-readable publisher name and some independent corroboration that the key is
yours — a signature or key fingerprint published on your own domain, repo, or
existing hub profile. Keep the private key off shared machines; rotation and
revocation are additive keyring commits, never history rewrites.

Publishing someone else's Apache-2.0 weights is fine and expected — sign it with
**your** key and do not imply publisher endorsement in the model name or card.

### Step 5. Add the `catalog.json` entry

One row per model in the index envelope
([`api/catalog.md`](./api/catalog.md), `catalog.Entry` in
[`internal/catalog/load.go`](./internal/catalog/load.go)):

```json
{
  "modelId": "Qwen/Qwen3-8B",
  "license": { "spdx": "Apache-2.0", "redistributable": true },
  "magnet": "magnet:?xt=urn:btmh:...",
  "sizeBytes": 131072,
  "manifest": "/models/Qwen/Qwen3-8B/manifest.json"
}
```

`sizeBytes` is the sum of the file sizes. The `license` block must match the
manifest exactly — the loader gates on both, and a mismatch is a review
rejection. Catalog metadata you contribute is **CC0-1.0**
([GOVERNANCE.md §5](./GOVERNANCE.md#5-licensing-model)); the weights keep their
own license.

### Step 6. Verify before you submit

```bash
make test                                     # manifest + catalog gates must pass
MT_CATALOG=path/to/catalog.json ./bin/mt get <org/name> /tmp/verify
```

A clean `mt get` proves the round trip: the catalog resolves, the swarm or
webseeds deliver bytes, per-file SHA-256 verifies, and a Hugging Face-layout
snapshot lands on disk for `huggingface_hub`, `transformers`, and `llama.cpp`.

Pull request checklist:

- [ ] SPDX id is on the allowlist, or a publisher opt-in is documented in the PR
- [ ] `license.redistributable` is `true` and matches the manifest and catalog entry
- [ ] every file has a lowercase hex SHA-256 and a correct byte size
- [ ] no archives; upstream bytes unmodified; base weights included where they exist
- [ ] manifest signature verifies against a keyring key (or the key is added in this PR)
- [ ] magnet is BitTorrent v2; webseeds, if any, actually resolve
- [ ] `mt get` round-trips and checksums verify
- [ ] you can seed it — see [Step 7](#step-7-seed-it)

Reviewers check the license gate, hash correctness, signature validity, and that
nothing has been re-packed. They **do not** judge whether a model is good,
tasteful, popular, or politically comfortable. If the license permits
redistribution and the bytes are what they claim, it belongs
([GOVERNANCE.md §8](./GOVERNANCE.md#8-catalog-neutrality-and-fork-rights)).

### Step 7. Seed it

Submitting a listing you will not seed produces exactly the zero-seed graveyard
we exist to fix. You do not need our VPS — anyone with the files can seed.
Short version: [docs/run-your-own-seeder.md](./docs/run-your-own-seeder.md).

```bash
./bin/mt seed ./downloads/<org>/<name>/snapshots/main   # seed what you have
./bin/mt pack popular                                   # seed the curated pack
```

Expectations:

- Seed your own submissions until the swarm has several independent seeders, and
  say in the PR roughly how long you can commit to.
- `mt get` leaves the torrent seeding by default; `--no-seed` exists but is not
  the norm.
- **Prioritize thin swarms.** An orphan torrent with two seeders needs you far
  more than the month's most popular 8B does.
- Institutional mirrors and long-horizon preservation are coordinated separately
  — see [Part 3](#part-3--mirrors-and-seeders).

---

## Part 2 — Code contributions

### The protocol wins

[PROTOCOL.md](./PROTOCOL.md) is **frozen at v1**. If your implementation cannot
satisfy the contract, the contract is right and the approach is wrong. Changing
the wire shape means a new versioned section and the Tier 2 process in
[GOVERNANCE.md §12](./GOVERNANCE.md#12-amendment-process) — it does not mutate
v1. Concretely: do not change the manifest JSON shape in
`internal/manifest`, the canonical-bytes rule, the SPDX gate, or the
content-addressing semantics as a side effect of a feature.

### Layout

| Path | What lives there |
|---|---|
| `cmd/mt` | The CLI: `get`, `seed`, `pack`, `catalog`. |
| `internal/manifest` | Manifest types, license gate, canonical bytes, Ed25519 verify. |
| `internal/blob` | SHA-256 content addressing. |
| `internal/torrentsvc` | Torrent create / seed / fetch, webseed fallback. |
| `internal/catalog` | Index, search, swarm health. Ranking stays local. |
| `internal/hflayout` | Hugging Face snapshot cache layout — the compatibility wedge. |
| `shim/` | Python `huggingface_hub` drop-in. |
| `web/` | Static catalog UI. No login, no build step required to try it. |
| `scripts/` | Fixture, catalog, and web-magnet generators (`make fixtures`). |

### Go

```bash
make test-go
make swarm-smoke      # focused localhost seed/fetch test
make tidy
gofmt -l .            # must be empty
```

Standard library first; a dependency in a distribution tool is a supply-chain
surface. Errors wrapped with context (`fmt.Errorf("pkg: op: %w", err)`), no
panics on user input, exit codes that scripts can rely on. New behavior arrives
with a test — ideally one written against the protocol before the code exists.

### Python shim

```bash
make test-py     # skipped, not failed, if pytest is absent
```

The shim's job is that existing code keeps working: `HF_ENDPOINT` pointed at a
mirror, and `huggingface_hub` / `transformers` call paths unchanged. It maps a
`repo_id` to a manifest, then hands the magnet and webseeds to `mt`. Do not
invent a new API surface people have to learn.

### Web

Static HTML/CSS/JS served from a directory (`python3 -m http.server 8080` in
`web/`). Non-negotiable properties:

- **No account, token, or phone number to browse or download.**
- License, per-file checksum, magnet, and peer count visible on the model page —
  the magnet is never hidden behind a proxy button.
- Search and ordering run locally over the cloned catalog. **No central ranking
  service, no paid placement, no vendor weighting**
  ([PROTOCOL.md §7](./PROTOCOL.md)).
- No analytics, no tracking, no fonts or scripts fetched from third parties.

### Client telemetry

Do not add it. Swarm health comes from trackers, mirrors, and manifests — never
from profiling users
([GOVERNANCE.md §11](./GOVERNANCE.md#11-security-and-abuse)).

### Patches, licensing, and review

- One logical change per pull request; explain *why*, and link the SCOPE.md or
  PROTOCOL.md section it serves.
- Add a `Signed-off-by` line (DCO). Inbound equals outbound: code under
  **AGPL-3.0-or-later**, catalog metadata under **CC0-1.0**. There is no
  copyright assignment — deliberately
  ([GOVERNANCE.md §5](./GOVERNANCE.md#5-licensing-model)).
- Report security issues privately first; see
  [GOVERNANCE.md §11](./GOVERNANCE.md#11-security-and-abuse).
- Releases are signed by a multi-signer quorum
  ([GOVERNANCE.md §7](./GOVERNANCE.md#7-releases-multi-signer-by-default)), so
  expect a second reviewer on anything that touches verification, signing, or
  the license gate.

---

## Part 3 — Mirrors and seeders

Replication is the actual defense against capture, so mirror operators are
first-class contributors
([GOVERNANCE.md §9](./GOVERNANCE.md#9-mirrors-and-preservation-no-kill-switch)).

- **Catalog mirror:** clone the catalog repository and serve byte-identical
  static JSON over HTTPS. A clone is a complete index; a `git remote` on a
  different host in a different jurisdiction is a preservation act.
- **Webseed:** expose the same files over plain HTTP(S) at stable paths and add
  the URL to the relevant manifests. Untrusted mirrors are safe to pull from,
  because clients verify per-file SHA-256.
- **Seeder:** run `mt seed` or `mt pack popular` on a box that stays up. Thin
  swarms first. `mt get` already seeds by default; a dedicated host is optional
  ([docs/run-your-own-seeder.md](./docs/run-your-own-seeder.md)).
- **Institutional partners** (university libraries, national archives, the
  Internet Archive) — please get in touch before mirroring at scale so we can
  supply the license provenance and written policy your legal team will want.
  This is precisely why the network carries only redistributable weights.

A useful standing test of whether the design still holds: **can a stranger with
only a clone and a magnet stand up a working index?** If the answer ever becomes
no, that is a release blocker.

---

## Questions that are genuinely open

Not settled, and pretending otherwise would waste your time:

- Publishing tooling (`mt keygen` / `mt manifest sign` / `mt publish`) does not
  exist yet; keyring discovery is a later wave.
- Which models make up the curated seed pack.
- Whether Llama- and Gemma-style licenses ever reach the allowlist, or stay
  per-publisher opt-in — that one needs counsel
  ([GOVERNANCE.md §14](./GOVERNANCE.md#14-open-questions-needing-a-human-or-legal-decision)).

Done is better than perfect. A working listing with real seeders beats a perfect
one that never ships.
