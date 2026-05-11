# Slycrel

A text-based RPG door game, reborn.

**Slycrel** was originally written in Pascal circa 1994-1996 by Dave Dolinar and Jeremy Stone for the [Hermes II BBS](https://www.hermesbbs.com) on Mac Classic. It was later ported to C++ by Bill Dolinar. This is a modern faithful recreation in Go, playable in your terminal.

## The Game

You are an adventurer in the realm of Slycrel. Create a character — Fighter, Thief, or Mage — and explore a medieval fantasy world:

- **Town Hub** — visit the Bank, Armory, Tavern, Inn, Common Guild, Blacksmith, Jail, Tower, Herbalist, and Arena
- **Wilderness Combat** — explore Forest, Mountain, and Swamp regions, battling creatures on a 12×48 ANSI terrain grid before engaging in narrative text combat
- **Arena PvP** — challenge other players to combat, set up gladiator fights, and bet on the outcomes
- **Tavern** — order drinks from Lynx the bartender, flirt with Corenne (good luck), and listen for rumors
- **Inn** — rent rooms, or become the innkeeper and manage the establishment
- **Level Up** — gain experience through combat, then visit the Common Guild to advance your class

The game preserves the original's mechanics, flavor text, and BBS-era charm — including the easter eggs.

## Headless Server (Browser Client)

`cmd/server` runs as a pure headless WebSocket server — no local terminal I/O, no static assets served. Browser (or any WebSocket client) connects to `/ws`, authenticates, and plays entirely via JSON message exchange. The local `cmd/slycrel` binary is a separate entry point for terminal play.

```bash
# Build and run the headless server
go build -o server ./cmd/server/
./server --port 8080 --data-dir ./data --state-dir ./state

# WebSocket endpoint
ws://localhost:8080/ws

# Health check
curl http://localhost:8080/health
```

See `internal/io/ws_protocol.go` for the full client↔server message schema.

### Docker

A static Linux image (~6.4 MB) for `cmd/server` is defined in `Dockerfile` —
multi-stage build with a `FROM scratch` final stage (no shell, no libc, just
the static Go binary and the `data/` assets).

```bash
docker build -t slycrel:dev .
docker run --rm -p 8080:8080 -v "$PWD/state:/state" slycrel:dev
```

`/state` inside the container is the writable directory for character and
game data. Bind-mount the gitignored host `state/` directory onto it for
local persistence.

## Godot Client

`clients/godot/` is a graphical client for the WebSocket server — a Godot 4.6 project that renders the ANSI menus with a DOS VGA font and draws grid-combat scenes with a custom tilemap. Open `clients/godot/project.godot` in Godot, set the server address in the connect screen, and play. See `clients/godot/QUICKSTART.md` for details.

## Web Client

`clients/web/` is a browser-only port of the game logic — no backend, designed for static hosting on Cloudflare Pages. Save state lives in `localStorage`. The Go server in `cmd/server/` remains the reference implementation; the web client mirrors its module layout (`model/`, `mechanics/`, `game/states/`) so logic can be cross-checked file-for-file. See `clients/web/README.md` to run it locally.

## Quick Start

```bash
# Build
go build -o slycrel ./cmd/slycrel/

# Play
./slycrel

# Terrain Editor
./slycrel --editor
```

Requires Go 1.21+ and a terminal that supports ANSI escape codes (virtually all modern terminals).

## Controls

### General
- Single keypress for all menu choices (no Enter needed)
- Ctrl-C to exit at any time

### Grid Combat
- **I/J/K/L** — Move up/left/down/right
- **A** — Attempt melee attack
- **F** — Fire ranged weapon
- **R** — Run away
- **P** — Pass turn
- **\*** — Redraw screen

### Text Combat
- **A** — Attack
- **B** — Block (reduce incoming damage)
- **P** — Parry (deal reduced damage + reduce incoming)
- **D** — Dodge (chance to avoid damage entirely)
- **S** — Special (don't ask)
- **R** — Run

### Terrain Editor
- **I/J/K/L** or **Arrow Keys** — Move cursor
- **0-9** — Paint terrain type
- **T** — Toggle trace mode (paint while moving)
- **S** — Save map
- **R** — Retrieve/load map
- **N** — New blank map
- **Q** — Quit

## Project Structure

```
cmd/
  slycrel/            Standalone console app (terminal play)
  server/             Headless WebSocket server (multi-client play)
  convert_ansi/       CP437 .ans → ANSI-escape converter tool
clients/
  godot/              Godot 4.6 graphical WebSocket client
internal/
  model/              Data structures (character, monster, weapon, armor, etc.)
  game/               Game engine, session, state machine
    states/           One file per game location/flow
  mechanics/          Pure game logic (combat, leveling, pathfinding, etc.)
  io/                 I/O abstraction (terminal + WebSocket impls, ANSI helpers)
  store/              Data persistence (JSON file store)
  editor/             Terrain map editor
data/
  monsters/           Monster definitions (forest, mountain, swamp)
  terrain/            12×48 combat terrain maps
  ansi/               ANSI art screens
  weapons.json        Weapon catalog
  armor.json          Armor catalog
  state/              Mutable game state (characters, inn, etc.)
orig/                 Original Pascal/C++ source (read-only reference)
```

## Architecture

The game is designed with clean separation for future expansion:

- **IOProvider interface** — abstracts all user I/O, replaceable for web/telnet/BBS backends
- **State machine** — faithful recreation of the original's `SetNextState`/`SetReturnState`/`PopReturnState` flow control
- **JSON data files** — human-readable and easy to extend with new monsters, weapons, terrain
- **Multi-player ready** — shared data store with mutex protection, session-per-user architecture

## History

- **~1994-1996** — Original Pascal version for Hermes II BBS (Dave Dolinar & Jeremy Stone)
- **~1997-2000** — C++ port by Bill Dolinar
- **2026** — Go recreation, faithful to the original

## Related Projects

- [Hermes BBS](https://www.hermesbbs.com) — the BBS system Slycrel originally ran on
- [hermesbbs-orig](https://github.com/slycrel/hermesbbs-orig) — original Hermes BBS source code (Pascal)

## Credits

- **Dave Dolinar** — co-creator, original programmer
- **Jeremy Stone** — co-creator, game design
- **Bill Dolinar** — C++ port

*"Spray Cheese & Ritz Crackers"* — from the original credits

## License

This is a personal project recreating a game from the mid-1990s. The original source code is included in `orig/` for reference.
