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
	instance := runtime.NewInstance("default", createInitialIntent())
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	opengl.Run(instance.Run)
}

func createInitialIntent() any {
	i := game.SelectIntent{}
	i = i.With("Adventure", func() any {
		return game.AdventureIntent{
			MapName: "intro", // map1
		}
	})

	fight := func(p rpg.PrimortalType) {
		i = i.With("Fight "+rpg.Primortals[p].Name, func() any {
			return game.CombatIntent{
				Opponent:   p,
				Background: "combat/background_sylvoria",
				OnComplete: func(r game.CombatIntentResult) {
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
