# Quick start

End-to-end smoke test: run the Go server with dev-mode auto-registration,
connect from the Godot client, log in as a new user, walk through
character creation, and try wilderness / grid combat.

## 1. Start the server

From the repo root:

```bash
go run ./cmd/server --port 8080 --data-dir ./data --state-dir ./state --allow-new-users
```

`--allow-new-users` auto-creates a blank character for any unknown
username at connect time. Dev-only; off by default.

Verify it's up:

```bash
curl http://localhost:8080/health   # → ok
```

## 2. Open the Godot client

```bash
/Applications/Godot.app/Contents/MacOS/Godot --path clients/godot/
```

(Or open the Godot app normally and import `clients/godot/project.godot`.)

Press **F5** / the Run button to launch the main scene.

## 3. Connect

In the login panel:

- **Server:** `ws://localhost:8080/ws` (prefilled)
- **Username:** any string (e.g. `testchar`) — will be auto-registered
- **Password:** any string (dev mode) — becomes the account password
- Click **Connect**

You should see:

1. The login panel disappears and the game panel appears.
2. The terminal (left) renders the opening screen + main menu
   (`[E]nter the Realm`, `[V]iew Guild Lists`, etc.).
3. Press **E** (or **C**) to start character creation. The server walks
   you through class selection, naming, stat rolling, etc. via prompts.

## 4. Test grid combat (once character is created)

From the town menu:
- **W** → Wilderness
- Choose a region (Forest / Mountain / Swamp)
- Explore until you hit an encounter → the TileMap panel on the right
  should render the 12×48 grid with player `@`, monster `M`, and HUD.

Controls during grid combat:
- **I/J/K/L** — move up/left/down/right
- **A** — attack (engage text combat)
- **F** — fire ranged weapon
- **R** — run
- **P** — pass
- **\*** — redraw

## 5. Known limitations (v1)

- `.ans` menu backgrounds currently render with escape sequences stripped
  (raw text only). Proper ANSI→BBCode rendering is Phase 2 polish.
- No reconnect on disconnect — close and reopen the app.
- Tileset is simple colored rects with characters on top. Sprite swap
  planned for Phase 5.
- No login persistence — credentials reset each launch.

## 6. Troubleshooting

- **"Connect failed"**: server not running, or URL typo. Check
  `curl http://localhost:8080/health`.
- **"authentication failed"**: make sure `--allow-new-users` is set on
  the server, or use credentials for a character that already exists.
- **Godot window is blank / errors in console**: check stderr for
  GDScript parse errors. The initial scene should load with nothing
  more than "Godot Engine v4.6.2.stable.official..." in the log.
- **Server log shows `already logged in`**: previous WS session still
  open; close Godot, wait 5 seconds for the server's idle-disconnect,
  reconnect.
