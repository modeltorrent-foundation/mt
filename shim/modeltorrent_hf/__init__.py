"""Hugging Face-compatible resolver for Model Torrent.

Set ``HF_ENDPOINT`` to point existing ``huggingface_hub`` / ``transformers`` code
at a Model Torrent catalog mirror. Wave 2 resolves manifests from catalog.json and
downloads via fixture webseeds or the ``mt`` CLI when available.
"""

from __future__ import annotations

import hashlib
import json
import os
import shutil
import subprocess
import tempfile
import urllib.error
import urllib.request
from pathlib import Path
from typing import Iterable

__all__ = ["resolve_endpoint", "snapshot_download", "DEFAULT_ENDPOINT"]

# Bootstrap catalog pointer we operate (Cloudflare Pages custom hostname).
# Mirrors: https://modeltorrent.pages.dev and https://modeltorrent-foundation.github.io/mt
# Override with HF_ENDPOINT.
DEFAULT_ENDPOINT = "https://modeltorrent.org"
_USER_AGENT = "modeltorrent-hf/0.2"


def resolve_endpoint(env: dict | None = None) -> str:
    """Return the catalog base URL, honoring ``HF_ENDPOINT`` when set."""
    env = os.environ if env is None else env
    base = env.get("HF_ENDPOINT", DEFAULT_ENDPOINT)
    return base.rstrip("/")


def _http_mode_active(env: dict | None = None, endpoint: str | None = None) -> bool:
    """True when downloads should use HTTP catalog + webseeds (explicit HF_ENDPOINT)."""
    if endpoint is not None:
        return endpoint.startswith(("http://", "https://"))
    env = os.environ if env is None else env
    hf = env.get("HF_ENDPOINT")
    if not hf:
        return False
    return hf.startswith(("http://", "https://"))


def _catalog_base(env: dict | None = None, endpoint: str | None = None) -> str:
    if endpoint is not None:
        return endpoint.rstrip("/")
    env = os.environ if env is None else env
    return env.get("HF_ENDPOINT", DEFAULT_ENDPOINT).rstrip("/")


def _find_catalog_path() -> Path | None:
    if p := os.environ.get("MT_CATALOG"):
        path = Path(p)
        if path.is_file():
            return path
    here = Path(__file__).resolve()
    for parent in [here.parent, *here.parents]:
        candidate = parent / "testdata" / "catalog.json"
        if candidate.is_file():
            return candidate
        if (parent / "go.mod").is_file():
            repo_candidate = parent / "testdata" / "catalog.json"
            if repo_candidate.is_file():
                return repo_candidate
    cwd = Path.cwd()
    for candidate in [cwd / "testdata" / "catalog.json", cwd / "catalog.json"]:
        if candidate.is_file():
            return candidate
    return None


def _load_catalog_local(catalog_path: Path) -> dict:
    with catalog_path.open(encoding="utf-8") as f:
        return json.load(f)


def _http_json(url: str) -> dict:
    req = urllib.request.Request(url, headers={"User-Agent": _USER_AGENT})
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.load(resp)


def _fetch_catalog_remote(base: str) -> dict:
    return _http_json(f"{base.rstrip('/')}/catalog.json")


def _load_manifest(catalog_root: Path, manifest_path: str) -> dict:
    rel = manifest_path.lstrip("/")
    local = catalog_root / rel
    if local.is_file():
        with local.open(encoding="utf-8") as f:
            return json.load(f)
    raise FileNotFoundError(f"manifest not found: {local}")


def _fetch_manifest_remote(base: str, manifest_path: str) -> dict:
    path = manifest_path if manifest_path.startswith("/") else f"/{manifest_path}"
    return _http_json(f"{base.rstrip('/')}{path}")


def _find_model(catalog: dict, repo_id: str) -> dict | None:
    for entry in catalog.get("models", []):
        if entry.get("modelId") == repo_id:
            return entry
    return None


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(65536), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _verify_sources(manifest: dict, sources: dict[str, Path]) -> None:
    for file_entry in manifest.get("files", []):
        rel_path = file_entry["path"]
        expected = file_entry["sha256"].lower()
        src = sources[rel_path]
        actual = _sha256_file(src)
        if actual != expected:
            raise RuntimeError(
                f"sha256 mismatch for {rel_path}: expected {expected}, got {actual}"
            )


