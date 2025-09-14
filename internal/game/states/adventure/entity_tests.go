package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/menu"
)

type EntityMenuTest struct {
	InnateEntity
}

func (e *EntityMenuTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityMenuTest) Interact(ctx *game.Context, adv *State, source Entity) {
	ctx.SwapActiveState(menu.New(adv))
}

type EntityCombatTest struct {
	InnateEntity
}

func (e *EntityCombatTest) RenderScene(target pixel.Target, matrix pixel.Matrix) {

}

func (e *EntityCombatTest) Interact(ctx *game.Context, adv *State, source Entity) {
	adv.TriggerCombat()
}
