package debugapi

import (
    "encoding/json"
    "net/http"
    "reflect"

    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/states/adventure"
    "fisherevans.com/project/f/internal/game/states/combat"
)

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        active := game.GetActiveState()
        typeName := reflect.TypeOf(active).String()

        info := StateInfo{
            Type: typeName,
        }

        switch st := active.(type) {
        case *adventure.State:
            info.Capabilities = []string{"entities", "teleport", "map", "command"}
            info.MapName = st.MapName()
            if p := st.Globals().Player(); p != nil {
                loc := p.GetPreciseLocation()
                info.PlayerPos = &Vec2{X: loc.X, Y: loc.Y}
            }
        case *combat.State:
            info.Capabilities = []string{"combat", "command"}
        default:
            info.Capabilities = []string{"command"}
        }

        return info, nil
    })
}

func (s *Server) handleGetSave(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        save := game.CurrentSave()
        data, err := json.Marshal(save)
        if err != nil {
            return nil, err
        }
        var result any
        json.Unmarshal(data, &result)
        return result, nil
    })
}
