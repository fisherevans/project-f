package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type Entity interface {
	GetId() string
	GetSystem() *EntitySystem

	// from guaranteed position

	GetLocation() MapLocation
	GetPreciseLocation() pixel.Vec
	IsMoving() bool
	CancelMovement()
	GetMovementState() types.MoveState
	AlterMovementState(state types.MoveState)
	GetFacingDirection() input.Direction
	SetFacingDirection(input.Direction)
	GetMovementSpeed(state types.MoveState) float64
	SetMovementSpeed(state types.MoveState, speed float64)
	Teleport(toLocation MapLocation)
	GetEntityContext() events.EntityContext
	SetEntityContextMetadata(key string, accessor func() any)

	// optional traits

	PushBehavior(behavior EntityBehavior)
	PopBehavior() (EntityBehavior, bool)
	GetBehavior() (EntityBehavior, bool)
	DisableBehavior(source string)
	EnableBehavior(source string)
	IsBehaviorEnabled() bool

	SetPresence(presence EntityPresence)
	GetPresence() (EntityPresence, bool)

	SetRenderer(renderer EntityRenderer)
	GetRenderer() (EntityRenderer, bool)

	SetState(state EntityState)
	GetState() (EntityState, bool)
	InteractsWith(location MapLocation)
}

type entityReference struct {
	system *EntitySystem
	id     string
}

func (e *entityReference) IsBehaviorEnabled() bool {
	return len(e.system.disabledBehaviors[e.id]) == 0
}

func (e *entityReference) DisableBehavior(source string) {
	if _, ok := e.system.disabledBehaviors[e.id][source]; !ok {
		e.system.disabledBehaviors[e.id] = map[string]struct{}{}
	}
	e.system.disabledBehaviors[e.id][source] = struct{}{}
}

func (e *entityReference) EnableBehavior(source string) {
	if _, ok := e.system.disabledBehaviors[e.id][source]; !ok {
		return
	}
	delete(e.system.disabledBehaviors[e.id], source)
}

func (e *entityReference) PushBehavior(behavior EntityBehavior) {
	e.system.behaviors[e.id] = append(e.system.behaviors[e.id], behavior)
}

func (e *entityReference) PopBehavior() (EntityBehavior, bool) {
	if len(e.system.behaviors[e.id]) == 0 {
		return nil, false
	}
	lastIndex := len(e.system.behaviors[e.id]) - 1
	popped := e.system.behaviors[e.id][lastIndex]
	e.system.behaviors[e.id] = e.system.behaviors[e.id][:lastIndex]
	return popped, true
}

func (e *entityReference) AlterMovementState(state types.MoveState) {
	if state == types.MoveStateIdle {
		e.CancelMovement()
		return
	}
	e.system.movements[e.id].MovementState = state
}

func (e *entityReference) SetMovementSpeed(state types.MoveState, speed float64) {
	if e.system.movements[e.id].MovementSpeeds == nil {
		e.system.movements[e.id].MovementSpeeds = map[types.MoveState]float64{}
	}
	e.system.movements[e.id].MovementSpeeds[state] = speed
}

func (e *entityReference) SetEntityContextMetadata(key string, accessor func() any) {
	e.system.contexts[e.id].WithMetadata(key, accessor)
}
func (e *entityReference) InteractsWith(location MapLocation) {
	direction := e.GetLocation().DirectionTowards(location)
	for targetEntityId := range e.system.interactableEntityIds(location) {
		e.system.state.eventDispatcher.Dispatch(events.EventOnInteract{
			SourceId:              e.id,
			SourceFacingDirection: direction,
			TargetId:              targetEntityId,
		})
	}
}

func (e *entityReference) GetEntityContext() events.EntityContext {
	return e.system.contexts[e.id]
}

func (e *entityReference) Teleport(toLocation MapLocation) {
	warnLog := log.Log().Str("id", e.id).Any("location", toLocation)
	if e.GetLocation() == toLocation {
		warnLog.Msgf("attempted teleportation to same location")
		return
	}
	if e.IsMoving() {
		e.CancelMovement()
	}
	e.system.occupations.Vacate(e.id, e.GetLocation())
	e.system.occupations.SetPosition(e.id, toLocation, true)
}

func (e *entityReference) movement() *EntityMovement {
	return e.system.movements[e.id]
}

func (e *entityReference) GetId() string {
	return e.id
}

func (e *entityReference) GetSystem() *EntitySystem {
	return e.system
}

func (e *entityReference) GetLocation() MapLocation {
	return e.system.occupations.GetPosition(e.id) // todo rename to location
}

func (e *entityReference) GetPreciseLocation() pixel.Vec {
	location := e.GetLocation().ToVec()
	if !e.IsMoving() {
		return location
	}
	m := e.movement()
	p := m.Progression
	if m.MovementState == types.MoveStateDashing {
		p = interp.Smootherstep(p)
	} else if m.AccumulatedMovement < 1 {
		p = interp.EaseInToLinear(p, 2)
	}
	movementDelta := m.TargetLocation.ToVec().Sub(location).Scaled(p)
	return location.Add(movementDelta)
}

func (e *entityReference) GetMovementState() types.MoveState {
	return e.system.movements[e.id].MovementState
}

func (e *entityReference) GetMovementSpeed(moveState types.MoveState) float64 {
	return e.system.movements[e.id].getSpeed(moveState)
}

func (e *entityReference) IsMoving() bool {
	return e.GetMovementState() != types.MoveStateIdle
}

func (e *entityReference) CancelMovement() {
	if !e.IsMoving() {
		return
	}
	m := e.movement()
	e.system.occupations.Vacate(e.id, m.TargetLocation)
	m.TargetLocation = e.GetLocation()
	m.MovementState = types.MoveStateIdle
	m.Progression = 0
	m.AccumulatedMovement = 0
	m.ProgressionScale = 1
}

func (e *entityReference) GetFacingDirection() input.Direction {
	return e.movement().FacingDirection
}

func (e *entityReference) SetFacingDirection(dir input.Direction) {
	e.movement().FacingDirection = dir
	behavior, ok := e.GetBehavior()
	if !ok {
		return
	}
	player, ok := behavior.(*PlayerBehavior)
	if !ok {
		return
	}
	player.intentDirection = dir
}

func (e *entityReference) GetBehavior() (EntityBehavior, bool) {
	if len(e.system.behaviors[e.id]) == 0 {
		return nil, false
	}
	return e.system.behaviors[e.id][len(e.system.behaviors[e.id])-1], true
}

func (e *entityReference) SetPresence(presence EntityPresence) {
	e.system.presences[e.id] = presence
}

func (e *entityReference) GetPresence() (EntityPresence, bool) {
	v, ok := e.system.presences[e.id]
	return v, ok
}

func (e *entityReference) SetRenderer(renderer EntityRenderer) {
	e.system.renderers[e.id] = renderer
}

func (e *entityReference) GetRenderer() (EntityRenderer, bool) {
	v, ok := e.system.renderers[e.id]
	return v, ok
}

func (e *entityReference) SetState(state EntityState) {
	e.system.states[e.id] = state
}

func (e *entityReference) GetState() (EntityState, bool) {
	v, ok := e.system.states[e.id]
	return v, ok
}
