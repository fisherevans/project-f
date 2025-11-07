package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
)

type MoveState int

const (
	MoveStateIdle MoveState = iota
	MoveStateWalking
	MoveStateRunning
	MoveStateDashing
)

type EntityMovement struct {
	System *EntitySystem
	Id     string

	FacingDirection input.Direction

	MovementState       MoveState
	TargetLocation      MapLocation
	Progression         float64
	ProgressionScale    float64 // used for moving more than 1 tile at a time (i.e. 4 at a time, scale == 1/4)
	MovementSpeeds      map[MoveState]float64
	AccumulatedMovement float64
}

func NewEntityMovement(id string, es *EntitySystem, location MapLocation) *EntityMovement {
	es.occupations.SetPosition(id, location, false)
	return &EntityMovement{
		System:           es,
		Id:               id,
		FacingDirection:  input.Down,
		ProgressionScale: 1,
	}
}

func (m *EntityMovement) getSpeed(moveState MoveState) float64 {
	speed, exists := m.MovementSpeeds[moveState]
	if !exists {
		speed = 1
	}
	return speed
}

func (ep *EntityMovement) ProgressMovement(timeDelta float64) float64 {
	if timeDelta <= 0 {
		return 0
	}
	if ep.MovementState == MoveStateIdle {
		ep.AccumulatedMovement = 0
		return 0
	}
	moveSpeed := ep.getSpeed(ep.MovementState)
	moveDelta := timeDelta * moveSpeed
	ep.AccumulatedMovement += moveDelta
	ep.Progression += moveDelta * ep.ProgressionScale
	if ep.Progression >= 1.0 {
		ep.System.occupations.Vacate(ep.Id, ep.System.occupations.GetPosition(ep.Id))
		newPrimary := ep.TargetLocation
		ep.TargetLocation = MapLocation{}
		ep.MovementState = MoveStateIdle
		remaining := ep.Progression - 1.0
		ep.Progression = 0
		ep.ProgressionScale = 1
		ep.System.occupations.SetPosition(ep.Id, newPrimary, false)
		//ep.System.occupations.EmitOccupyEvents(ep.Id, ep.Location, false)
		remainingTime := remaining / moveSpeed // movement is complete, but there is more time in the tick to move
		return remainingTime
	}
	return 0
}
