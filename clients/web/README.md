# Slycrel Web Client

Live at **https://slycrel-legacy.pages.dev**.

Browser-only port of Slycrel, hosted on Cloudflare Pages with a small
Pages-Functions API backed by D1 for shared world state. Players log in
with a BBS name + password and share the world (Guild Lists, Inn, Arena
fights, mail) the way the original BBS did.

## One-time D1 setup

All `wrangler` commands must be run from `clients/web/` so wrangler picks
up the `wrangler.toml` in that directory.

```bash
cd clients/web
npx wrangler d1 create slycrel-legacy
# Paste the returned database_id into wrangler.toml.

# Production database:
npx wrangler d1 execute slycrel-legacy --remote --file=migrations/0001_init.sql

# Local emulator (for `wrangler pages dev`):
npx wrangler d1 execute slycrel-legacy --local --file=migrations/0001_init.sql
```

## Run locally

`wrangler pages dev` serves the static files **and** runs the Functions
locally (with a SQLite emulator for D1). The client fetches `/data/...`
at absolute paths, so copy `data/` into the dev tree first.

```bash
# from the repo root:
cp -r data clients/web/data
cd clients/web
npx wrangler pages dev .
# open http://localhost:8788/
```

The plain `python3 -m http.server` no longer works — the API endpoints
won't be served.

## Layout

```
clients/web/
  index.html      Entry HTML
  styles.css      Terminal-style theme (monospace, line-height: 1.0)
  app.js          Bootstrap — mounts the root and starts the session
  wrangler.toml   Pages config + D1 binding
  migrations/     D1 schema (one file per version)
  functions/      Pages Functions (server-side endpoints)
    _lib/           auth (PBKDF2 + session cookie), json helpers
    api/            REST endpoints: login, me, character, inn, fights, …
  ui/             IO abstraction (mirrors internal/io)
  game/           State machine (mirrors internal/game)
  model/          Plain-data schemas (mirrors internal/model)
  mechanics/      Pure game logic (mirrors internal/mechanics)
  store/
    remote.js     Wraps /api/* — used by every state
```

## Cloudflare Pages deploy

`deploy.sh` at the repo root stages `data/` into `clients/web/data/` and
runs `wrangler pages deploy clients/web --project-name slycrel-legacy`.
First run creates the project at `slycrel-legacy.pages.dev`.

```bash
# from the repo root:
./deploy.sh
```

D1 binding lives in `wrangler.toml` (`binding = "DB"`); Functions reach it
as `env.DB`. The project is hard-coded to `slycrel-legacy` so we don't
accidentally publish to a different project from a stray flag.

## Auth model

`POST /api/login { bbsName, password }` does both login and first-time
registration — a previously-unused BBS name is claimed by whoever logs in
first (the supplied password becomes the lock). Sessions are stored
server-side in the `sessions` table, identified by an httpOnly cookie.

All shared-world writes (anyone's character, inn, mail) are accepted from
any authenticated user. The bar for anti-cheat is intentionally low — this
is a hobby BBS recreation, not a leaderboard game.
