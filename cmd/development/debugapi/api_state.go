package debugapi

import (
    "encoding/json"
    "net/http"
    "reflect"

    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/states/adventure"
)

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        active := game.GetActiveState()
        typeName := reflect.TypeOf(active).String()

        info := StateInfo{
            Type: typeName,
        }

        if advState, ok := active.(*adventure.State); ok {
            info.Capabilities = []string{"entities", "teleport", "map", "command"}
            info.MapName = advState.MapName()
            if p := advState.Globals().Player(); p != nil {
                loc := p.GetPreciseLocation()
                info.PlayerPos = &Vec2{X: loc.X, Y: loc.Y}
            }
        } else {
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
