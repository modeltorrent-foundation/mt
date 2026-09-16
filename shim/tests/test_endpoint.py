"""Tests for the Hugging Face-compatible shim."""

from __future__ import annotations

import json
import os
import shutil
import threading
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

import pytest

from modeltorrent_hf import DEFAULT_ENDPOINT, resolve_endpoint, snapshot_download


def test_default_endpoint_is_operated_bootstrap_domain():
    """DEFAULT_ENDPOINT must be a domain this project operates (no trailing slash)."""
    assert DEFAULT_ENDPOINT == "https://modeltorrent.org"
    assert not DEFAULT_ENDPOINT.endswith("/")


def test_resolve_endpoint_honors_hf_endpoint():
    assert resolve_endpoint({"HF_ENDPOINT": "https://mirror.example"}) == "https://mirror.example"


def test_snapshot_download_returns_local_dir(tmp_path, monkeypatch):
    monkeypatch.delenv("HF_ENDPOINT", raising=False)
    path = snapshot_download("Qwen/Qwen3-8B", cache_dir=str(tmp_path))
    assert os.path.isdir(path)
    snap = Path(path)
    assert (snap / "model.safetensors").exists()
    assert (snap / "tokenizer.json").exists()


def test_snapshot_download_via_hf_endpoint_http(tmp_path, monkeypatch):
    repo = Path(__file__).resolve().parents[2]
    src_fixtures = repo / "testdata" / "fixtures"
    src_manifest_path = repo / "testdata" / "models" / "Qwen" / "Qwen3-8B" / "manifest.json"
    if not src_fixtures.is_dir() or not src_manifest_path.is_file():
        pytest.skip("run `make fixtures` to generate testdata/fixtures")

    serve_root = tmp_path / "mirror"
    fixture_dest = serve_root / "fixtures" / "test-model"
    shutil.copytree(src_fixtures, fixture_dest)

    src_manifest = json.loads(src_manifest_path.read_text(encoding="utf-8"))
    manifest = {
        **src_manifest,
        "modelId": "test/model",
        "webseeds": ["/fixtures/test-model"],
    }
    manifest_path = serve_root / "models" / "test" / "model" / "manifest.json"
    manifest_path.parent.mkdir(parents=True, exist_ok=True)
    manifest_path.write_text(json.dumps(manifest), encoding="utf-8")

    catalog = {
        "schemaVersion": 1,
        "models": [
            {
                "modelId": "test/model",
                "manifest": "/models/test/model/manifest.json",
                "license": manifest["license"],
                "magnet": "fixture://test/model",
                "sizeBytes": 196608,
            }
        ],
    }
    (serve_root / "catalog.json").write_text(json.dumps(catalog), encoding="utf-8")

    class Handler(SimpleHTTPRequestHandler):
        def __init__(self, *args, **kwargs):
            super().__init__(*args, directory=str(serve_root), **kwargs)

    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    port = server.server_address[1]
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()

    try:
        base = f"http://127.0.0.1:{port}"
        monkeypatch.setenv("HF_ENDPOINT", base)
        monkeypatch.delenv("MT_CATALOG", raising=False)

        cache = tmp_path / "cache"
        path = snapshot_download("test/model", cache_dir=str(cache))
        snap = Path(path)

        assert snap.is_dir()
        assert (snap / "model.safetensors").is_file()
        assert (snap / "tokenizer.json").is_file()
        assert (snap / "Q4_K_M.gguf").is_file()

        layout_root = cache / "models--test--model"
        assert (layout_root / "refs" / "main").read_text(encoding="utf-8").strip() == "main"
        assert any((layout_root / "blobs").iterdir())
    finally:
        server.shutdown()
