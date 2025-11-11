package game

import (
	"image/color"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"

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
	OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64)
	OnEnter()
	OnExit()
}

type BaseState struct{}

func (s BaseState) ClearColor() color.Color {
	return color.Black
}

func (s BaseState) OnEnter() {}

func (s BaseState) OnExit() {}

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
