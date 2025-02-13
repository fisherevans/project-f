package frames

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
)

func Draw(target pixel.Target, frame *resources.SpriteFrame, rect pixel.Rect, matrix pixel.Matrix) {
	// Split the rect into 9 sub-rectangles
	top := float64(frame.CutMargin[resources.FrameTop])
	left := float64(frame.CutMargin[resources.FrameLeft])
	bottom := float64(frame.CutMargin[resources.FrameBottom])
	right := float64(frame.CutMargin[resources.FrameRight])

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

	// Draw each sub-rectangle
	for side, subRect := range subRects {
		sprite := frame.Sprites[side]
		if sprite == nil {
			continue
		}

		frameMode, ok := frame.FrameModes[side]
		if !ok {
			frameMode = resources.FrameModeStretch
		}
		switch frameMode {
		case resources.FrameModeStretch:
			scaleAround := matrix.Project(subRect.Center())
			switch side {
			case resources.FrameTop, resources.FrameBottom:
				sprite.Sprite.Draw(target, matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(subRect.W()/sprite.Bounds.W(), 1)))
			case resources.FrameLeft, resources.FrameRight:
				sprite.Sprite.Draw(target, matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(1, subRect.H()/sprite.Bounds.H())))
			case resources.FrameMiddle:
				sprite.Sprite.Draw(target, matrix.Moved(subRect.Center()).ScaledXY(scaleAround, pixel.V(subRect.W()/sprite.Bounds.W(), subRect.H()/sprite.Bounds.H())))
			default: // corners never scale
				sprite.Sprite.Draw(target, matrix.Moved(subRect.Center()))
			}
		case resources.FrameModeRepeat:
			drawRepeated(target, matrix, sprite, subRect)
		}
	}

}

func drawRepeated(target pixel.Target, matrix pixel.Matrix, sprite *resources.SpriteReference, rect pixel.Rect) {
	spriteWidth := sprite.Bounds.W()
	spriteHeight := sprite.Bounds.H()

	for x := rect.Min.X; x < rect.Max.X; x += spriteWidth {
		for y := rect.Min.Y; y < rect.Max.Y; y += spriteHeight {
			sprite.Sprite.Draw(target, matrix.Moved(pixel.V(x+spriteWidth/2, y+spriteHeight/2)))
		}
	}
}
