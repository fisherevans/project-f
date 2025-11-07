package adventure

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type renderLayer struct {
	tiles [][]pixelutil.BoundedDrawable
}

func (r renderLayer) Render(target pixel.Target, cameraDelta pixel.Vec, bounds MapBounds) {
	for x := bounds.MinX; x <= bounds.MaxX; x++ {
		for y := bounds.MinY; y <= bounds.MaxY; y++ {
			spriteRef := r.tiles[x][y]
			if spriteRef == nil {
				continue
			}
			renderDelta := cameraDelta.Add(pixel.V(float64(x), float64(y)).Scaled(resources.MapTileSize.Float()))
			roundedRenderDelta := pixel.Vec{
				X: math.Round(renderDelta.X),
				Y: math.Round(renderDelta.Y),
			}
			spriteRef.Draw(target, pixel.IM.Moved(roundedRenderDelta))
		}
	}
}

func normalizeRenderMoveDelta(vec pixel.Vec, tileSize resources.Pixels) pixel.Vec {
	return pixel.Vec{
		X: math.Round(vec.X * float64(tileSize)),
		Y: math.Round(vec.Y * float64(tileSize)),
	}
}
