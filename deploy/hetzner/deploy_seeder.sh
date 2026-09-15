#!/usr/bin/env bash
# Deploy the Model Torrent always-on seeder to a durable, non-sleeping host over
# SSH. Cross-compiles the mt CLI for the target, uploads the model file(s) that
# are already seeding locally (or downloaded into a staging dir), installs a
# systemd template unit, and enables one seeder instance per model.
#
# The host is NEVER hardcoded (this repo is public). Pass it via env:
#
#   MT_SEED_HOST=my-vps ./deploy/hetzner/deploy_seeder.sh
#
# where MT_SEED_HOST is an ssh(1) target (an alias from ~/.ssh/config or
# user@host). The remote user must be able to write /opt and manage systemd
# (root, or via sudo — set MT_SSH_SUDO=sudo).
#
# Model set + ports are declared in MODELS below (slug:file:port). The staging
# dir (MT_STAGING, default .work/dl) must contain <staging>/<slug>/<file>.
set -euo pipefail

HOST="${MT_SEED_HOST:?set MT_SEED_HOST to an ssh target (e.g. my-vps)}"
SUDO="${MT_SSH_SUDO:-}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${MT_STAGING:-$REPO_ROOT/.work/dl}"
REMOTE_ROOT="/opt/model-torrent"

# slug : filename-in-staging : fixed BitTorrent listen port
MODELS=(
  "qwen3-0.6b:Qwen3-0.6B-Q8_0.gguf:42413"
  "smollm2-360m:smollm2-360m-instruct-q8_0.gguf:42414"
  "qwen2.5-0.5b:qwen2.5-0.5b-instruct-q4_k_m.gguf:42415"
)

echo ">> cross-compiling mt for linux/amd64"
BIN="$REPO_ROOT/.work/mt-linux-amd64"
( cd "$REPO_ROOT" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o "$BIN" ./cmd/mt )

echo ">> preparing remote layout on $HOST"
ssh "$HOST" "$SUDO mkdir -p $REMOTE_ROOT/bin $REMOTE_ROOT/seed $REMOTE_ROOT/env"

echo ">> uploading mt binary"
scp -q "$BIN" "$HOST:/tmp/mt.new"
ssh "$HOST" "$SUDO install -m 0755 /tmp/mt.new $REMOTE_ROOT/bin/mt && rm -f /tmp/mt.new && $REMOTE_ROOT/bin/mt --help >/dev/null && echo '   mt installed:' \$($REMOTE_ROOT/bin/mt --help | head -1)"

echo ">> uploading models + writing per-instance env"
for entry in "${MODELS[@]}"; do
  IFS=':' read -r slug file port <<<"$entry"
  src="$STAGING/$slug/$file"
  if [[ ! -f "$src" ]]; then
    echo "!! missing $src — skipping $slug" >&2
    continue
  fi
  echo "   - $slug ($file) -> port $port"
  ssh "$HOST" "$SUDO mkdir -p $REMOTE_ROOT/seed/$slug"
  scp -q "$src" "$HOST:/tmp/$file"
  # A seed dir must contain exactly one file (single-file torrent name/infohash).
  ssh "$HOST" "$SUDO rm -f $REMOTE_ROOT/seed/$slug/* 2>/dev/null; $SUDO mv /tmp/$file $REMOTE_ROOT/seed/$slug/$file"
  printf 'MT_LISTEN_PORT=%s\n' "$port" | ssh "$HOST" "$SUDO tee $REMOTE_ROOT/env/$slug.env >/dev/null"
done

echo ">> installing systemd unit"
scp -q "$REPO_ROOT/deploy/hetzner/mt-seed@.service" "$HOST:/tmp/mt-seed@.service"
ssh "$HOST" "$SUDO install -m 0644 /tmp/mt-seed@.service /etc/systemd/system/mt-seed@.service && rm -f /tmp/mt-seed@.service && $SUDO systemctl daemon-reload"

echo ">> enabling + starting seeders (auto-restart, survives reboot)"
for entry in "${MODELS[@]}"; do
  IFS=':' read -r slug file port <<<"$entry"
  [[ -f "$STAGING/$slug/$file" ]] || continue
  ssh "$HOST" "$SUDO systemctl enable --now mt-seed@$slug"
done

echo ">> status"
ssh "$HOST" "$SUDO systemctl --no-pager --plain list-units 'mt-seed@*' | head -20; echo; ss -tulnp 2>/dev/null | grep -E ':(4241[0-9])' || true"
echo ">> done."
