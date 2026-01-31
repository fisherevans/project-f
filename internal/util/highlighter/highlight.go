package highlighter

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type Target struct {
	Area      pixel.Rect
	Message   string
	BadgeText string
	Flip      bool
}

type Drawer struct {
	bgColor pixel.RGBA
	atlas   *resources.Atlas
	frame   *frames.Instance
	text    *textbox.Instance

	fromArea pixel.Rect

	target        Target
	targetContent *textbox.Content
	targetBadge   *badges.ButtonAction

	transitionElapsed float64
	transitionTime    float64
}

func NewDrawer(atlas *resources.Atlas, initialArea pixel.Rect, transitionTime float64) *Drawer {
	return &Drawer{
		bgColor: colors.WithAlpha(colors.Black.RGBA, 0.5),
		atlas:   atlas,
		frame:   frames.New("common/highlight_frame", atlas),
		text: textbox.NewInstance(atlas.GetFont("ffont"), tbcfg.Config{
			HAlignment: tbcfg.HAlignCenter,
			VAlignment: tbcfg.VAlignBottom,
			Foreground: colors.White.RGBA,
			Origin:     gfx.BottomCenter,
		}),

		fromArea: initialArea,
		target: Target{
			Area: initialArea,
		},
		transitionElapsed: transitionTime,
		transitionTime:    transitionTime,
	}
}

func (d *Drawer) SetTargetArea(target Target, doTransition bool) {
	d.targetContent = nil
	if target.Message != "" {
		d.targetContent = d.text.NewComplexContent("{+o:#000}" + target.Message)
	}
	d.targetBadge = nil
	if target.BadgeText != "" {
		d.targetBadge = badges.Using(d.atlas).ButtonAction("a", target.BadgeText, badges.ButtonStyleStandard).Flipped()
	}
	if doTransition {
		d.fromArea = d.currentArea()
		d.target = target
		d.transitionElapsed = 0
		return
	}
	d.fromArea = target.Area
	d.target = target
	d.transitionElapsed = d.transitionTime
}

func (h *Drawer) Render(target pixel.Target, timeDelta float64) {
	h.transitionElapsed += timeDelta
	progress := h.transitionProgress()
	area := lerpRect(h.fromArea, h.target.Area, progress)
	frameMin := area.Min.Sub(gfx.IVec(
		h.frame.CutMargin[resources.FrameLeft],
		h.frame.CutMargin[resources.FrameBottom],
	))
	frameMax := area.Max.Add(gfx.IVec(
		h.frame.CutMargin[resources.FrameRight],
		h.frame.CutMargin[resources.FrameTop],
	))
	frameRect := pixel.R(frameMin.X, frameMin.Y, frameMax.X, frameMax.Y)
	h.frame.Draw(target, frameRect, pixel.IM)
	bgBlocks := []struct {
		bottomLeft pixel.Vec
		width      int
		height     int
	}{
		{bottomLeft: pixel.ZV, width: game.GameWidth, height: int(frameMin.Y)},
		{bottomLeft: pixel.V(0, frameMax.Y), width: game.GameWidth, height: game.GameHeight - int(frameMax.Y)},
		{bottomLeft: pixel.V(0, frameMin.Y), width: int(frameMin.X), height: int(frameMax.Y - frameMin.Y)},
		{bottomLeft: pixel.V(frameMax.X, frameMin.Y), width: game.GameWidth - int(frameMax.X), height: int(frameMax.Y - frameMin.Y)},
	}
	for _, block := range bgBlocks {
		gfx.DrawRect(h.atlas, target, pixel.IM.Moved(block.bottomLeft), gfx.BottomLeft, block.width, block.height, h.bgColor)
	}
	if progress < 1 {
		return
	}
	if h.targetBadge != nil {
		origin := gfx.TopRight
		matrix := pixel.IM.Moved(pixel.V(frameMax.X+3, frameMin.Y+6))
		if h.target.Flip {
			origin = gfx.BottomRight
			matrix = pixel.IM.Moved(pixel.V(frameMax.X+3, frameMax.Y-6))
		}
		h.targetBadge.Render(target, matrix, origin)
	}
	if h.targetContent != nil {
		vAlign := tbcfg.VAlignBottom
		matrix := pixel.IM.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMax.Y+2))
		if h.target.Flip {
			vAlign = tbcfg.VAlignTop
			matrix = pixel.IM.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMin.Y-2))
		}
		h.text.Render(target, matrix, h.targetContent, tbcfg.VAligned(vAlign))
	}
}

func (h *Drawer) transitionProgress() float64 {
	if h.transitionTime <= 0 {
		return 1
	}
	progress := h.transitionElapsed / h.transitionTime
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}
	return interp.Smootherstep(progress)
}

func (h *Drawer) currentArea() pixel.Rect {
	return lerpRect(h.fromArea, h.target.Area, h.transitionProgress())
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func lerpVec(a, b pixel.Vec, t float64) pixel.Vec {
	return pixel.V(lerp(a.X, b.X, t), lerp(a.Y, b.Y, t))
}

func lerpRect(a, b pixel.Rect, t float64) pixel.Rect {
	min := lerpVec(a.Min, b.Min, t)
	max := lerpVec(a.Max, b.Max, t)
	return pixel.R(math.Round(min.X), math.Round(min.Y), math.Round(max.X), math.Round(max.Y))
}
