# Model Torrent Protocol v1 (FROZEN in Wave 0)

This document is the wire contract. Implementers in later waves make the red
tests green **without changing this file**. If a body cannot satisfy the
contract, the contract wins and the task is wrong, not the protocol.

Status: `schemaVersion: 1`. Anything not specified here is an implementation
detail and may vary between clients.

---

## 1. Identity — content addressing

A model's canonical identity is **not** a URL, a domain, or a database row. It is
content-addressed:

- Every **file** has a lowercase hex **SHA-256** and a byte `size`.
- Every **model** is distributed as a **BitTorrent v2** torrent whose canonical
  reference is its **infohash** (v2, SHA-256 based).
- The shareable form of that infohash is a **magnet** URI. A magnet is enough to
  locate and verify a model with no server alive.

Because identity is content-derived, the same bytes are the same object
everywhere. A domain is a convenience pointer, never the source of truth.

### Quant rule (hash the file, not the folder)

Torrents are created so that **identical file bytes share one identity**. Two
"repos" that both contain the exact same `Q4_K_M.gguf` MUST yield the same
per-file SHA-256, so a single swarm serves both. Do not fold unrelated files
into one opaque archive; that fragments swarms (one of the named failure modes
in the r/LocalLLaMA `1w6bkkh` thread).

---

## 2. Manifest

The manifest is the signed, portable description of a model. It is JSON, UTF-8,
and travels inside the catalog and (optionally) alongside the torrent.

```json
{
  "schemaVersion": 1,
  "modelId": "Qwen/Qwen3-8B",
  "files": [
    { "path": "model.safetensors", "size": 65536, "sha256": "<hex>" },
    { "path": "tokenizer.json",     "size": 65536, "sha256": "<hex>" }
  ],
  "license": { "spdx": "Apache-2.0", "redistributable": true },
  "publisher": { "keyId": "ed25519:<hex-or-fingerprint>", "signature": "<base64>" },
  "magnet": "magnet:?xt=urn:btmh:...",
  "webseeds": ["https://mirror.example.org/Qwen/Qwen3-8B/"],
  "createdAt": "2026-09-15T00:00:00Z"
}
```

Field rules:

- `modelId` — `org/name`, human-facing, NOT an identity. Collisions are resolved
  by content hashes, not by first-writer-wins on this string.
- `files[]` — order-independent set; each entry is `{path, size, sha256}`.
- `license.spdx` — an SPDX identifier. `license.redistributable` MUST be
  consistent with the allowlist in §3.
- `publisher.signature` — Ed25519 signature over the **canonical bytes** of the
  manifest (§4), produced by the key named in `publisher.keyId`.
- `magnet` — the BT v2 magnet for the fileset.
- `webseeds[]` — zero or more HTTP fallbacks (§5).

---

## 3. License gate (protocol-level, not UI)

Ingest MUST reject any manifest whose `license.spdx` is not on the
redistributable allowlist, or whose `redistributable` flag is `false`. This is a
protocol rule: closed, leaked, or non-redistributable weights are refused before
they ever enter a catalog. This is what keeps universities, Internet Archive,
and ISPs able to seed (`1w6bkkh`: "I would totally seed something that was legal
to seed").

Wave-0 allowlist (extend deliberately, with counsel — do not silently widen):

```
Apache-2.0, MIT, BSD-2-Clause, BSD-3-Clause,
CC-BY-4.0, CC-BY-SA-4.0, CC0-1.0, OpenRAIL-permissive
```

Explicitly NOT redistributable-by-default in v1 (require publisher opt-in and
legal review before they can appear): Llama Community, Gemma, and any
"open weights" license with redistribution or field-of-use restrictions.

---

## 4. Canonical bytes & signatures

To sign or verify a manifest:

1. Take the manifest with `publisher.signature` set to the empty string.
2. Serialize deterministically (stable field order, no insignificant
   whitespace). This is `Manifest.Canonical()`.
3. `publisher.signature = base64(ed25519_sign(privkey, canonicalBytes))`.

Verification recomputes the canonical bytes (again with an empty signature
field) and checks the Ed25519 signature against the public key identified by
`publisher.keyId`. Key discovery / keyrings are a later wave; the signature math
is fixed here.

---

## 5. Webseeds

Any fast HTTP(S) mirror MAY back a torrent via BEP-19 (`url-list`). Webseeds are
**fallbacks for when the swarm is thin**, never the catalog and never identity.
A webseed can be a CDN we do not control (even ModelScope or Hugging Face) — the
per-file SHA-256 makes tampering detectable, so an untrusted mirror is safe to
pull from. Losing every webseed MUST NOT break identity or discovery.

---

## 6. Catalog transport

The catalog is a **static JSON index** (`catalog.json`) plus per-model
manifests. Identical bytes must serve equally from:

- a git repository (clone = full mirror),
- a plain HTTPS host / domain pointer,
- IPFS / any content-addressed store (later).

All catalog **read** paths are **anonymous**: no account, no token, no phone
number (`1w6bkkh`: ModelScope "demand mobile number… struggling to serve me
html"). The HTTP read endpoints consumed by the web UI and the Python shim are
documented in [api/catalog.md](api/catalog.md).

---

## 7. What is out of scope for the protocol

Ranking/recommendation, discussions, datasets, Spaces-style hosted compute, and
paid/gated private models are NOT part of this contract. Ranking in particular
must remain local and forkable so no operator (or chip vendor) can tilt
discoverability. If any of these needs a wire format later, it gets its own
versioned section — it does not mutate v1.
