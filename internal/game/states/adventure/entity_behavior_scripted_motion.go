package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/rs/zerolog/log"
)

type ScriptedMotionBehavior struct {
	*baseEntityBehavior
	target MotionTarget
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
	if s.target == nil {
		return
	}
	s.onComplete(true)
}

func (s *ScriptedMotionBehavior) SetTarget(target MotionTarget) {
	s.onComplete(true)
	s.target = target
}

func (s *ScriptedMotionBehavior) triggerMovement(position *EntityPosition) {
	if position.IsMoving() || s.target == nil {
		return
	}
	if position.GetPrimaryLocation() == s.target.TargetLocation(position.GetPrimaryLocation()) {
		s.onComplete(false)
		return
	}
	warnLog := log.Warn().Str("entity", s.id).Str("motion", s.target.MotionId())
	next, isTargetInvalid := s.target.NextLocation(position.GetPrimaryLocation())
	if isTargetInvalid {
		warnLog.Msgf("next location could not be computed, canceling motion")
		s.onComplete(true)
		return
	}
	movementStarted := s.system.AttemptMovement(s.id, next, types.MoveStateWalking)
	if !movementStarted {
		warnLog.Msgf("next location was invalid, canceling motion")
		log.Warn().Str("id", s.id).Msg("next move was invalid")
		s.onComplete(true)
	}
}

func (s *ScriptedMotionBehavior) onComplete(wasCanceled bool) {
	if s.target == nil {
		return
	}
	motionId := s.target.MotionId()
	s.target = nil
	if motionId == "" {
		return
	}
	s.system.state.eventDispatcher.Dispatch(events.EventScriptedMotionComplete{
		EntityId:    s.id,
		MotionId:    motionId,
		WasCanceled: wasCanceled,
	})
	s.system.state.planExecutor.MarkMotionComplete(motionId)
}

type MotionTarget interface {
	MotionId() string
	NextLocation(currentLocation MapLocation) (MapLocation, bool)
	TargetLocation(currentLocation MapLocation) MapLocation
}

type RelativeMotion struct {
	motionId  string
	direction input.Direction
	tiles     int
}

func NewRelativeMotion(motionId string, direction input.Direction, tiles int) *RelativeMotion {
	return &RelativeMotion{
		motionId:  motionId,
		direction: direction,
		tiles:     tiles,
	}
}

func (r *RelativeMotion) MotionId() string {
	return r.motionId
}

func (r *RelativeMotion) NextLocation(currentLocation MapLocation) (MapLocation, bool) {
	if r.tiles <= 0 {
		return currentLocation, false
	}
	r.tiles--
	return currentLocation.Moved(r.direction), false
}

func (r *RelativeMotion) TargetLocation(currentLocation MapLocation) MapLocation {
	dx, dy := r.direction.GetVector()
	return MapLocation{
		X: currentLocation.X + dx*r.tiles,
		Y: currentLocation.Y + dy*r.tiles,
	}
}

// todo, consider caching found paths
// - maybe recomputing every X steps, retrying if movement fails
// - currently doing it every step as it takes less than 1ms
type PathfindingMotion struct {
	motionId string
	es       *EntitySystem
	entityId string
	target   MapLocation
}

func NewPathfindingMotion(motionId string, es *EntitySystem, entityId string, to MapLocation) *PathfindingMotion {
	m := &PathfindingMotion{
		motionId: motionId,
		es:       es,
		entityId: entityId,
		target:   to,
	}
	return m
}

func (p *PathfindingMotion) MotionId() string {
	return p.motionId
}

func (p *PathfindingMotion) NextLocation(currentLocation MapLocation) (MapLocation, bool) {
	if currentLocation == p.target {
		return currentLocation, false
	}
	path := p.es.FindPath(currentLocation, p.target, p.entityId)
	if !path.PathFound || len(path.Tiles) == 0 {
		return currentLocation, true
	}
	return path.Tiles[0], false
}

func (p *PathfindingMotion) TargetLocation(currentLocation MapLocation) MapLocation {
	return p.target
}
