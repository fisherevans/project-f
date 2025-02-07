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
	if p.Moving {
		return
	}
	if adv.inputMode != inputModePlayerMovement {
		return
	}
	if ctx.Controls.ButtonA().JustPressedOrRepeated() {
		interactLocation := p.InteractLocation()
		if interactLocation != nil {
			targetEntityId, exists := adv.occupiedBy(*interactLocation)
			if exists {
				targetEntity := adv.entities[targetEntityId]
				targetEntity.Interact(ctx, adv, p)
			}
		}
	}
	if ctx.Controls.DPad().IsPressed() {
		direction := ctx.Controls.DPad().GetDirection()
		p.FacingDirection = direction
		if p.intentDirection != direction {
			p.intentDirection = direction
			p.intentDuration = 0
		}
		p.intentDuration += timeDelta
		if p.intentDuration > 0.05 {
			p.TriggerMovement(adv, ctx.Controls.DPad().GetDirection())
		}
	}
}
