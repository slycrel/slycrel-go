#!/usr/bin/env bash
# Deploy the Slycrel web client to Cloudflare Pages.
#
# Usage: ./deploy.sh [extra wrangler args]
#
# Requires wrangler (auto-installs via npx) and a one-time `wrangler login`.
# First run creates the "slycrel" Pages project; subsequent runs deploy
# new versions. Extra args are forwarded to wrangler — e.g. pass
# `--branch=preview` to deploy a preview environment.

set -euo pipefail

cd "$(dirname "$0")"

# Stage the deploy: clients/web/ is the deploy root, but the JS fetches
# /data/... at absolute paths, so copy /data alongside index.html.
rm -rf clients/web/data
cp -r data clients/web/data

# Clean up the staged copy on exit even if wrangler fails.
trap 'rm -rf clients/web/data' EXIT

npx wrangler pages deploy clients/web --project-name slycrel "$@"
