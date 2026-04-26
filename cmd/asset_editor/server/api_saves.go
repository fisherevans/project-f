package server

import (
    "encoding/json"
    "io"
    "net/http"
    "os"
    "sort"
)

func (s *Server) handleListSaves(w http.ResponseWriter, r *http.Request) {
    saves, err := s.saves.ListSaves()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    sort.Slice(saves, func(i, j int) bool { return saves[i].SaveId < saves[j].SaveId })
    writeRPGJSON(w, saves)
}

func (s *Server) handleGetSave(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    save, err := s.saves.GetSave(id)
    if err != nil {
        if os.IsNotExist(err) {
            http.NotFound(w, r)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    writeRPGJSON(w, save)
}

func (s *Server) handleSaveSave(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read body", http.StatusBadRequest)
        return
    }
    var req struct {
        RawYaml string `json:"raw_yaml"`
    }
    if err := json.Unmarshal(body, &req); err != nil {
        http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    if req.RawYaml == "" {
        http.Error(w, "missing raw_yaml", http.StatusBadRequest)
        return
    }
    if err := s.saves.SaveRaw(id, req.RawYaml); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteSave(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if err := s.saves.DeleteSave(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCloneSave(w http.ResponseWriter, r *http.Request) {
    srcId := r.PathValue("id")
    var req struct {
        NewId string `json:"new_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    if req.NewId == "" {
        http.Error(w, "missing new_id", http.StatusBadRequest)
        return
    }
    if err := s.saves.CloneSave(srcId, req.NewId); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
