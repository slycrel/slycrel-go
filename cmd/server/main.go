// cmd/server — headless WebSocket game server.
//
// Clients connect via WebSocket at /ws, authenticate with a hello message,
// and play the game entirely through JSON message exchange.
// See internal/io/ws_protocol.go for the full message schema.
//
// Usage:
//
//	go run ./cmd/server [flags]
//
// Flags:
//
//	--port      TCP port to listen on (default 8080)
//	--data-dir  Path to game asset directory containing monsters/, weapons/, ansi/ (default ./data)
//	--state-dir Path to writable directory for character/game state files (default ./state)
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/game/states"
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/model"
	"github.com/slycrel/slycrel/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "TCP port to listen on")
	dataDir := flag.String("data-dir", "./data", "game asset directory (monsters, weapons, ansi)")
	stateDir := flag.String("state-dir", "./state", "writable state directory (characters, inn, etc.)")
	allowNewUsers := flag.Bool("allow-new-users", false, "auto-register unknown usernames at handshake (dev only — connecting user's first password becomes their account password)")
	_ = flag.Bool("headless", true, "run as a headless server (browser/WebSocket clients only; no local terminal I/O) — this is the default and only mode for cmd/server")
	flag.Parse()

	if err := os.MkdirAll(*stateDir, 0755); err != nil {
		log.Fatalf("create state-dir: %v", err)
	}

	st, err := store.NewJSONStore(*dataDir, *stateDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	engine := game.NewGameEngine(st, *dataDir)
	engine.RegisterStates = states.RegisterAll

	auth := func(username, password string) error {
		err := st.AuthenticatePlayer(username, password)
		if err == nil {
			return nil
		}
		if !*allowNewUsers {
			return err
		}
		// Dev-mode auto-registration: if the user doesn't exist yet, create a
		// blank character with BBSName + password. The in-game state machine
		// routes them through create_character on first login, which fills in
		// Name / class / stats.
		if _, findErr := st.FindCharacterByBBSName(username); findErr != nil && errors.Is(findErr, store.ErrNotFound) {
			if err := st.AddCharacter(&model.Character{BBSName: username}); err != nil {
				return err
			}
			if err := st.SetPlayerPassword(username, password); err != nil {
				return err
			}
			log.Printf("auto-registered new user %q (dev mode)", username)
			return nil
		}
		return err
	}

	wsHandler := slyio.WSUpgrader(*dataDir, auth, func(sess *slyio.WSSession, username string) {
		if err := engine.RunWSSession(sess, username); err != nil {
			log.Printf("session %q ended: %v", username, err)
		}
	})

	http.Handle("/ws", wsHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("slycrel server listening on %s  (data=%s  state=%s)", addr, *dataDir, *stateDir)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", addr)
	log.Printf("Protocol version: %d  (see internal/io/ws_protocol.go)", slyio.WSProtocolVersion)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
