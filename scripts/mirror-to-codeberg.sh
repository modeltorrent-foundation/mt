#!/usr/bin/env bash
# Push a second git remote of this catalog/repo to Codeberg.
# Secrets stay in the environment; nothing here is committed.
set -euo pipefail

REMOTE_NAME="${CODEBERG_REMOTE:-codeberg}"
# Override after you create the repo, e.g.
#   git@codeberg.org:modeltorrent-foundation/mt.git
#   https://codeberg.org/modeltorrent-foundation/mt.git
CODEBERG_URL="${CODEBERG_URL:-git@codeberg.org:modeltorrent-foundation/mt.git}"

if git remote get-url "$REMOTE_NAME" >/dev/null 2>&1; then
  git remote set-url "$REMOTE_NAME" "$CODEBERG_URL"
else
  git remote add "$REMOTE_NAME" "$CODEBERG_URL"
fi

echo "Pushing HEAD to ${REMOTE_NAME} (${CODEBERG_URL})"
git push --mirror "$REMOTE_NAME"
