package debugapi

import (
    "net/http"

    "fisherevans.com/project/f/internal/game"
)

type LoadSaveRequest struct {
    Id string `json:"id"`
}

func (s *Server) handleLoadSave(w http.ResponseWriter, r *http.Request) {
    var req LoadSaveRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    if req.Id == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        if err := game.SwapSave(req.Id); err != nil {
            return nil, err
        }
        // Reset into the boot state so the new save takes full effect (the live
        // state holds references to the previous save's data).
        s.rt.Reset()
        return map[string]string{"status": "ok", "save": req.Id}, nil
    })
}
