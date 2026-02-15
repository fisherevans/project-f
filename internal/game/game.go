package game

import (
	"math"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/shaders"
	"github.com/google/uuid"
)

const (
	GameWidth  = 240
	GameHeight = 160
)

var InstanceId = uuid.New().String()

type DebugArea int

const (
	AreaTopLeft DebugArea = iota
	AreaTopRight
	AreaBottomLeft
	AreaBottomRight
)

type State interface {
	ClearColor() pixel.RGBA
	OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64)
	OnEnter(data any)
	OnExit()
	HandleConsoleInput(string) bool
}

type BaseState struct{}

func (s BaseState) ClearColor() pixel.RGBA {
	return colors.Black.RGBA
}

func (s BaseState) OnEnter(any) {}

func (s BaseState) OnExit() {}

func (s BaseState) HandleConsoleInput(string) bool {
	return false
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
		0.8,                  // radius (covering corners)
		20,                   // swirl radians
		0.6,                  // falloff
		float32(math.Min(1, s.elapsedSeconds/s.durationSeconds)), // progress
	)
}
