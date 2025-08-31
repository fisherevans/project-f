package game

import (
	"image/color"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
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
	OnTick(ctx *Context, target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64)
}

type BaseState struct{}

func (s *BaseState) ClearColor() color.Color {
	return color.Black
}

type Context struct {
	DebugInfo

	activeState  State
	customShader AppliedShader

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

type SwirlShader struct {
	durationSeconds float64
	elapsedSeconds  float64
}

func NewSwirlShader(durationSeconds float64) *SwirlShader {
	return &SwirlShader{
		durationSeconds: durationSeconds,
	}
}

func (s *SwirlShader) Apply(shader shaders.Options, timeDelta float64) {
	s.elapsedSeconds += timeDelta
	shader.SetSwirlShader(
		mgl32.Vec2{0.5, 0.5}, // center
		30,                   // radius (full safe radius)
		20,                   // swirl radians
		0.6,                  // falloff
		float32(math.Min(1, s.elapsedSeconds/s.durationSeconds)), // progress
	)
}
