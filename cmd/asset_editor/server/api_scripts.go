package server

import (
	"encoding/json"
	"io"
	"net/http"

	"fisherevans.com/project/f/internal/schema"
)

func (s *Server) handleGetScriptSchema(w http.ResponseWriter, r *http.Request) {
	ss, err := schema.LoadScriptSchema()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, ss)
}

func (s *Server) handleListScripts(w http.ResponseWriter, r *http.Request) {
	entries, err := s.scripts.ListScripts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, entries)
}

func (s *Server) handleGetScript(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	detail, err := s.scripts.GetScript(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleSaveScript(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	body, _ := io.ReadAll(r.Body)
	var req struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		req.Content = string(body)
	}

	if req.Content == "" {
		http.Error(w, "empty content", http.StatusBadRequest)
		return
	}

	if err := s.scripts.SaveScript(name, req.Content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	detail, err := s.scripts.GetScript(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleDeleteScript(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.scripts.DeleteScript(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
