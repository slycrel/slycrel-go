// Command slycrel-server runs the Slycrel game as an HTTP + WebSocket server.
// Players connect via a browser at / and play over WebSocket.
// Static files are served from web/.
// JSON data APIs are available at /api/*.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/game/states"
	slyio "github.com/slycrel/slycrel/internal/io"
	"github.com/slycrel/slycrel/internal/store"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// sessionCounter generates monotonically increasing session IDs.
var sessionCounter atomic.Int64

// peer tracks an active player connection.
// done is closed when the session goroutine exits.
type peer struct {
	terminal  *slyio.WebSocketTerminal
	sessionID string
	done      chan struct{}
}

// peerRegistry is the mutex-protected map of active WebSocket sessions.
// It prevents duplicate logins and enables transparent reconnection.
type peerRegistry struct {
	mu    sync.Mutex
	peers map[string]*peer
}

func newPeerRegistry() *peerRegistry {
	return &peerRegistry{peers: make(map[string]*peer)}
}

func (r *peerRegistry) get(username string) (*peer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.peers[username]
	return p, ok
}

func (r *peerRegistry) add(username string, p *peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[username] = p
}

func (r *peerRegistry) remove(username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, username)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dataDir := flag.String("data", "", "Path to data/ directory (auto-detected if empty)")
	stateDir := flag.String("state", "", "Path to mutable state directory (default: ./state)")
	webDir := flag.String("web", "", "Path to web/ directory (auto-detected if empty)")
	flag.Parse()

	// Auto-detect paths relative to the binary location.
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("os.Executable: %v", err)
	}
	binDir := filepath.Dir(execPath)
	repoRoot := findRepoRoot(binDir)

	if *dataDir == "" {
		*dataDir = filepath.Join(repoRoot, "data")
	}
	if *stateDir == "" {
		*stateDir = filepath.Join(repoRoot, "state")
	}
	if *webDir == "" {
		*webDir = filepath.Join(repoRoot, "web")
	}

	log.Printf("data dir  : %s", *dataDir)
	log.Printf("state dir : %s", *stateDir)
	log.Printf("web dir   : %s", *webDir)

	st, err := store.NewJSONStore(*dataDir, *stateDir)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	engine := game.NewGameEngine(st, *dataDir)
	engine.RegisterStates = states.RegisterAll

	registry := newPeerRegistry()

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/", http.FileServer(http.Dir(*webDir)))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("health: encode error: %v", err)
		}
	})

	// API endpoints
	mux.HandleFunc("/api/characters", apiHandler(func() (any, error) {
		return st.ListCharacters()
	}))
	mux.HandleFunc("/api/monsters", apiHandler(func() (any, error) {
		regions := []string{"forest", "mountain", "swamp"}
		result := map[string]any{}
		for _, region := range regions {
			m, err := st.LoadMonsters(region)
			if err != nil {
				return nil, fmt.Errorf("loading %s monsters: %w", region, err)
			}
			result[region] = m
		}
		return result, nil
	}))
	mux.HandleFunc("/api/weapons", apiHandler(func() (any, error) {
		return st.LoadWeapons()
	}))
	mux.HandleFunc("/api/armor", apiHandler(func() (any, error) {
		return st.LoadArmor()
	}))

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}

		log.Printf("ws: player %q connected from %s", username, r.RemoteAddr)

		// Reconnect path: swap the WS connection into the running session goroutine.
		if existing, ok := registry.get(username); ok {
			log.Printf("ws: reconnecting player %q (session %s)", username, existing.sessionID)
			existing.terminal.SwapConn(conn)
			existing.terminal.SendSessionRestored(username, existing.sessionID)
			// Hold the HTTP handler open until the session ends.
			<-existing.done
			log.Printf("ws: player %q session ended after reconnect", username)
			return
		}

		// New session: create terminal, register peer, start session goroutine.
		sessionID := newSessionID(username)
		terminal := slyio.NewWebSocketTerminal(conn, *dataDir)
		terminal.SendSessionInit(username, sessionID)

		done := make(chan struct{})
		registry.add(username, &peer{
			terminal:  terminal,
			sessionID: sessionID,
			done:      done,
		})

		go func() {
			defer func() {
				registry.remove(username)
				close(done)
				log.Printf("ws: player %q disconnected (session %s)", username, sessionID)
			}()

			if err := engine.RunLocalSession(terminal, username, sessionID); err != nil {
				log.Printf("ws: session error for %q: %v", username, err)
			}
			terminal.SendSessionEnd()
		}()

		// Block until session ends so the HTTP handler keeps the WS alive.
		<-done
	})

	log.Printf("Slycrel server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// newSessionID returns a unique session identifier for a player.
func newSessionID(username string) string {
	return fmt.Sprintf("%s-%d", username, sessionCounter.Add(1))
}

// apiHandler wraps a data-loading function and returns a JSON HTTP handler.
func apiHandler(fn func() (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		data, err := fn()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encErr != nil {
				log.Printf("apiHandler: encode error response: %v", encErr)
			}
			return
		}
		if encErr := json.NewEncoder(w).Encode(data); encErr != nil {
			log.Printf("apiHandler: encode data: %v", encErr)
		}
	}
}

// findRepoRoot walks up from dir looking for go.mod to find the repo root.
// Falls back to the current working directory.
func findRepoRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	cwd, _ := os.Getwd()
	return cwd
}
