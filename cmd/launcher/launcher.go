package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gopxl/pixel/v2/backends/opengl"

	_ "fisherevans.com/project/f/cmd/setup"
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/runtime"
	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/game/states/combat"
	"fisherevans.com/project/f/internal/game/states/menu"
	"fisherevans.com/project/f/internal/game/states/startup"
	"fisherevans.com/project/f/internal/game/states/state_selector"
	"fisherevans.com/project/f/internal/game/states/title"
	"fisherevans.com/project/f/internal/game/states/xenolog"
)

func main() {
	go func() { // expose pprof to diagnose memory usage
		http.ListenAndServe("localhost:6060", nil)
	}()
	game.RegisterStateFactory(adventure.New)
	game.RegisterStateFactory(combat.New)
	game.RegisterStateFactory(menu.New)
	game.RegisterStateFactory(xenolog.New)
	game.RegisterStateFactory(state_selector.New)
	game.RegisterStateFactory(title.New)
	game.RegisterStateFactory(startup.NewDevice)
	game.RegisterStateFactory(startup.NewCopyright)
	game.RegisterStateFactory(startup.NewDeveloper)
	opengl.Run(runtime.Run)
}
