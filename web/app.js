// Model Torrent — static catalog UI with in-browser WebTorrent downloads.
//
// Contract (see api/catalog.md):
//   GET catalog.json                         -> { models: [{ modelId, license, magnet, manifest, ... }] }
//   GET models/{org}/{name}/manifest.json    -> full signed manifest (file list)
//   GET models/{org}/{name}/health.json     -> { seeders, peers, checksumOk, ... }
//
// No framework, no backend, no login. Ranking stays local and forkable.

const CATALOG_URL = "catalog.json";
const NO_PEERS_GRACE_MS = 15_000;

/** @typedef {{ modelId: string, license?: { spdx?: string, redistributable?: boolean }, magnet?: string, sizeBytes?: number, manifest?: string }} CatalogEntry */
/** @typedef {{ path: string, size: number, sha256: string }} ManifestFile */
/** @typedef {{ seeders?: number, peers?: number, checksumOk?: boolean, lastWebseedOk?: string }} SwarmHealth */
/** @typedef {{ torrent: object, modelId: string, noPeersTimer?: ReturnType<typeof setTimeout>, blobUrls?: string[], uiAttached?: boolean }} ActiveDownload */

/** @type {object | null} */
let webTorrentClient = null;

/** @type {Map<string, ActiveDownload>} */
const activeDownloads = new Map();

/** Known placeholder hash fragments in demo magnets. */
const PLACEHOLDER_FRAGMENTS = [
  "1220a1b2c3d4",
  "1234567890abcdef",
  "0123456789abcdef",
  "deadbeefdeadbeef",
  "abcdef1234567890",
];

/** Load the static catalog index. Returns { models, error }. */
async function loadCatalog() {
  try {
    const res = await fetch(CATALOG_URL, { cache: "no-store" });
    if (!res.ok) {
      return { models: [], error: "missing" };
    }
    const data = await res.json();
    const models = Array.isArray(data.models) ? data.models : [];
    return { models, error: models.length === 0 ? "empty" : null };
  } catch {
    return { models: [], error: "missing" };
  }
}

