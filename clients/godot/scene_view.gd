extends Control
# SceneView draws grid-combat scenes received as MsgTypeScene.
# v1: simple custom-drawn cells (ColorRect + character). Upgrade path to
# Godot TileMap is straightforward — same data shape.

var current_scene: Dictionary = {}

# Cell size. Tweak together with viewport size in project.godot.
const CELL_W := 16.0
const CELL_H := 22.0
# HUD renders below the grid (the scene panel is tall-and-narrow, so "below"
# fits better than "beside"). HUD_TOP_PAD separates grid from HUD.
const HUD_TOP_PAD := 18.0
const HUD_COL_W := 220.0        # column width when laying out HUD fields side-by-side
const FONT_PX := 16
const HUD_FONT_PX := 14
const HUD_LINE_H := 18.0

# Terrain palette — mirrors mechanics.GetTerrainDisplay but in RGB so we can
# draw without an ANSI parser. Index matches terrain tile ID (0-9).
const TERRAIN := [
	{"ch": " ", "fg": Color(0.85, 0.85, 0.85), "bg": Color(0.02, 0.02, 0.02)}, # 0 grass
	{"ch": "^", "fg": Color(0.10, 0.70, 0.10), "bg": Color(0.02, 0.02, 0.02)}, # 1 forest
	{"ch": ".", "fg": Color(0.85, 0.60, 0.15), "bg": Color(0.10, 0.08, 0.02)}, # 2 sand/road
	{"ch": "~", "fg": Color(0.40, 0.40, 1.00), "bg": Color(0.05, 0.05, 0.45)}, # 3 water
	{"ch": "=", "fg": Color(0.85, 0.60, 0.15), "bg": Color(0.55, 0.45, 0.05)}, # 4 bridge
	{"ch": "O", "fg": Color(0.45, 0.45, 0.45), "bg": Color(0.02, 0.02, 0.02)}, # 5 boulder
	{"ch": "#", "fg": Color(0.00, 0.95, 0.00), "bg": Color(0.02, 0.02, 0.02)}, # 6 deep forest
	{"ch": "+", "fg": Color(0.00, 0.55, 0.00), "bg": Color(0.02, 0.02, 0.02)}, # 7 light forest
	{"ch": "%", "fg": Color(1.00, 0.90, 0.10), "bg": Color(0.12, 0.45, 0.05)}, # 8 swamp
	{"ch": "&", "fg": Color(0.10, 0.95, 0.10), "bg": Color(0.05, 0.35, 0.05)}, # 9 deep swamp
]

const ENTITY := {
	"player": {"ch": "@", "color": Color(0.45, 0.55, 1.00)},
	"monster": {"ch": "M", "color": Color(1.00, 0.35, 0.35)},
}

# HUD colors — loosely track the terminal ANSI layout (cyan names, green
# stats, magenta gear, red enemy).
const HUD_COLOR_NAME := Color(0.20, 0.85, 0.95)
const HUD_COLOR_STAT := Color(0.25, 0.85, 0.30)
const HUD_COLOR_GEAR := Color(0.90, 0.45, 0.95)
const HUD_COLOR_ENEMY := Color(0.95, 0.30, 0.30)
const HUD_COLOR_LOG := Color(0.60, 0.60, 0.60)

var _font: Font

func _ready() -> void:
	_font = ThemeDB.fallback_font

# Called by main.gd after updating current_scene.
func refresh() -> void:
	queue_redraw()

