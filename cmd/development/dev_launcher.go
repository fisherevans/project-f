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
	instance := runtime.NewInstance("default", func() any {
		return game.StartupDeviceIntent{}
	}).WithDevScenes(devScenes())
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	opengl.Run(instance.Run)
}

// devScenes returns the set of named entry points exposed in the overlay's
// SCENES section. Only reachable in dev builds.
func devScenes() []runtime.DevScene {
	// Each combat scenario needs an OnComplete; send them back to the device
	// screen so the overlay's scene picker can kick off another scenario.
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
