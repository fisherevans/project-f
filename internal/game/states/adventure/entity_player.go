package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
)

type Player struct {
	AnimatedMoveableEntity

	intentDirection input.Direction
	intentDuration  float64
}

func (p *Player) Update(ctx *game.Context, adv *State, timeDelta float64) {
	defer p.AnimatedMoveableEntity.Update(ctx, adv, timeDelta)
	// trigger running or face new direction after movement
	if p.IsMoving() {
		if ctx.Controls.DPad().IsPressed() {
			p.intentDirection = ctx.Controls.DPad().GetDirection()
		}
		if ctx.Controls.ButtonB().IsPressed() {
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
	if ctx.Controls.DPad().IsPressed() {
		direction := ctx.Controls.DPad().GetDirection()
		p.FacingDirection = direction
		if p.intentDirection != direction {
			p.intentDirection = direction
			p.intentDuration = 0
		}
		p.intentDuration += timeDelta
		if p.intentDuration > 0.05 {
			speed := MoveStateWalking
			if ctx.Controls.ButtonB().IsPressed() {
				speed = MoveStateRunning
			}
			p.TriggerMovement(adv, p.GetFacingLocation(), speed)
		}
	}
	// interact with item if player is pressing A
	if ctx.Controls.ButtonA().JustPressedOrRepeated() {
		interactLocation := p.InteractLocation()
		if interactLocation == nil {
			return
		}
		targetEntityId, entityExists := adv.occupiedBy(interactLocation.Location)
		if entityExists {
			targetEntity := adv.entities[targetEntityId]
			targetEntity.Interact(ctx, adv, p)
			return
		}
		doDash := false
		for {
			movementTile, movementExists := adv.movementRestrictions[interactLocation.Location]
			if !movementExists || movementTile.EntryAllowed() {
				if doDash {
					p.TriggerMovement(adv, interactLocation.Location, MoveStateDashing)
				}
				break
			}
			if movementTile.CanDashOver() {
				doDash = true
				interactLocation.NextTile()
				continue
			}
			break
		}
	}
}
