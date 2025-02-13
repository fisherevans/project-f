package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type MoveableEntity struct {
	EntityId
	CurrentLocation MapLocation
	TargetLocation  MapLocation
	Moving          bool
	MoveSpeed       float64
	MoveProgression float64
	FacingDirection input.Direction
}

func (m *MoveableEntity) Move(adv *State, timeDelta float64) float64 {
	if timeDelta < 0 {
		return 0
	}
	if !m.Moving {
		return 0
	}
	moveDelta := timeDelta * m.MoveSpeed
	m.MoveProgression += moveDelta
	if m.MoveProgression >= 0.5 {
		adv.unoccupy(m.CurrentLocation, m.EntityId)
	}
	if m.MoveProgression >= 1.0 {
		m.CurrentLocation = m.TargetLocation
		m.TargetLocation = MapLocation{}
		m.Moving = false
		remaining := m.MoveProgression - 1.0
		m.MoveProgression = 0
		return remaining / m.MoveSpeed // todo this is weird
	}
	return 0
}

func (m *MoveableEntity) TriggerMovement(adv *State, direction input.Direction) bool {
	dx, dy := direction.GetVector()
	if m.Moving || (dx == 0 && dy == 0) {
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
	m.Moving = true
	return true
}

func (m *MoveableEntity) RenderMapLocation() pixel.Vec {
	location := m.CurrentLocation.ToVec()
	if m.Moving {
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
	if m.Moving {
		return nil
	}
	dx, dy := m.FacingDirection.GetVector()
	return &MapLocation{
		X: m.CurrentLocation.X + dx,
		Y: m.CurrentLocation.Y + dy,
	}
}
