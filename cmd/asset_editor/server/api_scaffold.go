package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

func (s *Server) handleScaffold(w http.ResponseWriter, r *http.Request) {
	var req ScaffoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	args := []string{
		"run", "./cmd/sprite_new",
		"-name", req.Name,
		"-tile-width", fmt.Sprintf("%d", req.TileWidth),
		"-tile-height", fmt.Sprintf("%d", req.TileHeight),
		"-cols", fmt.Sprintf("%d", req.Cols),
		"-rows", fmt.Sprintf("%d", req.Rows),
	}
	if len(req.Sprites) > 0 {
		args = append(args, "-sprites", strings.Join(req.Sprites, ","))
	}
	if req.Force {
		args = append(args, "-force")
	}

	cmd := exec.Command("go", args...)
	// Run from the project root (two levels up from assets dir).
	cmd.Dir = s.assetsDir + "/.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("scaffold failed: %s\n%s", err, output), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"status": "ok",
		"output": string(output),
	})
}
