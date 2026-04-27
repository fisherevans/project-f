package server

import (
	"io"
	"net/http"
)

func (s *Server) handleListOverlays(w http.ResponseWriter, r *http.Request) {
	flows, err := s.overlays.ListFlows()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if flows == nil {
		flows = []OverlayFlowEntry{}
	}
	writeJSON(w, flows)
}

func (s *Server) handleGetOverlay(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	detail, err := s.overlays.GetFlow(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleSaveOverlay(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if err := s.overlays.SaveFlow(name, string(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteOverlay(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.overlays.DeleteFlow(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetNamedRects(w http.ResponseWriter, r *http.Request) {
	detail, err := s.overlays.GetNamedRects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleSaveNamedRects(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if err := s.overlays.SaveNamedRects(string(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