def _write_hf_layout(dest: Path, manifest: dict, blob_sources: dict[str, Path]) -> Path:
    """Materialize blobs/ + snapshots/main/ under dest (mirrors internal/hflayout)."""
    revision = "main"
    blobs_dir = dest / "blobs"
    snap_root = dest / "snapshots" / revision
    blobs_dir.mkdir(parents=True, exist_ok=True)
    snap_root.mkdir(parents=True, exist_ok=True)

    _verify_sources(manifest, blob_sources)

    for file_entry in manifest.get("files", []):
        rel_path = file_entry["path"]
        sha = file_entry["sha256"]
        src = blob_sources[rel_path]
        blob_dest = blobs_dir / sha
        if not blob_dest.exists():
            shutil.copy2(src, blob_dest)
        snap_path = snap_root / rel_path
        snap_path.parent.mkdir(parents=True, exist_ok=True)
        if snap_path.exists() or snap_path.is_symlink():
            snap_path.unlink()
        try:
            os.symlink(os.path.relpath(blob_dest, snap_path.parent), snap_path)
        except OSError:
            shutil.copy2(blob_dest, snap_path)

    refs_dir = dest / "refs"
    refs_dir.mkdir(parents=True, exist_ok=True)
    (refs_dir / "main").write_text(revision, encoding="utf-8")
    return snap_root


def _fixture_root_from_webseeds(webseeds: Iterable[str], catalog_root: Path | None) -> Path | None:
    for ws in webseeds:
        if ws.startswith("file://fixtures/") and catalog_root is not None:
            rel = ws.removeprefix("file://")
            root = catalog_root / rel
            if root.is_dir():
                return root
        if ws.startswith("file://"):
            raw = ws.removeprefix("file://")
            root = Path(raw if raw.startswith("/") else f"/{raw}")
            if root.is_dir():
                return root
    return None


def _copy_from_fixture_webseed(manifest: dict, catalog_root: Path) -> dict[str, Path]:
    webseeds = manifest.get("webseeds") or []
    fixture_root = _fixture_root_from_webseeds(webseeds, catalog_root)
    if fixture_root is None or not fixture_root.is_dir():
        raise RuntimeError("no file:// fixture webseed in manifest")

    sources: dict[str, Path] = {}
    for file_entry in manifest.get("files", []):
        rel = file_entry["path"]
        src = fixture_root / rel
        if not src.is_file():
            raise FileNotFoundError(f"fixture file missing: {src}")
        sources[rel] = src
    return sources


def _file_download_urls(base: str, webseeds: list[str], file_path: str) -> list[str]:
    """Candidate HTTP URLs for a manifest file, most preferred first."""
    rel_path = file_path.lstrip("/")
    urls: list[str] = []

    for ws in webseeds:
        if ws.startswith(("http://", "https://")):
            urls.append(f"{ws.rstrip('/')}/{rel_path}")
        elif ws.startswith("file://fixtures/"):
            rel = ws.removeprefix("file://").lstrip("/")
            urls.append(f"{base.rstrip('/')}/{rel}/{rel_path}")
        elif ws.startswith("/"):
            urls.append(f"{base.rstrip('/')}{ws.rstrip('/')}/{rel_path}")
        elif ws and "://" not in ws:
            urls.append(f"{base.rstrip('/')}/{ws.strip('/')}/{rel_path}")

    urls.append(f"{base.rstrip('/')}/{rel_path}")
    return urls


def _download_http_verified(url: str, expected_sha256: str, dest: Path) -> Path:
    req = urllib.request.Request(url, headers={"User-Agent": _USER_AGENT})
    with urllib.request.urlopen(req, timeout=120) as resp:
        data = resp.read()
    actual = hashlib.sha256(data).hexdigest()
    expected = expected_sha256.lower()
    if actual != expected:
        raise RuntimeError(
            f"sha256 mismatch for {url}: expected {expected}, got {actual}"
        )
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(data)
    return dest


