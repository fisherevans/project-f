package debugapi

import (
    "net/http"

    "fisherevans.com/project/f/internal/util/highlighter"
)

type HighlightController interface {
    Show(targets []highlighter.Target, durationPerStep float64)
    Dismiss()
}

type Server struct {
    mux       *http.ServeMux
    queue     *CommandQueue
    highlight HighlightController
}

func NewServer(queue *CommandQueue, highlight HighlightController) *Server {
    s := &Server{
        mux:       http.NewServeMux(),
        queue:     queue,
        highlight: highlight,
    }
    s.routes()
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
    s.mux.HandleFunc("GET /api/v1/debug/state", s.handleGetState)
    s.mux.HandleFunc("GET /api/v1/debug/save", s.handleGetSave)

    s.mux.HandleFunc("GET /api/v1/debug/globals", s.handleListGlobals)
    s.mux.HandleFunc("GET /api/v1/debug/globals/{key...}", s.handleGetGlobal)
    s.mux.HandleFunc("POST /api/v1/debug/globals/{key...}", s.handleSetGlobal)
    s.mux.HandleFunc("DELETE /api/v1/debug/globals/{key...}", s.handleDeleteGlobal)

    s.mux.HandleFunc("POST /api/v1/debug/command", s.handleCommand)
    s.mux.HandleFunc("POST /api/v1/debug/teleport", s.handleTeleport)
    s.mux.HandleFunc("POST /api/v1/debug/map", s.handleLoadMap)

    s.mux.HandleFunc("GET /api/v1/debug/entities", s.handleListEntities)
    s.mux.HandleFunc("GET /api/v1/debug/entities/{id}", s.handleGetEntity)
    s.mux.HandleFunc("GET /api/v1/debug/teleports", s.handleListTeleports)
    s.mux.HandleFunc("GET /api/v1/debug/zones", s.handleListZones)

    s.mux.HandleFunc("POST /api/v1/debug/highlight", s.handleSetHighlight)
    s.mux.HandleFunc("DELETE /api/v1/debug/highlight", s.handleDismissHighlight)
}
