package gfx

import (
	"github.com/gopxl/pixel/v2"
)

func R(width int, height int) pixel.Rect {
	return pixel.R(0, 0, float64(width), float64(height))
}
