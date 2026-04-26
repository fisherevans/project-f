package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleListTiledUsages(w http.ResponseWriter, r *http.Request) {
	usages, err := s.tiled.ListHandlerUsages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usages)
}
