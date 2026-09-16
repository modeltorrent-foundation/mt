# Model Torrent — system scope

Limewire-simple model downloads. Hugging Face catalog UX. BitTorrent underneath. A hub nobody can buy.

This is the north-star for the repo. It is written against live demand on r/LocalLLaMA after NVIDIA agreed to acquire Hugging Face for **$12.93B** (announced 2026-09-03, expected close 1H 2027).

**Primary evidence (this thread, scraped over headed Chrome CDP on 2026-09-15, not a second-hand reconstruction):**

- [“ModelScope” Is a Hugging Face Alternative now that Nvidias deal is a Go](https://old.reddit.com/r/LocalLLaMA/comments/1w6bkkh/modelscope_is_a_hugging_face_alternative_now_that/?limit=500) (`1w6bkkh`)
- Rendered URL: `old.reddit.com` … `?limit=500`. www.reddit.com loaded in CDP but the new UI navigated into a comment permalink (`/comment/p7lzi7o/`) when expanders were clicked.
- Capture: **116 comments of 118 listed**, 28 top-level, 56 authors, 0 remaining “load more comments”, 0 login/captcha blockers. The 2-comment gap is consistent with `[deleted]` nodes still counted by Reddit.
- Artifacts: `~/Desktop/web-interact/localllama-1w6bkkh/2026-09-15T08-49-33Z/` (`extract.json`, `body.txt`, `page.html`, `page.png`).

Sibling / earlier threads are still useful, but they are **not** this scrape:

- Sibling “what reliable alternatives exist?” (`1w6986r`) — second-hand until similarly scraped
- June 2026 torrent wave (`1u4gto1`, 897 points; `1uhevvf` / [modelregistry.io](https://modelregistry.io); `1oy9w39`) — IPFS cross-seed, magnet UX, 0-seed graveyard

## The one-sentence bet

Hugging Face won because it felt like GitHub for models. NVIDIA can buy the company. Nobody can buy a swarm that auto-seeds, speaks the old APIs, and keeps the catalog in git.

## What `1w6bkkh` actually said

OP **Hannibalj2ca** (228 points, flair Resources) posted ModelScope (`modelscope.cn` / `modelscope.ai`) as a backup “if things go south,” and said today’s NVIDIA wants power consolidation. The comment tree did **not** treat ModelScope as the answer. The highest-scoring comment is a torrent request.

### OP

> I liked the Nvidia that focused on just GPUs for gaming, not on the Nvidia of today which seem want power consolidation. Modelscope is another platform for those that simply want to know an alternative if things go south.

### Highest-signal quotes from this thread

| Score | Author | Quote | Product implication |
|---|---|---|---|
| 118 | Cereal_Grapeist | “I really hope that some kind of torrent alternative pops up instead to get rid of these points of failure. Maybe it's just under heavy load but Modelscope is pretty janky right now.” | The named “HF alternative” is already a point of failure. P2P is the top demand, not another CDN. |
| 48 | PaceZealousideal6091 | “If you have a problem with nvidia owned Hugging face then you should also keep away from Ali Baba owned modelscope. There's a need for an independent repository.” | Independence is the product. Alibaba is not a hedge against NVIDIA. |
| 33 | 1-800-methdyke | “It would be so obvious for something like LMStudio that already has model discovery and library management to incorporate a tracker to seed downloaded models. The question is whether users would turn on seeding.” | Seed from the client people already use. Default on; do not hope they open qBittorrent. |
| 30 | CatzRuleZWorld | “lots of people would like seeding especially since it’s not pirated material.” | Legal-only catalog is an adoption feature, not just a lawyer feature. |
| 19 | PaceZealousideal6091 | “It's about power corporations owning these repositories. They get the power to decide what can be hosted and what not.” | Governance = who can unlist a model. |
| 12 | whichsideisup | “its alibaba and the chinese government. we need independent companies hosting this.” | State-bound clouds fail the same test as chip-vendor clouds. |
| 12 | makingnoise | ModelScope is “VERY much targeted at Chinese speakers… slow AF (at least on the East Coast of the US).” | English catalog + Western/EU webseeds. Translation is not a community. |
| 10 | PM_ME_DEAD_CEOS | “Torrent is like 1% of the value of a HF type of site. HF is a huge resource for information, discussion, free compute, etc.” | Blobs without cards/search/discussion will not pull people off HF. Steal the catalog feeling. |
| 9 | makingnoise | “I would totally seed something that was legal to seed.” (won’t seed “seven seas” even with a VPN) | SPDX gate + no pirate branding. |
| 6 | SkyFeistyLlama8 | “falls under Chinese government regulations and Communist Party surveillance, like anything with a .cn domain. If you're thinking that an Nvidia-linked HF would be bad for uncensored models, then ModelScope isn't much better.” | Uncensored/experimental tail is a hard requirement. |
| 5 | cniinc | Hosting “isn't viable for an independent company… maybe [Usenet] would be a better system.” / later: Anna’s Archive–style contribution-based speed. | Nonprofit + contribution bandwidth, not ads. |
| 4 | noiserr | Torrent is “a wrong usage model because the individual files you download tend not to result in many seeders, because there are so many different quants.” Seed the original, or people re-quant locally. | Content-address **files**. Prefer popular GGUFs in the seed pack; also pin base safetensors. |
| 4 | makingnoise | “I need a non-profit organization to fill this huge gap and be immune from commercial incentives or corporate takeover.” | 501(c)3 / asset lock. Wiki / archive.org as the comparison, not a startup. |
| 3 | eto-bleh | ModelScope won’t host “QwenniePooh-3.8-Horny-UNCENSORED-Abliterated-27B… it's definitely not an alternative.” GitHub/MSFT “was also (and still) diabolical.” | License-allowable community fine-tunes stay. GitHub-under-MSFT is **not** comfort. |
| 2 | a_beautiful_rhind | ModelScope “slow as molasses… demand mobile number or the right email… downloading 100gb of weights from a site struggling to serve me html.” | No account to download. Resume must work. |
| 1 | Clueless_Nooblet | “No uncensored models allowed -> not an alternative.” | Same as SkyFeistyLlama8, compressed. |
| 1 | datbackup | “Devote the energy to a truly decentralized model storage network, not just another company that will get bought or shitty or bought then shitty.” | No acquisition path. |
| 1 | makingnoise | “a proper 501(c)3 … can be structured in such a way to literally make it illegal to transfer the assets to a for profit.” Wiki foundation, archive.org. | Kill the $12.9B payout path in the charter. |
| 0 | EugenePopcorn | “Wed probably be better off torrenting with them as a web seed.” | HTTP webseeds from whoever is fast, including CN mirrors, without making them the catalog. |
| 0 | 1-800-methdyke | Private trackers reward **orphan** torrents (≤5 seeds) more than popular ones. | Swarm-health UI + seed-pack should prioritize thin swarms, not only Qwen-8B. |
| −1 | Yorn2 | Still centralized; wants pay-for-host storage/compute (Lightning, Nostr). | Out of v1. Note the demand; don’t block on crypto markets. |

### Demand map (this thread only)

| Theme | Gravity in `1w6bkkh` | Notes |
|---|---|---|
| Torrent / client-side seeding | Dominant (118-pt top comment + 33-pt LMStudio tracker reply) | Distinct from June’s “magnet links when” — here it is a **rejection of ModelScope**, not a hobby tracker. |
| Independent / unbuyable host | Very high (48-pt Alibaba symmetry, 19-pt “who decides what can be hosted”) | Nonprofit + asset lock shows up unprompted. |
| China / Alibaba / censorship / phone signup | High, contested | English UX, `.cn` surveillance, uncensored tail. A minority says anti-China talking points are cringe; the upvote pattern still rejects ModelScope. |
| NVIDIA/HF will be fine (GitHub, monopoly optics) | Medium (24 “Feels premature”, 12 “won’t rock the boat”, 8 GitHub comparison) | Do not wait for a ban. Build during the close window anyway. |
| HF is more than blobs | Medium (10-pt “torrent is 1%”) | Catalog, cards, discussion. Defer Spaces/compute. |
| Speed / resume | Present | ModelScope jank + HF session/resume breakage (`wget`/FDM workarounds). |
| Quant fragmentation | Present (4-pt noiserr) | Same-file hashing; don’t make one torrent per quant folder if the GGUF is identical. |
| IPFS, Ollama, Civitai, HF API drop-in | **Not in this thread** | Those signals are June / sibling / product inference. Do not cite them as `1w6bkkh`. |

Live check of [modelregistry.io](https://modelregistry.io) on 2026-09-15: most listed models show **0 seeds**. This thread independently names [llama.garden](https://llama.garden/) (de4dee, 10 pts) and [duckweights.com](https://duckweights.com) (Relevant-Magic-Card; P2P + seed credits + homelab RAID). A torrent index without a seeding default is still a graveyard.

The quieter NVIDIA risk is not bans. Several commenters (SporksInjected, DeepOrangeSky, fallingdowndizzyvr) argue NVIDIA **wants** models to run because it sells the GPUs, and may leave HF alone for monopoly optics. The counter from this thread is still decisive: **whoever owns the repository decides what can be hosted.** Ranking here must be local, forkable, and hardware-neutral.

## Domain name

**The protocol does not need a domain. Humans do.**

- Canonical identity is a **magnet / infohash / CID**, not a URL.
- The catalog is a **git repo** anyone can clone and mirror.
- A memorable domain (`modeltorrent.org`) is a bootstrap pointer: search, docs, WebTorrent UI. Mirrors remain `modeltorrent.pages.dev` and `modeltorrent-foundation.github.io/mt`.
- Run at least two independent HTTPS mirrors plus a git remote. A domain seizure must be a DNS event, not death.

If the only copy of the index lives at one domain, we rebuilt Hugging Face with extra steps.

## Legal line

Limewire is the **UX metaphor**, not the business model.

v1 redistributes only weights whose licenses allow redistribution (Apache-2.0, MIT, and similarly permissive terms; SPDX on every manifest). Reject leaked, closed, and non-redistributable checkpoints. Some “open weight” licenses are **not** free to rehost — those stay off the network unless the publisher signs them in.

That is how universities, Internet Archive, and ISPs can participate. “AI Pirate Bay” branding kills those partners on day one.

## Architecture

```
[browser] [mt CLI] [huggingface_hub shim] [llama.cpp / Ollama]
                    |
                    v
            Catalog (git, signed manifests)
                    |
         +----------+----------+
         v          v          v
    BitTorrent   HTTP      IPFS CID
    v2 swarm    webseeds   (optional)
```

| Layer | Job | Must not |
|---|---|---|
| 0. Governance | Foundation / nonprofit. AGPL hub software. No acquisition path. | Become a VC-backed company a chip vendor can write a check for. |
| 1. Blobs | BitTorrent v2 pieces + HTTP webseeds + optional IPFS CID. | One CDN as the only copy of the weights. |
| 2. Manifests | Signed model cards, SPDX license, SHA256, publisher keys. | Unsigned mystery bins. |
| 3. Catalog | Search, tags, hardware, GGUF vs safetensors, swarm health. | A ranking algorithm a vendor can tilt. |
| 4. Compatibility | `huggingface_hub` drop-in. llama.cpp and Ollama keep working. | Force a new ecosystem. |
| 5. Clients | Browser WebTorrent, `mt get`, always-on seeder daemon. | Require a desktop app or an account to download. |
| 6. Seeding | Auto-seed after download. Seed-packs. Uni / Archive.org mirrors. | Hope datahoarders remember to open qBittorrent. |

### Why BitTorrent *and* IPFS *and* HTTP

LocalLLaMA already litigated this (June threads for IPFS; `1w6bkkh` for seeding + webseeds + quant split):

- Torrents are the first alternative people named when ModelScope showed up janky. LMStudio-class clients should seed after download.
- Quant fragmentation is the in-thread objection: one torrent per GGUF flavor starves seeders. Hash the **file**. Pin popular quants **and** the base safetensors so people can re-quant.
- IPFS (June threads, not `1w6bkkh`) wins on same-file cross-seeding. Same requirement, different wire.
- HTTP webseeds (BEP-19) are how a swarm survives 0 peers. EugenePopcorn even suggested using ModelScope as a webseed, not as the catalog.

v1: BitTorrent v2 + HTTP webseeds, with per-file SHA256 that can later pin to an IPFS CID without rewriting the catalog.

## Hugging Face parity

Steal the feeling. Do not clone the company.

| HF surface | When | How |
|---|---|---|
| Model pages, cards, file lists | v1 | Git markdown + blob map |
| Search / tags / task filters | v1 | Local index over catalog git |
| `huggingface_hub` / `hf download` | v1 | Compatible HTTP + P2P backend |
| Checksums / revisions | v1 | Content hash **is** the revision |
| Orgs / publishers | v1 | Signing keys, not a user database |
| Datasets | v1.5 | Same blob protocol, later catalog |
| Discussions | v1.5 | Nostr or git issues, not a central DB |
| Spaces / Inference / AutoTrain | Defer | Cloud products. Out of scope. |
| Gated / paid private models | Never | If it needs permission, it is not this network. |

## Hyper-accessibility

People will not adopt a new religion. They will adopt a faster download that still answers the old names.

1. **No account** to download.
2. **Browser**: click a model, WebTorrent starts, magnet is visible, license + checksum + peer count on the same page.
3. **CLI**: `mt get org/name` writes a Hugging Face-layout folder and **leaves the torrent seeding**.
4. **Existing clients**: LM Studio (named in-thread, 33 pts) and later llama.cpp/Ollama should seed what they already downloaded. Do not require a new desktop app as the only seeder.
5. **Drop-in**: `HF_ENDPOINT=https://<mirror>` so existing Python stays.
6. **Seed pack**: `mt pack popular` — the one-click seeder bundle the registry thread asked for.
7. **Small first**: 4–8B GGUFs are the on-ramp. Base safetensors are the preservation layer (community can re-quant if quants vanish).
8. **Resume + verify** on potato internet. Offline after first get.
9. **Publishers** (Unsloth, Qwen, Mistral) upload once, publish a signed magnet, and can delete their origin copy after N seeders.

## 90-day wedge

Done is better than perfect. The June threads already said that.

1. Protocol spec: magnet = identity, SPDX gate, publisher signatures, webseed fallback.
2. Ten popular GGUFs + two base safetensors, hashed and magnetized.
3. `mt get` / `mt seed` / `mt pack`. Seeding on by default.
4. HF Hub-compatible HTTP shim.
5. Static site: search, card, files, swarm health, magnet. No login.
6. Foundation papers + at least two catalog mirrors so a domain is optional.

### Explicit non-goals for v1

No Spaces. No inference marketplace. No accounts required to download. No closed-weight leaking. No token-gated dumps the license forbids redistributing.

## Competitive landscape (do not ignore)

| Project | What it got right | Why it is not enough |
|---|---|---|
| ModelScope | Exists, two landing pages (`modelscope.ai` English-friendlier). Named in OP. | Alibaba. `.cn` surveillance. Chinese-language comments. Phone signup reports. Slow in US/EU. Uncensored tail rejected in-thread. Another company. |
| modelregistry.io | Magnets + HF webseeds | 15 GitHub stars; live listings often 0 seeds; no HF API; no auto-seed client. |
| llama.garden | Named in `1w6bkkh` (10 pts) as an existing torrent list | A list of torrents is not a hub. Same seed-default problem. |
| duckweights.com | Named repeatedly in `1w6bkkh`; P2P + seed credits + homelab RAID | Early; account/credits; still a company-shaped site. Watch, don’t clone the incentive theater. |
| Lemonade | One comment: “provides it by default,” hopes Unsloth follows | Client default-source is distribution, not a catalog. |
| Ollama | One command, local-first | **Not mentioned in `1w6bkkh`.** Still a company registry. |
| Academic Torrents / IA | Proven preservation; archive.org named as nonprofit template | archive.org also called slow/unusable in-thread. Not a model hub. |
| IPFS one-offs | Content addressing | **Not mentioned in `1w6bkkh`.** noiserr’s quant-fragmentation point is the local equivalent. |

The gap is **seeding default + HF-shaped UX + uncensorable catalog**. Not another index page.

## Governance

NVIDIA paid $12.9B because a company can be bought. GitHub-under-Microsoft is the comfort take in `1w6bkkh` (Signature97, fallingdowndizzyvr) and the horror take in the same thread (eto-bleh). makingnoise’s bar is sharper: a 501(c)3 structured so transferring assets to a for-profit is **illegal**. LandscapePenguin asked who would refuse a $12B check — that is the threat model.

- Software: AGPL (or equivalent copyleft) so a fork stays free.
- Catalog: CC0 metadata, licenses stay with the artifacts.
- Operator: foundation / nonprofit, multi-signer releases, asset lock against sale.
- Kill switch: there isn’t one. Mirrors continue from git + magnets.

## Open questions (do not block v1)

- Exact trademark / name (Model Torrent vs something less pirate-coded for university partners).
- Which 10 GGUFs are the seed pack (ask LocalLLaMA; default to whatever is actually running this month).
- How hard to be on restrictive-but-common licenses (Llama / Gemma style). Start strict; widen with counsel.
- Whether to webseed from Hugging Face **during** the close window (useful, but a chokepoint we must outgrow).
