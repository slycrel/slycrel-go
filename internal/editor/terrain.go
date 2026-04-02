package editor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// TerrainEditor is an interactive terrain map editor.
// Faithful port of the original Slycrel Terrain Editor external.
type TerrainEditor struct {
	term     *slyio.LocalTerminal
	terrain  model.TerrainMap
	cursorR  int
	cursorC  int
	brush    model.TerrainCell
	trace    bool // paint while moving
	dataDir  string
	fileName string
	dirty    bool
}

// RunTerrainEditor launches the interactive terrain editor.
func RunTerrainEditor(dataDir string) error {
	term := slyio.NewLocalTerminal(dataDir)
	if err := term.EnableRawMode(); err != nil {
		return fmt.Errorf("enabling raw mode: %w", err)
	}
	defer term.RestoreTerminal()

	ed := &TerrainEditor{
		term:    term,
		dataDir: dataDir,
		cursorR: 0,
		cursorC: 0,
		brush:   model.TerrainEmpty,
	}

	ed.drawFullScreen()
	ed.mainLoop()
	return nil
}

func (ed *TerrainEditor) mainLoop() {
	for ed.term.IsConnected() {
		ed.drawCursor()

		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			return
		}

		ch := b[0]

		// Handle escape sequences (arrow keys)
		if ch == 27 {
			next := make([]byte, 2)
			os.Stdin.Read(next)
			if next[0] == '[' {
				switch next[1] {
				case 'A': // up
					ch = 'i'
				case 'B': // down
					ch = 'k'
				case 'C': // right
					ch = 'l'
				case 'D': // left
					ch = 'j'
				}
			}
		}

		switch {
		case ch == 'i' || ch == 'I':
			ed.eraseCursor()
			if ed.cursorR > 0 {
				ed.cursorR--
			}
			if ed.trace {
				ed.paint()
			}
		case ch == 'k' || ch == 'K':
			ed.eraseCursor()
			if ed.cursorR < model.GridRows-1 {
				ed.cursorR++
			}
			if ed.trace {
				ed.paint()
			}
		case ch == 'j' || ch == 'J':
			ed.eraseCursor()
			if ed.cursorC > 0 {
				ed.cursorC--
			}
			if ed.trace {
				ed.paint()
			}
		case ch == 'l' || ch == 'L':
			ed.eraseCursor()
			if ed.cursorC < model.GridCols-1 {
				ed.cursorC++
			}
			if ed.trace {
				ed.paint()
			}
		case ch >= '0' && ch <= '9':
			ed.brush = model.TerrainCell(ch - '0')
			ed.paint()
			ed.updateBrushDisplay()
		case ch == 't' || ch == 'T':
			ed.trace = !ed.trace
			ed.updateTraceDisplay()
		case ch == 'c' || ch == 'C':
			// Clear map
			ed.terrain.Cells = [model.GridRows][model.GridCols]model.TerrainCell{}
			ed.dirty = true
			ed.drawGrid()
		case ch == 's' || ch == 'S':
			ed.savePrompt()
		case ch == 'r' || ch == 'R':
			ed.loadPrompt()
		case ch == 'n' || ch == 'N':
			ed.newMapPrompt()
		case ch == 'q' || ch == 'Q':
			if ed.dirty {
				ed.statusMsg("Unsaved changes! Press Q again to quit, S to save.")
				b2 := make([]byte, 1)
				os.Stdin.Read(b2)
				if b2[0] == 'q' || b2[0] == 'Q' {
					return
				}
				if b2[0] == 's' || b2[0] == 'S' {
					ed.savePrompt()
				}
				continue
			}
			return
		case ch == 3 || ch == 4: // Ctrl-C / Ctrl-D
			return
		case ch == ' ':
			// Paint current cell
			ed.paint()
		}
	}
}

func (ed *TerrainEditor) paint() {
	ed.terrain.Cells[ed.cursorR][ed.cursorC] = ed.brush
	ed.dirty = true
	ed.drawTerrainCell(ed.cursorR, ed.cursorC)
}

