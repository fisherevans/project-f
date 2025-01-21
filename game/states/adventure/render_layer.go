package adventure

import (
	"fisherevans.com/project/f/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type renderLayer struct {
	tiles [][]*resources.SpriteReference
}

func (r renderLayer) Render(canvas *opengl.Canvas, cameraMatrix pixel.Matrix, from, to MapLocation) {
	batch := pixel.NewBatch(&pixel.TrianglesData{}, resources.SpriteAtlas)
	for x := from.X; x <= to.X; x++ {
		for y := from.Y; y <= to.Y; y++ {
			spriteRef := r.tiles[x][y]
			if spriteRef == nil {
				continue
			}
			spriteRef.Sprite.Draw(batch, cameraMatrix.Moved(pixel.V(float64(x), float64(y)).Scaled(resources.TileSizeF64)))
		}
	}
	batch.Draw(canvas)
}
