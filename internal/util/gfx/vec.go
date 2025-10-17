package gfx

import "github.com/gopxl/pixel/v2"

func IVec(x, y int) pixel.Vec {
	return pixel.V(float64(x), float64(y))
}

func Moved(x, y int) pixel.Matrix {
	return pixel.IM.Moved(IVec(x, y))
}
