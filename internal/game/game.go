package game

import (
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
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
	OnTick(ctx *Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64)
}
type Context struct {
	DebugInfo

	activeState State

	CanvasScale         float64
	CanvasMousePosition pixel.Vec
	MouseInCanvas       bool
	Controls            *input.Controls
}

func NewContext(initialActiveState State) *Context {
	return &Context{
		activeState: initialActiveState,
		CanvasScale: 1.0,
		Controls:    input.NewControls(),
	}
}

func (c *Context) Update(window *opengl.Window) {
	c.Controls.Update(window)
}

func (c *Context) GetActiveState() State {
	return c.activeState
}

func (c *Context) SwapActiveState(newState State) State {
	oldState := c.activeState
	c.activeState = newState
	return oldState
}
