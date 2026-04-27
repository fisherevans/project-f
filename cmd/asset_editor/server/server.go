package server

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
)

// Server is the asset editor HTTP server.
type Server struct {
	mux         *http.ServeMux
	assetsDir   string
	devMode     bool
	sprites     *SpriteService
	audio       *AudioService
	scripts     *ScriptService
	rpg         *RPGService
	saves       *SaveService
	overlays    *OverlayService
	tiled       *TiledService
	tiledBridge *TiledBridge
	hub         *Hub
	frontendFS  fs.FS
}

// New creates a configured Server. The assetsDir is the root of the game's
// assets/ folder on disk. frontendFS should be the embedded frontend/dist
// filesystem (or nil in dev mode).
func New(assetsDir string, devMode bool, frontendFS fs.FS) *Server {
	s := &Server{
		mux:         http.NewServeMux(),
		assetsDir:   assetsDir,
		devMode:     devMode,
		sprites:     NewSpriteService(assetsDir),
		audio:       NewAudioService(assetsDir),
		scripts:     NewScriptService(assetsDir),
		rpg:         NewRPGService(assetsDir),
		saves:       NewSaveService(assetsDir),
		overlays:    NewOverlayService(assetsDir),
		tiled:       NewTiledService(assetsDir),
		tiledBridge: NewTiledBridge(),
		hub:         NewHub(assetsDir),
		frontendFS:  frontendFS,
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

	s.mux.HandleFunc("GET /api/v1/audio", s.handleListAudio)
	s.mux.HandleFunc("GET /api/v1/audio-files/{path...}", s.handleServeAudioFile)
	s.mux.HandleFunc("GET /api/v1/audio/{name...}", s.handleGetAudio)
	s.mux.HandleFunc("PUT /api/v1/audio/{name...}", s.handleSaveAudio)

	s.mux.HandleFunc("POST /api/v1/scripts/_validate-expr", s.handleValidateExpr)
	s.mux.HandleFunc("GET /api/v1/script-schema", s.handleGetScriptSchema)
	s.mux.HandleFunc("GET /api/v1/scripts", s.handleListScripts)
	s.mux.HandleFunc("GET /api/v1/scripts/{name...}", s.handleGetScript)
	s.mux.HandleFunc("PUT /api/v1/scripts/{name...}", s.handleSaveScript)
	s.mux.HandleFunc("DELETE /api/v1/scripts/{name...}", s.handleDeleteScript)

	s.mux.HandleFunc("GET /api/v1/templates", s.handleListPropertyTemplates)

	s.mux.HandleFunc("GET /api/v1/overlays", s.handleListOverlays)
	s.mux.HandleFunc("GET /api/v1/overlays/rects", s.handleGetNamedRects)
	s.mux.HandleFunc("PUT /api/v1/overlays/rects", s.handleSaveNamedRects)
	s.mux.HandleFunc("GET /api/v1/overlays/{name...}", s.handleGetOverlay)
	s.mux.HandleFunc("PUT /api/v1/overlays/{name...}", s.handleSaveOverlay)
	s.mux.HandleFunc("DELETE /api/v1/overlays/{name...}", s.handleDeleteOverlay)

	s.mux.HandleFunc("GET /api/v1/rpg/skills", s.handleListSkills)
	s.mux.HandleFunc("GET /api/v1/rpg/skills/{id}", s.handleGetSkill)
	s.mux.HandleFunc("PUT /api/v1/rpg/skills/{id}", s.handleSaveSkill)
	s.mux.HandleFunc("DELETE /api/v1/rpg/skills/{id}", s.handleDeleteSkill)
	s.mux.HandleFunc("GET /api/v1/rpg/primortals", s.handleListPrimortals)
	s.mux.HandleFunc("GET /api/v1/rpg/primortals/{type}", s.handleGetPrimortal)
	s.mux.HandleFunc("PUT /api/v1/rpg/primortals/{type}", s.handleSavePrimortal)
	s.mux.HandleFunc("DELETE /api/v1/rpg/primortals/{type}", s.handleDeletePrimortal)
	s.mux.HandleFunc("GET /api/v1/rpg/combat", s.handleGetCombatInfo)

	s.mux.HandleFunc("GET /api/v1/saves", s.handleListSaves)
	s.mux.HandleFunc("GET /api/v1/saves/{id}", s.handleGetSave)
	s.mux.HandleFunc("PUT /api/v1/saves/{id}", s.handleSaveSave)
	s.mux.HandleFunc("DELETE /api/v1/saves/{id}", s.handleDeleteSave)
	s.mux.HandleFunc("POST /api/v1/saves/{id}/clone", s.handleCloneSave)

	s.mux.HandleFunc("GET /api/v1/tiled/usages", s.handleListTiledUsages)

	s.mux.HandleFunc("POST /api/v1/tiled-bridge/selection", s.handleTiledBridgeSelection)
	s.mux.HandleFunc("GET /api/v1/tiled-bridge/selection", s.handleTiledBridgeGetSelection)
	s.mux.HandleFunc("GET /api/v1/tiled-bridge/commands", s.handleTiledBridgeCommands)
	s.mux.HandleFunc("POST /api/v1/tiled-bridge/commands", s.handleTiledBridgeEnqueueCommand)
	s.mux.HandleFunc("POST /api/v1/tiled-bridge/heartbeat", s.handleTiledBridgeHeartbeat)
	s.mux.HandleFunc("POST /api/v1/tiled-bridge/ack", s.handleTiledBridgeAck)
	s.mux.HandleFunc("GET /api/v1/tiled-bridge/status", s.handleTiledBridgeStatus)

	s.mux.HandleFunc("GET /api/v1/ws", s.handleWebSocket)

	// Frontend
	if s.devMode {
		s.mux.Handle("/", viteProxy())
	} else {
		if s.frontendFS == nil {
			log.Fatal("no embedded frontend and not in dev mode")
		}
		s.mux.Handle("/", spaHandler(s.frontendFS))
	}
}

func safePath(root, name string) (string, error) {
	joined := filepath.Join(root, name)
	cleaned := filepath.Clean(joined)
	if !strings.HasPrefix(cleaned, filepath.Clean(root)+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", name)
	}
	return cleaned, nil
}

func viteProxy() http.Handler {
	target, _ := url.Parse("http://localhost:5173")
	return httputil.NewSingleHostReverseProxy(target)
}

// spaHandler serves static files from the embedded FS, falling back to
// index.html for any path that doesn't match a real file. This lets
// React Router handle client-side routes like /sprites/foo/bar.
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(fsys, path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
