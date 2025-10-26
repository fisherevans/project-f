package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/rs/zerolog/log"
)

type ScriptedMotionBehavior struct {
	*baseEntityBehavior
	currentMotionId string
	targetLocation  MapLocation
}

func NewScriptedMotionBehavior(id string, system *EntitySystem) *ScriptedMotionBehavior {
	return &ScriptedMotionBehavior{
		baseEntityBehavior: newBaseEntityBehavior(id, system),
	}
}

func (s *ScriptedMotionBehavior) MovementComplete(dispatcher Dispatcher) {
	s.triggerMovement(s.system.positions[s.id])
}

func (s *ScriptedMotionBehavior) Update(timeDelta float64, position *EntityPosition, dispatcher Dispatcher) {
	s.triggerMovement(position)
}

func (s *ScriptedMotionBehavior) Reset() {
	if s.currentMotionId != "" {
		s.onComplete(true)
	}
}

func (s *ScriptedMotionBehavior) SetTarget(motionId string, location MapLocation) {
	if s.currentMotionId != "" {
		s.onComplete(true)
	}
	s.currentMotionId = motionId
	s.targetLocation = location
}

func (s *ScriptedMotionBehavior) triggerMovement(position *EntityPosition) {
	if position.IsMoving() || s.currentMotionId == "" {
		return
	}
	if position.Location == s.targetLocation {
		if s.currentMotionId != "" {
			s.onComplete(false)
		}
		return
	}
	dx := s.targetLocation.X - position.Location.X
	dy := s.targetLocation.Y - position.Location.Y

	var moveDirection, backupDirection input.Direction
	if dy != 0 {
		if dy > 0 {
			moveDirection = input.Up
		} else {
			moveDirection = input.Down
		}
		// Backup is horizontal
		if dx > 0 {
			backupDirection = input.Right
		} else if dx < 0 {
			backupDirection = input.Left
		}
	} else {
		if dx > 0 {
			moveDirection = input.Right
		} else {
			moveDirection = input.Left
		}
	}
	if s.system.AttemptMovement(s.id, position.Location.Moved(moveDirection), types.MoveStateWalking) {
		return
	}
	if backupDirection != input.NotPressed && s.system.AttemptMovement(s.id, position.Location.Moved(backupDirection), types.MoveStateWalking) {
		return
	}
	log.Warn().Str("id", s.id).
		Any("target", s.targetLocation).
		Str("motionId", s.currentMotionId).
		Msg("failed to trigger movement to target")
	s.onComplete(true)
}

func (s *ScriptedMotionBehavior) onComplete(wasCanceled bool) {
	s.system.state.eventDispatcher.Dispatch(events.EventScriptedMotionComplete{
		EntityId:    s.id,
		MotionId:    s.currentMotionId,
		WasCanceled: wasCanceled,
	})
	s.system.state.planExecutor.MarkMotionComplete(s.currentMotionId)
	s.currentMotionId = ""
}