func (ed *TerrainEditor) drawFullScreen() {
	fmt.Print(slyio.ClearScreenCode())

	// Title bar
	fmt.Printf("%s1;1H%s=--=-- Slycrel Terrain Editor ---=--=%s", slyio.ESC, slyio.ESC+"1;33m", slyio.Reset)

	// Column numbers
	fmt.Printf("%s2;2H%s", slyio.ESC, slyio.ESC+"0;36m")
	for c := 0; c < model.GridCols; c++ {
		fmt.Print(string(rune('1' + c%10)))
	}

	// Draw grid border and content
	ed.drawGrid()

	// Legend
	row := model.GridRows + 5
	fmt.Printf("%s%d;1H%s", slyio.ESC, row, slyio.ESC+"0;36m")
	fmt.Print("I,J,K,L/Arrows - Move   T - Trace   Space - Paint   C - Clear\r\n")
	fmt.Print("0-Nothing  1-Plains  2-Road  3-Water  4-Bridge\r\n")
	fmt.Print("5-Boulder  6-Forest  7-Deep Forest  8-Swamp  9-Deep Swamp\r\n")
	fmt.Print("S - Save   R - Retrieve/Load   N - New   Q - Quit\r\n")

	// Info panel
	ed.updateBrushDisplay()
	ed.updateTraceDisplay()
	ed.updateFileDisplay()
}

func (ed *TerrainEditor) drawGrid() {
	// Top border
	fmt.Printf("%s3;1H%s+%s+", slyio.ESC, slyio.ESC+"1;37m", strings.Repeat("-", model.GridCols))

	for r := 0; r < model.GridRows; r++ {
		fmt.Printf("%s%d;1H%s|", slyio.ESC, r+4, slyio.ESC+"1;37m")
		for c := 0; c < model.GridCols; c++ {
			td := mechanics.GetTerrainDisplay(ed.terrain.Cells[r][c])
			fmt.Printf("%s%s", slyio.ESC+td.ANSIColor, td.Char)
		}
		fmt.Printf("%s|", slyio.ESC+"1;37m")
	}

	// Bottom border
	fmt.Printf("%s%d;1H%s+%s+", slyio.ESC, model.GridRows+4, slyio.ESC+"1;37m", strings.Repeat("-", model.GridCols))
}

func (ed *TerrainEditor) drawTerrainCell(r, c int) {
	td := mechanics.GetTerrainDisplay(ed.terrain.Cells[r][c])
	fmt.Printf("%s%d;%dH%s%s", slyio.ESC, r+4, c+2, slyio.ESC+td.ANSIColor, td.Char)
}

func (ed *TerrainEditor) drawCursor() {
	// Draw blinking cursor
	fmt.Printf("%s%d;%dH%s\u2588", slyio.ESC, ed.cursorR+4, ed.cursorC+2, slyio.ESC+"1;5;37m")
	// Update position display
	infoCol := model.GridCols + 4
	fmt.Printf("%s3;%dH%sPos: %2d,%2d  ", slyio.ESC, infoCol, slyio.ESC+"0;36m", ed.cursorR+1, ed.cursorC+1)
}

func (ed *TerrainEditor) eraseCursor() {
	ed.drawTerrainCell(ed.cursorR, ed.cursorC)
}

func (ed *TerrainEditor) updateBrushDisplay() {
	infoCol := model.GridCols + 4
	td := mechanics.GetTerrainDisplay(ed.brush)

	terrNames := [10]string{
		"Nothing", "Plains", "Road", "Water", "Bridge",
		"Boulder", "Forest", "Deep Forest", "Swamp", "Deep Swamp",
	}
	name := "?"
	if int(ed.brush) < len(terrNames) {
		name = terrNames[ed.brush]
	}

	fmt.Printf("%s1;%dH%sBrush: %s%s %s%s%s    ",
		slyio.ESC, infoCol, slyio.ESC+"0;36m",
		slyio.ESC+td.ANSIColor, td.Char,
		slyio.ESC+"0;36m", name, slyio.Reset)
}

func (ed *TerrainEditor) updateTraceDisplay() {
	infoCol := model.GridCols + 4
	if ed.trace {
		fmt.Printf("%s2;%dH%sTrace: %sON!%s  ",
			slyio.ESC, infoCol, slyio.ESC+"0;36m", slyio.ESC+"1;37;41m", slyio.Reset)
	} else {
		fmt.Printf("%s2;%dH%sTrace: %s<Off>%s",
			slyio.ESC, infoCol, slyio.ESC+"0;36m", slyio.ESC+"0;37;44m", slyio.Reset)
	}
}