/** Fetch JSON from a relative path; null on failure. */
async function fetchJSON(url) {
  try {
    const res = await fetch(url, { cache: "no-store" });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

/** Derive health.json path from modelId (org/name). */
function healthPath(modelId) {
  return `models/${modelId}/health.json`;
}

/** Case-insensitive substring match over modelId. Local, forkable ranking. */
function filterModels(models, query) {
  const q = query.trim().toLowerCase();
  if (!q) return models;
  return models.filter((m) => (m.modelId || "").toLowerCase().includes(q));
}

/** Human-readable byte size. */
function formatBytes(n) {
  if (typeof n !== "number" || n < 0) return "—";
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)} MB`;
  return `${(n / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

/** Human-readable download speed. */
function formatSpeed(bytesPerSec) {
  if (!bytesPerSec || bytesPerSec <= 0) return "0 B/s";
  return `${formatBytes(bytesPerSec)}/s`;
}

/** Extract btih/btmh from a magnet URI. */
function parseMagnetHashes(magnet) {
  const btih = magnet.match(/[?&]xt=urn:btih:([^&]+)/i)?.[1] || "";
  const btmh = magnet.match(/[?&]xt=urn:btmh:([^&]+)/i)?.[1] || "";
  return {
    btih: decodeURIComponent(btih).toLowerCase(),
    btmh: decodeURIComponent(btmh).toLowerCase(),
  };
}

/** True when btih/btmh looks like a demo placeholder, not a live swarm. */
function isPlaceholderMagnet(magnet) {
  if (!magnet || !magnet.startsWith("magnet:")) return true;
  const { btih, btmh } = parseMagnetHashes(magnet);
  const combined = `${btih}${btmh}`;
  if (!combined) return true;

  for (const fragment of PLACEHOLDER_FRAGMENTS) {
    if (combined.includes(fragment.toLowerCase())) return true;
  }

  // Repeated single hex digit or very low entropy (e.g. aaaa…, 121212…)
  if (/^(.)\1{15,}$/.test(btih) || /^(.)\1{15,}$/.test(btmh)) return true;

  return false;
}

/** Lazy singleton WebTorrent client; null when CDN failed to load. */
function getWebTorrentClient() {
  if (typeof WebTorrent === "undefined") return null;
  if (!webTorrentClient) {
    webTorrentClient = new WebTorrent();
  }
  return webTorrentClient;
}

/** Copy text to clipboard with a brief visual confirmation. */
async function copyToClipboard(text, buttonEl) {
  try {
    await navigator.clipboard.writeText(text);
    if (buttonEl) {
      const prev = buttonEl.textContent;
      buttonEl.textContent = "Copied!";
      setTimeout(() => {
        buttonEl.textContent = prev;
      }, 1500);
    }
    return true;
  } catch {
    return false;
  }
}

/**
 * @param {CatalogEntry} entry
 * @param {ManifestFile[]} files
 * @param {SwarmHealth | null} health
 */
function renderCard(entry, files, health) {
  const tmpl = document.getElementById("model-card");
  const node = tmpl.content.cloneNode(true);
  const card = /** @type {HTMLElement} */ (node.querySelector(".model"));

  card.dataset.modelId = entry.modelId || "";

  node.querySelector(".model-id").textContent = entry.modelId || "unknown";

  const spdx = entry.license?.spdx || "unknown license";
  const redist = entry.license?.redistributable;
  const licenseEl = node.querySelector(".license");
  licenseEl.textContent =
    redist === true ? `${spdx} · redistributable` : redist === false ? `${spdx} · not redistributable` : spdx;

  const magnetUri = entry.magnet || "";
  const placeholder = isPlaceholderMagnet(magnetUri);

  const magnetUriEl = node.querySelector(".magnet-uri");
  const copyBtn = /** @type {HTMLButtonElement} */ (node.querySelector(".btn-copy-magnet"));
  const externalMagnet = /** @type {HTMLAnchorElement} */ (node.querySelector(".magnet-external"));

  if (magnetUri) {
    magnetUriEl.textContent = magnetUri;
    magnetUriEl.title = "Click to copy magnet URI";
    externalMagnet.href = magnetUri;
    copyBtn.disabled = false;
  } else {
    magnetUriEl.textContent = "No magnet available";
    magnetUriEl.classList.add("magnet-uri--missing");
    externalMagnet.href = "#";
    externalMagnet.classList.add("magnet-external--disabled");
    copyBtn.disabled = true;
  }

  const filesEl = node.querySelector(".files");
  if (files.length === 0) {
    const li = document.createElement("li");
    li.className = "files-empty";
    li.textContent = "No file list (manifest unavailable)";
    filesEl.appendChild(li);
  } else {
    for (const f of files) {
      const li = document.createElement("li");
      li.className = "file-row";
      li.innerHTML = `
        <div class="file-path"><code>${escapeHtml(f.path)}</code> <span class="file-meta">${formatBytes(f.size)}</span></div>
        <code class="sha256" title="Click to copy SHA-256">${escapeHtml(f.sha256)}</code>
      `;
      const shaEl = li.querySelector(".sha256");
      shaEl.addEventListener("click", () => copyToClipboard(f.sha256, null));
      filesEl.appendChild(li);
    }
  }

  const seeders = health?.seeders;
  const peers = health?.peers;
  const checksumOk = health?.checksumOk;

  node.querySelector(".seeders").textContent =
    typeof seeders === "number" ? `${seeders} seeders` : "— seeders";
  node.querySelector(".peers").textContent =
    typeof peers === "number" ? `${peers} peers` : "— peers";

  const livePeersEl = node.querySelector(".live-peers");
  livePeersEl.textContent = "live: —";
  livePeersEl.hidden = true;

  const checksumEl = node.querySelector(".checksum");
  checksumEl.textContent =
    checksumOk === true ? "checksum OK" : checksumOk === false ? "checksum failed" : "checksum unknown";
  checksumEl.classList.toggle("checksum--ok", checksumOk === true);
  checksumEl.classList.toggle("checksum--bad", checksumOk === false);

  const downloadBtn = /** @type {HTMLButtonElement} */ (node.querySelector(".btn-download"));
  const downloadHint = node.querySelector(".download-hint");
  const progressWrap = /** @type {HTMLElement} */ (node.querySelector(".download-progress"));

  if (!magnetUri) {
    downloadBtn.disabled = true;
    downloadHint.textContent = "No magnet URI in catalog.";
  } else if (placeholder) {
    downloadBtn.disabled = true;
    downloadHint.textContent = "Placeholder magnet — not yet magnetized. Use CLI when available.";
  } else if (!getWebTorrentClient()) {
    downloadBtn.disabled = true;
    downloadHint.textContent = "WebTorrent failed to load from CDN.";
  } else {
    downloadHint.textContent =
      "Requires a live WebRTC/WSS seed or webseed. Plain CLI seeds may not connect to the browser.";
  }

  return { node, card, entry, magnetUri, placeholder, downloadBtn, downloadHint, progressWrap, copyBtn, magnetUriEl, livePeersEl };
}

/** Update progress UI for an active torrent on a card. */
function updateProgressUI(card, torrent) {
  const progressWrap = card.querySelector(".download-progress");
  const fill = /** @type {HTMLElement} */ (card.querySelector(".progress-fill"));
  const bar = /** @type {HTMLElement} */ (card.querySelector(".progress-bar"));
  const stats = card.querySelector(".progress-stats");
  const status = card.querySelector(".download-status");
  const livePeersEl = card.querySelector(".live-peers");

  progressWrap.hidden = false;

  const pct = torrent.length ? Math.min(100, (torrent.downloaded / torrent.length) * 100) : 0;
  fill.style.width = `${pct.toFixed(1)}%`;
  bar.setAttribute("aria-valuenow", String(Math.round(pct)));

  stats.textContent = `${pct.toFixed(1)}% · ${formatBytes(torrent.downloaded)} / ${formatBytes(torrent.length)} · ${torrent.numPeers} peer${torrent.numPeers === 1 ? "" : "s"} · ${formatSpeed(torrent.downloadSpeed)}`;

  livePeersEl.hidden = false;
  livePeersEl.textContent = `live: ${torrent.numPeers} peer${torrent.numPeers === 1 ? "" : "s"}`;

  if (torrent.done) {
    status.textContent = "Download complete.";
    status.classList.remove("download-status--warn");
    status.classList.add("download-status--ok");
  }
}

/** Render save-file links after torrent completes. */
function renderSaveLinks(card, torrent, modelId) {
  const list = card.querySelector(".save-links");
  list.replaceChildren();

  const active = activeDownloads.get(modelId);
  if (active?.blobUrls) {
    for (const url of active.blobUrls) {
      URL.revokeObjectURL(url);
    }
    active.blobUrls = [];
  }

  for (const file of torrent.files) {
    const li = document.createElement("li");
    const link = document.createElement("a");
    link.textContent = `Save ${file.name}`;
    link.className = "save-link";
    link.download = file.name;
    li.appendChild(link);
    list.appendChild(li);

    file.getBlobURL((err, url) => {
      if (err || !url) {
        link.textContent = `Save ${file.name} (failed)`;
        link.classList.add("save-link--error");
        return;
      }
      link.href = url;
      const activeEntry = activeDownloads.get(modelId);
      if (activeEntry) {
        if (!activeEntry.blobUrls) activeEntry.blobUrls = [];
        activeEntry.blobUrls.push(url);
      }
    });
  }
}

/** Wire download button and restore any in-flight download for this model. */
function setupDownloadHandlers(card, entry, magnetUri, placeholder, downloadBtn, downloadHint, progressWrap, copyBtn, magnetUriEl) {
  const modelId = entry.modelId || "";

  copyBtn.addEventListener("click", () => {
    if (magnetUri) copyToClipboard(magnetUri, copyBtn);
  });
  magnetUriEl.addEventListener("click", () => {
    if (magnetUri) copyToClipboard(magnetUri, copyBtn);
  });
  magnetUriEl.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      if (magnetUri) copyToClipboard(magnetUri, copyBtn);
    }
  });

  if (placeholder || !magnetUri) return;

  const existing = activeDownloads.get(modelId);
  if (existing?.torrent) {
    attachTorrentToCard(card, existing, downloadBtn, downloadHint);
    return;
  }

  downloadBtn.addEventListener("click", () => {
    startBrowserDownload(card, entry, magnetUri, downloadBtn, downloadHint, progressWrap);
  });
}

