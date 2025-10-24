package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
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
	targetEntity, ok := h.State.entities[EntityId(event.TargetId)]
	if !ok {
		log.Warn().Msgf("system event: no target entity with id %s", event.TargetId)
		return nil
	}
	if event.SourceId != string(h.State.player.GetEntityId()) {
		log.Warn().Msgf("system event: got interact event for a non-player entity %s", ctx.Id())
		return nil
	}
	switch castEntity := targetEntity.(type) {
	case *EntityDashGap:
		return h.onInteractDashGap(castEntity, event.SourceDirection)
	}
	return nil
}

func (h *SystemEventHandler) onInteractDashGap(dashGap *EntityDashGap, interactDirection input.Direction) *events.HandlerOutput {
	h.State.player.DashTowards(h.State, interactDirection, dashGap.Location())
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
