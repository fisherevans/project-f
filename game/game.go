package game

import (
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
	OnTick(ctx *Context, win *opengl.Window, canvas *opengl.Canvas, timeDelta float64)
}
type Context struct {
	activeState State
	debugLines  map[DebugArea][]string

	CanvasScale         float64
	CanvasMousePosition pixel.Vec
	MouseInCanvas       bool
}

func NewContext(initialActiveState State) *Context {
	return &Context{
		activeState: initialActiveState,
		CanvasScale: 1.0,
	}
}

func (c *Context) GetActiveState() State {
	return c.activeState
}

func (c *Context) SwapActiveState(newState State) State {
	oldState := c.activeState
	c.activeState = newState
	return oldState
}
