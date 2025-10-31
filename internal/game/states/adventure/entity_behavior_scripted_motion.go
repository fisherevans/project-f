package adventure

import (
	"math"

	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
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
	s.triggerMovement(0)
}

func (s *ScriptedMotionBehavior) Update(timeDelta float64, dispatcher Dispatcher) {
	s.triggerMovement(timeDelta)
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

func (s *ScriptedMotionBehavior) triggerMovement(timeDelta float64) {
	if s.entity.IsMoving() || s.target == nil {
		return
	}
	if s.target.IsTarget(s.entity.GetLocation()) {
		s.onComplete(false)
		return
	}
	warnLog := log.Warn().Str("entity", s.entity.GetId()).Str("motion", s.target.MotionId())
	next, isTargetInvalid := s.target.NextLocation(s.entity.GetLocation(), timeDelta)
	if isTargetInvalid {
		s.target.NextLocationWasValid(false)
		warnLog.Msgf("next location could not be computed, canceling motion")
		s.onComplete(true)
		return
	}
	movementStarted := s.entity.GetSystem().AttemptMovement(s.entity.GetId(), next, types.MoveStateWalking)
	s.target.NextLocationWasValid(movementStarted)
	if !movementStarted {
		warnLog.Msgf("next location was invalid")
		//s.onComplete(true) // todo configure this behavior
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
	NextLocation(currentLocation MapLocation, timeDelta float64) (MapLocation, bool)
	NextLocationWasValid(bool)
	IsTarget(currentLocation MapLocation) bool
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

func (r *RelativeMotion) NextLocation(currentLocation MapLocation, timeDelta float64) (MapLocation, bool) {
	if r.tiles <= 0 {
		return currentLocation, false
	}
	r.tiles--
	return currentLocation.Moved(r.direction), false
}

func (r *RelativeMotion) NextLocationWasValid(valid bool) {
}

func (r *RelativeMotion) IsTarget(currentLocation MapLocation) bool {
	dx, dy := r.direction.GetVector()
	return currentLocation == MapLocation{
		X: currentLocation.X + dx*r.tiles,
		Y: currentLocation.Y + dy*r.tiles,
	}
}

// todo, consider caching found paths
// - maybe recomputing every X steps, retrying if movement fails
// - currently doing it every step as it takes less than 1ms
type PathfindingMotion struct {
	motionId             string
	entity               Entity
	target               MapLocation
	giveUpAfter          float64
	timeStuck            float64
	path                 *Path
	isStable             bool
	secondsSinceLastCalc float64
	timeUnstable         int
}

// todo jitter
const stableRecalcInterval = 1.
const invalidRecalcInitialInterval = 0.1
const recalcMaxInterval = 10.

func NewPathfindingMotion(motionId string, entity Entity, to MapLocation) *PathfindingMotion {
	m := &PathfindingMotion{
		motionId:    motionId,
		entity:      entity,
		target:      to,
		isStable:    true,
		giveUpAfter: -1, // all this to be configured
	}
	return m
}

func (p *PathfindingMotion) MotionId() string {
	return p.motionId
}

func (p *PathfindingMotion) NextLocation(currentLocation MapLocation, timeDelta float64) (MapLocation, bool) {
	p.secondsSinceLastCalc += timeDelta
	if currentLocation == p.target {
		return currentLocation, false
	}
	recalcAfter := stableRecalcInterval
	if !p.isStable || p.path == nil || len(p.path.Tiles) == 0 || p.path.Tiles[0] != currentLocation {
		recalcAfter = invalidRecalcInitialInterval + math.Pow(float64(p.timeUnstable), 2)*invalidRecalcInitialInterval
	}
	recalcAfter = min(recalcAfter, recalcMaxInterval)
	if p.path == nil || recalcAfter <= p.secondsSinceLastCalc {
		p.path = util.Ptr(p.entity.GetSystem().FindPath(currentLocation, p.target, p.entity))
	}

	if p.path.PathFound && len(p.path.Tiles) >= 2 { // 2 for current + next, if next is target
		p.timeStuck = 0
		if p.path.Tiles[0] != currentLocation {
			log.Fatal().Str("motionId", p.motionId).Str("entity", p.entity.GetId()).Msg("motion path is invalid")
		}
		return p.path.Tiles[1], false
	}
	p.timeStuck += timeDelta
	if p.giveUpAfter >= 0 && p.timeStuck > p.giveUpAfter {
		log.Warn().Str("motionId", p.motionId).Str("entity", p.entity.GetId()).Msg("motion stuck, giving up")
		return currentLocation, true
	}
	return currentLocation, false
}

func (p *PathfindingMotion) NextLocationWasValid(valid bool) {
	p.isStable = valid
	if valid {
		p.path.Tiles = p.path.Tiles[1:]
		p.timeUnstable = 0
	} else {
		p.timeUnstable++
	}
}

func (p *PathfindingMotion) IsTarget(currentLocation MapLocation) bool {
	return currentLocation == p.target
}
