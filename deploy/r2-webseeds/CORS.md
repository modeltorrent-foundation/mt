# CORS for browser WebTorrent

WebTorrent in the browser issues cross-origin `GET` / `Range` requests against
BEP-19 webseeds. Without CORS headers those fetches are silently blocked, so
the catalog button hangs even though `curl` against the same URL returns 200.

Rules live in [`cors.json`](./cors.json) (GET/HEAD, `Range`, expose
`Content-Length` / `Content-Range`). Apply after creating the public bucket:

```bash
npx wrangler r2 bucket cors set mt-webseeds --file deploy/r2-webseeds/cors.json --force
npx wrangler r2 bucket cors list mt-webseeds
```

Verify from a browser origin (not just `curl -I`):

```bash
curl -sI -H 'Origin: https://modeltorrent.org' \
  "$MT_WEBSEED_BASE/demo/tiny-gguf/Q4_K_M.gguf" \
  | grep -iE 'HTTP/|access-control-allow-origin|accept-ranges'
```

`Access-Control-Allow-Origin: *` (or the Pages origin) plus `Accept-Ranges:
bytes` is the signal the catalog UI can fetch over HTTP. A memory-capped
`mt-webtorrent-hybrid` unit on Hetzner additionally seeds the three small
GGUFs over WebRTC/WSS; 8B stays on R2 + `mt-seed`.
