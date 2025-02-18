package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
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

func (m *MoveableEntity) TriggerMovement(adv *State, direction input.Direction, running bool) bool {
	dx, dy := direction.GetVector()
	if m.IsMoving() || (dx == 0 && dy == 0) {
		return false
	}
	if dx > 1 || dy > 1 || dx < -1 || dy < -1 || (dx != 0 && dy != 0) {
		log.Error().Msgf("got an unexpect move: %d,%d", dx, dy)
		return false
	}
	m.FacingDirection = direction
	newLocation := MapLocation{
		X: m.CurrentLocation.X + dx,
		Y: m.CurrentLocation.Y + dy,
	}
	if !adv.attemptToOccupy(newLocation, m.EntityId) {
		return false
	}
	m.TargetLocation = newLocation
	if running {
		m.MoveState = MoveStateRunning
	} else {
		m.MoveState = MoveStateWalking
	}
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

// InteractLocation returns the map location in front of the entity if they are not currently moving
func (m *MoveableEntity) InteractLocation() *MapLocation {
	if m.IsMoving() {
		return nil
	}
	dx, dy := m.FacingDirection.GetVector()
	return &MapLocation{
		X: m.CurrentLocation.X + dx,
		Y: m.CurrentLocation.Y + dy,
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
