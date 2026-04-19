extends Control
# Slycrel Godot client — connects to cmd/server over WebSocket and drives the
# login / terminal / scene UI from incoming JSON messages.
# Protocol: see ../../internal/io/ws_protocol.go.

const PROTOCOL_VERSION := 1

# Server palette (IOProvider color index → hex).
# 0 is default (no color). 1=cyan, 2=white, 3=green, 4=yellow, 5=magenta, 6=red.
const PALETTE := {
	0: "",
	1: "#00cccc",
	2: "#ffffff",
	3: "#66ff66",
	4: "#ffff66",
	5: "#ff66ff",
	6: "#ff6666",
}

@onready var login_panel: Control = $LoginPanel
@onready var game_panel: Control = $GamePanel
@onready var server_input: LineEdit = $LoginPanel/VBox/ServerRow/ServerInput
@onready var user_input: LineEdit = $LoginPanel/VBox/UserRow/UserInput
@onready var pass_input: LineEdit = $LoginPanel/VBox/PassRow/PassInput
@onready var connect_button: Button = $LoginPanel/VBox/ConnectButton
@onready var status_label: Label = $LoginPanel/VBox/StatusLabel
@onready var terminal_text: RichTextLabel = $GamePanel/MainSplit/TerminalPanel/TerminalScroll/TerminalText
@onready var scene_view: Control = $GamePanel/MainSplit/ScenePanel/SceneView
@onready var prompt_label: Label = $GamePanel/PromptRow/PromptLabel
@onready var prompt_input: LineEdit = $GamePanel/PromptRow/PromptInput

var ws := WebSocketPeer.new()
var pending_hello := false
var handshake_done := false

# Current prompt state.
var prompt_seq := -1
var prompt_kind := ""
var prompt_allowed := ""
var prompt_capitalize := false
var prompt_show_input := true
var prompt_min := 0
var prompt_max := 0

func _ready() -> void:
	connect_button.pressed.connect(_on_connect_pressed)
	prompt_input.text_submitted.connect(_on_prompt_submitted)
	pass_input.text_submitted.connect(func(_t): _on_connect_pressed())
	set_process(true)
	user_input.grab_focus()

	# Monospace font for the terminal so .ans column alignment works and the
	# 80-col opening art doesn't wrap arbitrarily. The font size is tuned so
	# a standard 80-char line fits in ~780 logical pixels (the terminal panel
	# width) — roughly 9px per char.
	var mono := SystemFont.new()
	mono.font_names = PackedStringArray(["Menlo", "Monaco", "SF Mono", "Consolas", "Courier New"])
	terminal_text.add_theme_font_override("normal_font", mono)
	terminal_text.add_theme_font_override("bold_font", mono)
	terminal_text.add_theme_font_override("italics_font", mono)
	terminal_text.add_theme_font_override("mono_font", mono)
	terminal_text.add_theme_font_size_override("normal_font_size", 14)
	terminal_text.add_theme_font_size_override("bold_font_size", 14)
	terminal_text.add_theme_font_size_override("mono_font_size", 14)

func _process(_delta: float) -> void:
	ws.poll()
	var state := ws.get_ready_state()

	if pending_hello and state == WebSocketPeer.STATE_OPEN:
		pending_hello = false
		_send({
			"v": PROTOCOL_VERSION,
			"type": "hello",
			"username": user_input.text,
			"password": pass_input.text,
		})

	if state == WebSocketPeer.STATE_OPEN:
		while ws.get_available_packet_count() > 0:
			var packet := ws.get_packet()
			var text := packet.get_string_from_utf8()
			_handle_message(text)
	elif state == WebSocketPeer.STATE_CLOSED:
		if handshake_done:
			_show_status("connection closed")
			handshake_done = false
			game_panel.visible = false
			login_panel.visible = true
			set_process(false)

# ──────────────────────────────────────────────────────────────────────────
# WebSocket + connection flow
# ──────────────────────────────────────────────────────────────────────────

func _on_connect_pressed() -> void:
	var url := server_input.text.strip_edges()
	if url == "" or user_input.text == "":
		_show_status("fill server + username")
		return
	_show_status("connecting…")
	var err := ws.connect_to_url(url)
	if err != OK:
		_show_status("connect failed: %s" % err)
		return
	pending_hello = true

func _send(msg: Dictionary) -> void:
	var data := JSON.stringify(msg)
	var err := ws.send_text(data)
	if err != OK:
		push_warning("ws send failed: %s" % err)

func _show_status(text: String) -> void:
	status_label.text = text

# ──────────────────────────────────────────────────────────────────────────
# Message dispatch
# ──────────────────────────────────────────────────────────────────────────

