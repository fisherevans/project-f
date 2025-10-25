package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/rs/zerolog/log"
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
	case *events.EventOnInteract:
		return h.onInteract(ctx, world, state, e)
	}
	return nil
}

func (h *SystemEventHandler) onInteract(ctx events.EntityContext, world events.WorldStateReader, _ any, event *events.EventOnInteract) *events.HandlerOutput {
	if event.SourceId != h.State.player {
		log.Warn().Msgf("system event: got interact event for a non-player entity %s", ctx.Id())
	}

	// todo dash
	//switch castEntity := targetEntity.(type) {
	//case *EntityDashGap:
	//	return h.onInteractDashGap(castEntity, event.SourceDirection)
	//}
	return nil
}

//func (h *SystemEventHandler) onInteractDashGap(dashGap *EntityDashGap, interactDirection input.Direction) *events.HandlerOutput {
//	h.State.player.DashTowards(h.State, interactDirection, dashGap.Location())
//	return nil
//}

func (h *SystemEventHandler) onZoneActivity(ctx events.EntityContext, world events.WorldStateReader, _ any, event *events.EventEntityZoneActivity) *events.HandlerOutput {
	if !event.IsEntering || event.EntityId != h.State.player {
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
