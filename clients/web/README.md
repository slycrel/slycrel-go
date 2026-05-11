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
