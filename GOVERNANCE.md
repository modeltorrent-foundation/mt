# Model Torrent — Charter and Governance

**Status: DRAFT. Not legal advice.** This document states design intent for the
entity that will operate Model Torrent. Every clause in [§3 Legal
structure](#3-legal-structure-draft--requires-counsel) is a *specification to
hand to counsel*, not a filed instrument. Nothing here is binding until an
attorney in the chosen jurisdiction has drafted articles and bylaws and they
have been adopted. Where this document says "illegal" or "prohibited," read
"we intend the governing documents to make this so, and counsel must confirm
it is achievable."

Read alongside:

- [SCOPE.md](./SCOPE.md) — system design, the demand evidence, the 90-day wedge.
- [PROTOCOL.md](./PROTOCOL.md) — the frozen v1 wire contract. Governance does
  not get to quietly reinterpret it.
- [CONTRIBUTING.md](./CONTRIBUTING.md) — how a model or a patch actually lands.

---

## 1. Mission

Keep open-weight models downloadable, verifiable, and re-hostable by anyone,
permanently, without depending on the continued goodwill of any company or
state.

Concretely: a catalog anyone can clone, content-addressed identities that
survive every server dying, and enough independent mirrors and seeders that no
single actor can make a model disappear.

We are building infrastructure, not a protest. See [§13](#13-name-and-framing)
on why that distinction is operationally load-bearing.

## 2. The non-acquisition principle

NVIDIA agreed to pay $12.93B for Hugging Face because **a company can be
bought**. Swapping one corporate owner for another — a chip vendor, a cloud, a
state-adjacent platform — fails the same test twice. The design goal is not
"a nicer host." It is an operator with **no acquisition path and no unlisting
power worth capturing**.

Three defenses, in order of how much we trust them:

1. **Structural (weakest link, strongest intent):** an asset lock in the
   governing documents, so there is no legal route by which assets or control
   pass to a for-profit. See [§3](#3-legal-structure-draft--requires-counsel).
2. **Licensing:** AGPL software and CC0 metadata, so a fork is a complete,
   lawful replacement rather than a partial copy. See [§5](#5-licensing-model).
3. **Topological (strongest):** the valuable thing — weights, hashes, magnets,
   catalog history — is already replicated outside our control. Buying the
   foundation would buy a domain, a trademark, and a git remote that a hundred
   people already have a full copy of. See
   [§9](#9-mirrors-and-preservation-no-kill-switch).

If we ever have to rely on defense 1 alone, we have already failed. The
foundation should be **structurally boring to acquire and technically pointless
to acquire.**

Corollary, from the sharpest point in the originating thread — *whoever owns the
repository decides what can be hosted*: the foundation must be built so that it
**does not hold that power in the first place**. A governance body that can be
trusted not to abuse unlisting is still a governance body worth buying. See
[§8](#8-catalog-neutrality-and-fork-rights).

## 3. Legal structure (DRAFT — requires counsel)

**Intended form:** a US 501(c)(3) public charity, or a functionally equivalent
non-profit in another jurisdiction (Dutch *stichting*, German *e.V.*, or
fiscal sponsorship under an existing charity such as the Software Freedom
Conservancy while we are small). Comparators we are deliberately imitating:
the **Wikimedia Foundation** and the **Internet Archive** — long-lived,
donation-funded, mirror-friendly. Not a startup, not a foundation that is a
shell for a commercial entity.

**Charitable purpose (draft language for counsel):** preservation of and public
access to openly licensed machine-learning artifacts, and development of free
software for their distribution — framed as educational, scientific, and
archival.

**Asset lock — the clauses we are asking counsel to draft:**

| # | Intent | Notes for counsel |
|---|---|---|
| A1 | Assets (domains, trademarks, catalog repositories, signing keys, funds, infrastructure) may not be sold, licensed exclusively, or transferred to any for-profit entity. | The "no $12.9B payout" clause. Narrow, enumerated exceptions only: arms-length purchase of ordinary services (hosting, audits, legal). |
| A2 | On dissolution, all assets pass only to another non-profit with a substantially identical purpose and an equivalent asset lock. | Standard 501(c)(3) dissolution language, tightened to require the lock to survive. Named fallback beneficiaries (e.g. an archival institution) should be identified in advance. |
| A3 | The entity may not convert to, merge into, or be reorganized as a for-profit, and may not create a for-profit subsidiary that holds any core asset. | The OpenAI-shaped failure mode. Counsel should advise on what actually binds here versus what a future board can amend. |
| A4 | No member, director, officer, or contributor holds an equity-like interest, and no compensation may be contingent on a transfer of control. | Removes the personal incentive to sell. |
| A5 | The clauses above are entrenched: amendment requires the elevated process in [§12](#12-amendment-process), and any amendment that weakens A1–A4 is void. | **Counsel must tell us how strong this really is.** Self-entrenchment against a future board is the single most uncertain item in this document. |

**Known open legal questions — do not paper over these:**

- How durable is an asset lock against a determined future board? A
  supermajority can usually amend bylaws; the honest answer may be "quite
  durable in the articles, weak in the bylaws." This drives which clauses go
  where.
- Jurisdiction. US 501(c)(3) buys donor deductibility and a familiar template;
  an EU foundation buys distance from US-centric pressure. A two-entity
  structure (US charity + EU foundation, mirroring each other) is plausible but
  more overhead than an unfunded project can carry.
- Trademark ownership and the enforcement posture (see [§13](#13-name-and-framing)).
- Safe-harbor and notice-and-takedown obligations per jurisdiction for an
  index that stores hashes, magnets, and metadata but may host no weights
  itself.
- Whether signing keys can be held such that no single officer, and no
  successor entity, can unilaterally issue releases.

Until an entity exists, **the project is governed by this document as a
convention among maintainers**, and by the licenses on the files. Nobody should
claim charity status, tax-deductibility, or an operative asset lock before it is
filed and confirmed.

## 4. What the operator controls — and what it must not

Enumerating the powers we *withhold* matters more than enumerating the ones we
keep, because the withheld ones are what an acquirer would be buying.

**The foundation may:**

- publish releases of the reference software and the catalog snapshot;
- maintain the publisher keyring and the SPDX allowlist process;
- operate mirrors, trackers, and webseeds, and coordinate seed-packs;
- hold the trademark and the bootstrap domain;
- accept donations and pay for infrastructure, audits, and legal work.

**The foundation may not:**

- operate a ranking, recommendation, or "featured" algorithm that any funder,
  vendor, or officer can tilt — ranking is local and forkable
  ([PROTOCOL.md §7](./PROTOCOL.md));
- require an account, token, phone number, or any identification to *download*
  ([PROTOCOL.md §6](./PROTOCOL.md));
- host gated, paid, or permissioned models. "If it needs permission, it is not
  this network" ([SCOPE.md](./SCOPE.md));
- unlist a model on any ground outside the published, narrow list in
  [§8](#8-catalog-neutrality-and-fork-rights);
- make any mirror, CDN, or webseed the sole source of a model's bytes;
- sell placement, sell exclusivity, or condition distribution on hardware
  vendor, cloud, or model family;
- collect per-user telemetry from clients as a condition of use.

**There is no kill switch.** Not "we promise not to use it" — the architecture
does not provide one. See [§9](#9-mirrors-and-preservation-no-kill-switch).

## 5. Licensing model

Three layers, three different licenses, on purpose.

| Layer | License | Why |
|---|---|---|
| Software (`cmd/`, `internal/`, `shim/`, `web/`, scripts) | **AGPL-3.0-or-later** | Strong copyleft, including over a network. A hosted fork stays free. If someone runs a nicer Model Torrent, users can get its source. |
| Catalog metadata (`catalog.json`, manifests, model cards, tags, hashes) | **CC0-1.0** | Metadata must be trivially re-hostable, re-indexable, and forkable with zero legal thought. A fork of the catalog is a lawful clone, not a derivative-work argument. |
| Model weights | **Whatever the publisher chose.** We never relicense. | We are a distributor, not an owner. Every manifest carries the SPDX id, and only redistributable licenses pass the gate. |

**The SPDX gate.** Ingest rejects any manifest whose license is not on the
redistributable allowlist, or whose `redistributable` flag is false. This is a
protocol rule enforced in code, not a UI hint — see
[PROTOCOL.md §3](./PROTOCOL.md) and `RedistributableLicenses` in
`internal/manifest/manifest.go`. Closed, leaked, and non-redistributable
checkpoints are refused before they reach a catalog.

This is a governance decision, not only a legal one. Universities, the Internet
Archive, and ISPs can only seed a network whose contents are lawful to seed, and
ordinary users seed more willingly when nothing on the network is stolen —
*"lots of people would like seeding especially since it's not pirated material."*
Legal-only redistribution is an **adoption feature**.

**Widening the allowlist** (e.g. toward Llama- or Gemma-style community
licenses) requires: a written analysis of the license's redistribution and
field-of-use terms, review by counsel, a public comment period, and an
explicit maintainer vote. It is never a silent commit. A specific publisher may
alternatively opt their own weights in by signing them, which is a narrower and
safer path than a blanket allowlist change. Start strict; widen deliberately.

Contributions are inbound=outbound: patches under the layer's license above,
with a `Signed-off-by` line (DCO). No copyright assignment to the foundation —
an entity that owns all the copyright is an entity whose acquisition means
something.

## 6. Publisher identity

**Publishers are keys, not accounts.** There is no user database to subpoena,
sell, or leak; there is nothing to migrate if the foundation disappears.

- A publisher is an **Ed25519** key, referenced as `ed25519:<hex-or-fingerprint>`
  in `publisher.keyId`.
- Signatures cover the manifest's canonical bytes exactly as fixed in
  [PROTOCOL.md §4](./PROTOCOL.md). Verifiers recompute the canonical form; the
  signature math does not change with governance.
- The **keyring** — the mapping from `keyId` to a public key and a human-readable
  publisher name — lives in the catalog repository, in git, under the same CC0
  terms. Its history is auditable by anyone with a clone.
- Adding a key is a reviewed catalog change: the key, a name, and some
  independent corroboration that the key belongs to who it claims to (a
  signature published on the publisher's own domain, repo, or existing hub
  presence).
- **Rotation and revocation** are additive keyring commits, never rewrites.
  Revoking a compromised key does not invalidate history; it marks the key
  untrusted from a stated point, and affected manifests must be re-signed.
  Content already downloaded and verified stays verifiable by hash regardless.
- A `keyId` is not a namespace monopoly. `modelId` is a human label; identity is
  the content hash ([PROTOCOL.md §1](./PROTOCOL.md)). Two publishers of the same
  bytes converge on the same swarm — that is a feature, not a conflict.
- Downloading requires **no** key, identity, or account. Signing is for
  publishers and verifiers.

Open question for maintainers: how much corroboration is enough for a key
claiming to be a well-known org, and who arbitrates a disputed claim without
recreating a naming authority. Current bias: attest weakly, display honestly
("unverified publisher"), never adjudicate more than we must.

## 7. Releases: multi-signer by default

No single person — and no single compromised laptop — can ship a release that
clients trust.

- Releases of the reference software and of catalog snapshots are signed by a
  **quorum of release signers: at least 2 of N, N ≥ 3**, with signers based in
  more than one jurisdiction and not all employed by the same organization.
- Signer keys are held individually (hardware tokens preferred). There is no
  single "foundation master key" whose custody transfers with the entity.
- Tags are signed; release artifacts carry per-file SHA-256; builds should be
  reproducible far enough that a second signer can independently rebuild before
  signing rather than rubber-stamping a binary.
- Adding or removing a signer is a public, quorum-approved change with a
  minimum notice period, recorded in the repository.
- If quorum becomes unreachable (people leave, keys are lost), the recovery path
  is a **publicly announced re-keying with a comment period** — not a quiet
  fallback to one trusted person.

Emergency exception, deliberately narrow: a single signer may ship a fix for an
actively exploited security issue, and must publish the diff and obtain
retroactive quorum sign-off within seven days. Repeated use of this exception is
a governance failure and should be treated as one.

## 8. Catalog neutrality and fork rights

The reason to have this network is that no owner should decide what may be
hosted. So the operator's discretion is enumerated and small.

**A model may be removed or refused only because:**

1. it fails the license gate — non-redistributable, missing, or misdeclared
   SPDX ([PROTOCOL.md §3](./PROTOCOL.md));
2. its bytes do not match its manifest hashes, or its signature does not verify;
3. it ships malware, an executable payload, or a deserialization exploit
   masquerading as weights;
4. it contains content whose *possession or distribution* is unlawful in the
   mirror's jurisdiction (CSAM and similar), or it is a verified copyright
   claim against the checkpoint itself;
5. the signing publisher requests removal of their own release.

**Not grounds for removal:** the model is uncensored, abliterated, rude, weird,
politically inconvenient, embarrassing, commercially threatening to a funder,
badly named, low quality, unpopular, or a fine-tune somebody dislikes. If the
license permits redistribution and the bytes are what they claim to be, it
stays. The community fine-tune tail is not a tolerated externality; it is a
stated requirement of the project.

**Every removal is logged** — model id, hash, ground from the list above, date,
requester class, and who decided — in a public, append-only record in the
catalog repository. A removal that cannot be justified in one line from the list
above should not happen. Publishers may appeal to the maintainers, and the
appeal and outcome are logged too.

**Ranking is not a governance lever.** Search and ordering run locally over a
cloned catalog and are forkable by construction
([PROTOCOL.md §7](./PROTOCOL.md), `internal/catalog`). There is no central
relevance service to tune, no paid placement, and no hardware-vendor
weighting. Where the project surfaces curation at all — seed-packs, swarm
health — the bias is explicit and mechanical: **thin swarms first**, because an
orphan torrent with three seeds needs help and Qwen-8B does not.

**Fork rights are explicit and unconditional.** Anyone may clone the catalog,
the software, and the keyring; run their own mirror, tracker, or index; apply
their own curation policy, stricter or looser; and use the protocol without
permission, attribution to the foundation, or notice. The only things the
foundation asserts are AGPL reciprocity on software and trademark on the name.
**A fork that disagrees with our curation is a success condition of the design,
not a threat to it.**

## 9. Mirrors and preservation (no kill switch)

Survival is a property of replication, not of promises. The architecture must
make "shut it down" a meaningless instruction.

**Standing requirements:**

- **≥ 2 independent git remotes** for the catalog, on different hosts in
  different jurisdictions, plus the expectation that every contributor's clone
  is a full mirror. A clone of the catalog is a complete, working index.
- **≥ 2 independent HTTPS catalog mirrors**, operated by different people or
  institutions, serving byte-identical JSON
  ([PROTOCOL.md §6](./PROTOCOL.md)).
- **≥ 1 institutional preservation partner** as a standing goal — a university
  library, a national archive, the Internet Archive — holding the base weights,
  not only the popular quants. Institutions can only participate in a legal-only
  network, which is why [§5](#5-licensing-model) is upstream of this section.
- **Webseeds are fallbacks, never authority.** Any fast HTTP mirror may back a
  swarm, including one we do not control and do not trust, because per-file
  SHA-256 makes tampering detectable ([PROTOCOL.md §5](./PROTOCOL.md)). Losing
  every webseed must not affect identity or discovery.
- **Seeding is on by default** in first-party clients, because a torrent index
  without seeders is a graveyard — the observable failure mode of existing
  model-torrent sites, most of whose listings show zero seeds.

**The domain is a bootstrap pointer, not a point of control.** Losing it must be
a DNS inconvenience: magnets still resolve, clones still work, mirrors still
serve. If the only usable copy of the index lives at one domain, we rebuilt
Hugging Face with extra steps.

**Preservation bias:** pin the base safetensors as well as the popular quants.
Quants can be regenerated from base weights; base weights cannot be regenerated
from quants.

**Continuity:** if the foundation dissolves, is seized, or is captured, the
intended outcome is a non-event — mirrors keep serving, magnets keep resolving,
and a successor group publishes from a forked catalog under a new keyring. The
[A2 dissolution clause](#3-legal-structure-draft--requires-counsel) covers
formal assets; replication covers everything that actually matters. Maintainers
should periodically verify this by test: **can a stranger with only a clone and
a magnet stand up a working index?** If not, that is a release blocker, not a
philosophical concern.

## 10. Decision-making, roles, and vendor neutrality

Small and legible, sized for a project that does not exist yet rather than for
the org chart we hope to need.

- **Contributors** — anyone who sends a patch, a manifest, or a mirror.
- **Maintainers** — review and merge in a defined area (protocol, CLI, catalog,
  web, shim, infra). Decisions by lazy consensus on the public record; sustained
  disagreement escalates to a maintainer vote, not to whoever is loudest.
- **Release signers** — the quorum in [§7](#7-releases-multi-signer-by-default).
  Overlapping with maintainers is fine; being the *same single person* is not.
- **Board / stewards** (once an entity exists) — legal, financial, and
  trademark responsibility. Explicitly **not** an editorial body: the board does
  not decide what may be hosted, and cannot direct a removal outside
  [§8](#8-catalog-neutrality-and-fork-rights).

**Vendor neutrality and conflicts:**

- No single employer, funder, or state-affiliated organization may hold a
  majority of maintainer or board seats — and specifically not a GPU vendor, a
  hyperscaler, or a model lab.
- Affiliations are disclosed publicly and stated in any decision they touch.
- No funder gets placement, ranking influence, exclusive webseed status, or
  advance notice of policy changes.
- Recusal is required on decisions touching an employer's models or products.

**Funding.** Donations, grants, and in-kind infrastructure. No ads, no paid
placement, no exclusive hosting deals, no pay-to-rank, no gated tiers, and no
"contribute bandwidth to unlock speed" credit economy that turns access into a
market. Contributed bandwidth is welcome; **rationed access is not**. Individual
sources above a published threshold are disclosed, and no single source should
be structurally load-bearing — funding concentration is an acquisition path with
extra steps.

## 11. Security and abuse

- Coordinated disclosure with a published contact and a stated response window;
  fixes ship under the [§7](#7-releases-multi-signer-by-default) emergency rule
  when actively exploited.
- Signing-key compromise is handled by keyring revocation
  ([§6](#6-publisher-identity)) and a public incident note. Never a silent
  rotation.
- Clients must verify per-file SHA-256 before writing a Hugging Face-layout
  directory, and must surface verification failures rather than degrade quietly.
- Malicious-payload reports (pickle exploits, trojaned quants) are handled as
  ground 3 in [§8](#8-catalog-neutrality-and-fork-rights) and logged like any
  other removal.
- Client telemetry is not a condition of use. Swarm health is measured from
  trackers and mirrors, not by profiling users.

## 12. Amendment process

Three tiers, because not every line here deserves the same friction.

**Tier 1 — ordinary changes** (wording, process detail, adding a mirror, adding
a role): pull request, 7 days for comment, maintainer lazy consensus.

**Tier 2 — substantive governance changes** (allowlist widening, removal
grounds, quorum size, funding policy, keyring rules): pull request with written
rationale, 21 days public comment, explicit maintainer vote with a recorded
tally, and counsel review where legal exposure changes. Protocol changes get
their own versioned section and **do not mutate v1** — the frozen contract wins
over convenience ([PROTOCOL.md](./PROTOCOL.md)).

**Tier 3 — entrenched core.** These exist to be hard to change, and a proposal
to relax any of them should be read first as evidence of capture:

1. the asset lock and no-for-profit-transfer principle
   ([§3](#3-legal-structure-draft--requires-counsel));
2. copyleft software and CC0 metadata ([§5](#5-licensing-model));
3. no account required to download; no gated, paid, or permissioned models
   ([§4](#4-what-the-operator-controls--and-what-it-must-not));
4. the enumerated removal grounds and unconditional fork rights
   ([§8](#8-catalog-neutrality-and-fork-rights));
5. no central tiltable ranking ([§8](#8-catalog-neutrality-and-fork-rights));
6. the replication minimums and "no kill switch"
   ([§9](#9-mirrors-and-preservation-no-kill-switch)).

Tier 3 requires: 60 days public comment, a two-thirds supermajority of
maintainers **and** of the board once one exists, counsel review, and a
published statement of what the change makes newly possible and who benefits.
Amendments that weaken the asset lock are intended to be void under
[A5](#3-legal-structure-draft--requires-counsel) — with the honest caveat that
[§3](#3-legal-structure-draft--requires-counsel) flags self-entrenchment as the
least certain claim in this document.

All governance changes are public commits in the repository. There is no private
governance channel and no decision that only exists in a call.

## 13. Name and framing

Model Torrent distributes **legally redistributable open-weight models**. It is
not a piracy tool, and the branding must not read like one.

This is operational, not squeamish. Universities, national archives, the
Internet Archive, and ISPs are the mirror partners who make preservation
survive a bad decade — and none of them can sign a partnership with something
positioned as "the AI Pirate Bay." The same framing loses the ordinary
contributors who will happily seed lawful weights and will not touch anything
that smells like warez: *"I would totally seed something that was legal to
seed."* One pirate-coded tagline costs more distribution than it ever wins in
attention.

So: no jolly-roger iconography, no "unblockable," no leaked-checkpoint
lore, no mystery-bin uploads. Serious infrastructure vocabulary — preservation,
verification, mirrors, provenance, availability. Limewire is the *UX metaphor*
for how easy a download should be, never the business model.

Whether the project name itself should be less torrent-coded for institutional
partners remains an open question ([§14](#14-open-questions-needing-a-human-or-legal-decision)).

## 14. Open questions needing a human or legal decision

Listed because pretending they are settled is how governance documents become
fiction. None of these block v1.

**Legal (needs counsel):**

1. Jurisdiction and form — US 501(c)(3), an EU foundation, both, or fiscal
   sponsorship first. Affects donor deductibility, exposure, and how credible
   the asset lock is.
2. How strong an asset lock is actually achievable against a future board, and
   which clauses belong in articles versus bylaws.
3. Safe-harbor and takedown obligations for a hash-and-magnet index across the
   jurisdictions where mirrors will live.
4. Whether the allowlist can ever include Llama- or Gemma-style community
   licenses, or whether those stay per-publisher opt-in permanently.
5. Trademark: who holds it, and whether we enforce against forks that use the
   name (current bias: register defensively, enforce only against
   impersonation).
6. Liability for institutional mirrors — what indemnity or written policy a
   university library needs before it will mirror.

**Governance (needs a human decision, no lawyer required):**

7. Initial maintainer and release-signer roster, and how the first N are chosen
   legitimately when the project is three people.
8. Corroboration standard for a publisher key claiming a well-known org, and
   who arbitrates disputed claims without becoming a naming authority.
9. Whether to accept corporate in-kind infrastructure (bandwidth, storage) and
   under what disclosure and exit terms — free CDN capacity is the friendliest
   acquisition vector there is.
10. Whether to webseed from Hugging Face during the acquisition close window:
    useful now, a chokepoint we must outgrow.
11. The project name and its institutional acceptability
    ([§13](#13-name-and-framing)).
12. Funding-concentration threshold that triggers a disclosed dependency
    warning.

---

*This charter is a draft under [§12 Tier 3](#12-amendment-process) rules once
adopted. Until an entity exists, it is a convention among maintainers and a
specification for counsel. See [CONTRIBUTING.md](./CONTRIBUTING.md) to
participate.*
