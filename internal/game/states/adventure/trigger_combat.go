package adventure

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
)

func (adv *State) TriggerCombat(onComplete func(*game.Context, *State)) {
	if adv.enteringCombat {
		return
	}
	adv.enteringCombat = true
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
			s.enteringCombat = false
			ctx.RemoveCustomShader()
			options := []rpg.PrimortalType{
				rpg.Primortal_Volteel.Type,
				rpg.Primortal_Toxmidge.Type,
				rpg.Primortal_Scintail.Type,
				rpg.Primortal_Myceli.Type,
				rpg.Primortal_Pumbl.Type,
			}
			opponent := options[rand.Intn(len(options))]
			ctx.SetActiveStateIntent(game.CombatIntent{
				Run:      &rpg.Run{},
				Opponent: opponent,
				OnComplete: func(ctx *game.Context, r game.CombatIntentResult) {
					ctx.Notify("Combat complete!")
					if !r.PlayerWon {
						ctx.SetActiveStateIntent(game.InitialState())
						return
					}
					ctx.SetActiveStateIntent(game.SwapStateIntent{
						State: adv,
					})
					ctx.GameSave.Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
					if onComplete != nil {
						onComplete(ctx, adv)
					}
					// uncomment to fully recover after battle
					//adv.animech.CurrentShield = adv.animech.GetMaxShield()
					//for _, p := range adv.animech.DeployedPrimortals {
					//	p.CurrentSync = p.GetMaxSync()
					//}
				},
			})
		}),
	))
}
