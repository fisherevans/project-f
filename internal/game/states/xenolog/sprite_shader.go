package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type spriteShader struct {
	canvas *shaders.Canvas
}

func newSpriteShader() *spriteShader {
	s := &spriteShader{
		canvas: shaders.NewCanvas(screenWidth, screenHeight),
	}
	s.canvas.SetXbitShaderWithThresholds(
		screenClear,
		[]shaders.ColorThreshold{
			{Threshold: 0.15, Color: maskText},
			{Threshold: 0.5, Color: maskTextHighlight},
		},
		0.01,
	)
	return s
}

func (s *spriteShader) Clear() {
	s.canvas.Clear(pixel.Alpha(0))
}

func (s *spriteShader) DrawSprite(drawable pixelutil.BoundedDrawable, matrix pixel.Matrix) {
	drawable.Draw(s.canvas, matrix)
}

func (s *spriteShader) Render(target pixel.Target) {
	s.canvas.Draw(target, pixel.IM.Moved(s.canvas.Bounds().Center()))
}
