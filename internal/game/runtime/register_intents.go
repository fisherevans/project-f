package runtime

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/states/adventure"
	"fisherevans.com/project/f/internal/game/states/combat"
	"fisherevans.com/project/f/internal/game/states/menu"
	"fisherevans.com/project/f/internal/game/states/startup"
	"fisherevans.com/project/f/internal/game/states/state_selector"
	"fisherevans.com/project/f/internal/game/states/title"
	"fisherevans.com/project/f/internal/game/states/transition"
	"fisherevans.com/project/f/internal/game/states/xenolog"
)

func registerIntents() {
	game.RegisterStateFactory(adventure.New)
	game.RegisterStateFactory(combat.New)
	game.RegisterStateFactory(menu.New)
	game.RegisterStateFactory(xenolog.New)
	game.RegisterStateFactory(state_selector.New)
	game.RegisterStateFactory(title.New)
	game.RegisterStateFactory(startup.NewDevice)
	game.RegisterStateFactory(startup.NewCopyright)
	game.RegisterStateFactory(startup.NewDeveloper)
	game.RegisterStateFactory(startup.NewControls)
	game.RegisterStateFactory(game.DoSwapStateIntent)
	game.RegisterStateFactory(transition.NewFadeTransition)
	game.RegisterStateFactory(transition.NewSwirlTransition)
	game.RegisterStateFactory(transition.NewGlitchTransition)
	game.RegisterStateFactory(transition.NewFlushTransition)
}
