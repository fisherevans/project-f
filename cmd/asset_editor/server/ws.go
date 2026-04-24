package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// FileEvent is broadcast to connected WebSocket clients when a sprite file
// changes on disk.
type FileEvent struct {
	Type string `json:"type"` // "change", "create", "remove"
	Path string `json:"path"` // relative to assets dir
}

// Hub manages WebSocket clients and filesystem watching.
type Hub struct {
	assetsDir string

	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}

	broadcast chan FileEvent
}

func NewHub(assetsDir string) *Hub {
	return &Hub{
		assetsDir: assetsDir,
		clients:   make(map[*websocket.Conn]struct{}),
		broadcast: make(chan FileEvent, 64),
	}
}

// Run starts the filesystem watcher and broadcast loop. It blocks forever.
func (h *Hub) Run() {
	go h.watch()

	for ev := range h.broadcast {
		data, err := json.Marshal(ev)
		if err != nil {
			continue
		}

		h.mu.Lock()
		for conn := range h.clients {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				conn.Close()
				delete(h.clients, conn)
			}
		}
		h.mu.Unlock()
	}
}

func (h *Hub) addClient(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

func (h *Hub) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("fsnotify: %v", err)
		return
	}
	defer watcher.Close()

	spritesDir := filepath.Join(h.assetsDir, "sprites")
	if err := addDirRecursive(watcher, spritesDir); err != nil {
		log.Printf("fsnotify add: %v", err)
	}

	// Debounce: collapse rapid events on the same path.
	const debounce = 200 * time.Millisecond
	pending := map[string]*time.Timer{}
	var mu sync.Mutex

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}

			ext := strings.ToLower(filepath.Ext(ev.Name))
			if ext != ".png" && ext != ".yaml" && ext != ".yml" {
				continue
			}

			rel, _ := filepath.Rel(h.assetsDir, ev.Name)

			mu.Lock()
			if t, exists := pending[rel]; exists {
				t.Stop()
			}
			pending[rel] = time.AfterFunc(debounce, func() {
				mu.Lock()
				delete(pending, rel)
				mu.Unlock()

				evType := "change"
				if ev.Has(fsnotify.Create) {
					evType = "create"
				} else if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
					evType = "remove"
				}
				h.broadcast <- FileEvent{Type: evType, Path: rel}
			})
			mu.Unlock()

			// If a new directory was created, watch it too.
			if ev.Has(fsnotify.Create) {
				_ = addDirRecursive(watcher, ev.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("fsnotify error: %v", err)
		}
	}
}

func addDirRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		return w.Add(path)
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}

	s.hub.addClient(conn)

	// Read loop - just drain messages so we detect close.
	go func() {
		defer s.hub.removeClient(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}
