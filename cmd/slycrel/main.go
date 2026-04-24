// cmd/slycrel — standalone single-player console app.
//
// Runs the game against the local terminal (stdin/stdout with ANSI escape
// codes). The username prompt at launch is just a save-file key — there's no
// password check because this is a single-player local binary. For a
// networked multi-user setup with authentication, use cmd/server plus a
// WebSocket client (clients/godot, etc.).
//
// Usage:
//
//	go build -o slycrel ./cmd/slycrel
//	./slycrel                # play the game
//	./slycrel --editor       # launch the terrain editor
//
// Flags:
//
//	--data-dir   game asset directory (default ./data)
//	--state-dir  writable state directory (default ./state)
//	--editor     launch the terrain editor instead of the game
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slycrel/slycrel/internal/editor"
	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/game/states"
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/store"
)

func main() {
	dataDir := flag.String("data-dir", "./data", "game asset directory (monsters, weapons, ansi)")
	stateDir := flag.String("state-dir", "./state", "writable state directory (characters, inn, etc.)")
	runEditor := flag.Bool("editor", false, "launch the terrain editor instead of the game")
	flag.Parse()

	if *runEditor {
		if err := editor.RunTerrainEditor(*dataDir); err != nil {
			log.Fatalf("editor: %v", err)
		}
		return
	}

	if err := os.MkdirAll(*stateDir, 0o755); err != nil {
		log.Fatalf("create state-dir: %v", err)
	}

	st, err := store.NewJSONStore(*dataDir, *stateDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	engine := game.NewGameEngine(st, *dataDir)
	engine.RegisterStates = states.RegisterAll

	term := slyio.NewLocalTerminal(*dataDir)
	if err := term.EnableRawMode(); err != nil {
		log.Fatalf("raw mode: %v", err)
	}
	defer term.RestoreTerminal()

	username := strings.TrimSpace(term.ReadLine("Name:", 20))
	if username == "" {
		return
	}

	if err := engine.RunLocalSession(term, username); err != nil {
		term.RestoreTerminal()
		fmt.Fprintf(os.Stderr, "session: %v\n", err)
		os.Exit(1)
	}
}
