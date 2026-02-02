package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
)

type EffectTriggerCombat struct {
	CombatId         string `auto_generate:"true"`
	Opponent         *game.CombatOpponent
	Player           *game.CombatPlayer
	Background       string
	TrainingSequence *string
}

func (e *EffectTriggerCombat) CompletionID() string {
	if e.CombatId == "" {
		return ""
	}
	return "combat:" + e.CombatId
}

func (e *EffectTriggerCombat) Process(source EntityReader, s *State) bool {
	if s.enteringCombat {
		return false
	}
	s.enteringCombat = true

	postCombat := func(r game.CombatIntentResult) {
		game.DebugNotificationf("Combat complete!")
		if !r.PlayerWon {
			game.SetActiveStateIntent(game.StartupDeviceIntent{})
			return
		}
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s,
		})
		game.CurrentSave().Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
		s.ExecuteSystemEffects(NewDeactivateFadeEffect("combat_fade"))
		s.eventDispatcher.Dispatch(&EventCombatComplete{
			CombatId: e.CombatId,
			Result:   "completed", // TODO: serialize result properly
		})
		s.planExecutor.MarkComplete(e.CompletionID())
	}

	s.ExecuteSystemEffectsInOrder(
		NewMutateEntityBehaviorEffect(s.player).WithDisableBy("combat"),
		&EffectFade{
			DurationSeconds: 1,
			AutoDeactivate:  util.Ptr(true),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     6,
		},
		&EffectFade{
			FadeId:          "combat_fade",
			DurationSeconds: 3,
			AutoDeactivate:  util.Ptr(false),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     1,
		},
		NewFunctionEffect(func(*State) {
			game.SetCustomShader(game.NewSwirlShader(3))
		}),
		NewMutateEntityBehaviorEffect(s.player).WithEnableBy("combat"),
		NewFunctionEffect(func(*State) {
			s.enteringCombat = false
			game.RemoveCustomShader()
			var opponent game.CombatOpponent
			if e.Opponent != nil {
				opponent = *e.Opponent
			} else {
				options := []rpg.PrimortalType{
					rpg.Primortal_Volteel.Type,
					rpg.Primortal_Toxmidge.Type,
					rpg.Primortal_Scintail.Type,
					rpg.Primortal_Myceli.Type,
					rpg.Primortal_Pumbl.Type,
				}
				opponent = game.CombatOpponent{
					Type: options[rand.Intn(len(options))],
				}
			}
			var player game.CombatPlayer
			if e.Player != nil {
				player = *e.Player
			} else {
				player = game.NewCombatPlayer(game.CurrentSave().Animech)
			}
			intent := game.CombatIntent{
				Player:     player,
				Opponent:   opponent,
				OnComplete: postCombat,
				Background: e.Background,
			}
			if e.TrainingSequence != nil {
				intent.TrainingSequence = *e.TrainingSequence
			}
			game.SetActiveStateIntent(intent)
		}),
	)
	return true
}

type EffectLoadMap struct {
	instantEffect
	MapName string
}

func (e *EffectLoadMap) Process(source EntityReader, s *State) bool {
	if e == nil {
		return false
	}
	if err := game.CurrentSave().Save(); err != nil {
		game.DebugNotificationf("failed to save: %v", err)
	}
	s.ExecuteSystemEffectsInOrder(
		NewMutateEntityBehaviorEffect(s.player).WithDisableBy("map_load"),
		NewFadeEffect(1, 1).
			WithAutoDeactivate(false).
			WithFromColor("#0000").
			WithToColor("#000f"),
		NewFunctionEffect(func(*State) {
			game.SetActiveStateIntent(game.AdventureIntent{
				MapName: e.MapName,
			})
		}),
	)
	return true
}
