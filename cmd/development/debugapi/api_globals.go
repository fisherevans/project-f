package debugapi

import (
    "fmt"
    "net/http"
    "sort"

    "fisherevans.com/project/f/internal/game"
)

func (s *Server) handleListGlobals(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        prefix := r.URL.Query().Get("prefix")
        globals := game.CurrentSave().Globals
        keys := globals.KeysWithPrefix(prefix)
        sort.Strings(keys)

        entries := make([]GlobalEntry, 0, len(keys))
        for _, k := range keys {
            v := globals.Get(k)
            entries = append(entries, GlobalEntry{
                Key:    k,
                Value:  v.Value(),
                Exists: v.Exists(),
            })
        }
        return entries, nil
    })
}

func (s *Server) handleGetGlobal(w http.ResponseWriter, r *http.Request) {
    key := r.PathValue("key")
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        v := game.CurrentSave().Globals.Get(key)
        return GlobalEntry{
            Key:    key,
            Value:  v.Value(),
            Exists: v.Exists(),
        }, nil
    })
}

func (s *Server) handleSetGlobal(w http.ResponseWriter, r *http.Request) {
    key := r.PathValue("key")
    var req SetGlobalRequest
    if err := readJSON(r, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
        return
    }
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        val, err := coerceGlobalValue(req.Type, req.Value)
        if err != nil {
            return map[string]string{"error": err.Error()}, nil
        }
        game.CurrentSave().Globals.Set(key, val)
        v := game.CurrentSave().Globals.Get(key)
        return GlobalEntry{
            Key:    key,
            Value:  v.Value(),
            Exists: v.Exists(),
        }, nil
    })
}

func (s *Server) handleDeleteGlobal(w http.ResponseWriter, r *http.Request) {
    key := r.PathValue("key")
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        old := game.CurrentSave().Globals.Delete(key)
        return GlobalEntry{
            Key:    key,
            Value:  old.Value(),
            Exists: old.Exists(),
        }, nil
    })
}

func coerceGlobalValue(typ string, raw any) (any, error) {
    switch typ {
    case "string":
        s, ok := raw.(string)
        if !ok {
            return nil, fmt.Errorf("value must be a string")
        }
        return s, nil
    case "int":
        f, ok := raw.(float64)
        if !ok {
            return nil, fmt.Errorf("value must be a number")
        }
        return int(f), nil
    case "float":
        f, ok := raw.(float64)
        if !ok {
            return nil, fmt.Errorf("value must be a number")
        }
        return f, nil
    case "bool":
        b, ok := raw.(bool)
        if !ok {
            return nil, fmt.Errorf("value must be a boolean")
        }
        return b, nil
    default:
        return nil, fmt.Errorf("unknown type %q (expected string, int, float, bool)", typ)
    }
}
