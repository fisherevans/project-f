package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

// MovementController handles tile-to-tile movement for all entities
type MovementController struct {
	states map[EntityId]*MovementState
}

func NewMovementController() *MovementController {
	return &MovementController{
		states: make(map[EntityId]*MovementState),
	}
}

// MovementState represents the movement state of a single entity
type MovementState struct {
	entityId         EntityId
	currentLocation  MapLocation
	targetLocation   MapLocation
	moveState        MoveState
	progression      float64
	constantMovement float64
	facingDirection  input.Direction
	speeds           map[MoveState]float64
	noClip           bool
}

func (mc *MovementController) Register(entityId EntityId, location MapLocation, speeds map[MoveState]float64) *MovementState {
	state := &MovementState{
		entityId:        entityId,
		currentLocation: location,
		moveState:       MoveStateIdle,
		speeds:          speeds,
		facingDirection: input.Down,
	}
	mc.states[entityId] = state
	return state
}

func (mc *MovementController) Unregister(entityId EntityId) {
	delete(mc.states, entityId)
}

func (mc *MovementController) Get(entityId EntityId) *MovementState {
	return mc.states[entityId]
}

// Update processes movement for all entities, handling time delta chaining
func (mc *MovementController) Update(s *State, timeDelta float64) {
	for id, state := range mc.states {
		remaining := timeDelta
		for remaining > 0 {
			elapsed := mc.updateSingleMovement(s, state, remaining)
			remaining -= elapsed

			// If movement completed and there's time left, check if entity wants to continue
			if remaining > 0 && state.moveState == MoveStateIdle {
				// Call behavior's OnMovementComplete if it exists
				if behavior, ok := s.behaviors[id]; ok {
					if !behavior.OnMovementComplete(s, state) {
						break // No more movement this frame
					}
				} else {
					break // No behavior, stop
				}
			}
		}
	}
}

// updateSingleMovement updates a single movement step, returns elapsed time
func (mc *MovementController) updateSingleMovement(s *State, state *MovementState, timeDelta float64) float64 {
	if timeDelta < 0 {
		return 0
	}
	if !state.IsMoving() {
		state.constantMovement = 0
		return timeDelta // Consume all time, no movement
	}

	moveSpeed := state.GetCurrentSpeed()
	moveDelta := timeDelta * moveSpeed
	state.constantMovement += moveDelta
	state.progression += moveDelta

	// Handle tile transition at 0.5 progression
	if state.progression >= 0.5 {
		if s.GetTileState(state.currentLocation).RemoveEntity(state.entityId) {
			state.emitZoneEvents(s, state.currentLocation, false)
		}
		if s.GetTileState(state.targetLocation).AddEntity(state.entityId) {
			state.emitZoneEvents(s, state.targetLocation, true)
			// When entering a new tile, cap progression to prevent skipping
			if state.progression >= 1 {
				state.progression = 1
			}
		}
	}

	// Check if movement is complete
	if state.progression >= 1.0 {
		state.currentLocation = state.targetLocation
		state.targetLocation = MapLocation{}
		state.moveState = MoveStateIdle
		remaining := state.progression - 1.0
		state.progression = 0
		state.emitZoneEvents(s, state.currentLocation, true)

		// Return remaining time for potential chained movement
		remainingTime := remaining / moveSpeed
		return timeDelta - remainingTime
	}

	return timeDelta // All time consumed
}

func (ms *MovementState) emitZoneEvents(s *State, loc MapLocation, isEntering bool) {
	for _, zoneId := range s.zones.ZonesAt(loc) {
		s.eventDispatcher.Dispatch(events.EventEntityZoneActivity{
			EntityId:   string(ms.entityId),
			ZoneId:     zoneId,
			IsEntering: isEntering,
		})
	}
}

// TriggerMovement attempts to start movement to a new location
func (mc *MovementController) TriggerMovement(s *State, entityId EntityId, target MapLocation, desiredMoveState MoveState) bool {
	state := mc.states[entityId]
	if state == nil {
		return false
	}

	if target == state.currentLocation {
		return false
	}
	if state.IsMoving() {
		return false
	}
	if !state.isValidMovement(s, state.currentLocation, target) {
		return false
	}

	state.targetLocation = target
	state.moveState = desiredMoveState
	state.facingDirection = state.currentLocation.DirectionTowards(target)
	return true
}

// TeleportTo instantly moves an entity to a new location
func (mc *MovementController) TeleportTo(s *State, entityId EntityId, newLocation MapLocation) bool {
	state := mc.states[entityId]
	if state == nil {
		return false
	}

	if state.moveState != MoveStateIdle {
		log.Warn().Msgf("cannot teleport '%s' while moving '%d'", entityId, state.moveState)
		return false
	}

	oldLocation := state.currentLocation
	s.GetTileState(newLocation).AddEntity(entityId)
	s.GetTileState(oldLocation).RemoveEntity(entityId)
	state.currentLocation = newLocation
	return true
}

// MovementState methods

func (ms *MovementState) IsMoving() bool {
	return ms.moveState != MoveStateIdle
}

func (ms *MovementState) GetCurrentSpeed() float64 {
	speed, exists := ms.speeds[ms.moveState]
	if !exists {
		return 0
	}
	return speed
}

func (ms *MovementState) SetMoveState(newState MoveState) {
	if ms.IsMoving() {
		ms.moveState = newState
	}
}

func (ms *MovementState) Location() MapLocation {
	return ms.currentLocation
}

func (ms *MovementState) FacingDirection() input.Direction {
	return ms.facingDirection
}

func (ms *MovementState) SetFacingDirection(dir input.Direction) {
	ms.facingDirection = dir
}

func (ms *MovementState) MoveState() MoveState {
	return ms.moveState
}

func (ms *MovementState) ConstantMovement() float64 {
	return ms.constantMovement
}

// PreciseLocation returns the interpolated position between tiles
func (ms *MovementState) PreciseLocation() pixel.Vec {
	location := ms.currentLocation.ToVec()
	if ms.IsMoving() {
		p := ms.progression
		if ms.moveState == MoveStateDashing {
			p = interp.Smootherstep(p)
		} else if ms.constantMovement < 1 {
			p = interp.EaseInToLinear(p, 2)
		}
		delta := ms.targetLocation.ToVec().Sub(ms.currentLocation.ToVec()).Scaled(p)
		location = location.Add(delta)
	}
	return location
}

func (ms *MovementState) GetLocationInDirection(dir input.Direction) MapLocation {
	return ms.currentLocation.Moved(dir.GetVector())
}

func (ms *MovementState) GetFacingLocation() MapLocation {
	return ms.GetLocationInDirection(ms.facingDirection)
}

func (ms *MovementState) isValidMovement(s *State, from, to MapLocation) bool {
	if ms.noClip {
		return true
	}

	movementDirection := from.DirectionTowards(to)
	if !s.CanEntityLeave(ms.entityId, movementDirection, from) {
		return false
	}
	if !s.CanEntityEnter(ms.entityId, movementDirection, to) {
		return false
	}
	return true
}

// InteractLocation returns the location in front of the entity if not moving
func (ms *MovementState) InteractLocation() *InteractionTarget {
	if ms.IsMoving() {
		return nil
	}
	dx, dy := ms.facingDirection.GetVector()
	return &InteractionTarget{
		Location:  ms.currentLocation.Moved(dx, dy),
		Direction: ms.facingDirection,
	}
}
