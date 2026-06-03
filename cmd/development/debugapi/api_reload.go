package debugapi

import (
    "net/http"
)

// ReloadFunc reloads a named content kind (e.g. "scripts", "rpg", "overlays")
// from disk into the running process. Runs on the game thread. Returns a human
// readable status message.
type ReloadFunc func(kind string) (string, error)

var reloadFn ReloadFunc

// SetReloadFunc installs the content-reload implementation. Wired by the dev
// launcher so debugapi stays decoupled from the game/adventure packages.
func SetReloadFunc(fn ReloadFunc) {
    reloadFn = fn
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
    kind := r.PathValue("kind")
    if reloadFn == nil {
        writeJSON(w, http.StatusNotImplemented, ReloadResponse{
            Kind:   kind,
            Status: "error",
            Message: "reload not wired",
        })
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        msg, err := reloadFn(kind)
        if err != nil {
            return ReloadResponse{Kind: kind, Status: "error", Message: err.Error()}, nil
        }
        return ReloadResponse{Kind: kind, Status: "ok", Message: msg}, nil
    })
}
