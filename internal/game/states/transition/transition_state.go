package transition

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"
)

type baseState struct {
	game.BaseState
	fromState, toState       game.State
	fromSnapshot, toSnapshot *opengl.Canvas
}

func newBaseState(intent game.BaseTransitionIntent) *baseState {
	toState := intent.ToState
	if toState == nil {
		var err error
		toState, err = game.CreateStateFromIntent(intent.ToIntent)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create state from intent")
		}
	}
	s := &baseState{
		fromState:    intent.From,
		fromSnapshot: takeSnapshot(intent.From),

		toState:    toState,
		toSnapshot: takeSnapshot(toState),
	}
	return s
}

func takeSnapshot(s game.State) *opengl.Canvas {
	snapshot := opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
	s.OnTick(snapshot, pixel.R(0, 0, game.GameWidth, game.GameHeight), 1.0/1000.0)
	return snapshot
}
