package debugapi

import (
    "fmt"
    "net/http"

    "fisherevans.com/project/f/internal/game"
)

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
    var req CommandRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        output := game.Console().RunCapture(func() {
            game.GetActiveState().HandleConsoleInput(req.Command)
        })
        return CommandResponse{Output: output}, nil
    })
}

func (s *Server) handleTeleport(w http.ResponseWriter, r *http.Request) {
    var req TeleportRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        var cmd string
        if req.Target != "" {
            cmd = "tp " + req.Target
        } else if req.X != nil && req.Y != nil {
            cmd = fmt.Sprintf("tp %d %d", *req.X, *req.Y)
        } else {
            return map[string]string{"error": "provide target or x/y"}, nil
        }
        output := game.Console().RunCapture(func() {
            game.GetActiveState().HandleConsoleInput(cmd)
        })
        return CommandResponse{Output: output}, nil
    })
}

func (s *Server) handleLoadMap(w http.ResponseWriter, r *http.Request) {
    var req MapRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        cmd := "map " + req.Name
        if req.Waypoint != "" {
            cmd += " " + req.Waypoint
        }
        output := game.Console().RunCapture(func() {
            game.GetActiveState().HandleConsoleInput(cmd)
        })
        return CommandResponse{Output: output}, nil
    })
}
