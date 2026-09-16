#!/usr/bin/env bash
# Install a memory-capped WebTorrent-hybrid (WebRTC/WSS) seeder on the durable
# host. Complements mt-seed@* (TCP/UDP). Does NOT seed Qwen3-8B — that file is
# too large for webtorrent RSS on a 3.7Gi shared VPS.
#
#   MT_SEED_HOST=hetzner ./deploy/hetzner/deploy_hybrid.sh
set -euo pipefail

HOST="${MT_SEED_HOST:?set MT_SEED_HOST to an ssh target (e.g. hetzner)}"
SUDO="${MT_SSH_SUDO:-}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REMOTE_ROOT="/opt/model-torrent"
NODE_VER="${MT_NODE_VER:-v22.18.0}"
NODE_TARBALL="node-${NODE_VER}-linux-x64.tar.xz"

echo ">> installing node ${NODE_VER} under ${REMOTE_ROOT}/node (if missing)"
ssh "$HOST" "set -euo pipefail
if [[ -x $REMOTE_ROOT/node/bin/node ]]; then
  echo '   node already present:' \$($REMOTE_ROOT/node/bin/node -v)
  exit 0
fi
$SUDO mkdir -p $REMOTE_ROOT
tmp=\$(mktemp -d)
trap 'rm -rf \"\$tmp\"' EXIT
curl -fsSL -o \"\$tmp/$NODE_TARBALL\" \"https://nodejs.org/dist/${NODE_VER}/$NODE_TARBALL\"
tar -xJf \"\$tmp/$NODE_TARBALL\" -C \"\$tmp\"
$SUDO rm -rf $REMOTE_ROOT/node
$SUDO mv \"\$tmp/node-${NODE_VER}-linux-x64\" $REMOTE_ROOT/node
echo '   installed' \$($REMOTE_ROOT/node/bin/node -v)
"

echo ">> uploading hybrid seeder + catalog torrents"
ssh "$HOST" "$SUDO mkdir -p $REMOTE_ROOT/hybrid/torrents"
scp -q "$REPO_ROOT/deploy/hetzner/webtorrent-hybrid/package.json" "$REPO_ROOT/deploy/hetzner/webtorrent-hybrid/seed.js" "$HOST:/tmp/"
ssh "$HOST" "$SUDO install -m 0644 /tmp/package.json $REMOTE_ROOT/hybrid/package.json && $SUDO install -m 0755 /tmp/seed.js $REMOTE_ROOT/hybrid/seed.js && rm -f /tmp/package.json /tmp/seed.js"

copy_torrent() {
  local slug="$1" rel="$2"
  scp -q "$REPO_ROOT/web/models/$rel/publish.torrent" "$HOST:/tmp/${slug}.torrent"
  ssh "$HOST" "$SUDO install -m 0644 /tmp/${slug}.torrent $REMOTE_ROOT/hybrid/torrents/${slug}.torrent && rm -f /tmp/${slug}.torrent"
}
copy_torrent qwen3-0.6b "Qwen/Qwen3-0.6B-GGUF"
copy_torrent smollm2-360m "HuggingFaceTB/SmolLM2-360M-Instruct-GGUF"
copy_torrent qwen2.5-0.5b "Qwen/Qwen2.5-0.5B-Instruct-GGUF"

echo ">> npm install (native wrtc, on the target)"
ssh "$HOST" "set -euo pipefail
export PATH=$REMOTE_ROOT/node/bin:\$PATH
cd $REMOTE_ROOT/hybrid
npm install --omit=dev --no-fund --no-audit
"

echo ">> installing systemd unit"
scp -q "$REPO_ROOT/deploy/hetzner/mt-webtorrent-hybrid.service" "$HOST:/tmp/mt-webtorrent-hybrid.service"
ssh "$HOST" "$SUDO install -m 0644 /tmp/mt-webtorrent-hybrid.service /etc/systemd/system/mt-webtorrent-hybrid.service && rm -f /tmp/mt-webtorrent-hybrid.service && $SUDO systemctl daemon-reload && $SUDO systemctl enable --now mt-webtorrent-hybrid"

echo ">> status"
ssh "$HOST" "$SUDO systemctl --no-pager --plain status mt-webtorrent-hybrid | head -25; echo; $SUDO journalctl -u mt-webtorrent-hybrid -n 30 --no-pager"
echo ">> done."
