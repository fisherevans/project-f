package debugapi

import (
    "net/http"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    ready, state, frame := s.rt.Health()
    writeJSON(w, http.StatusOK, HealthResponse{
        Ready: ready,
        State: state,
        Frame: frame,
    })
}

func (s *Server) handleScreenshot(w http.ResponseWriter, r *http.Request) {
    layer := r.URL.Query().Get("layer")
    if layer == "" {
        layer = "scene"
    }
    png, err := s.rt.CapturePNG(layer)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "image/png")
    w.WriteHeader(http.StatusOK)
    w.Write(png)
}

func (s *Server) handleInput(w http.ResponseWriter, r *http.Request) {
    var req InputRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    frames := req.Frames
    if frames <= 0 {
        frames = 1
    }
    s.rt.InjectInput(req.A, req.B, req.Start, req.Select, req.Dir, frames)
    writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "frames": frames})
}

func (s *Server) handleClearInput(w http.ResponseWriter, r *http.Request) {
    s.rt.ClearInput()
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTime(w http.ResponseWriter, r *http.Request) {
    var req TimeRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    paused := req.Mode == "pause"
    s.rt.SetTime(paused, req.Speed)
    writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "paused": paused})
}

func (s *Server) handleStep(w http.ResponseWriter, r *http.Request) {
    var req StepRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    frames := req.Frames
    if frames <= 0 {
        frames = 1
    }
    s.rt.Step(frames, req.Dt)
    writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "frames": frames})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        s.rt.Reset()
        return map[string]string{"status": "ok"}, nil
    })
}
