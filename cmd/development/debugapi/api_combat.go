package debugapi

import (
    "fmt"
    "net/http"
    "reflect"

    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/states/combat"
)

func (s *Server) handleGetCombat(w http.ResponseWriter, r *http.Request) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        active := game.GetActiveState()
        cs, ok := active.(*combat.State)
        if !ok {
            return nil, fmt.Errorf("requires combat state (active: %s)", reflect.TypeOf(active).String())
        }
        return cs.DebugInfo(), nil
    })
}
