package gfx

import (
	"github.com/gopxl/pixel/v2"
)

type OriginLocation int

const (
	Centered OriginLocation = iota
	BottomLeft
	TopLeft
	BottomRight
	TopRight
	LeftCenter
	TopCenter
	RightCenter
	BottomCenter
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
	return l.AlignFrom(Centered, w, h)
}

func (l OriginLocation) AlignFrom(original OriginLocation, w, h float64) pixel.Vec {
	halfW, halfH := w/2, h/2
	var delta pixel.Vec
	// move from original to bottom left
	switch original {
	case Centered:
		delta = pixel.V(halfW, halfH)
	case TopLeft:
		delta = pixel.V(0, h)
	case BottomLeft:
		// already bottom left
	case TopRight:
		delta = pixel.V(-w, h)
	case BottomRight:
		delta = pixel.V(-w, 0)
	case LeftCenter:
		delta = pixel.V(0, halfH)
	case TopCenter:
		delta = pixel.V(-halfW, h)
	case RightCenter:
		delta = pixel.V(-w, halfH)
	case BottomCenter:
		delta = pixel.V(-halfW, 0)
	}
	// now move from bottom left to target (l)
	switch l {
	case Centered:
		delta = delta.Add(pixel.V(-halfW, -halfH))
	case TopLeft:
		delta = delta.Add(pixel.V(0, -h))
	case BottomLeft:
		// already bottom left
	case TopRight:
		delta = delta.Add(pixel.V(-w, -h))
	case BottomRight:
		delta = delta.Add(pixel.V(-w, 0))
	case LeftCenter:
		delta = delta.Add(pixel.V(0, -halfH))
	case TopCenter:
		delta = delta.Add(pixel.V(-halfW, -h))
	case RightCenter:
		delta = delta.Add(pixel.V(-w, -halfH))
	case BottomCenter:
		delta = delta.Add(pixel.V(-halfW, 0))
	}
	return delta
}
