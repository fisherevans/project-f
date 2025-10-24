package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
)

type Player struct {
	AnimatedMoveableEntity

	intentDirection input.Direction
	intentDuration  float64
}

func (p *Player) Update(adv *State, timeDelta float64) {
	defer p.AnimatedMoveableEntity.Update(adv, timeDelta)
	// trigger running or face new direction after movement
	if p.IsMoving() {
		if game.Controls[*State]().DPad().IsPressed() {
			p.intentDirection = game.Controls[*State]().
				DPad().GetDirection()
		}
		if game.Controls[*State]().ButtonB().IsPressed() {
			if p.MoveState == MoveStateWalking {
				p.MoveState = MoveStateRunning
			}
		} else {
			if p.MoveState == MoveStateRunning {
				p.MoveState = MoveStateWalking
			}
		}
		return
	}
	// once player is done moving, stop accepting input if player does not have input priority
	if adv.inputMode() != inputModePlayerMovement {
		return
	}
	// face direction of intent after movement
	if p.intentDirection != input.NotPressed && p.intentDirection != p.FacingDirection {
		p.FacingDirection = p.intentDirection
	}
	// trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().
			DPad().GetDirection()
		p.FacingDirection = direction
		if p.intentDirection != direction {
			p.intentDirection = direction
			p.intentDuration = 0
		}
		p.intentDuration += timeDelta
		if p.intentDuration > 0.075 {
			speed := MoveStateWalking
			if game.Controls[*State]().ButtonB().IsPressed() {
				speed = MoveStateRunning
			}
			p.TriggerMovement(adv, p.GetFacingLocation(), speed)
		}
	}
	// interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() {
		interactLocation := p.InteractLocation()
		if interactLocation == nil {
			return
		}
		for _, entityIdAtLocation := range adv.GetTileState(interactLocation.Location).EntitiesWithin {
			adv.eventDispatcher.Dispatch(events.EventOnInteract{
				SourceId:        string(p.id),
				SourceDirection: interactLocation.Direction,
				TargetId:        string(entityIdAtLocation),
			})
		}
	}
}

func (p *Player) DashTowards(s *State, direction input.Direction, location MapLocation) {
	location = p.findDashDestination(s, direction, location)
	p.TriggerMovement(s, location, MoveStateDashing)
}

func (p *Player) findDashDestination(s *State, direction input.Direction, location MapLocation) MapLocation {
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
