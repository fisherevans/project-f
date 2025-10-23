package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
)

type SystemEventHandler struct {
	State *State
}

func newSystemEventHandler(s *State) events.EventHandler {
	return &SystemEventHandler{
		State: s,
	}
}

func (h *SystemEventHandler) Init(ctx events.EntityContext, world events.WorldStateReader, state any) *events.HandlerOutput {
	return nil
}

func (h *SystemEventHandler) HandleEvent(ctx events.EntityContext, world events.WorldStateReader, state any, event any) *events.HandlerOutput {
	if event == nil {
		return nil
	}
	switch e := event.(type) {
	case *events.EventEntityZoneActivity:
		return h.onZoneActivity(ctx, world, state, e)
	}
	return nil
}

func (h *SystemEventHandler) onZoneActivity(ctx events.EntityContext, world events.WorldStateReader, _ any, event *events.EventEntityZoneActivity) *events.HandlerOutput {
	if !event.IsEntering || event.EntityId != string(h.State.player.GetEntityId()) {
		return nil
	}
	tele, ok := h.State.teleports[TeleportReference(event.ZoneId)]
	if !ok {
		return nil
	}
	ref := string(tele.Destination)
	if ref == "teleport:" { // todo this is gross
		return nil
	}
	return &events.HandlerOutput{
		Effects: []events.Effect{
			{
				TeleportPlayer: &events.EffectTeleportPlayer{
					ToReference: &ref,
				},
			},
		},
	}
}
