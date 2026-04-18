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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slycrel/slycrel/internal/game"
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/game/states"
	"github.com/slycrel/slycrel/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "TCP port to listen on")
	dataDir := flag.String("data-dir", "./data", "game asset directory (monsters, weapons, ansi)")
	stateDir := flag.String("state-dir", "./state", "writable state directory (characters, inn, etc.)")
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
		return st.AuthenticatePlayer(username, password)
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
