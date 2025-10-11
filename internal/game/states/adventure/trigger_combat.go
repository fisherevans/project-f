package adventure

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
)

func (adv *State) TriggerCombat(opponentIntent *rpg.PrimortalType, background string, onComplete func(*State)) {
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
		NewSimpleAction(func(s *State) {
			s.overlays.Add(NewFadeOverlay(
				pixel.RGBA{A: 0},
				pixel.RGBA{A: 1},
				1,
				NewBaseOverlay(swirlDuration, true, nil),
			))
			game.SetCustomShader(game.NewSwirlShader(swirlDuration))
		}),
		NewSleepAction(swirlDuration),
		NewSimpleAction(func(s *State) {
			s.blockInput = false
			s.enteringCombat = false
			game.RemoveCustomShader()
			var opponent rpg.PrimortalType
			if opponentIntent != nil {
				opponent = *opponentIntent
			} else {
				options := []rpg.PrimortalType{
					rpg.Primortal_Volteel.Type,
					rpg.Primortal_Toxmidge.Type,
					rpg.Primortal_Scintail.Type,
					rpg.Primortal_Myceli.Type,
					rpg.Primortal_Pumbl.Type,
				}
				opponent = options[rand.Intn(len(options))]
			}
			game.SetActiveStateIntent(game.CombatIntent{
				Run:        &rpg.Run{},
				Opponent:   opponent,
				Background: background,
				OnComplete: func(r game.CombatIntentResult) {
					game.DebugNotification("Combat complete!")
					if !r.PlayerWon {
						game.SetActiveStateIntent(game.InitialState())
						return
					}
					game.SetActiveStateIntent(game.SwapStateIntent{
						State: adv,
					})
					game.CurrentSave().Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
					if onComplete != nil {
						onComplete(adv)
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
