package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/shaders"
)

type spriteShader struct {
	canvas *shaders.Canvas
	batch  *pixel.Batch
}

func newSpriteShader() *spriteShader {
	s := &spriteShader{
		canvas: shaders.NewCanvas(screenWidth, screenHeight),
		batch:  atlas.NewBatch(),
	}
	s.canvas.SetXbitShaderWithThresholds(
		colorClear,
		[]shaders.ColorThreshold{
			{Threshold: 0.15, Color: colorText},
			{Threshold: 0.45, Color: colorHighlight},
		},
		0.01,
	)
	return s
}

func (s *spriteShader) Clear() {
	s.canvas.Clear(pixel.Alpha(0))
}

func (s *spriteShader) Target() pixel.Target {
	return s.batch
}

func (s *spriteShader) Render(target pixel.Target) {
	s.batch.Draw(s.canvas)
	s.canvas.Draw(target, pixel.IM.Moved(s.canvas.Bounds().Center()))
	s.batch.Clear()
}
