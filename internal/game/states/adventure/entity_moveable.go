package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"slices"
)

type MoveState int

const (
	MoveStateIdle MoveState = iota
	MoveStateWalking
	MoveStateRunning
	MoveStateDashing
)

type MoveableEntity struct {
	EntityId
	CurrentLocation MapLocation
	TargetLocation  MapLocation
	MoveState       MoveState
	MoveSpeeds      map[MoveState]float64
	MoveProgression float64
	FacingDirection input.Direction
}

func (m *MoveableEntity) Move(adv *State, timeDelta float64) float64 {
	if timeDelta < 0 {
		return 0
	}
	if !m.IsMoving() {
		return 0
	}
	moveSpeed := m.GetCurrentSpeed()
	moveDelta := timeDelta * moveSpeed
	m.MoveProgression += moveDelta
	if m.MoveProgression >= 0.5 {
		adv.unoccupy(m.CurrentLocation, m.EntityId)
	}
	if m.MoveProgression >= 1.0 {
		m.CurrentLocation = m.TargetLocation
		m.TargetLocation = MapLocation{}
		m.MoveState = MoveStateIdle
		remaining := m.MoveProgression - 1.0
		m.MoveProgression = 0
		return remaining / moveSpeed // todo this is weird
	}
	return 0
}

func (m *MoveableEntity) GetLocationInDirection(dir input.Direction) MapLocation {
	return m.CurrentLocation.Moved(dir.GetVector())
}

func (m *MoveableEntity) GetFacingLocation() MapLocation {
	return m.GetLocationInDirection(m.FacingDirection)
}

var validMovementMoveStates = []MoveState{MoveStateWalking, MoveStateRunning, MoveStateDashing}

func (m *MoveableEntity) TriggerMovement(adv *State, newLocation MapLocation, desiredMoveState MoveState) bool {
	if newLocation == m.CurrentLocation {
		return false
	}
	if !slices.Contains(validMovementMoveStates, desiredMoveState) {
		return false
	}
	if m.IsMoving() {
		return false
	}
	if !adv.attemptToOccupy(newLocation, m.EntityId) {
		return false
	}
	m.TargetLocation = newLocation
	m.MoveState = desiredMoveState
	return true
}

func (m *MoveableEntity) RenderMapLocation() pixel.Vec {
	location := m.CurrentLocation.ToVec()
	if m.IsMoving() {
		delta := m.TargetLocation.ToVec().Sub(m.CurrentLocation.ToVec()).Scaled(m.MoveProgression)
		location = location.Add(delta)
	}
	return location
}

func (m *MoveableEntity) Location() MapLocation {
	return m.CurrentLocation
}

func (m *MoveableEntity) Interact(ctx *game.Context, adv *State, source Entity) {

}

type InteractionTarget struct {
	Location  MapLocation
	Direction input.Direction
}

func (it *InteractionTarget) NextTile() {
	dx, dy := it.Direction.GetVector()
	it.Location = it.Location.Moved(dx, dy)
}

// InteractLocation returns the map location in front of the entity if they are not currently moving
func (m *MoveableEntity) InteractLocation() *InteractionTarget {
	if m.IsMoving() {
		return nil
	}
	dx, dy := m.FacingDirection.GetVector()
	return &InteractionTarget{
		Location:  m.CurrentLocation.Moved(dx, dy),
		Direction: m.FacingDirection,
	}
}

func (m *MoveableEntity) IsMoving() bool {
	return m.MoveState != MoveStateIdle
}

func (m *MoveableEntity) GetCurrentSpeed() float64 {
	speed, exists := m.MoveSpeeds[m.MoveState]
	if !exists {
		speed = 0
	}
	return speed
}