def _fetch_sources_http(base: str, manifest: dict, staging: Path) -> dict[str, Path]:
    webseeds = manifest.get("webseeds") or []
    sources: dict[str, Path] = {}

    for file_entry in manifest.get("files", []):
        rel = file_entry["path"]
        sha = file_entry["sha256"]
        staging_path = staging / sha
        if staging_path.is_file() and _sha256_file(staging_path) == sha.lower():
            sources[rel] = staging_path
            continue

        last_err: Exception | None = None
        for url in _file_download_urls(base, webseeds, rel):
            try:
                sources[rel] = _download_http_verified(url, sha, staging_path)
                last_err = None
                break
            except (urllib.error.URLError, urllib.error.HTTPError, TimeoutError, RuntimeError) as exc:
                last_err = exc
        if last_err is not None:
            raise RuntimeError(f"failed to download {rel!r} from {base}: {last_err}") from last_err

    return sources


def _try_mt_get(repo_id: str, dest: str) -> bool:
    mt_bin = os.environ.get("MT_BIN", "mt")
    try:
        proc = subprocess.run(
            [mt_bin, "get", repo_id, dest, "--no-seed"],
            capture_output=True,
            text=True,
            timeout=120,
            check=False,
        )
        return proc.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return False


def snapshot_download(
    repo_id: str,
    *,
    revision: str = "main",
    cache_dir: str | None = None,
    endpoint: str | None = None,
) -> str:
    """Download a model by ``repo_id`` and return the local snapshot directory."""
    _ = revision  # content-addressed; revision is always main in Wave 2
    base = _catalog_base(endpoint=endpoint)
    cache = Path(cache_dir or os.path.join(os.path.expanduser("~"), ".cache", "modeltorrent"))
    cache.mkdir(parents=True, exist_ok=True)

    layout_root = cache / f"models--{repo_id.replace('/', '--')}"
    snap_dir = layout_root / "snapshots" / "main"
    if snap_dir.is_dir() and any(snap_dir.iterdir()):
        return str(snap_dir)

    if _http_mode_active(endpoint=endpoint):
        try:
            catalog = _fetch_catalog_remote(base)
        except (urllib.error.URLError, urllib.error.HTTPError, TimeoutError, json.JSONDecodeError) as exc:
            raise RuntimeError(f"could not load catalog from {base}: {exc}") from exc

        entry = _find_model(catalog, repo_id)
        if entry is None:
            raise RuntimeError(f"model {repo_id!r} not found in catalog")

        manifest_path = entry.get("manifest", "")
        manifest = _fetch_manifest_remote(base, manifest_path)

        if _try_mt_get(repo_id, str(layout_root)):
            if snap_dir.is_dir() and any(snap_dir.iterdir()):
                return str(snap_dir)

        with tempfile.TemporaryDirectory(prefix="modeltorrent-hf-") as tmp:
            sources = _fetch_sources_http(base, manifest, Path(tmp))
            snap = _write_hf_layout(layout_root, manifest, sources)
        return str(snap)

    catalog_path = _find_catalog_path()
    if catalog_path is None:
        raise RuntimeError(
            "no local catalog found; set HF_ENDPOINT to an http(s) catalog mirror "
            "or MT_CATALOG / testdata/catalog.json"
        )

    catalog = _load_catalog_local(catalog_path)
    catalog_root = catalog_path.parent

    entry = _find_model(catalog, repo_id)
    if entry is None:
        raise RuntimeError(f"model {repo_id!r} not found in catalog")

    manifest_path = entry.get("manifest", "")
    manifest = _load_manifest(catalog_root, manifest_path)

    if _try_mt_get(repo_id, str(layout_root)):
        if snap_dir.is_dir() and any(snap_dir.iterdir()):
            return str(snap_dir)

    sources = _copy_from_fixture_webseed(manifest, catalog_root)
    snap = _write_hf_layout(layout_root, manifest, sources)
    return str(snap)
