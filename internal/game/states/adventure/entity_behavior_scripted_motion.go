package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/rs/zerolog/log"
)

type ScriptedMotionBehavior struct {
	entity Entity
	target MotionTarget
}

func AttachScriptedMotionBehavior(entity Entity) *ScriptedMotionBehavior {
	behavior := &ScriptedMotionBehavior{
		entity: entity,
	}
	entity.PushBehavior(behavior)
	return behavior
}

func (s *ScriptedMotionBehavior) MovementComplete(dispatcher Dispatcher) {
	s.triggerMovement()
}

func (s *ScriptedMotionBehavior) Update(timeDelta float64, dispatcher Dispatcher) {
	s.triggerMovement()
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

func (s *ScriptedMotionBehavior) triggerMovement() {
	if s.entity.IsMoving() || s.target == nil {
		return
	}
	if s.entity.GetLocation() == s.target.TargetLocation(s.entity.GetLocation()) {
		s.onComplete(false)
		return
	}
	warnLog := log.Warn().Str("entity", s.entity.GetId()).Str("motion", s.target.MotionId())
	next, isTargetInvalid := s.target.NextLocation(s.entity.GetLocation())
	if isTargetInvalid {
		warnLog.Msgf("next location could not be computed, canceling motion")
		s.onComplete(true)
		return
	}
	movementStarted := s.entity.GetSystem().AttemptMovement(s.entity.GetId(), next, types.MoveStateWalking)
	if !movementStarted {
		warnLog.Msgf("next location was invalid, canceling motion")
		log.Warn().Str("id", s.entity.GetId()).Msg("next move was invalid")
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
	s.entity.GetSystem().state.eventDispatcher.Dispatch(events.EventScriptedMotionComplete{
		EntityId:    s.entity.GetId(),
		MotionId:    motionId,
		WasCanceled: wasCanceled,
	})
	s.entity.GetSystem().state.planExecutor.MarkMotionComplete(motionId)
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
	entity   Entity
	target   MapLocation
}

func NewPathfindingMotion(motionId string, entity Entity, to MapLocation) *PathfindingMotion {
	m := &PathfindingMotion{
		motionId: motionId,
		entity:   entity,
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
	path := p.entity.GetSystem().FindPath(currentLocation, p.target, p.entity)
	if !path.PathFound || len(path.Tiles) == 0 {
		return currentLocation, true
	}
	return path.Tiles[0], false
}

func (p *PathfindingMotion) TargetLocation(currentLocation MapLocation) MapLocation {
	return p.target
}
