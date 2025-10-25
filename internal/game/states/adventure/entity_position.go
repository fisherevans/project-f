package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
)

type EntityPosition struct {
	System *EntitySystem
	Id     string

	Location        MapLocation
	FacingDirection input.Direction

	MovementState            types.MoveState
	MovementTargetLocation   MapLocation
	MovementProgression      float64
	MovementProgressionScale float64
	MovementSpeeds           map[types.MoveState]float64
	AccumulatedMovement      float64
}

func NewEntityPosition(id string, es *EntitySystem, location MapLocation) *EntityPosition {
	return &EntityPosition{
		System:                   es,
		Id:                       id,
		Location:                 location,
		FacingDirection:          input.Down,
		MovementProgressionScale: 1,
	}
}

func (ep *EntityPosition) GetCurrentSpeed() float64 {
	speed, exists := ep.MovementSpeeds[ep.MovementState]
	if !exists {
		speed = 1
	}
	return speed
}

func (ep *EntityPosition) ProgressMovement(timeDelta float64) float64 {
	if timeDelta <= 0 {
		return 0
	}
	if !ep.IsMoving() {
		ep.AccumulatedMovement = 0
		return 0
	}
	moveSpeed := ep.GetCurrentSpeed()
	moveDelta := timeDelta * moveSpeed
	ep.AccumulatedMovement += moveDelta
	ep.MovementProgression += moveDelta * ep.MovementProgressionScale
	if ep.MovementProgression >= 1.0 {
		ep.System.vacateAndEmitEvent(ep.Id, ep.Location, false)
		ep.Location = ep.MovementTargetLocation
		ep.MovementTargetLocation = MapLocation{}
		ep.MovementState = types.MoveStateIdle
		remaining := ep.MovementProgression - 1.0
		ep.MovementProgression = 0
		ep.MovementProgressionScale = 1
		ep.System.emitEntityLocationEnterEvents(ep.Id, ep.Location, false)
		remainingTime := remaining / moveSpeed // movement is complete, but there is more time in the tick to move
		return remainingTime
	}
	return 0
}

func (ep *EntityPosition) CancelMovement() {
	if !ep.IsMoving() {
		return
	}
	ep.System.vacateAndEmitEvent(ep.Id, ep.MovementTargetLocation, false)
	ep.MovementTargetLocation = ep.Location
	ep.MovementState = types.MoveStateIdle
	ep.MovementProgression = 0
	ep.AccumulatedMovement = 0
	ep.MovementProgressionScale = 1
}

func (ep *EntityPosition) PreciseLocation() pixel.Vec {
	location := ep.Location.ToVec()
	if !ep.IsMoving() {
		return location
	}
	p := ep.MovementProgression
	if ep.MovementState == types.MoveStateDashing {
		p = interp.Smootherstep(p)
	} else if ep.AccumulatedMovement < 1 {
		p = interp.EaseInToLinear(p, 2)
	}
	movementDelta := ep.MovementTargetLocation.ToVec().Sub(ep.Location.ToVec()).Scaled(p)
	return location.Add(movementDelta)
}

func (ep *EntityPosition) IsMoving() bool {
	return ep.MovementState != types.MoveStateIdle
}
