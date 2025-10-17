package gfx

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
)

func DrawRect(atlas *resources.Atlas, target pixel.Target, matrix pixel.Matrix, origin OriginLocation, width, height int, color pixel.RGBA) {
	sprite := atlas.GetSprite("2x2")
	spriteW := sprite.Bounds().H()
	spriteH := sprite.Bounds().W()
	fw := float64(width)
	fh := float64(height)
	scaleX := fw / spriteW
	scaleY := fh / spriteH
	renderMatrix := pixel.IM.ScaledXY(pixel.ZV, pixel.V(scaleX, scaleY))
	switch origin {
	case BottomLeft:
		renderMatrix = renderMatrix.Moved(pixel.V(fw/2, fh/2))
	case TopLeft:
		renderMatrix = renderMatrix.Moved(pixel.V(fw/2, -fh/2))
	case BottomRight:
		renderMatrix = renderMatrix.Moved(pixel.V(-fw/2, fh/2))
	case TopRight:
		renderMatrix = renderMatrix.Moved(pixel.V(-fw/2, -fh/2))
	case Centered: // do nothing
	case LeftCenter:
		renderMatrix = renderMatrix.Moved(pixel.V(fw/2, 0))
	case TopCenter:
		renderMatrix = renderMatrix.Moved(pixel.V(0, -fh/2))
	case RightCenter:
		renderMatrix = renderMatrix.Moved(pixel.V(-fw/2, 0))
	case BottomCenter:
		renderMatrix = renderMatrix.Moved(pixel.V(0, fh/2))
	default:
		panic(fmt.Sprintf("unknown origin location: %d", origin))
	}
	sprite.DrawColorMask(target, renderMatrix.Chained(matrix), color)
}
