#!/usr/bin/env node
// Seed existing catalog torrents over WebRTC + WSS so a browser WebTorrent
// client can find a peer. Files stay on disk; we never create a new infohash.
// webtorrent 1.9.7 is the hybrid equivalent: webtorrent@2 + parse-torrent@11
// crash on Uint8Array vs hex infoHash (arr2hex(undefined)).
'use strict'

const fs = require('fs')
const path = require('path')
const WebTorrent = require('webtorrent')
const wrtc = require('@roamhq/wrtc')

const SEED_ROOT = process.env.MT_SEED_ROOT || '/opt/model-torrent'
const WSS_TRACKERS = [
  'wss://tracker.openwebtorrent.com',
  'wss://tracker.webtorrent.dev',
]

// Three small Apache-2.0 GGUFs only. Qwen3-8B (~4.68GiB) stays on mt-seed + R2:
// loading it into webtorrent on a 3.7Gi shared VPS would OOM LinkedIn Chrome.
const MODELS = [
  {
    id: 'Qwen/Qwen3-0.6B-GGUF',
    slug: 'qwen3-0.6b',
    infoHash: '23bbb1ba4c30d3c844e753182fabedd528983511',
  },
  {
    id: 'HuggingFaceTB/SmolLM2-360M-Instruct-GGUF',
    slug: 'smollm2-360m',
    infoHash: 'cf34e9c6295bfc2f6f43bbf040537a16a5658d05',
  },
  {
    id: 'Qwen/Qwen2.5-0.5B-Instruct-GGUF',
    slug: 'qwen2.5-0.5b',
    infoHash: 'b030b1ceb2459628be5e575481df189ba7c71d6b',
  },
]

function torrentPath(slug) {
  return path.join(SEED_ROOT, 'hybrid', 'torrents', slug + '.torrent')
}

function seedDir(slug) {
  return path.join(SEED_ROOT, 'seed', slug)
}

function log(...args) {
  fs.writeSync(1, args.map(String).join(' ') + '\n')
}

const client = new WebTorrent({
  wrtc,
  maxConns: 20,
  dht: false,
  utp: false,
  webSeeds: false,
})

client.on('error', (err) => {
  fs.writeSync(2, 'webtorrent error: ' + err + '\n')
})

log('hybrid starting', MODELS.length, 'torrents')

let ready = 0
for (const model of MODELS) {
  const torrentFile = torrentPath(model.slug)
  const dir = seedDir(model.slug)
  if (!fs.existsSync(torrentFile)) {
    fs.writeSync(2, 'missing torrent: ' + torrentFile + '\n')
    process.exit(1)
  }
  if (!fs.existsSync(dir)) {
    fs.writeSync(2, 'missing seed dir: ' + dir + '\n')
    process.exit(1)
  }
  const buf = fs.readFileSync(torrentFile)
  log('adding', model.id, 'bytes', buf.length)
  client.add(
    buf,
    {
      path: dir,
      announce: WSS_TRACKERS,
    },
    (torrent) => {
      const got = String(torrent.infoHash).toLowerCase()
      const want = model.infoHash.toLowerCase()
      if (got !== want) {
        fs.writeSync(2, 'infohash mismatch for ' + model.id + ' got ' + got + ' want ' + want + '\n')
        process.exit(1)
      }
      torrent.on('error', (err) => {
        fs.writeSync(2, 'torrent error ' + model.id + ' ' + err + '\n')
      })
      log(
        'seeding',
        model.id,
        'btih',
        got,
        'peers',
        torrent.numPeers,
        'files',
        torrent.files.map((f) => f.name).join(','),
      )
      ready++
      if (ready === MODELS.length) {
        log('hybrid ready:', ready, 'torrents (WebRTC/WSS)')
      }
    },
  )
}

setInterval(() => {
  const rows = client.torrents.map((t) => {
    return String(t.infoHash).slice(0, 8) + ':' + t.numPeers + 'p'
  })
  log('peers', rows.join(' ') || '(none yet)')
}, 60000).unref()

function shutdown() {
  client.destroy(() => process.exit(0))
}
process.on('SIGTERM', shutdown)
process.on('SIGINT', shutdown)
