package server

import (
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "fisherevans.com/project/f/internal/schema"
)

func (s *Server) handleListAudio(w http.ResponseWriter, r *http.Request) {
    entries, err := s.audio.ListAudio()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, entries)
}

func (s *Server) handleGetAudio(w http.ResponseWriter, r *http.Request) {
    name := r.PathValue("name")
    detail, err := s.audio.GetAudio(name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    writeJSON(w, detail)
}

func (s *Server) handleSaveAudio(w http.ResponseWriter, r *http.Request) {
    name := r.PathValue("name")

    var meta schema.AudioMetadata
    if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
        http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
        return
    }

    if err := s.audio.SaveAudio(name, &meta); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    detail, err := s.audio.GetAudio(name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, detail)
}

func (s *Server) handleServeAudioFile(w http.ResponseWriter, r *http.Request) {
    rel := r.PathValue("path")

    clean := filepath.Clean(rel)
    if strings.Contains(clean, "..") {
        http.Error(w, "invalid path", http.StatusBadRequest)
        return
    }

    abs := s.audio.AudioFilePath(clean)
    if abs == "" {
        http.NotFound(w, r)
        return
    }

    if _, err := os.Stat(abs); err != nil {
        http.NotFound(w, r)
        return
    }

    w.Header().Set("Cache-Control", "no-cache")
    http.ServeFile(w, r, abs)
}
