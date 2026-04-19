package io

import "github.com/slycrel/slycrel/internal/model"

// Scene is a structured grid-based display surface for tilemap-aware clients
// (e.g., Godot TileMap). It carries enough information to fully rebuild the
// grid-combat display without any ANSI escape sequences.
//
// LocalTerminal.RenderScene converts a Scene back to ANSI for terminal play.
// WSSession.RenderScene ships it to the Godot client as MsgTypeScene.
//
// Full-scene replacement on every call — 12×48 is tiny, so dirty-cell diffing
// is unnecessary.
type Scene struct {
	Kind     string        `json:"kind"` // "grid_combat"
	Cols     int           `json:"cols"`
	Rows     int           `json:"rows"`
	Terrain  [][]int       `json:"terrain"` // [rows][cols] tile IDs (0-9, see mechanics/terrain.go)
	Entities []SceneEntity `json:"entities"`
	HUD      SceneHUD      `json:"hud"`
}

// SceneEntity is a sprite overlay on the terrain grid (player, monster, etc.).
// Char is the fallback ASCII glyph so terminal clients (and bare-bones Godot
// tilesets) have something to draw without sprite art.
type SceneEntity struct {
	Kind string `json:"kind"` // "player" | "monster"
	Row  int    `json:"row"`
	Col  int    `json:"col"`
	Char string `json:"char"`
}

// SceneHUD is the stats panel drawn alongside grid combat.
type SceneHUD struct {
	CharName    string `json:"char_name"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	Movement    int    `json:"mv"`
	Psyche      int    `json:"ps"`
	MaxPsyche   int    `json:"max_ps"`
	Weapon1     string `json:"w1"`
	Weapon2     string `json:"w2"`
	Armor       string `json:"ar"`
	MonsterName string `json:"monster_name,omitempty"`
	MonsterHP   int    `json:"monster_hp,omitempty"`

	// Log is the recent status-line history (most-recent-last).
	// In terminal play these render in a fixed 5-row pad below the grid;
	// in Godot these render in a log panel alongside the TileMap.
	Log []string `json:"log,omitempty"`
}

// NewGridCombatScene builds a Scene from raw grid-combat state.
// Callers in game/states assemble this without importing the io internals.
func NewGridCombatScene(terrain *model.TerrainMap, userR, userC, monsR, monsC int, hud SceneHUD) Scene {
	rows := model.GridRows
	cols := model.GridCols
	grid := make([][]int, rows)
	for r := 0; r < rows; r++ {
		row := make([]int, cols)
		for c := 0; c < cols; c++ {
			row[c] = int(terrain.Cells[r][c])
		}
		grid[r] = row
	}
	return Scene{
		Kind:    "grid_combat",
		Cols:    cols,
		Rows:    rows,
		Terrain: grid,
		Entities: []SceneEntity{
			{Kind: "player", Row: userR, Col: userC, Char: "@"},
			{Kind: "monster", Row: monsR, Col: monsC, Char: "M"},
		},
		HUD: hud,
	}
}
