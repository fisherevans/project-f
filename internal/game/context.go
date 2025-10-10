package game

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
)

var Ctx *Context

type Context struct {
	DebugInfo

	activeState  State
	customShader AppliedShader

	stateIntent any

	CanvasScale         float64
	CanvasMousePosition pixel.Vec
	MouseInCanvas       bool
	Controls            *input.Controls
	Window              *opengl.Window

	GameSave *rpg.GameSave

	Elapsed      float64
	Utils        ContextUtils
	DebugToggles *DebugToggles
}

func NewContext(window *opengl.Window, saveId string) *Context {
	saves, err := rpg.LoadGameSaves()
	if err != nil {
		panic(err)
	}
	save, ok := saves[saveId]
	if !ok {
		panic("Save not found: " + saveId)
	}
	ctx := &Context{
		stateIntent:  InitialState(),
		CanvasScale:  1.0,
		Controls:     input.NewControls(),
		Window:       window,
		GameSave:     save,
		DebugToggles: newToggles(),
	}
	ctx.Utils = ContextUtils{
		ctx: ctx,
	}
	return ctx

}

func (c *Context) Update(window *opengl.Window, timeDelta float64) {
	c.Controls.Update(window)
	c.DebugToggles.update(window)
	c.Elapsed += timeDelta
}

func (c *Context) GetActiveState() State {
	return c.activeState
}

func (c *Context) SetActiveStateIntent(intent any) {
	if c.stateIntent != nil {
		log.Warn().Msgf("something is overriding an existing state intent")
	}
	c.stateIntent = intent
}

func (c *Context) ApplyIntent() {
	if c.stateIntent == nil {
		return
	}
	s, err := createState(c, c.stateIntent)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new state from intent")
	}
	c.activeState = s
	c.stateIntent = nil
}

func (c *Context) WithNoControls() *Context {
	without := *c
	without.Controls = input.NewControls()
	return &without
}

func (c *Context) SetCustomShader(shader AppliedShader) {
	c.customShader = shader
}

func (c *Context) GetCustomShader() AppliedShader {
	return c.customShader
}

func (c *Context) RemoveCustomShader() {
	c.customShader = nil
}

type AppliedShader interface {
	Apply(shaderOptions shaders.Options, timeDelta float64)
}
