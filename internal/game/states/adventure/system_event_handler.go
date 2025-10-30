package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
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

func (h *SystemEventHandler) onInteractDashGap(dashGapEntity Entity, state DashGapState, event *events.EventOnInteract) *events.HandlerOutput {
	location := findDashDestination(h.State, event.SourceFacingDirection, dashGapEntity.GetLocation())
	to := events.Location{X: location.X, Y: location.Y}
	return events.NewOutput().WithEffects(
		events.NewTriggerMovementEffect(event.SourceId).WithLocation(to).WithMoveState(types.MoveStateDashing),
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

func (h *SystemEventHandler) onZoneActivity(ctx events.EntityContext, world events.WorldStateReader, _ any, event *events.EventEntityZoneActivity) *events.HandlerOutput {
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
	return events.NewOutput().WithEffects(events.NewTeleportPlayerEffect().WithToReference(ref))
}