/** Bind torrent events to a card's progress UI. */
function attachTorrentToCard(card, active, downloadBtn, downloadHint) {
  const { torrent, modelId } = active;
  downloadBtn.disabled = true;
  downloadBtn.textContent = torrent.done ? "Downloaded" : "Downloading…";
  downloadHint.textContent = torrent.done
    ? "Files ready below."
    : "Downloading via WebTorrent (WebRTC/WSS peers only).";

  if (active.uiAttached) {
    updateProgressUI(card, torrent);
    if (torrent.done) renderSaveLinks(card, torrent, modelId);
    return;
  }
  active.uiAttached = true;

  const onUpdate = () => updateProgressUI(card, torrent);
  torrent.on("download", onUpdate);
  torrent.on("wire", onUpdate);
  onUpdate();

  if (torrent.done) {
    renderSaveLinks(card, torrent, modelId);
    return;
  }

  torrent.once("done", () => {
    updateProgressUI(card, torrent);
    renderSaveLinks(card, torrent, modelId);
    downloadBtn.textContent = "Downloaded";
    downloadHint.textContent = "Files ready below.";
    if (active.noPeersTimer) {
      clearTimeout(active.noPeersTimer);
      active.noPeersTimer = undefined;
    }
  });
}

/** Start a WebTorrent download for one catalog entry. */
function startBrowserDownload(card, entry, magnetUri, downloadBtn, downloadHint, progressWrap) {
  const client = getWebTorrentClient();
  if (!client) return;

  const modelId = entry.modelId || "";
  const statusEl = card.querySelector(".download-status");

  downloadBtn.disabled = true;
  downloadBtn.textContent = "Downloading…";
  downloadHint.textContent = "Connecting to swarm…";
  progressWrap.hidden = false;
  statusEl.textContent = "Searching for peers…";
  statusEl.classList.remove("download-status--ok", "download-status--warn");
  card.querySelector(".save-links").replaceChildren();

  /** @type {ActiveDownload} */
  const active = { torrent: null, modelId };
  activeDownloads.set(modelId, active);

  active.noPeersTimer = setTimeout(() => {
    if (!active.torrent || active.torrent.done) return;
    if (active.torrent.numPeers === 0) {
      statusEl.textContent = `No live seeders — try \`mt get ${modelId}\` from CLI`;
      statusEl.classList.add("download-status--warn");
      downloadHint.textContent =
        "Browser downloads need WebRTC/WSS peers or a WSS-capable webseed. A plain BitTorrent CLI seed won't connect without a WebSocket tracker.";
    }
  }, NO_PEERS_GRACE_MS);

  client.add(magnetUri, (torrent) => {
    active.torrent = torrent;
    attachTorrentToCard(card, active, downloadBtn, downloadHint);

    torrent.on("download", () => {
      if (torrent.numPeers > 0 && statusEl.classList.contains("download-status--warn")) {
        statusEl.textContent = "Connected — downloading…";
        statusEl.classList.remove("download-status--warn");
      }
    });
  }, (err) => {
    activeDownloads.delete(modelId);
    if (active.noPeersTimer) clearTimeout(active.noPeersTimer);
    downloadBtn.disabled = false;
    downloadBtn.textContent = "Download in browser";
    statusEl.textContent = err?.message || "Failed to start download.";
    statusEl.classList.add("download-status--warn");
    downloadHint.textContent = `Try \`mt get ${modelId}\` from CLI instead.`;
  });
}

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function renderEmptyState(error) {
  const results = document.getElementById("results");
  const msg = document.createElement("div");
  msg.className = "empty-state";
  msg.innerHTML = `
    <p><strong>No models in the catalog.</strong></p>
    <p>This static site reads <code>catalog.json</code> from the same directory.
       To populate demo entries:</p>
    <ol>
      <li>Generate tiny fixture files: <code>make fixtures</code></li>
      <li>Build or serve the catalog: <code>mt catalog</code> (when implemented) or use the bundled <code>web/catalog.json</code></li>
    </ol>
    <p class="empty-hint">${error === "missing" ? "Could not load catalog.json — are you serving from the <code>web/</code> directory?" : "The catalog loaded but contains zero models."}</p>
  `;
  results.replaceChildren(msg);
}

