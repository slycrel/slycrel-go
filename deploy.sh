#!/usr/bin/env bash
# Deploy the Slycrel web client to Cloudflare Pages.
#
# Usage: ./deploy.sh [extra wrangler args]
#
# Requires wrangler (auto-installs via npx) and a one-time `wrangler login`.
# First run creates the "slycrel-legacy" Pages project; subsequent runs deploy
# new versions. Extra args are forwarded to wrangler — e.g. pass
# `--branch=preview` to deploy a preview environment.
#
# Prerequisites for the D1-backed shared world:
#   cd clients/web
#   wrangler d1 create slycrel-legacy
#   # paste returned database_id into wrangler.toml
#   wrangler d1 execute slycrel-legacy --remote --file=migrations/0001_init.sql

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$REPO_ROOT"

# Stage the deploy: clients/web/ is the deploy root, but the JS fetches
# /data/... at absolute paths, so copy /data alongside index.html.
rm -rf clients/web/data
cp -r data clients/web/data

# Clean up the staged copy on exit even if wrangler fails (absolute path
# so cd later in this script doesn't break the trap).
trap "rm -rf '$REPO_ROOT/clients/web/data'" EXIT

# wrangler must run from the deploy directory so it finds wrangler.toml
# (and the Functions in ./functions/). Otherwise the Functions bundle is
# silently skipped and /api/* fall through to the static index.html.
cd "$REPO_ROOT/clients/web"
npx wrangler pages deploy . --commit-dirty=true "$@"
