package debugapi

import (
	"fmt"
	"net/http"

	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/overlays"
	"fisherevans.com/project/f/internal/schema"
)

type HighlightRequest struct {
	FlowName string                        `json:"flowName,omitempty"`
	Flow     *schema.OverlayFlow           `json:"flow,omitempty"`
	Rects    map[string]schema.OverlayRect `json:"rects,omitempty"`
}

func (s *Server) handleSetHighlight(w http.ResponseWriter, r *http.Request) {
	var req HighlightRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
		var flow schema.OverlayFlow
		if req.FlowName != "" {
			f, ok := overlays.GetFlow(req.FlowName)
			if !ok {
				return nil, fmt.Errorf("unknown flow: %s", req.FlowName)
			}
			flow = f
		} else if req.Flow != nil {
			flow = *req.Flow
		} else {
			return nil, fmt.Errorf("provide flowName or flow")
		}
		targets, err := overlays.ResolveTargetsWithExtra(flow, req.Rects)
		if err != nil {
			return nil, fmt.Errorf("resolve targets: %w", err)
		}
		advState.SetHighlightSequence(targets)
		return map[string]string{
			"status":      "ok",
			"targetCount": fmt.Sprintf("%d", len(targets)),
		}, nil
	})
}

func (s *Server) handleDismissHighlight(w http.ResponseWriter, r *http.Request) {
	s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
		advState.DismissHighlight()
		return map[string]string{"status": "ok"}, nil
	})
}