/** @param {CatalogEntry[]} models */
async function render(models) {
  const results = document.getElementById("results");
  results.replaceChildren();

  if (models.length === 0) return;

  const enriched = await Promise.all(
    models.map(async (entry) => {
      const manifestPath = entry.manifest || `models/${entry.modelId}/manifest.json`;
      const manifest = await fetchJSON(manifestPath);
      const files = Array.isArray(manifest?.files) ? manifest.files : [];
      const health = await fetchJSON(healthPath(entry.modelId));
      return { entry, files, health };
    })
  );

  for (const { entry, files, health } of enriched) {
    const rendered = renderCard(entry, files, health);
    results.appendChild(rendered.node);

    const cardEl = /** @type {HTMLElement} */ (results.querySelector(`[data-model-id="${CSS.escape(entry.modelId || "")}"]`));
    if (cardEl) {
      setupDownloadHandlers(
        cardEl,
        entry,
        rendered.magnetUri,
        rendered.placeholder,
        rendered.downloadBtn,
        rendered.downloadHint,
        rendered.progressWrap,
        rendered.copyBtn,
        rendered.magnetUriEl
      );
    }
  }
}

async function main() {
  const { models, error } = await loadCatalog();

  if (models.length === 0) {
    renderEmptyState(error);
    return;
  }

  await render(models);

  const search = document.getElementById("search");
  search.addEventListener("input", async () => {
    const filtered = filterModels(models, search.value);
    if (filtered.length === 0) {
      const results = document.getElementById("results");
      const msg = document.createElement("p");
      msg.className = "no-match";
      msg.textContent = `No models match “${search.value.trim()}”.`;
      results.replaceChildren(msg);
      return;
    }
    await render(filtered);
  });
}

if (typeof document !== "undefined") {
  document.addEventListener("DOMContentLoaded", main);
}
