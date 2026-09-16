# Catalog git mirror (Codeberg)

The catalog is git. GitHub is one remote, not the identity. A second forge in a
different jurisdiction is a preservation act ([GOVERNANCE.md §9](../GOVERNANCE.md)).

## If you already have a Codeberg account

1. Create the org `modeltorrent-foundation` (or a user-owned repo) on
   [codeberg.org](https://codeberg.org).
2. Create an empty public repository named `mt` (no README, no license file —
   this push is a mirror).
3. Add an SSH key or a [Codeberg access token](https://codeberg.org/user/settings/applications)
   with repo write. Do **not** commit the token.
4. From a clone of this repo:

```bash
export CODEBERG_URL=git@codeberg.org:modeltorrent-foundation/mt.git
./scripts/mirror-to-codeberg.sh
```

HTTPS with a token (token is only in the env, never in git):

```bash
export CODEBERG_URL="https://${CODEBERG_USER}:${CODEBERG_TOKEN}@codeberg.org/modeltorrent-foundation/mt.git"
./scripts/mirror-to-codeberg.sh
```

`tea` (the Codeberg CLI) can create the empty repo if you are already logged in:

```bash
tea login add --url https://codeberg.org --token "$CODEBERG_TOKEN"
tea repos create --name mt --owner modeltorrent-foundation --private=false
```

Keep the mirror in sync with a scheduled `git push --mirror` (GitHub Action
`workflow_dispatch` or a cron on a box you control). This script is
intentionally not a GitHub Action: putting a Codeberg token in Actions secrets
is fine later, but the first mirror should be a human-owned credential.

## If you do not have a Codeberg account yet

1. Open https://codeberg.org/user/sign_up and register (no phone number).
2. Confirm email, add an SSH key under **Settings → SSH / GPG Keys**.
3. Create org + empty `mt` repo as above, then run the script.

GitLab (`glab`) is an acceptable second choice if Codeberg is blocked for you:
create `modeltorrent-foundation/mt` on gitlab.com and point `CODEBERG_URL` at
that HTTPS/SSH remote (the script name is historical). Do not make Hugging Face
a catalog remote.

## What this is not

- Not a webseed. Weights stay on BitTorrent + R2 (`deploy/r2-webseeds/`).
- Not a domain. Humans can clone `https://codeberg.org/.../mt` if GitHub is
  down; magnets still work with no git host at all.
