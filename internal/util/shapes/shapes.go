package shapes

import (
	"fisherevans.com/project/f/internal/resources"
	"fmt"
	"github.com/gopxl/pixel/v2"
)

type OriginLocation int

const (
	BottomLeft OriginLocation = iota
	TopLeft
	BottomRight
	TopRight
)

var sprite = resources.Sprites["2x2"]
var spriteW = sprite.Sprite.Frame().W()
var spriteH = sprite.Sprite.Frame().H()

func DrawRect(target pixel.Target, origin pixel.Vec, originLocation OriginLocation, width, height int, color pixel.RGBA) {
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
	default:
		panic(fmt.Sprintf("unknown origin location: %d", originLocation))
	}
	sprite.Sprite.DrawColorMask(target, matrix.Moved(origin), color)
}

func DrawRect2(target pixel.Target, inputMatrix pixel.Matrix, originLocation OriginLocation, width, height int, color pixel.RGBA) {
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
	default:
		panic(fmt.Sprintf("unknown origin location: %d", originLocation))
	}
	sprite.Sprite.DrawColorMask(target, matrix.Chained(inputMatrix), color)
}
