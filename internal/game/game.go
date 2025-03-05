package game

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"image/color"
)

const (
	GameWidth  = 240
	GameHeight = 160
)

type DebugArea int

const (
	AreaTopLeft DebugArea = iota
	AreaTopRight
	AreaBottomLeft
	AreaBottomRight
)

type State interface {
	ClearColor() color.Color
	OnTick(ctx *Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64)
}

type BaseState struct{}

func (s *BaseState) ClearColor() color.Color {
	return color.Black
}

type Context struct {
	DebugInfo

	activeState State

	CanvasScale         float64
	CanvasMousePosition pixel.Vec
	MouseInCanvas       bool
	Controls            *input.Controls

	DebugToggles *DebugToggles

	GameSave *rpg.GameSave
}

func NewContext(initialActiveState State, saveId string) *Context {
	saves, err := rpg.LoadGameSaves()
	if err != nil {
		panic(err)
	}
	save, ok := saves[saveId]
	if !ok {
		panic("Save not found: " + saveId)
	}
	return &Context{
		activeState:  initialActiveState,
		CanvasScale:  1.0,
		Controls:     input.NewControls(),
		GameSave:     save,
		DebugToggles: newToggles(),
	}
}

func (c *Context) Update(window *opengl.Window) {
	c.Controls.Update(window)
	c.DebugToggles.update(window)
}

func (c *Context) GetActiveState() State {
	return c.activeState
}

func (c *Context) SwapActiveState(newState State) State {
	oldState := c.activeState
	c.activeState = newState
	return oldState
}
