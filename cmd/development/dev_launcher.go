package main

import (
	// used in dev builds to expose profiler
	"net/http"
	_ "net/http/pprof"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/setup"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

func main() {
	setup.SetupLogging()
	instance := runtime.NewInstance("default", createInitialIntent)
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	opengl.Run(instance.Run)
}

func createInitialIntent() any {
	i := game.SelectIntent{}
	i = i.With("Adventure: Intro", func() any {
		return game.AdventureIntent{
			MapName: "intro",
		}
	})
	i = i.With("Adventure: HQ", func() any {
		return game.AdventureIntent{
			MapName: "hq",
		}
	})
	i = i.With("Adventure: Map 1", func() any {
		return game.AdventureIntent{
			MapName: "map1",
		}
	})

	i = i.With("Training Combat One Hit Win", func() any {
		return game.CombatIntent{
			Opponent: game.CombatOpponent{
				Type:      rpg.Primortal_Dummy.Type,
				Archetype: "onehit",
			},
			Player: game.CombatPlayer{
				SkillSet: &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				},
				InitialSync:   25,
				MaxSync:       25,
				InitialShield: 15,
				MaxShield:     15,
			},
			Reward: game.CombatReward{
				ExperiencePoints: 10,
				ResearchPoints:   2,
				ResearchType:     rpg.Primortal_Dummy.Type,
			},
			TrainingSequence: "none",
			Background:       rpg.CombatBGSpaceBase,
			OnComplete: func(_ game.State, r game.CombatIntentResult) {
				game.SetActiveStateIntent(createInitialIntent())
			},
		}
	})

	i = i.With("Training Combat One Hit Loss", func() any {
		return game.CombatIntent{
			Opponent: game.CombatOpponent{
				Type:      rpg.Primortal_Dummy.Type,
				Archetype: "aggressive",
			},
			Player: game.CombatPlayer{
				SkillSet: &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				},
				InitialSync:   1,
				MaxSync:       25,
				InitialShield: 1,
				MaxShield:     15,
			},
			Reward: game.CombatReward{
				ExperiencePoints: 10,
				ResearchPoints:   2,
				ResearchType:     rpg.Primortal_Dummy.Type,
			},
			TrainingSequence: "none",
			Background:       rpg.CombatBGSpaceBase,
			OnComplete: func(_ game.State, r game.CombatIntentResult) {
				game.SetActiveStateIntent(createInitialIntent())
			},
		}
	})

	i = i.With("Training Combat 1", func() any {
		return game.CombatIntent{
			Opponent: game.CombatOpponent{
				Type:      rpg.Primortal_Dummy.Type,
				Archetype: "training.1",
			},
			Player: game.CombatPlayer{
				SkillSet: &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				},
				InitialSync:   25,
				MaxSync:       25,
				InitialShield: 15,
				MaxShield:     15,
			},
			TrainingSequence: "training.1",
			Background:       rpg.CombatBGSpaceBase,
			OnComplete: func(_ game.State, r game.CombatIntentResult) {
				game.SetActiveStateIntent(createInitialIntent())
			},
		}
	})

	i = i.With("Training Combat 2", func() any {
		return game.CombatIntent{
			Opponent: game.CombatOpponent{
				Type:      rpg.Primortal_Toxmidge.Type,
				Archetype: "training.2",
			},
			Player: game.CombatPlayer{
				SkillSet: &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				},
				InitialSync:   25,
				MaxSync:       25,
				InitialShield: 15,
				MaxShield:     15,
			},
			TrainingSequence: "training.2",
			Background:       rpg.CombatBGSpaceBase,
			OnComplete: func(_ game.State, r game.CombatIntentResult) {
				game.SetActiveStateIntent(createInitialIntent())
			},
		}
	})

	fight := func(p rpg.PrimortalType) {
		i = i.With("Fight "+rpg.Primortals[p].Name, func() any {
			return game.CombatIntent{
				Opponent:   game.NewCombatOpponent(p, ""),
				Player:     game.NewCombatPlayer(game.CurrentSave().Animech),
				Background: "combat/background_sylvoria",
				OnComplete: func(_ game.State, r game.CombatIntentResult) {
					game.SetActiveStateIntent(createInitialIntent())
				},
			}
		})
	}
	fight(rpg.Primortal_Dummy.Type)
	fight(rpg.Primortal_Pumbl.Type)
	fight(rpg.Primortal_Myceli.Type)
	fight(rpg.Primortal_Scintail.Type)
	fight(rpg.Primortal_Toxmidge.Type)
	fight(rpg.Primortal_Volteel.Type)
	return i
}
