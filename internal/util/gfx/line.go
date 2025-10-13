package gfx

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
)

func DrawLine(batch pixel.Target, atlas *resources.Atlas, from, to pixel.Vec, thickness float64, mask pixel.RGBA) {
	sprite := atlas.GetSprite("1x1")
	delta := to.Sub(from)
	length := delta.Len()
	if length == 0 {
		return
	}

	angle := math.Atan2(delta.Y, delta.X)
	mat := pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(length, thickness)). // stretch 1×1 pixel
		Rotated(pixel.ZV, angle).
		Moved(from.Add(to).Scaled(0.5)) // move to midpoint

	sprite.DrawColorMask(batch, mat, mask)
}
