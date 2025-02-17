package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/combat"
	"github.com/gopxl/pixel/v2"
)

type EntityCombat struct {
	InnateEntity
}

func (e *EntityCombat) Render(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityCombat) Interact(ctx *game.Context, adv *State, source Entity) {
	ctx.SwapActiveState(combat.New(adv.animech, func(ctx *game.Context, combatState *combat.State) {
		ctx.Notify("Combat complete!")
		ctx.SwapActiveState(adv)
	}))
}
