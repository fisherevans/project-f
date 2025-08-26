package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/combat"
	"fisherevans.com/project/f/internal/game/states/menu"
)

type EntityMenuTest struct {
	InnateEntity
}

func (e *EntityMenuTest) Render(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityMenuTest) Interact(ctx *game.Context, adv *State, source Entity) {
	ctx.SwapActiveState(menu.New(adv))
}

type EntityCombatTest struct {
	InnateEntity
}

func (e *EntityCombatTest) Render(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityCombatTest) Interact(ctx *game.Context, adv *State, source Entity) {
	ctx.SwapActiveState(combat.New(adv.animech, func(ctx *game.Context, combatState *combat.State) {
		ctx.Notify("Combat complete!")
		ctx.SwapActiveState(adv)
		// reset health for now
		adv.animech.CurrentShield = adv.animech.GetMaxShield()
		for _, p := range adv.animech.DeployedPrimortals {
			p.CurrentSync = p.GetMaxSync()
		}
	}))
}
