# Slycrel Web Client

Browser-only port of Slycrel, targeting Cloudflare Pages for hosting. No
backend — all game state lives in the browser (`localStorage`).

## Run locally

The client fetches the BBS-era `.ans` screens from `/data/ansi/...`, so the
dev server has to serve the repo root (not just `clients/web/`):

```bash
# From the repo root:
python3 -m http.server 8000
# then open http://localhost:8000/clients/web/
```

Any static file server works — `npx serve .`, `caddy file-server`, etc.

## Layout

```
clients/web/
  index.html      Entry HTML
  styles.css      Terminal-style theme (monospace, line-height: 1.0)
  app.js          Bootstrap — mounts the root and starts the session
  ui/             IO abstraction (mirrors internal/io)
    ansi.js         ansi_up wrapper (renders ANSI escapes to HTML)
    io.js           clear / println / showAnsiFile / lettersPrompt
  game/           State machine (mirrors internal/game)
    session.js      Session + state-machine driver
    states/         One file per state (begin.js, town.js, ...)
  model/          Plain-data schemas (mirrors internal/model)  -- TODO
  mechanics/      Pure game logic (mirrors internal/mechanics) -- TODO
  store/          localStorage save/load                       -- TODO
```

`ansi_up` is loaded from jsdelivr's ESM CDN for now — see `ui/ansi.js`. If
we ever want offline-safe deploys, vendor the file under `vendor/` and
swap the import.

## Cloudflare Pages deploy

The client fetches `data/ansi/*`, `data/monsters/*`, `data/terrain/*`, etc.
at absolute `/data/...` paths. For local dev that resolves naturally
because we serve from the repo root. For Cloudflare Pages, the build
needs to copy `data/` next to `index.html` so the same absolute paths
resolve from the deploy root.

**Dashboard setup** (one-time):
1. Connect the repo to a new Pages project.
2. Production branch: `webapp` (or whichever you merge into).
3. Build command: `cp -r data clients/web/data`
4. Build output directory: `clients/web`
5. Deploy. Subsequent pushes auto-build.

**CLI deploy** (no GitHub integration needed):
```bash
cp -r data clients/web/data
npx wrangler pages deploy clients/web --project-name slycrel
```

Cloudflare Pages free tier covers any realistic traffic for this hobby
game — no Workers Paid required since there's no server.
