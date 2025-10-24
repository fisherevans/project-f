package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
)

type PlayerBehavior struct {
	entityId        EntityId
	intentDirection input.Direction
	intentDuration  float64
}

func NewPlayerBehavior(entityId EntityId) *PlayerBehavior {
	return &PlayerBehavior{
		entityId: entityId,
	}
}

func (pb *PlayerBehavior) Update(s *State, timeDelta float64) {
	movement := s.movementController.Get(pb.entityId)
	if movement == nil {
		return
	}
	
	// Handle input during movement
	if movement.IsMoving() {
		if game.Controls[*State]().DPad().IsPressed() {
			pb.intentDirection = game.Controls[*State]().DPad().GetDirection()
		}
		// Handle speed changes while moving
		if game.Controls[*State]().ButtonB().IsPressed() {
			if movement.MoveState() == MoveStateWalking {
				movement.SetMoveState(MoveStateRunning)
			}
		} else {
			if movement.MoveState() == MoveStateRunning {
				movement.SetMoveState(MoveStateWalking)
			}
		}
		return
	}
	
	// Once player is done moving, stop accepting input if player does not have input priority
	if s.inputMode() != inputModePlayerMovement {
		return
	}
	
	// Face direction of intent after movement
	if pb.intentDirection != input.NotPressed && pb.intentDirection != movement.FacingDirection() {
		movement.SetFacingDirection(pb.intentDirection)
	}
	
	// Trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().DPad().GetDirection()
		movement.SetFacingDirection(direction)
		
		if pb.intentDirection != direction {
			pb.intentDirection = direction
			pb.intentDuration = 0
		}
		pb.intentDuration += timeDelta
		
		if pb.intentDuration > 0.075 {
			speed := MoveStateWalking
			if game.Controls[*State]().ButtonB().IsPressed() {
				speed = MoveStateRunning
			}
			s.movementController.TriggerMovement(s, pb.entityId, movement.GetFacingLocation(), speed)
		}
	}
	
	// Interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() {
		interactLocation := movement.InteractLocation()
		if interactLocation == nil {
			return
		}
		for _, entityIdAtLocation := range s.GetTileState(interactLocation.Location).EntitiesWithin {
			s.eventDispatcher.Dispatch(events.EventOnInteract{
				SourceId:        string(pb.entityId),
				SourceDirection: interactLocation.Direction,
				TargetId:        string(entityIdAtLocation),
			})
		}
	}
}

func (pb *PlayerBehavior) OnMovementComplete(s *State, movement *MovementState) bool {
	// Player doesn't auto-continue movement
	return false
}

func (pb *PlayerBehavior) DashTowards(s *State, direction input.Direction, location MapLocation) {
	movement := s.movementController.Get(pb.entityId)
	if movement == nil {
		return
	}
	
	location = pb.findDashDestination(s, direction, location)
	s.movementController.TriggerMovement(s, pb.entityId, location, MoveStateDashing)
}

func (pb *PlayerBehavior) findDashDestination(s *State, direction input.Direction, location MapLocation) MapLocation {
	for {
		ts := s.GetTileState(location)
		moved := false
		for _, entityId := range ts.EntitiesWithin {
			if entity, ok := s.entities[entityId]; ok {
				if _, ok := entity.(*EntityDashGap); ok {
					moved = true
					location = location.Moved(direction.GetVector())
				}
			}
		}
		if !moved {
			return location
		}
	}
}
