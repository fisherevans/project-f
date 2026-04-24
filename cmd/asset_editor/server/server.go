package server

import (
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Server is the asset editor HTTP server.
type Server struct {
	mux        *http.ServeMux
	assetsDir  string
	devMode    bool
	sprites    *SpriteService
	hub        *Hub
	frontendFS fs.FS
}

// New creates a configured Server. The assetsDir is the root of the game's
// assets/ folder on disk. frontendFS should be the embedded frontend/dist
// filesystem (or nil in dev mode).
func New(assetsDir string, devMode bool, frontendFS fs.FS) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		assetsDir:  assetsDir,
		devMode:    devMode,
		sprites:    NewSpriteService(assetsDir),
		hub:        NewHub(assetsDir),
		frontendFS: frontendFS,
	}
	s.routes()
	go s.hub.Run()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.devMode {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// API routes
	s.mux.HandleFunc("GET /api/v1/sprites", s.handleListSprites)
	s.mux.HandleFunc("GET /api/v1/sprites/{name...}", s.handleGetSprite)
	s.mux.HandleFunc("PUT /api/v1/sprites/{name...}", s.handleSaveSprite)
	s.mux.HandleFunc("DELETE /api/v1/sprites/{name...}", s.handleDeleteSpriteSidecar)

	s.mux.HandleFunc("GET /api/v1/images/{path...}", s.handleServeImage)

	s.mux.HandleFunc("POST /api/v1/sprites/_scaffold", s.handleScaffold)

	s.mux.HandleFunc("GET /api/v1/ws", s.handleWebSocket)

	// Frontend
	if s.devMode {
		s.mux.Handle("/", viteProxy())
	} else {
		if s.frontendFS == nil {
			log.Fatal("no embedded frontend and not in dev mode")
		}
		s.mux.Handle("/", http.FileServer(http.FS(s.frontendFS)))
	}
}

func viteProxy() http.Handler {
	target, _ := url.Parse("http://localhost:5173")
	return httputil.NewSingleHostReverseProxy(target)
}
