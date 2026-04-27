package debugapi

import (
	"fmt"
	"net/http"

	"fisherevans.com/project/f/internal/overlays"
	"fisherevans.com/project/f/internal/schema"
)

type HighlightRequest struct {
	FlowName string                        `json:"flowName,omitempty"`
	Flow     *schema.OverlayFlow           `json:"flow,omitempty"`
	Rects    map[string]schema.OverlayRect `json:"rects,omitempty"`
	Duration float64                       `json:"duration,omitempty"`
}

func (s *Server) handleSetHighlight(w http.ResponseWriter, r *http.Request) {
	var req HighlightRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	var flow schema.OverlayFlow
	if req.FlowName != "" {
		f, ok := overlays.GetFlow(req.FlowName)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown flow: " + req.FlowName})
			return
		}
		flow = f
	} else if req.Flow != nil {
		flow = *req.Flow
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provide flowName or flow"})
		return
	}
	targets, err := overlays.ResolveTargetsWithExtra(flow, req.Rects)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	duration := req.Duration
	if duration <= 0 {
		duration = 5
	}
	s.highlight.Show(targets, duration)
	writeJSON(w, http.StatusOK, map[string]string{
		"status":      "ok",
		"targetCount": fmt.Sprintf("%d", len(targets)),
	})
}

func (s *Server) handleDismissHighlight(w http.ResponseWriter, r *http.Request) {
	s.highlight.Dismiss()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
