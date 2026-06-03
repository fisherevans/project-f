package debugapi

import (
    "net/http"

    "fisherevans.com/project/f/internal/util/highlighter"
)

type HighlightController interface {
    Show(targets []highlighter.Target, durationPerStep float64)
    Dismiss()
}

// RuntimeControl is the loop-coupled control surface the debug API drives:
// time stepping, synthetic input, frame capture, status, and full reset.
// Implemented by *runtime.Instance (delegating to its Harness).
type RuntimeControl interface {
    Health() (ready bool, state string, frame int64)
    CapturePNG(layer string) ([]byte, error)
    SetTime(paused bool, speed float64)
    Step(frames int, dt float64)
    InjectInput(a, b, start, sel bool, dir string, frames int)
    ClearInput()
    Reset()
}

type Server struct {
    mux       *http.ServeMux
    queue     *CommandQueue
    highlight HighlightController
    rt        RuntimeControl
}

func NewServer(queue *CommandQueue, highlight HighlightController, rt RuntimeControl) *Server {
    s := &Server{
        mux:       http.NewServeMux(),
        queue:     queue,
        highlight: highlight,
        rt:        rt,
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
    s.mux.HandleFunc("GET /api/v1/debug/health", s.handleHealth)
    s.mux.HandleFunc("GET /api/v1/debug/state", s.handleGetState)
    s.mux.HandleFunc("GET /api/v1/debug/save", s.handleGetSave)

    s.mux.HandleFunc("GET /api/v1/debug/globals", s.handleListGlobals)
    s.mux.HandleFunc("GET /api/v1/debug/globals/{key...}", s.handleGetGlobal)
    s.mux.HandleFunc("POST /api/v1/debug/globals/{key...}", s.handleSetGlobal)
    s.mux.HandleFunc("DELETE /api/v1/debug/globals/{key...}", s.handleDeleteGlobal)

    s.mux.HandleFunc("POST /api/v1/debug/command", s.handleCommand)
    s.mux.HandleFunc("POST /api/v1/debug/teleport", s.handleTeleport)
    s.mux.HandleFunc("POST /api/v1/debug/map", s.handleLoadMap)

    s.mux.HandleFunc("GET /api/v1/debug/combat", s.handleGetCombat)
    s.mux.HandleFunc("GET /api/v1/debug/entities", s.handleListEntities)
    s.mux.HandleFunc("GET /api/v1/debug/entities/{id}", s.handleGetEntity)
    s.mux.HandleFunc("GET /api/v1/debug/teleports", s.handleListTeleports)
    s.mux.HandleFunc("GET /api/v1/debug/zones", s.handleListZones)

    s.mux.HandleFunc("POST /api/v1/debug/highlight", s.handleSetHighlight)
    s.mux.HandleFunc("DELETE /api/v1/debug/highlight", s.handleDismissHighlight)

    // Runtime control: capture, input, time, reset, reload.
    s.mux.HandleFunc("GET /api/v1/debug/screenshot", s.handleScreenshot)
    s.mux.HandleFunc("POST /api/v1/debug/input", s.handleInput)
    s.mux.HandleFunc("DELETE /api/v1/debug/input", s.handleClearInput)
    s.mux.HandleFunc("POST /api/v1/debug/time", s.handleTime)
    s.mux.HandleFunc("POST /api/v1/debug/step", s.handleStep)
    s.mux.HandleFunc("POST /api/v1/debug/reset", s.handleReset)
    s.mux.HandleFunc("POST /api/v1/debug/reload/{kind}", s.handleReload)
}
