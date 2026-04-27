package debugapi

import (
    "fmt"
    "net/http"
    "reflect"

    "fisherevans.com/project/f/internal/game"
    "fisherevans.com/project/f/internal/game/states/adventure"
)

func (s *Server) requireAdventure(w http.ResponseWriter, r *http.Request, fn func(*adventure.State) (any, error)) {
    s.queue.handleOnGameThread(w, r, func() (any, error) {
        active := game.GetActiveState()
        advState, ok := active.(*adventure.State)
        if !ok {
            return nil, fmt.Errorf("requires adventure state (active: %s)", reflect.TypeOf(active).String())
        }
        return fn(advState)
    })
}

func (s *Server) handleListEntities(w http.ResponseWriter, r *http.Request) {
    s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
        ids := advState.EntityIds()
        globals := advState.Globals()
        playerId := advState.PlayerId()
        snapshots := make([]EntitySnapshot, 0, len(ids))
        for _, id := range ids {
            reader, ok := globals.GetEntityReader(id)
            if !ok {
                continue
            }
            loc := reader.GetLocation()
            snapshots = append(snapshots, EntitySnapshot{
                Id:         id,
                X:          loc.X,
                Y:          loc.Y,
                IsMoving:   reader.IsMoving(),
                Direction:  reader.GetFacingDirection().String(),
                IsPlayer:   id == playerId,
                DebugType:  advState.EntityDebugType(id),
                HandlerRef: advState.EntityHandlerRef(id),
            })
        }
        mapW, mapH := advState.MapSize()
        return EntityListResponse{
            Entities:  snapshots,
            MapWidth:  mapW,
            MapHeight: mapH,
        }, nil
    })
}

func (s *Server) handleGetEntity(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
        reader, ok := advState.Globals().GetEntityReader(id)
        if !ok {
            return map[string]string{"error": "entity not found"}, nil
        }
        loc := reader.GetLocation()
        precise := reader.GetPreciseLocation()
        _, hasBehavior := reader.GetBehavior()
        _, hasPresence := reader.GetPresence()
        _, hasRenderer := reader.GetRenderer()
        detail := EntityDetail{
            EntitySnapshot: EntitySnapshot{
                Id:        id,
                X:         loc.X,
                Y:         loc.Y,
                IsMoving:  reader.IsMoving(),
                Direction: reader.GetFacingDirection().String(),
            },
            PreciseX:        precise.X,
            PreciseY:        precise.Y,
            BehaviorEnabled: reader.IsBehaviorEnabled(),
            HasBehavior:     hasBehavior,
            HasPresence:     hasPresence,
            HasRenderer:     hasRenderer,
            SoundEnabled:    reader.IsSoundEnabled(),
        }
        md := reader.GetMetadata()
        if md != nil {
            detail.Metadata = md
        }
        return detail, nil
    })
}

func (s *Server) handleListZones(w http.ResponseWriter, r *http.Request) {
    s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
        rects := advState.ZoneRects()
        zones := make([]ZoneRect, 0, len(rects))
        for _, z := range rects {
            zones = append(zones, ZoneRect{
                Id: z.ZoneId,
                X:  z.X,
                Y:  z.Y,
                W:  z.W,
                H:  z.H,
            })
        }
        return zones, nil
    })
}

func (s *Server) handleListTeleports(w http.ResponseWriter, r *http.Request) {
    s.requireAdventure(w, r, func(advState *adventure.State) (any, error) {
        teleports := advState.TeleportEntries()
        entries := make([]TeleportEntry, 0, len(teleports))
        for ref, tp := range teleports {
            entries = append(entries, TeleportEntry{
                Reference:      string(ref),
                DestinationRef: string(tp.Destination),
                X:              tp.Location.X,
                Y:              tp.Location.Y,
                ExitDirection:  tp.ExitDirection.String(),
            })
        }
        return entries, nil
    })
}
