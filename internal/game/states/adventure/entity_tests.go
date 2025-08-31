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
	adv.blockInput = true
	flashDuration := 1.0
	swirlDuration := 3.0
	adv.overlays.Add(NewFadeOverlay(
		pixel.RGBA{A: 0},
		pixel.RGBA{A: 1},
		6,
		NewBaseOverlay(flashDuration, true, nil),
	))
	adv.actions.Add(NewSerialActions(
		NewSleepAction(flashDuration),
		NewSimpleAction(func(ctx *game.Context, s *State) {
			s.overlays.Add(NewFadeOverlay(
				pixel.RGBA{A: 0},
				pixel.RGBA{A: 1},
				1,
				NewBaseOverlay(swirlDuration, true, nil),
			))
			ctx.SetCustomShader(game.NewSwirlShader(swirlDuration))
		}),
		NewSleepAction(swirlDuration),
		NewSimpleAction(func(ctx *game.Context, s *State) {
			s.blockInput = false
			ctx.RemoveCustomShader()
			ctx.SwapActiveState(combat.New(adv.animech, func(ctx *game.Context, combatState *combat.State) {
				ctx.Notify("Combat complete!")
				ctx.SwapActiveState(adv)
				// reset health for now
				adv.animech.CurrentShield = adv.animech.GetMaxShield()
				for _, p := range adv.animech.DeployedPrimortals {
					p.CurrentSync = p.GetMaxSync()
				}
			}))
		}),
	))
}
