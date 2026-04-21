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
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat One Hit Loss", Factory: func() any {
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
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat 1", Factory: func() any {
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
				OnComplete:       backToDevice,
			}
		}},
		{Name: "Training Combat 2", Factory: func() any {
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
				OnComplete:       backToDevice,
			}
		}},
	}

	for _, p := range []rpg.PrimortalType{
		rpg.Primortal_Dummy.Type,
		rpg.Primortal_Pumbl.Type,
		rpg.Primortal_Myceli.Type,
		rpg.Primortal_Scintail.Type,
		rpg.Primortal_Toxmidge.Type,
		rpg.Primortal_Volteel.Type,
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
