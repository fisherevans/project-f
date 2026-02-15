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
	Reward           *game.CombatReward
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

	postCombat := func(combatState game.State, r game.CombatIntentResult) {
		if !r.PlayerWon {
			game.SetActiveStateIntent(game.StartupDeviceIntent{})
			return
		}
		game.SetActiveStateIntent(game.TransitionFlushIntent{
			BaseTransitionIntent: game.BaseTransitionIntent{
				From:    combatState,
				ToState: s,
			},
			Duration: 1.5,
		})
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
		NewMutateEntityBehaviorEffect(s.player).WithEnableBy("combat"),
		NewFunctionEffect(func(*State) {
			s.enteringCombat = false
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
			var reward game.CombatReward
			if e.Reward != nil {
				reward = *e.Reward
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
				Reward:     reward,
				OnComplete: postCombat,
				Background: e.Background,
			}
			if e.TrainingSequence != nil {
				intent.TrainingSequence = *e.TrainingSequence
			}
			game.SetActiveStateIntent(game.TransitionGlitchIntent{
				BaseTransitionIntent: game.BaseTransitionIntent{
					From:     s,
					ToIntent: intent,
				},
				Duration: 2.5,
			})
		}),
	)
	return true
}

type EffectLoadMap struct {
	instantEffect
	MapName  string
	Waypoint *string
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
			waypoint := "default"
			if e.Waypoint != nil {
				waypoint = *e.Waypoint
			}
			game.SetActiveStateIntent(game.AdventureIntent{
				MapName:  e.MapName,
				Waypoint: waypoint,
			})
		}),
	)
	return true
}
