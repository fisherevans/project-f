package adventure

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
)

type renderLayer struct {
	tiles [][]*resources.SpriteReference
}

func (r renderLayer) Render(target pixel.Target, cameraMatrix pixel.Matrix, bounds MapBounds) {
	for x := bounds.MinX; x <= bounds.MaxX; x++ {
		for y := bounds.MinY; y <= bounds.MaxY; y++ {
			spriteRef := r.tiles[x][y]
			if spriteRef == nil {
				continue
			}
			spriteRef.Sprite.Draw(target, cameraMatrix.Moved(pixel.V(float64(x), float64(y)).Scaled(resources.TileSizeF64)))
		}
	}
}
