package frames

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
)

type Instance struct {
	*resources.SpriteFrame
	name  string
	atlas *resources.Atlas
}

func New(name string, atlas *resources.Atlas) *Instance {
	return &Instance{
		SpriteFrame: resources.GetFrame(name),
		name:        name,
		atlas:       atlas,
	}
}

func (i *Instance) Draw(target pixel.Target, rect pixel.Rect, matrix pixel.Matrix, opts ...Opt) {
	// Split the rect into 9 sub-rectangles
	top := float64(i.CutMargin[resources.FrameTop])
	left := float64(i.CutMargin[resources.FrameLeft])
	bottom := float64(i.CutMargin[resources.FrameBottom])
	right := float64(i.CutMargin[resources.FrameRight])

	options := &frameOptions{
		color: pixel.RGBA{1, 1, 1, 1},
	}
	for _, opt := range opts {
		opt(options)
	}

	subRects := map[resources.FrameSide]pixel.Rect{
		resources.FrameTopLeft:     pixel.R(rect.Min.X, rect.Max.Y-top, rect.Min.X+left, rect.Max.Y),
		resources.FrameTop:         pixel.R(rect.Min.X+left, rect.Max.Y-top, rect.Max.X-right, rect.Max.Y),
		resources.FrameTopRight:    pixel.R(rect.Max.X-right, rect.Max.Y-top, rect.Max.X, rect.Max.Y),
		resources.FrameLeft:        pixel.R(rect.Min.X, rect.Min.Y+bottom, rect.Min.X+left, rect.Max.Y-top),
		resources.FrameMiddle:      pixel.R(rect.Min.X+left, rect.Min.Y+bottom, rect.Max.X-right, rect.Max.Y-top),
		resources.FrameRight:       pixel.R(rect.Max.X-right, rect.Min.Y+bottom, rect.Max.X, rect.Max.Y-top),
		resources.FrameBottomLeft:  pixel.R(rect.Min.X, rect.Min.Y, rect.Min.X+left, rect.Min.Y+bottom),
		resources.FrameBottom:      pixel.R(rect.Min.X+left, rect.Min.Y, rect.Max.X-right, rect.Min.Y+bottom),
		resources.FrameBottomRight: pixel.R(rect.Max.X-right, rect.Min.Y, rect.Max.X, rect.Min.Y+bottom),
	}

	// DrawColorMask() each sub-rectangl, options.colore
	for side, subRect := range subRects {
		sprite := i.atlas.GetFrameSprite(i.name, side)

		frameMode, ok := i.FrameModes[side]
		if !ok {
			frameMode = resources.FrameModeStretch
		}
		switch frameMode {
		case resources.FrameModeStretch:
			scaleAround := matrix.Project(subRect.Center())
			drawMatrix := matrix
			switch side {
			case resources.FrameTop, resources.FrameBottom:
				drawMatrix = matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(subRect.W()/sprite.Bounds().W(), 1))
			case resources.FrameLeft, resources.FrameRight:
				drawMatrix = matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(1, subRect.H()/sprite.Bounds().H()))
			case resources.FrameMiddle:
				drawMatrix = matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(subRect.W()/sprite.Bounds().W(), subRect.H()/sprite.Bounds().H()))
			default: // corners never scale
				drawMatrix = matrix.Moved(subRect.Center())
			}
			sprite.DrawColorMask(target, drawMatrix, options.color)
		case resources.FrameModeRepeat:
			drawRepeated(target, matrix, sprite, subRect, options)
		}
	}

}

type frameOptions struct {
	color pixel.RGBA
}

type Opt func(*frameOptions)

func WithColor(color pixel.RGBA) Opt {
	return func(o *frameOptions) {
		o.color = color
	}
}

func drawRepeated(target pixel.Target, matrix pixel.Matrix, sprite pixelutil.BoundedDrawable, rect pixel.Rect, options *frameOptions) {
	spriteWidth := sprite.Bounds().W()
	spriteHeight := sprite.Bounds().H()

	for x := rect.Min.X; x < rect.Max.X; x += spriteWidth {
		for y := rect.Min.Y; y < rect.Max.Y; y += spriteHeight {
			sprite.DrawColorMask(target, matrix.Moved(pixel.V(x+spriteWidth/2, y+spriteHeight/2)), options.color)
		}
	}
}
