package adventure

import (
	"slices"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/interp"
)

type MoveState int

const (
	MoveStateIdle MoveState = iota
	MoveStateWalking
	MoveStateRunning
	MoveStateDashing
)

type MoveableEntity struct {
	BaseEntity
	CurrentLocation  MapLocation
	TargetLocation   MapLocation
	MoveState        MoveState
	MoveSpeeds       map[MoveState]float64
	MoveProgression  float64
	FacingDirection  input.Direction
	NoClip           bool
	ConstantMovement float64
}

func (m *MoveableEntity) Move(adv *State, timeDelta float64) float64 {
	if timeDelta < 0 {
		return 0
	}
	if !m.IsMoving() {
		m.ConstantMovement = 0
		return 0
	}
	moveSpeed := m.GetCurrentSpeed()
	moveDelta := timeDelta * moveSpeed
	m.ConstantMovement += moveDelta
	m.MoveProgression += moveDelta
	if m.MoveProgression >= 0.5 && m.IsPassable() {
		adv.unoccupy(m.CurrentLocation, m.Id)
	}
	if m.MoveProgression >= 1.0 {
		m.CurrentLocation = m.TargetLocation
		m.TargetLocation = MapLocation{}
		m.MoveState = MoveStateIdle
		remaining := m.MoveProgression - 1.0
		m.MoveProgression = 0
		if newTile, newTileExists := adv.movementRestrictions[m.CurrentLocation]; newTileExists {
			newTile.OnEntryComplete(adv, m.Id)
		}
		remainingTime := remaining / moveSpeed // movement is complete, but there is more time in the tick to move
		return remainingTime
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
	if !m.IsPassable() && !adv.attemptToOccupy(newLocation, m.Id) {
		return false
	}
	m.TargetLocation = newLocation
	m.MoveState = desiredMoveState
	if newTile, newTileExists := adv.movementRestrictions[newLocation]; newTileExists {
		newTile.OnEntryBegin(adv, m.Id)
	}
	return true
}

func (m *MoveableEntity) PreciseMapLocation() pixel.Vec {
	location := m.CurrentLocation.ToVec()
	if m.IsMoving() {
		p := m.MoveProgression
		if m.MoveState == MoveStateDashing {
			p = interp.Smootherstep(p)
		} else if m.ConstantMovement < 1 {
			p = interp.EaseInToLinear(p, 2)
		}
		delta := m.TargetLocation.ToVec().Sub(m.CurrentLocation.ToVec()).Scaled(p)
		location = location.Add(delta)
	}
	return location
}

func (m *MoveableEntity) RenderMapLocation() pixel.Vec {
	return m.PreciseMapLocation()
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

func (m *MoveableEntity) TeleportTo(s *State, location MapLocation) {
	if m.MoveState != MoveStateIdle {
		log.Warn().Msgf("cannot teleport '%s' while moving '%d'", m.Id, m.MoveState)
		return
	}
	if !m.IsPassable() {
		m.CurrentLocation = location
	}
	currentLocation := m.CurrentLocation
	if !s.attemptToOccupy(location, m.Id) {
		log.Warn().Msgf("cannot teleport '%s' to '%s' because it is occupied", m.Id, location)
		return
	}
	m.CurrentLocation = location
	s.unoccupy(currentLocation, m.Id)
}
