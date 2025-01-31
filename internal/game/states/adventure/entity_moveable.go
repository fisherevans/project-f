package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fmt"
	"github.com/gopxl/pixel/v2"
)

type MoveableEntity struct {
	EntityReference
	CurrentLocation MapLocation
	TargetLocation  MapLocation
	Moving          bool
	MoveSpeed       float64
	MoveProgression float64
	LastDirection   input.Direction
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
		adv.unoccupy(m.CurrentLocation, m.EntityReference)
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

func (m *MoveableEntity) TriggerMovement(adv *State, dx, dy int) bool {
	if m.Moving || (dx == 0 && dy == 0) {
		return false
	}
	if dx > 1 || dy > 1 || dx < -1 || dy < -1 || (dx != 0 && dy != 0) {
		fmt.Printf("got an unexpect move: %d,%d\n", dx, dy)
		return false
	}
	newLocation := MapLocation{
		X: m.CurrentLocation.X + dx,
		Y: m.CurrentLocation.Y + dy,
	}
	if !adv.attemptToOccupy(newLocation, m.EntityReference) {
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
