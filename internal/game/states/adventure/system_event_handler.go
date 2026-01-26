package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
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

func (h *SystemEventHandler) Init(thisEntity EntityReader, globals StateGlobalsReader, state any) *HandlerOutput {
	return NewOutput().WithEffects(
		NewFadeEffect(1, 1).
			WithFromColor("#000f").
			WithToColor("#0000"))
}

func (h *SystemEventHandler) HandleEvent(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput {
	if event == nil {
		return nil
	}
	switch e := event.(type) {
	case *EventEntityZoneActivity:
		return h.onZoneActivity(thisEntity, globals, state, e)
	case *EventOnInteract:
		return h.onInteract(thisEntity, globals, state, e)
	}
	return nil
}

func (h *SystemEventHandler) onInteract(thisEntity EntityReader, globals StateGlobalsReader, _ any, event *EventOnInteract) *HandlerOutput {
	if event.SourceId != h.State.player {
		log.Warn().Msgf("system event: got interact event for a non-player entity %s", thisEntity.GetId())
	}
	targetEntity, ok := h.State.entities.GetEntity(event.TargetId)
	if !ok {
		log.Warn().Msgf("system event: got interact event for a non-existent entity %s", event.TargetId)
		return nil
	}
	if DashGapMetadataKey.Exists(targetEntity) {
		return h.onInteractDashGap(targetEntity, event)
	}
	return nil
}

func (h *SystemEventHandler) onInteractDashGap(dashGapEntity Entity, event *EventOnInteract) *HandlerOutput {
	location := findDashDestination(h.State, event.SourceFacingDirection, dashGapEntity.GetLocation())
	return NewOutput().WithEffects(
		NewTriggerMovementEffect(event.SourceId).WithLocation(location).WithMoveState(MoveStateDashing),
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
			if !DashGapMetadataKey.Exists(entity) {
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

func (h *SystemEventHandler) onZoneActivity(thisEntity EntityReader, globals StateGlobalsReader, _ any, event *EventEntityZoneActivity) *HandlerOutput {
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
