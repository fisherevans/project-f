package adventure

import (
	"hash/fnv"
	"math"

	"fisherevans.com/project/f/internal/game/input"
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

func (s *ScriptedMotionBehavior) MovementComplete() {
	s.triggerMovement()
}

func (s *ScriptedMotionBehavior) Update(timeDelta float64) {
	if s.target != nil {
		s.target.Update(timeDelta)
	}
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
	if s.target.IsTarget(s.entity.GetLocation()) {
		s.onComplete(false)
		return
	}
	warnLog := log.Warn().Str("entity", s.entity.GetId()).Str("motion", s.target.MotionId())
	next, isTargetInvalid := s.target.NextLocation(s.entity.GetLocation())
	if isTargetInvalid {
		s.target.NextLocationWasValid(false)
		warnLog.Msgf("next location could not be computed, canceling motion")
		s.onComplete(true)
		return
	}
	movementStarted := s.entity.AttemptMovement(next, MoveStateWalking)
	s.target.NextLocationWasValid(movementStarted)
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
	s.entity.GetSystem().state.eventDispatcher.Dispatch(EventScriptedMotionComplete{
		EntityId:    s.entity.GetId(),
		MotionId:    motionId,
		WasCanceled: wasCanceled,
	})
	s.entity.GetSystem().state.planExecutor.MarkMotionComplete(motionId)
}

type MotionTarget interface {
	MotionId() string
	Update(timeDelta float64)
	NextLocation(currentLocation MapLocation) (MapLocation, bool)
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

func (r *RelativeMotion) Update(timeDelta float64) {}

func (r *RelativeMotion) NextLocation(currentLocation MapLocation) (MapLocation, bool) {
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

type PathfindingMotion struct {
	motionId    string
	entity      Entity
	target      MapLocation
	giveUpAfter float64
	jitterScale float64

	path *Path

	isStable             bool
	secondsSinceLastCalc float64
	timeStuck            float64
}

const stableRecalcInterval = 5.
const invalidRecalcInitialInterval = 0.1
const recalcMaxInterval = 5.

func NewPathfindingMotion(motionId string, entity Entity, to MapLocation) *PathfindingMotion {
	m := &PathfindingMotion{
		motionId:    motionId,
		entity:      entity,
		target:      to,
		isStable:    true,
		giveUpAfter: -1, // all this to be configured
		jitterScale: getRecalcJitterScale(entity.GetId()),
	}
	return m
}

func (p *PathfindingMotion) MotionId() string {
	return p.motionId
}

func (p *PathfindingMotion) Update(timeDelta float64) {
	p.secondsSinceLastCalc += timeDelta
	if p.isStable {
		p.timeStuck = 0
	} else {
		p.timeStuck += timeDelta
	}
}

func (p *PathfindingMotion) NextLocation(currentLocation MapLocation) (MapLocation, bool) {
	if currentLocation == p.target {
		return currentLocation, false
	}
	recalcAfter := stableRecalcInterval * p.jitterScale
	if !p.isStable || p.path == nil || len(p.path.Tiles) == 0 {
		recalcAfter = math.Pow(p.timeStuck, 2) * invalidRecalcInitialInterval * p.jitterScale
	}
	recalcAfter = min(recalcAfter, recalcMaxInterval)

	if p.path == nil || recalcAfter <= p.secondsSinceLastCalc {
		log.Debug().Str("motionId", p.motionId).Str("entity", p.entity.GetId()).Msg("recalculating path")
		p.path = util.Ptr(p.entity.GetSystem().FindPath(currentLocation, p.target, p.entity))
		p.secondsSinceLastCalc = 0
	}

	if p.path.PathFound && len(p.path.Tiles) >= 2 { // 2 for current + next, if next is target
		if p.path.Tiles[0] != currentLocation {
			log.Fatal().Str("motionId", p.motionId).Str("entity", p.entity.GetId()).Msg("motion path is invalid")
		}
		return p.path.Tiles[1], false
	}
	if p.giveUpAfter >= 0 && p.timeStuck > p.giveUpAfter {
		log.Warn().Str("motionId", p.motionId).Str("entity", p.entity.GetId()).Msg("motion stuck, giving up")
		return currentLocation, true
	}
	return currentLocation, false
}

func getRecalcJitterScale(entityId string) float64 {
	h := fnv.New64a()
	h.Write([]byte(entityId))
	h.Write([]byte("recalc-jitter")) // Different from path jitter
	hash := h.Sum64()
	// Return jitter in range [0.8, 1.2] to vary timing by ±20%
	normalized := float64(hash%1000) / 1000.0 // [0, 1)
	return 0.8 + normalized*0.4               // [0.8, 1.2]
}

func (p *PathfindingMotion) NextLocationWasValid(valid bool) {
	p.isStable = valid
	if valid {
		p.path.Tiles = p.path.Tiles[1:]
	}
}

func (p *PathfindingMotion) IsTarget(currentLocation MapLocation) bool {
	return currentLocation == p.target
}
