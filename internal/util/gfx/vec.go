package gfx

import "github.com/gopxl/pixel/v2"

func IVec(x, y int) pixel.Vec {
	return pixel.V(float64(x), float64(y))
}
