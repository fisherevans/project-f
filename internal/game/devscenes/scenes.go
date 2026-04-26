package devscenes

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/runtime"
)

// Scenes returns the named entry points shown in the overlay SCENES section.
func Scenes() []runtime.DevScene {
	backToDevice := func(_ game.State, _ game.CombatIntentResult) {
		game.SetActiveStateIntent(game.StartupDeviceIntent{})
	}

	scenes := []runtime.DevScene{
		{Name: "Travel: Sylvoria", Factory: func() any {
			return game.TravelIntent{
				ToIntent: game.AdventureIntent{
					MapName:  "sylvoria",
					Waypoint: "new_beginnings",
				},
				PlanetSpriteName: "computer/planet_1",
			}
		}},
		{Name: "Adventure: Intro", Factory: func() any {
			return game.AdventureIntent{MapName: "intro"}
		}},
		{Name: "Adventure: HQ", Factory: func() any {
			return game.AdventureIntent{MapName: "hq"}
		}},
		{Name: "Adventure: Map 1", Factory: func() any {
			return game.AdventureIntent{MapName: "map1"}
		}},
		{Name: "Training Combat One Hit Win", Factory: func() any {
			return game.CombatIntent{
				Opponent: game.CombatOpponent{
					Type:      "dummy",
					Archetype: "onehit",
				},
				Player: game.CombatPlayer{
					SkillSet: &rpg.SkillSet{
						Skill1: "tackle",
						Skill2: "guard",
					},
					InitialSync:   25,
					MaxSync:       25,
					InitialShield: 15,
					MaxShield:     15,
				},
				Reward: game.CombatReward{
					ExperiencePoints: 10,
					ResearchPoints:   2,
					ResearchType:     "dummy",
				},
				TrainingSequence: "none",
				Background:       rpg.CombatBGSpaceBase,
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat One Hit Loss", Factory: func() any {
			return game.CombatIntent{
				Opponent: game.CombatOpponent{
					Type:      "dummy",
					Archetype: "aggressive",
				},
				Player: game.CombatPlayer{
					SkillSet: &rpg.SkillSet{
						Skill1: "tackle",
						Skill2: "guard",
					},
					InitialSync:   1,
					MaxSync:       25,
					InitialShield: 1,
					MaxShield:     15,
				},
				Reward: game.CombatReward{
					ExperiencePoints: 10,
					ResearchPoints:   2,
					ResearchType:     "dummy",
				},
				TrainingSequence: "none",
				Background:       rpg.CombatBGSpaceBase,
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat 1", Factory: func() any {
			return game.CombatIntent{
				Opponent: game.CombatOpponent{
					Type:      "dummy",
					Archetype: "training.1",
				},
				Player: game.CombatPlayer{
					SkillSet: &rpg.SkillSet{
						Skill1: "tackle",
						Skill2: "guard",
					},
					InitialSync:   25,
					MaxSync:       25,
					InitialShield: 15,
					MaxShield:     15,
				},
				TrainingSequence: "training.1",
				Background:       rpg.CombatBGSpaceBase,
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat 2", Factory: func() any {
			return game.CombatIntent{
				Opponent: game.CombatOpponent{
					Type:      "toxmidge",
					Archetype: "training.2",
				},
				Player: game.CombatPlayer{
					SkillSet: &rpg.SkillSet{
						Skill1: "tackle",
						Skill2: "guard",
					},
					InitialSync:   25,
					MaxSync:       25,
					InitialShield: 15,
					MaxShield:     15,
				},
				TrainingSequence: "training.2",
				Background:       rpg.CombatBGSpaceBase,
				OnComplete:       backToDevice,
			}
		}},
	}

	for _, p := range []rpg.PrimortalType{
		"dummy",
		"pumbl",
		"myceli",
		"scintail",
		"toxmidge",
		"volteel",
	} {
		p := p
		scenes = append(scenes, runtime.DevScene{
			Name: "Fight " + rpg.Primortals[p].Name,
			Factory: func() any {
				return game.CombatIntent{
					Opponent:   game.NewCombatOpponent(p, ""),
					Player:     game.NewCombatPlayer(game.CurrentSave().Animech),
					Background: "combat/background_sylvoria",
					OnComplete: backToDevice,
				}
			},
		})
	}

	return scenes
}
