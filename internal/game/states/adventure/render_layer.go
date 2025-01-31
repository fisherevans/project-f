package adventure

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
)

type renderLayer struct {
	tiles [][]*resources.SpriteReference
}

func (r renderLayer) Render(target pixel.Target, cameraMatrix pixel.Matrix, from, to MapLocation) {
	for x := from.X; x <= to.X; x++ {
		for y := from.Y; y <= to.Y; y++ {
			spriteRef := r.tiles[x][y]
			if spriteRef == nil {
				continue
			}
			//spriteRef.Sprite.Draw(r.batch, cameraMatrix.Moved(pixel.V(float64(x), float64(y)).Scaled(resources.TileSizeF64)))
			spriteRef.Sprite.Draw(target, cameraMatrix.Moved(pixel.V(float64(x), float64(y)).Scaled(resources.TileSizeF64)))
		}
	}
}
