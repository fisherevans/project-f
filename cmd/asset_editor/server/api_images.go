package server

import (
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func (s *Server) handleServeImage(w http.ResponseWriter, r *http.Request) {
    rel := r.PathValue("path")

    clean := filepath.Clean(rel)
    if strings.Contains(clean, "..") {
        http.Error(w, "invalid path", http.StatusBadRequest)
        return
    }

    abs := filepath.Join(s.assetsDir, "sprites", clean+".png")

    if _, err := os.Stat(abs); err != nil {
        http.NotFound(w, r)
        return
    }

    w.Header().Set("Cache-Control", "no-cache")
    http.ServeFile(w, r, abs)
}
