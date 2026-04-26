package server

import (
    "encoding/json"
    "net/http"
    "os"
    "sort"

    "fisherevans.com/project/f/internal/game/rpg"
    "fisherevans.com/project/f/internal/schema"
)

// ---- Handlers ----

func (s *Server) handleListSkills(w http.ResponseWriter, r *http.Request) {
    skills, err := s.rpg.ListSkills()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
    writeRPGJSON(w, skills)
}

func (s *Server) handleGetSkill(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    skill, err := s.rpg.GetSkill(id)
    if err != nil {
        if os.IsNotExist(err) {
            http.NotFound(w, r)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    writeRPGJSON(w, skill)
}

func (s *Server) handleSaveSkill(w http.ResponseWriter, r *http.Request) {
    var skill schema.RPGSkill
    if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
        http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    if skill.Id == "" {
        http.Error(w, "missing skill id", http.StatusBadRequest)
        return
    }
    if err := s.rpg.SaveSkill(skill); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteSkill(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if err := s.rpg.DeleteSkill(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListPrimortals(w http.ResponseWriter, r *http.Request) {
    primortals, err := s.rpg.ListPrimortals()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    sort.Slice(primortals, func(i, j int) bool { return primortals[i].Name < primortals[j].Name })
    writeRPGJSON(w, primortals)
}

func (s *Server) handleGetPrimortal(w http.ResponseWriter, r *http.Request) {
    type_ := r.PathValue("type")
    p, err := s.rpg.GetPrimortal(type_)
    if err != nil {
        if os.IsNotExist(err) {
            http.NotFound(w, r)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }
    writeRPGJSON(w, p)
}

func (s *Server) handleSavePrimortal(w http.ResponseWriter, r *http.Request) {
    var p schema.RPGPrimortal
    if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
        http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    if p.Type == "" {
        http.Error(w, "missing primortal type", http.StatusBadRequest)
        return
    }
    if err := s.rpg.SavePrimortal(p); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeletePrimortal(w http.ResponseWriter, r *http.Request) {
    type_ := r.PathValue("type")
    if err := s.rpg.DeletePrimortal(type_); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetCombatInfo(w http.ResponseWriter, r *http.Request) {
    type StatusEntry struct {
        Type        string `json:"type"`
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    type StanceEntry struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    type CombatInfo struct {
        Statuses []StatusEntry `json:"statuses"`
        Stances  []StanceEntry `json:"stances"`
    }
    statuses := []StatusEntry{
        {Type: string(rpg.StatusWarded), Name: rpg.StatusWarded.CurrentTense(), Description: rpg.StatusWarded.Description()},
        {Type: string(rpg.StatusBurning), Name: rpg.StatusBurning.CurrentTense(), Description: rpg.StatusBurning.Description()},
        {Type: string(rpg.StatusPoisoned), Name: rpg.StatusPoisoned.CurrentTense(), Description: rpg.StatusPoisoned.Description()},
        {Type: string(rpg.StatusIonized), Name: rpg.StatusIonized.CurrentTense(), Description: rpg.StatusIonized.Description()},
        {Type: string(rpg.StatusMending), Name: rpg.StatusMending.CurrentTense(), Description: rpg.StatusMending.Description()},
    }
    stances := []StanceEntry{
        {Name: rpg.TickStanceDefending.String(), Description: rpg.TickStanceDefending.Description()},
        {Name: rpg.TickStanceReflecting.String(), Description: rpg.TickStanceReflecting.Description()},
        {Name: rpg.TickStanceVulnerable.String(), Description: rpg.TickStanceVulnerable.Description()},
        {Name: rpg.TickStanceExposed.String(), Description: rpg.TickStanceExposed.Description()},
    }
    writeRPGJSON(w, CombatInfo{Statuses: statuses, Stances: stances})
}

func writeRPGJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