func (ed *TerrainEditor) updateFileDisplay() {
	infoCol := model.GridCols + 4
	name := ed.fileName
	if name == "" {
		name = "(untitled)"
	}
	dirtyMark := ""
	if ed.dirty {
		dirtyMark = " *"
	}
	fmt.Printf("%s4;%dH%sFile: %s%s%s    ",
		slyio.ESC, infoCol, slyio.ESC+"0;36m", slyio.ESC+"1;37m", name+dirtyMark, slyio.Reset)
}

func (ed *TerrainEditor) statusMsg(msg string) {
	row := model.GridRows + 10
	fmt.Printf("%s%d;1H%s%-70s", slyio.ESC, row, slyio.ESC+"1;33m", msg)
}

func (ed *TerrainEditor) clearStatus() {
	row := model.GridRows + 10
	fmt.Printf("%s%d;1H%s%-70s", slyio.ESC, row, slyio.Reset, "")
}

func (ed *TerrainEditor) savePrompt() {
	ed.statusMsg("Save as (filename, no .json): ")

	// Read filename in raw mode
	name := ed.readString(30)
	if name == "" {
		if ed.fileName != "" {
			name = ed.fileName
		} else {
			ed.statusMsg("Save cancelled.")
			return
		}
	}

	ed.terrain.Name = name
	ed.fileName = name

	data, err := json.MarshalIndent(ed.terrain, "", "  ")
	if err != nil {
		ed.statusMsg(fmt.Sprintf("Error: %v", err))
		return
	}

	path := filepath.Join(ed.dataDir, "terrain", name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		ed.statusMsg(fmt.Sprintf("Error saving: %v", err))
		return
	}

	ed.dirty = false
	ed.statusMsg(fmt.Sprintf("Saved to %s.json!", name))
	ed.updateFileDisplay()
}

func (ed *TerrainEditor) loadPrompt() {
	// List available terrain files
	entries, err := os.ReadDir(filepath.Join(ed.dataDir, "terrain"))
	if err != nil {
		ed.statusMsg("No terrain directory found!")
		return
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}

	if len(names) == 0 {
		ed.statusMsg("No terrain files found!")
		return
	}

	ed.statusMsg(fmt.Sprintf("Available: %s", strings.Join(names, ", ")))

	// Wait a beat then prompt
	row := model.GridRows + 11
	fmt.Printf("%s%d;1H%sLoad filename: ", slyio.ESC, row, slyio.ESC+"0;36m")

	name := ed.readString(30)
	// Clear the extra line
	fmt.Printf("%s%d;1H%-70s", slyio.ESC, row, "")

	if name == "" {
		ed.clearStatus()
		return
	}

	path := filepath.Join(ed.dataDir, "terrain", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		ed.statusMsg(fmt.Sprintf("Error loading: %v", err))
		return
	}

	var terrain model.TerrainMap
	if err := json.Unmarshal(data, &terrain); err != nil {
		ed.statusMsg(fmt.Sprintf("Error parsing: %v", err))
		return
	}

	ed.terrain = terrain
	ed.fileName = name
	ed.dirty = false
	ed.drawGrid()
	ed.updateFileDisplay()
	ed.statusMsg(fmt.Sprintf("Loaded %s.json!", name))
}

func (ed *TerrainEditor) newMapPrompt() {
	if ed.dirty {
		ed.statusMsg("Unsaved changes! Save first? (Y/N)")
		b := make([]byte, 1)
		os.Stdin.Read(b)
		if b[0] == 'y' || b[0] == 'Y' {
			ed.savePrompt()
		}
	}

	ed.terrain = model.TerrainMap{}
	ed.fileName = ""
	ed.dirty = false
	ed.cursorR = 0
	ed.cursorC = 0
	ed.drawGrid()
	ed.updateFileDisplay()
	ed.clearStatus()
}

// readString reads a string in raw mode with backspace support.
func (ed *TerrainEditor) readString(maxLen int) string {
	var buf []byte
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			return string(buf)
		}

		ch := b[0]
		if ch == '\r' || ch == '\n' {
			return string(buf)
		}
		if ch == 27 || ch == 3 {
			return ""
		}
		if ch == 127 || ch == 8 {
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Print("\b \b")
			}
			continue
		}
		if ch >= 32 && ch < 127 && len(buf) < maxLen {
			buf = append(buf, ch)
			fmt.Print(string(ch))
		}
	}
}
