package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/rs/zerolog/log"
)

type SystemEventHandler struct {
	State *State
}

func newSystemEventHandler(s *State) EventHandler {
	return &SystemEventHandler{
		State: s,
	}
}

func (h *SystemEventHandler) Init(ctx EntityContext, gameState GameState, state any) *HandlerOutput {
	return nil
}

func (h *SystemEventHandler) HandleEvent(ctx EntityContext, gameState GameState, state any, event any) *HandlerOutput {
	if event == nil {
		return nil
	}
	switch e := event.(type) {
	case *EventEntityZoneActivity:
		return h.onZoneActivity(ctx, gameState, state, e)
	case *EventOnInteract:
		return h.onInteract(ctx, gameState, state, e)
	}
	return nil
}

func (h *SystemEventHandler) onInteract(ctx EntityContext, gameState GameState, _ any, event *EventOnInteract) *HandlerOutput {
	if event.SourceId != h.State.player {
		log.Warn().Msgf("system event: got interact event for a non-player entity %s", ctx.EntityId())
	}
	targetEntity, ok := h.State.entities.GetEntity(event.TargetId)
	if !ok {
		log.Warn().Msgf("system event: got interact event for a non-existent entity %s", event.TargetId)
		return nil
	}
	state, ok := targetEntity.GetState()
	if !ok {
		return nil
	}
	switch state := state.(type) {
	case DashGapState:
		return h.onInteractDashGap(targetEntity, state, event)
	}
	return nil
}

func (h *SystemEventHandler) onInteractDashGap(dashGapEntity Entity, state DashGapState, event *EventOnInteract) *HandlerOutput {
	location := findDashDestination(h.State, event.SourceFacingDirection, dashGapEntity.GetLocation())
	return NewOutput().WithEffects(
		NewTriggerMovementEffect(event.SourceId).WithLocation(location).WithMoveState(types.MoveStateDashing),
	)
}

func findDashDestination(s *State, direction input.Direction, location MapLocation) MapLocation {
	for {
		moved := false
		for _, entityId := range s.entities.occupations.OccupyingEntityList(location) {
			entity, ok := s.entities.GetEntity(entityId)
			if !ok {
				continue
			}
			state, ok := entity.GetState()
			if !ok {
				continue
			}
			_, ok = state.(DashGapState) // todo consider direction config in state once supported
			if !ok {
				continue
			}
			moved = true
			location = location.MovedDelta(direction.GetVector())
		}
		if !moved {
			return location
		}
	}
}

func (h *SystemEventHandler) onZoneActivity(ctx EntityContext, gameState GameState, _ any, event *EventEntityZoneActivity) *HandlerOutput {
	if !event.IsEntering || event.EntityId != h.State.player || event.WasTeleported {
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
	return NewOutput().WithEffects(NewTeleportPlayerEffect().WithToReference(ref))
}
