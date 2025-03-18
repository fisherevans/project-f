package gfx

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
	"math"
)

type OriginLocation int

const (
	Centered OriginLocation = iota
	BottomLeft
	TopLeft
	BottomRight
	TopRight
)

type Bounded interface {
	Bounds() pixel.Rect
}

func (l OriginLocation) Align(bounded Bounded) pixel.Vec {
	return l.AlignRect(bounded.Bounds())
}

func (l OriginLocation) AlignRect(bounds pixel.Rect) pixel.Vec {
	return l.AlignF64(bounds.W(), bounds.H())
}

func (l OriginLocation) AlignInt(w, h int) pixel.Vec {
	return l.AlignF64(float64(w), float64(h))
}

func (l OriginLocation) AlignF64(w, h float64) pixel.Vec {
	halfW, halfH := w/2, h/2
	switch l {
	case Centered:
		return pixel.ZV
	case TopLeft:
		return pixel.V(math.Floor(halfW), -math.Floor(halfH))
	case BottomLeft:
		return pixel.V(math.Floor(halfW), math.Ceil(halfH))
	case TopRight:
		return pixel.V(-math.Floor(halfW), -math.Floor(halfH))
	case BottomRight:
		return pixel.V(-math.Floor(halfW), math.Ceil(halfH))
	}
	panic(fmt.Sprintf("invalid origin location %d", l))
}