func _handle_message(text: String) -> void:
	var parsed = JSON.parse_string(text)
	if typeof(parsed) != TYPE_DICTIONARY:
		push_warning("bad JSON from server: %s" % text)
		return
	match parsed.get("type", ""):
		"hello":
			_on_hello_ack()
		"output":
			_on_output(parsed)
		"prompt":
			_on_prompt(parsed)
		"ansi_art":
			_on_ansi_art(parsed)
		"scene":
			_on_scene(parsed)
		"error":
			_on_error(parsed)
		_:
			push_warning("unknown msg type: %s" % parsed)

func _on_hello_ack() -> void:
	handshake_done = true
	login_panel.visible = false
	game_panel.visible = true
	terminal_text.clear()

func _on_output(msg: Dictionary) -> void:
	# Server has eliminated all live ANSI except ClearScreen, which arrives as
	# {ansi:"2J"} — handle that explicitly.
	var ansi: String = msg.get("ansi", "")
	if ansi == "2J":
		terminal_text.clear()
		return
	if ansi != "":
		# Unexpected post-refactor; log for visibility but don't crash.
		push_warning("unhandled ANSI from server: %s" % ansi)
		return

	var text: String = msg.get("text", "")
	var color: int = msg.get("color", 0)
	var hex: String = PALETTE.get(color, "")
	# Server text contains literal square brackets like "[E]nter the Realm" —
	# if we pass that through append_text with bbcode_enabled, Godot parses
	# "[E]" as a tag and swallows the content. Use push_color + add_text so
	# the text is never BBCode-parsed.
	if hex != "":
		terminal_text.push_color(Color(hex))
		terminal_text.add_text(text)
		terminal_text.pop()
	else:
		terminal_text.add_text(text)
	if msg.get("newline", false):
		terminal_text.add_text("\n")

func _on_ansi_art(msg: Dictionary) -> void:
	# .ans files ship as base64-encoded text with literal "\e[..." escape
	# sequences. Convert the SGR + cursor-forward subset into BBCode so
	# colors + column alignment render in the monospace terminal panel.
	var b64: String = msg.get("ansi_art_data", "")
	if b64 == "":
		return
	var bytes := Marshalls.base64_to_raw(b64)
	var content := bytes.get_string_from_utf8()
	var result := _ansi_to_bbcode(content)
	if result["cleared"]:
		terminal_text.clear()
	terminal_text.append_text(result["bbcode"])
	if not (result["bbcode"] as String).ends_with("\n"):
		terminal_text.add_text("\n")

func _on_scene(msg: Dictionary) -> void:
	var scene = msg.get("scene", {})
	if typeof(scene) != TYPE_DICTIONARY:
		return
	scene_view.set("current_scene", scene)
	if scene_view.has_method("refresh"):
		scene_view.refresh()

func _on_error(msg: Dictionary) -> void:
	var err: String = msg.get("err", "unknown error")
	_show_status(err)
	push_warning("server error: %s" % err)
	# If we weren't in-game yet, the connection's about to close anyway.

# ──────────────────────────────────────────────────────────────────────────
# Prompt handling
# ──────────────────────────────────────────────────────────────────────────

func _on_prompt(msg: Dictionary) -> void:
	prompt_seq = msg.get("seq", 0)
	prompt_kind = msg.get("kind", "")
	prompt_allowed = msg.get("allowed", "")
	prompt_capitalize = msg.get("capitalize", false)
	prompt_show_input = msg.get("show_input", true)
	prompt_min = msg.get("min", 0)
	prompt_max = msg.get("max", 0)

	var label_text: String = msg.get("prompt", "")
	if label_text == "":
		label_text = "›"
	prompt_label.text = label_text + " "
	prompt_input.clear()

	match prompt_kind:
		"letters", "yesno", "pause":
			# Single-key: listen via _unhandled_input. Don't grab focus so
			# the LineEdit doesn't swallow the key.
			prompt_input.editable = false
			prompt_input.placeholder_text = "(press a key)"
			if has_focus():
				release_focus()
		"number":
			prompt_input.editable = true
			prompt_input.placeholder_text = "%d – %d" % [prompt_min, prompt_max]
			prompt_input.grab_focus()
		"readline":
			prompt_input.editable = true
			prompt_input.placeholder_text = ""
			prompt_input.grab_focus()

