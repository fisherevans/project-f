package util

import (
	"github.com/gopxl/pixel/v2"
	"math"
)

func Rounded(v pixel.Vec) pixel.Vec {
	return pixel.V(
		math.Round(v.X),
		math.Round(v.Y),
	)
}
