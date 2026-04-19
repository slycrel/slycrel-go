# Slycrel Godot Client — Port Plan

Living document. Update as decisions change.

## Goal

Full working port of Slycrel to Godot 4.x as a WebSocket client against the
existing headless Go server (`cmd/server`). "Working" means: log in, navigate
menus, play through wilderness encounters including grid combat and text
combat, with gameplay feel intact. Terminal play (`cmd/slycrel`) must keep
working unchanged.

## Branch / base

- Branch: `godot-client`, cut from `headless-server-run7`.
- Inherits the WS server + `hello` handshake + session lifecycle already on
  that branch.

## Architecture decision — split renderers

Slycrel mixes two fundamentally different rendering modes:

1. **Text / menu output.** ANSI art screens (13 `.ans` files), colored text
   lines, prompts. Naturally a terminal-style display.
2. **Grid combat.** 12×48 terrain with entity overlays, HUD panel, painted
   today via ~40 runtime `ANSICode()` calls per frame. Naturally a tilemap.

Decision: **use both renderers in Godot, route each kind of scene to the
right one.**

- Text/menu path → Godot `RichTextLabel` (BBCode for colors). Server keeps
  emitting `MsgTypeOutput` + `MsgTypePrompt` + `MsgTypeANSIArt` as today.
- Grid scene path → Godot `TileMap` + entity sprites. Server emits a new
  `MsgTypeScene` with structured grid/entities/HUD data.

Rejected alternatives:
- **Server-side ANSI parsing → ship pre-rendered cells.** Would still require
  Godot to handle runtime ANSI for grid combat, giving us two parsers.
- **Godot as pure terminal emulator (parse all ANSI client-side).** Feasible
  but wastes Godot's actual strengths on grid combat, and blocks future
  sprite/animation upgrades.

## Canonical tileset (single source of truth)

`internal/mechanics/terrain.go` already caps terrain at **10 types** (IDs
0-9), which matches the terrain editor's `0-9` paint keys. This is our
tileset vocabulary — editor, server, and client all agree on it.

| ID | Meaning       | Char | Notes              |
|----|---------------|------|--------------------|
| 0  | Empty / grass | ` `  | passable (cost 1)  |
| 1  | Forest        | `^`  | cost 2             |
| 2  | Sand / road   | `.`  | cost 1             |
| 3  | Water         | `~`  | impassable (death) |
| 4  | Bridge        | `=`  | cost 1             |
| 5  | Boulder       | `O`  | impassable         |
| 6  | Deep forest   | `#`  | cost 3             |
| 7  | Light forest  | `+`  | cost 3 (legacy)    |
| 8  | Swamp         | `%`  | cost 3             |
| 9  | Deep swamp    | `&`  | cost 5             |

Godot `TileSet` mirrors these IDs. v1 visuals: char on colored background
(faithful to current look). v2: swap in real sprites — zero protocol change,
just art.

Entity sprites (not part of terrain tileset): player `@` blue, monster `M`
red. Kept as overlay nodes so entity animation is independent of terrain.

## New WS protocol addition: `MsgTypeScene`

```go
MsgTypeScene MsgType = "scene"  // RenderScene — structured grid display

type ServerMsg struct {
    // existing fields...

    // scene fields (MsgTypeScene)
    SceneKind    string  `json:"scene_kind,omitempty"`   // "grid_combat"
    SceneCols    int     `json:"scene_cols,omitempty"`   // 48
    SceneRows    int     `json:"scene_rows,omitempty"`   // 12
    SceneTerrain [][]int `json:"scene_terrain,omitempty"`// [rows][cols] tile IDs
    SceneEntities []SceneEntity `json:"scene_entities,omitempty"`
    SceneHUD     SceneHUD `json:"scene_hud,omitempty"`
}

type SceneEntity struct {
    Kind string `json:"kind"` // "player" | "monster"
    Row  int    `json:"row"`
    Col  int    `json:"col"`
    Char string `json:"char,omitempty"` // fallback rendering if no sprite
}

type SceneHUD struct {
    CharName  string `json:"char_name"`
    HP, MaxHP int    `json:"hp"`, `json:"max_hp"`
    Movement  int    `json:"mv"`
    Psyche    int    `json:"ps"`
    MaxPsyche int    `json:"max_ps"`
    Weapon1   string `json:"w1"`
    Weapon2   string `json:"w2"`
    Armor     string `json:"ar"`
    MonsterName string `json:"monster_name,omitempty"`
    MonsterHP   int    `json:"monster_hp,omitempty"`
}
```

Full scene replacement on every `RenderScene` call (simple + reliable; 12×48
is tiny — no need for dirty-cell diffing).

## IOProvider extension

```go
// IOProvider adds:
RenderScene(scene Scene) // or similar typed struct
```

- `LocalTerminal.RenderScene`: reimplements the existing ANSI grid rendering
  — no behavior change for terminal play.
- `WSSession.RenderScene`: emits `MsgTypeScene`.

## Godot project layout