func _unhandled_input(event: InputEvent) -> void:
	if prompt_seq == -1:
		return
	if not (event is InputEventKey and event.pressed and not event.echo):
		return
	var key := event as InputEventKey
	match prompt_kind:
		"letters":
			if key.unicode == 0:
				return
			var ch := char(key.unicode)
			if prompt_capitalize:
				ch = ch.to_upper()
			if prompt_allowed != "" and not _allowed_match(ch, prompt_allowed):
				return
			_submit_prompt(ch)
			get_viewport().set_input_as_handled()
		"yesno":
			if key.keycode == KEY_Y:
				_submit_prompt("y")
				get_viewport().set_input_as_handled()
			elif key.keycode == KEY_N:
				_submit_prompt("n")
				get_viewport().set_input_as_handled()
		"pause":
			_submit_prompt("")
			get_viewport().set_input_as_handled()

func _on_prompt_submitted(text: String) -> void:
	if prompt_seq == -1:
		return
	match prompt_kind:
		"number", "readline":
			_submit_prompt(text)

func _submit_prompt(value: String) -> void:
	_send({
		"v": PROTOCOL_VERSION,
		"type": "input",
		"seq": prompt_seq,
		"value": value,
	})
	prompt_seq = -1
	prompt_kind = ""
	prompt_input.clear()
	prompt_input.editable = false
	prompt_input.placeholder_text = "(waiting for server)"
	prompt_label.text = "› "

# ──────────────────────────────────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────────────────────────────────

func _allowed_match(ch: String, allowed: String) -> bool:
	# The server's allowed string is case-insensitive-ish; compare uppercased.
	return allowed.to_upper().find(ch.to_upper()) != -1

# ──────────────────────────────────────────────────────────────────────────
# ANSI → BBCode converter for .ans menu files
# ──────────────────────────────────────────────────────────────────────────

# Standard xterm-ish palette for the SGR 30-37 / 1+30-37 range.
const ANSI_FG_NORMAL := {
	"30": "#2a2a2a", "31": "#b04040", "32": "#40a040", "33": "#a07030",
	"34": "#4060b0", "35": "#a040a0", "36": "#40a0a0", "37": "#b0b0b0",
}
const ANSI_FG_BOLD := {
	"30": "#606060", "31": "#ff6060", "32": "#60ff60", "33": "#ffd060",
	"34": "#6090ff", "35": "#ff60ff", "36": "#60ffff", "37": "#ffffff",
}

# Returns {"bbcode": String, "cleared": bool}.
# cleared=true if the content contained a \e[2J clear-screen command (caller
# should clear the terminal buffer before appending bbcode).
func _ansi_to_bbcode(s: String) -> Dictionary:
	var out := ""
	var cleared := false
	var color_open := false
	var i := 0
	while i < s.length():
		# Literal "\e[...<letter>" escape sequence?
		if i + 1 < s.length() and s[i] == "\\" and s[i + 1] == "e":
			var j := i + 2
			if j < s.length() and s[j] == "[":
				j += 1
			var params := ""
			var term := ""
			while j < s.length():
				var c: String = s[j]
				if (c >= "A" and c <= "Z") or (c >= "a" and c <= "z"):
					term = c
					j += 1
					break
				params += c
				j += 1
			match term:
				"m":
					# SGR: close any open color span, then open a new one
					# (or leave closed for reset).
					if color_open:
						out += "[/color]"
						color_open = false
					var hex := _sgr_to_hex(params)
					if hex != "":
						out += "[color=" + hex + "]"
						color_open = true
				"C":
					# Cursor forward N columns → emit N spaces (works under
					# monospace, which is what the terminal panel uses).
					var n: int = int(params) if params != "" else 1
					for k in n:
						out += " "
				"J":
					if params == "2":
						cleared = true
						out = ""
						if color_open:
							color_open = false  # caller-side clear resets state
				# "H", "f" absolute positioning intentionally dropped — rare in
				# our menus and would require a full terminal emulator.
			i = j
			continue
		# Ordinary character. Escape BBCode brackets so "[E]nter" doesn't
		# become an unknown tag.
		var ch: String = s[i]
		if ch == "[":
			out += "[lb]"
		elif ch == "]":
			out += "[rb]"
		else:
			out += ch
		i += 1
	if color_open:
		out += "[/color]"
	return {"bbcode": out, "cleared": cleared}

func _sgr_to_hex(params: String) -> String:
	if params == "" or params == "0":
		return ""
	var parts := params.split(";")
	var bold := false
	var fg := ""
	for p in parts:
		var n: int = int(p)
		if n == 0:
			bold = false
			fg = ""
		elif n == 1:
			bold = true
		elif n >= 30 and n <= 37:
			fg = str(n)
	if fg == "":
		return ""
	var table = ANSI_FG_BOLD if bold else ANSI_FG_NORMAL
	return table.get(fg, "")
