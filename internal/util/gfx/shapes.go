package gfx

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
)

func DrawRect(atlas *resources.Atlas, target pixel.Target, inputMatrix pixel.Matrix, originLocation OriginLocation, width, height int, color pixel.RGBA) {
	sprite := atlas.GetSprite("2x2")
	spriteW := sprite.Bounds().H()
	spriteH := sprite.Bounds().W()
	fw := float64(width)
	fh := float64(height)
	scaleX := fw / spriteW
	scaleY := fh / spriteH
	matrix := pixel.IM.ScaledXY(pixel.ZV, pixel.V(scaleX, scaleY))
	switch originLocation {
	case BottomLeft:
		matrix = matrix.Moved(pixel.V(fw/2, fh/2))
	case TopLeft:
		matrix = matrix.Moved(pixel.V(fw/2, -fh/2))
	case BottomRight:
		matrix = matrix.Moved(pixel.V(-fw/2, fh/2))
	case TopRight:
		matrix = matrix.Moved(pixel.V(-fw/2, -fh/2))
	case Centered: // do nothing
	case LeftCenter:
		matrix = matrix.Moved(pixel.V(fw/2, 0))
	case TopCenter:
		matrix = matrix.Moved(pixel.V(0, -fh/2))
	case RightCenter:
		matrix = matrix.Moved(pixel.V(-fw/2, 0))
	case BottomCenter:
		matrix = matrix.Moved(pixel.V(0, fh/2))
	default:
		panic(fmt.Sprintf("unknown origin location: %d", originLocation))
	}
	sprite.DrawColorMask(target, matrix.Chained(inputMatrix), color)
}