```
clients/godot/
  project.godot
  PLAN.md                     # this file
  scenes/
    main.tscn                 # root: split between terminal + scene panels
    login.tscn                # credential entry (pre-hello)
  scripts/
    ws_client.gd              # WebSocket + hello + send/receive
    message_router.gd         # dispatches ServerMsg by type
    terminal_panel.gd         # RichTextLabel + BBCode coloring
    scene_panel.gd            # TileMap + entity sprites + HUD
    prompt_handler.gd         # 5 prompt kinds → input capture
  tilesets/
    terrain.tres              # 10-tile TileSet
  assets/
    (placeholder sprites)
```

## Implementation phases

### Phase 1 — Server changes ✓

- [x] Add `MsgTypeScene` + Scene struct fields to `internal/io/ws_protocol.go`.
- [x] Add `RenderScene(Scene)` to `IOProvider`.
- [x] Implement `LocalTerminal.RenderScene` (mirrors pre-port ANSI layout).
- [x] Implement `WSSession.RenderScene` (ship `MsgTypeScene`).
- [x] Refactor `combat_grid.go` to Scene-based rendering. All incremental
      draw helpers (drawPlayerOnGrid, drawTerrainCell, etc.) replaced with
      full-scene re-renders via `renderGridScene(s)`.
- [x] Kill live `ANSICode()` usage from game code. 8 SGR calls in
      `character_create.go` (hair color picker) refactored to use the
      `Outln` palette. Only ANSI surviving is `.ans` file shipping via
      `ShowANSIFile`. **Godot needs zero runtime ANSI parsing for gameplay.**
- [ ] Verify terminal play unchanged end-to-end: `./slycrel` still
      renders grid combat (deferred to Phase 2 e2e test).

### Phase 2 — Godot client

- [ ] Scaffold Godot 4.6 project at `clients/godot/`.
- [ ] Login scene: username/password → `ws_client.connect_to_url(...)` + hello.
- [ ] Main scene: terminal panel (top, scroll) + scene panel (bottom,
      hidden until `MsgTypeScene` arrives).
- [ ] Terminal panel: render `MsgTypeOutput` (append text with color), handle
      `MsgTypeANSIArt` (one-shot full clear + render .ans content), flush
      ANSI SGR+cursor codes in the subset still used (see Phase 3).
- [ ] Scene panel: TileSet with 10 terrain tiles + player/monster sprites +
      HUD. Handle `MsgTypeScene` by rebuilding terrain + repositioning
      entities + updating HUD.
- [ ] Prompt handler: dispatch on `Kind`. Letters/readline/number → LineEdit;
      yesno/pause → keypress listener. Send `ClientMsg{type:input, seq:N,
      value:...}`.

### Phase 3 — Remaining ANSI scope (RESOLVED)

Audit done: zero `ANSICode()` calls remain in the game code after the
Phase 1 refactor. Live output uses only `Outln` (color index) and prompts.

Remaining ANSI surface:
- **`.ans` menu screens** — 13 files, shipped via `ShowANSIFile`. These use
  SGR (`Nm`), cursor position (`r;cH`), cursor forward (`NC`), and clear
  (`2J H`). Godot must handle this subset for menu backgrounds.
- **Nothing else.**

Godot implements a one-shot ANSI renderer for `MsgTypeANSIArt` only.
Live text output is plain colored lines → `RichTextLabel` with BBCode.

### Phase 4 — Editor path

**Not in scope for this session.** Existing Go terminal editor
(`./slycrel --editor`) keeps working — it reads/writes
`data/terrain/*.json` directly, independent of the WS layer.

Future (cheap, once TileSet exists): port editor to Godot using built-in
tilemap painting. IJKL cursor, 0-9 paint, S save, R load — maps 1:1 to
Godot input actions. This is the kind of thing Godot excels at and will be
almost free once the tileset is defined.

### Phase 5 — Visual upgrades (aspirational)

- Swap char-on-bg tiles for real sprites.
- Tween entity movement between cells instead of instant teleport.
- Camera follow + zoom.
- Optional audio on combat events.

## Open questions / TBD

- **Window layout.** Terminal panel + scene panel split: side by side,
  stacked, or tab/modal? Starting stacked (terminal on top). Revisit after
  first playtest.
- **Font.** Any monospace with good UTF-8 coverage. Default to Godot's
  bundled fonts; consider a pixel font later for BBS feel.
- **Connection recovery.** v1: close app on disconnect. Reconnect flow later.
- **Multi-character selection.** Server-side auth flow handles one account;
  Godot login just passes it through.

## Success criteria for this session

Able to, with the Godot client running against `./server`:
1. Log in as a test character.
2. See the main menu rendered.
3. Navigate into Wilderness.
4. See grid combat render in the TileMap panel.
5. Move the player with I/J/K/L.
6. Engage a monster → text combat shown in terminal panel.
7. Exit back to main menu.

If all 7 pass, it's a working port. Visuals can be ugly; mechanics must work.
