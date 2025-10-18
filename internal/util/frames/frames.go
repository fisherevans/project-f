package frames

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type Instance struct {
	*resources.SpriteFrame
	name  string
	atlas *resources.Atlas
	opts  []Opt
}

func New(name string, atlas *resources.Atlas, opts ...Opt) *Instance {
	return &Instance{
		SpriteFrame: resources.GetFrame(name),
		name:        name,
		atlas:       atlas,
		opts:        opts,
	}
}

func (i *Instance) Draw(target pixel.Target, rect pixel.Rect, matrix pixel.Matrix, opts ...Opt) {
	// TODO add check that we're drawing to batch and warn if it's not
	// Split the rect into 9 sub-rectangles
	top := float64(i.CutMargin[resources.FrameTop])
	left := float64(i.CutMargin[resources.FrameLeft])
	bottom := float64(i.CutMargin[resources.FrameBottom])
	right := float64(i.CutMargin[resources.FrameRight])

	options := &frameOptions{
		color:        pixel.RGBA{1, 1, 1, 1},
		renderOrigin: gfx.BottomLeft,
	}
	for _, opt := range append(i.opts, opts...) {
		opt(options)
	}

	switch options.renderOrigin {
	case gfx.TopLeft:
		matrix = matrix.Moved(gfx.IVec(0, int(-rect.H())))
	case gfx.TopRight:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()), int(-rect.H())))
	case gfx.BottomLeft:
		matrix = matrix.Moved(gfx.IVec(0, 0))
	case gfx.BottomRight:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()), 0))
	case gfx.Centered:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()/2), int(-rect.H()/2)))
	case gfx.LeftCenter:
		matrix = matrix.Moved(gfx.IVec(0, int(-rect.H()/2)))
	case gfx.TopCenter:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()/2), int(-rect.H())))
	case gfx.RightCenter:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()), int(-rect.H()/2)))
	case gfx.BottomCenter:
		matrix = matrix.Moved(gfx.IVec(int(-rect.W()/2), 0))
	default:
		panic("invalid origin")
	}

	type frameSide struct {
		side resources.FrameSide
		rect pixel.Rect
	}
	frameSides := []frameSide{
		{side: resources.FrameTopLeft, rect: pixel.R(rect.Min.X, rect.Max.Y-top, rect.Min.X+left, rect.Max.Y)},
		{side: resources.FrameTop, rect: pixel.R(rect.Min.X+left, rect.Max.Y-top, rect.Max.X-right, rect.Max.Y)},
		{side: resources.FrameTopRight, rect: pixel.R(rect.Max.X-right, rect.Max.Y-top, rect.Max.X, rect.Max.Y)},
		{side: resources.FrameLeft, rect: pixel.R(rect.Min.X, rect.Min.Y+bottom, rect.Min.X+left, rect.Max.Y-top)},
		{side: resources.FrameMiddle, rect: pixel.R(rect.Min.X+left, rect.Min.Y+bottom, rect.Max.X-right, rect.Max.Y-top)},
		{side: resources.FrameRight, rect: pixel.R(rect.Max.X-right, rect.Min.Y+bottom, rect.Max.X, rect.Max.Y-top)},
		{side: resources.FrameBottomLeft, rect: pixel.R(rect.Min.X, rect.Min.Y, rect.Min.X+left, rect.Min.Y+bottom)},
		{side: resources.FrameBottom, rect: pixel.R(rect.Min.X+left, rect.Min.Y, rect.Max.X-right, rect.Min.Y+bottom)},
		{side: resources.FrameBottomRight, rect: pixel.R(rect.Max.X-right, rect.Min.Y, rect.Max.X, rect.Min.Y+bottom)},
	}

	// DrawColorMask() each sub-rectangle
	drawCount := 0
	for _, s := range frameSides {
		side := s.side
		subRect := s.rect
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
			drawCount++
		case resources.FrameModeRepeat:
			drawRepeated(target, matrix, sprite, subRect, options)
			drawCount++
		}
	}
}

type frameOptions struct {
	color        pixel.RGBA
	renderOrigin gfx.OriginLocation
}

type Opt func(*frameOptions)

func WithColor(color pixel.RGBA) Opt {
	return func(o *frameOptions) {
		o.color = color
	}
}

func WithRenderOrigin(origin gfx.OriginLocation) Opt {
	return func(o *frameOptions) {
		o.renderOrigin = origin
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