func _draw() -> void:
	if current_scene.is_empty() or _font == null:
		return

	var cols: int = int(current_scene.get("cols", 48))
	var rows: int = int(current_scene.get("rows", 12))
	var terrain = current_scene.get("terrain", [])

	# Terrain cells
	for r in rows:
		if r >= terrain.size():
			break
		var row = terrain[r]
		for c in cols:
			if c >= row.size():
				break
			var id: int = int(row[c])
			if id < 0 or id >= TERRAIN.size():
				id = 0
			var def = TERRAIN[id]
			var cell_pos := Vector2(c * CELL_W, r * CELL_H)
			draw_rect(Rect2(cell_pos, Vector2(CELL_W, CELL_H)), def["bg"])
			if def["ch"] != " ":
				draw_string(
					_font,
					cell_pos + Vector2(3, CELL_H - 5),
					def["ch"],
					HORIZONTAL_ALIGNMENT_LEFT,
					-1,
					FONT_PX,
					def["fg"],
				)

	# Entities overlay
	var entities = current_scene.get("entities", [])
	for e in entities:
		var kind: String = e.get("kind", "")
		var row: int = int(e.get("row", 0))
		var col: int = int(e.get("col", 0))
		var fallback_ch: String = e.get("char", "?")
		var info = ENTITY.get(kind, {"ch": fallback_ch, "color": Color.WHITE})
		var cell_pos := Vector2(col * CELL_W, row * CELL_H)
		draw_string(
			_font,
			cell_pos + Vector2(3, CELL_H - 5),
			info["ch"],
			HORIZONTAL_ALIGNMENT_LEFT,
			-1,
			FONT_PX,
			info["color"],
		)

	# HUD renders below the grid in a 3-column layout:
	#   col 1: player stats (name, HP, Mv, Ps)
	#   col 2: gear (W1, W2, Ar) + monster info
	#   col 3: combat log
	var hud = current_scene.get("hud", {})
	if typeof(hud) != TYPE_DICTIONARY or hud.is_empty():
		return

	var hud_y0 := rows * CELL_H + HUD_TOP_PAD
	var step := HUD_LINE_H

	# Column 1: player stats
	var x := 8.0
	var y := hud_y0
	_hud_line(Vector2(x, y), str(hud.get("char_name", "")), HUD_COLOR_NAME); y += step
	_hud_line(Vector2(x, y), "HP: %d/%d" % [int(hud.get("hp", 0)), int(hud.get("max_hp", 0))], HUD_COLOR_STAT); y += step
	_hud_line(Vector2(x, y), "Mv: %d" % int(hud.get("mv", 0)), HUD_COLOR_STAT); y += step
	_hud_line(Vector2(x, y), "Ps: %d/%d" % [int(hud.get("ps", 0)), int(hud.get("max_ps", 0))], HUD_COLOR_STAT)

	# Column 2: gear + enemy
	x += HUD_COL_W
	y = hud_y0
	var w1 := str(hud.get("w1", ""))
	if w1 == "": w1 = "Hands"
	_hud_line(Vector2(x, y), "W1: " + w1, HUD_COLOR_GEAR); y += step
	var w2 := str(hud.get("w2", ""))
	if w2 == "": w2 = "---"
	_hud_line(Vector2(x, y), "W2: " + w2, HUD_COLOR_GEAR); y += step
	var ar := str(hud.get("ar", ""))
	if ar == "": ar = "None"
	_hud_line(Vector2(x, y), "Ar: " + ar, HUD_COLOR_GEAR); y += step
	var mname := str(hud.get("monster_name", ""))
	if mname != "":
		y += step * 0.5
		_hud_line(Vector2(x, y), mname, HUD_COLOR_ENEMY); y += step
		_hud_line(Vector2(x, y), "HP: %d" % int(hud.get("monster_hp", 0)), HUD_COLOR_ENEMY)

	# Column 3: combat log
	x += HUD_COL_W
	y = hud_y0
	var log_lines = hud.get("log", [])
	if typeof(log_lines) == TYPE_ARRAY:
		for line in log_lines:
			_hud_line(Vector2(x, y), str(line), HUD_COLOR_LOG)
			y += step

func _hud_line(pos: Vector2, text: String, color: Color) -> void:
	draw_string(
		_font,
		pos,
		text,
		HORIZONTAL_ALIGNMENT_LEFT,
		-1,
		HUD_FONT_PX,
		color,
	)
