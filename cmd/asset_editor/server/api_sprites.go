package server

import (
	"encoding/json"
	"log"
	"net/http"

	"fisherevans.com/project/f/internal/schema"
)

func (s *Server) handleListSprites(w http.ResponseWriter, r *http.Request) {
	entries, err := s.sprites.ListSprites()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, entries)
}

func (s *Server) handleGetSprite(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	detail, err := s.sprites.GetSprite(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleSaveSprite(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var meta schema.SpriteMetadata
	if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.sprites.SaveSprite(name, &meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	detail, err := s.sprites.GetSprite(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) handleDeleteSpriteSidecar(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.sprites.DeleteSpriteSidecar(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode: %v", err)
	}
}
