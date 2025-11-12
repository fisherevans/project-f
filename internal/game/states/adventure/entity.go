package adventure

import (
	"slices"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntityReader interface {
	GetId() string

	GetLocation() MapLocation
	GetPreciseLocation() pixel.Vec

	IsMoving() bool
	GetMovementState() MoveState
	GetMovementSpeed(state MoveState) float64
	GetFacingDirection() input.Direction

	GetMetadata() *EntityMetadata

	GetBehavior() (EntityBehavior, bool)
	IsBehaviorEnabled() bool

	IsSoundEnabled() bool

	GetPresence() (EntityPresence, bool)

	GetRenderer() (EntityRenderer, bool)
}

type Entity interface {
	EntityReader
	GetSystem() *EntitySystem

	Teleport(toLocation MapLocation)

	CancelMovement()
	AttemptMovement(targetLocation MapLocation, movementState MoveState) bool
	AlterMovementState(state MoveState)
	SetMovementSpeed(state MoveState, speed float64)
	SetFacingDirection(input.Direction)

	PushBehavior(behavior EntityBehavior)
	PopBehavior() (EntityBehavior, bool)
	DisableBehavior(source string)
	EnableBehavior(source string)

	AddSoundProvider(soundProvider EntitySoundProvider)
	DisableSound(source string)
	EnableSound(source string)

	SetPresence(presence EntityPresence)

	SetRenderer(renderer EntityRenderer)

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
	if _, ok := e.system.disabledBehaviors[e.id]; !ok {
		e.system.disabledBehaviors[e.id] = map[string]struct{}{}
	}
	e.system.disabledBehaviors[e.id][source] = struct{}{}
}

func (e *entityReference) EnableBehavior(source string) {
	if _, ok := e.system.disabledBehaviors[e.id]; !ok {
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

func (e *entityReference) AlterMovementState(state MoveState) {
	if state == MoveStateIdle {
		e.CancelMovement()
		return
	}
	e.system.movements[e.id].MovementState = state
}

func (e *entityReference) SetMovementSpeed(state MoveState, speed float64) {
	if e.system.movements[e.id].MovementSpeeds == nil {
		e.system.movements[e.id].MovementSpeeds = map[MoveState]float64{}
	}
	e.system.movements[e.id].MovementSpeeds[state] = speed
}

func (e *entityReference) InteractsWith(location MapLocation) {
	direction := e.GetLocation().DirectionTowards(location)
	for targetEntityId := range e.system.interactableEntityIds(location) {
		e.system.state.eventDispatcher.Dispatch(EventOnInteract{
			SourceId:              e.id,
			SourceFacingDirection: direction,
			TargetId:              targetEntityId,
		})
	}
}

func (e *entityReference) AttemptMovement(targetLocation MapLocation, movementState MoveState) bool {
	debugLog := log.Debug().Str("id", e.id).Any("state", movementState).Any("target", targetLocation)
	if !slices.Contains(validAttemptMovementStates, movementState) {
		debugLog.Msgf("movement state not valid")
		return false
	}
	isValid, movementDirection := e.system.isMovementValid(e, targetLocation)
	if !isValid {
		debugLog.Msgf("movement not valid")
		return false
	}
	e.system.occupations.Occupy(e.GetId(), targetLocation) // emit enter events within movement update - don't "enter" until primary is updated
	// todo messy
	movement := e.system.movements[e.GetId()]
	movement.TargetLocation = targetLocation
	movement.MovementState = movementState
	movement.FacingDirection = movementDirection
	distance := e.GetLocation().DistanceTo(targetLocation)
	movement.ProgressionScale = 1.0 / distance
	return true
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
	if m.MovementState == MoveStateDashing {
		p = interp.Smootherstep(p)
	} else if m.AccumulatedMovement < 1 {
		p = interp.EaseInToLinear(p, 2)
	}
	movementDelta := m.TargetLocation.ToVec().Sub(location).Scaled(p)
	return location.Add(movementDelta)
}

func (e *entityReference) GetMovementState() MoveState {
	return e.system.movements[e.id].MovementState
}

func (e *entityReference) GetMovementSpeed(moveState MoveState) float64 {
	return e.system.movements[e.id].getSpeed(moveState)
}

func (e *entityReference) IsMoving() bool {
	return e.GetMovementState() != MoveStateIdle
}

func (e *entityReference) CancelMovement() {
	if !e.IsMoving() {
		return
	}
	m := e.movement()
	e.system.occupations.Vacate(e.id, m.TargetLocation)
	m.TargetLocation = e.GetLocation()
	m.MovementState = MoveStateIdle
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
func (e *entityReference) GetMetadata() *EntityMetadata {
	if _, ok := e.system.metadata[e.id]; !ok {
		e.system.metadata[e.id] = &EntityMetadata{}
	}
	return e.system.metadata[e.id]
}

func (e *entityReference) GetBehavior() (EntityBehavior, bool) {
	if len(e.system.behaviors[e.id]) == 0 {
		return nil, false
	}
	return e.system.behaviors[e.id][len(e.system.behaviors[e.id])-1], true
}

func (e *entityReference) AddSoundProvider(soundProvider EntitySoundProvider) {
	e.system.soundProviders[e.id] = append(e.system.soundProviders[e.id], soundProvider)
}

func (e *entityReference) DisableSound(source string) {
	if _, ok := e.system.disabledBehaviors[e.id]; !ok {
		e.system.disabledSounds[e.id] = map[string]struct{}{}
	}
	e.system.disabledSounds[e.id][source] = struct{}{}
}

func (e *entityReference) EnableSound(source string) {
	if _, ok := e.system.disabledSounds[e.id]; !ok {
		return
	}
	delete(e.system.disabledSounds[e.id], source)
}

func (e *entityReference) IsSoundEnabled() bool {
	return len(e.system.disabledSounds[e.id]) == 0
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
