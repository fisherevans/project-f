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

/*
TODO:
- add dismiss state + animation + initial intro area
- add config:
  - width
  - height
  - im
*/

type BadgePlacement int

const (
	BadgeInTopRight BadgePlacement = iota
	BadgeInBottomRight
	BadgeOnTopMiddle
	BadgeOnBottomMiddle
)

type TargetBadge struct {
	Label     string
	Placement BadgePlacement
}

func NewBadge(placement BadgePlacement) *TargetBadge {
	return &TargetBadge{
		Label:     "Next",
		Placement: placement,
	}
}

func (b *TargetBadge) WithLabel(label string) *TargetBadge {
	b.Label = label
	return b
}

type MessagePlacement int

const (
	MessageOnBottom MessagePlacement = iota
	MessageOnTop
	MessageOnLeft
	MessageOnRight
)

type TargetMessage struct {
	Text          string
	Placement     MessagePlacement
	AutoWrapWidth int
}

func NewMessage(text string, placement MessagePlacement) *TargetMessage {
	return &TargetMessage{
		Text:      text,
		Placement: placement,
	}
}

func (m *TargetMessage) Wrapped(width int) *TargetMessage {
	m.AutoWrapWidth = width
	return m
}

type Target struct {
	Area    pixel.Rect
	Padding bool

	Message *TargetMessage
	Badge   *TargetBadge

	dismiss bool
}

func NewTarget(area pixel.Rect) Target {
	return Target{
		Area:    area,
		Padding: true,
	}
}

func (t Target) NoPadding() Target {
	t.Padding = false
	return t
}

func (t Target) WithBadge(badge *TargetBadge) Target {
	t.Badge = badge
	return t
}

func (t Target) WithMessage(message *TargetMessage) Target {
	t.Message = message
	return t
}

type Opt func(drawer *Drawer)

func WithMatrix(m pixel.Matrix) func(drawer *Drawer) {
	return func(drawer *Drawer) {
		drawer.matrix = m
	}
}

func WithScreenSize(width, height int) func(drawer *Drawer) {
	return func(drawer *Drawer) {
		drawer.screenWidth = width
		drawer.screenHeight = height
	}
}

type Drawer struct {
	matrix                    pixel.Matrix
	screenWidth, screenHeight int

	bgColor pixel.RGBA
	atlas   *resources.Atlas
	frame   *frames.Instance
	text    *textbox.Instance

	fromArea pixel.Rect

	target            Target
	targetContent     *textbox.Content
	targetContentOpts []tbcfg.ConfigOpt
	targetBadge       *badges.ButtonAction

	transitionElapsed float64
	transitionTime    float64
}

func NewDrawer(atlas *resources.Atlas, font string) *Drawer {
	d := &Drawer{
		matrix:       pixel.IM,
		screenWidth:  game.GameWidth,
		screenHeight: game.GameHeight,
		bgColor:      colors.WithAlpha(colors.Black.RGBA, 0.5),
		atlas:        atlas,
		frame:        frames.New("common/highlight_frame", atlas),
		text: textbox.NewInstance(atlas.GetFont(font), tbcfg.NewConfig(game.GameWidth, game.GameHeight,
			tbcfg.Foreground(colors.White.RGBA),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
			tbcfg.RenderFrom(gfx.BottomCenter))),
		transitionElapsed: 0.25,
		transitionTime:    0.25,
	}
	d.Dismiss(false)
	return d
}

func (d *Drawer) SetTargetArea(target Target, doTransition bool) {
	d.targetContent = nil
	d.targetContentOpts = nil
	if target.Message != nil {
		var opts []textbox.ContentOpt
		if target.Message.AutoWrapWidth > 0 {
			opts = append(opts, textbox.WithAutoWrapWidth(target.Message.AutoWrapWidth))
		}
		switch target.Message.Placement {
		case MessageOnLeft:
			d.targetContentOpts = append(d.targetContentOpts, tbcfg.HAligned(tbcfg.HAlignRight))
		case MessageOnRight:
			d.targetContentOpts = append(d.targetContentOpts, tbcfg.HAligned(tbcfg.HAlignLeft))
		}
		d.targetContent = d.text.NewComplexContent("{+o:#000}"+target.Message.Text, opts...)
	}
	d.targetBadge = nil
	if target.Badge != nil {
		d.targetBadge = badges.Using(d.atlas).ButtonAction("a", target.Badge.Label, badges.ButtonStyleStandard).Flipped()
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

func (d *Drawer) Dismiss(doTransition bool) {
	d.SetTargetArea(Target{
		Area:    pixel.R(0, 0, float64(d.screenWidth), float64(d.screenHeight)),
		dismiss: true,
	}, doTransition)
}

func (d *Drawer) IsDismissed() bool {
	if !d.IsTransitioned() {
		return false
	}
	return d.target.dismiss
}

func (d *Drawer) Render(target pixel.Target, timeDelta float64) {
	d.transitionElapsed += timeDelta
	progress := d.transitionProgress()
	mask, bgColor := colors.White.RGBA, d.bgColor
	if d.target.dismiss {
		mask = colors.WithAlpha(mask, 1.0-progress)
		bgColor = colors.LayerAlpha(bgColor, 1.0-progress)
	}
	area := lerpRect(d.fromArea, d.target.Area, progress)
	frameMin := area.Min
	frameMax := area.Max
	if d.target.Padding {
		frameMin = frameMin.Sub(gfx.IVec(
			d.frame.CutMargin[resources.FrameLeft],
			d.frame.CutMargin[resources.FrameBottom],
		))
		frameMax = frameMax.Add(gfx.IVec(
			d.frame.CutMargin[resources.FrameRight],
			d.frame.CutMargin[resources.FrameTop],
		))
	}
	frameRect := pixel.R(frameMin.X, frameMin.Y, frameMax.X, frameMax.Y)
	d.frame.Draw(target, frameRect, d.matrix, frames.WithColor(mask))
	bgBlocks := []struct {
		bottomLeft pixel.Vec
		width      int
		height     int
	}{
		{bottomLeft: pixel.ZV, width: d.screenWidth, height: int(frameMin.Y)},
		{bottomLeft: pixel.V(0, frameMax.Y), width: d.screenWidth, height: d.screenHeight - int(frameMax.Y)},
		{bottomLeft: pixel.V(0, frameMin.Y), width: int(frameMin.X), height: int(frameMax.Y - frameMin.Y)},
		{bottomLeft: pixel.V(frameMax.X, frameMin.Y), width: d.screenWidth - int(frameMax.X), height: int(frameMax.Y - frameMin.Y)},
	}
	for _, block := range bgBlocks {
		gfx.DrawRect(d.atlas, target, d.matrix.Moved(block.bottomLeft), gfx.BottomLeft, block.width, block.height, bgColor)
	}

	var messageMatrix pixel.Matrix
	var messageOrigin gfx.OriginLocation
	var messageVAlign tbcfg.VAlignment
	if d.target.Message != nil {
		switch d.target.Message.Placement {
		case MessageOnTop:
			messageMatrix = d.matrix.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMax.Y+1))
			messageOrigin = gfx.BottomCenter
			messageVAlign = tbcfg.VAlignBottom
		case MessageOnBottom:
			messageMatrix = d.matrix.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMin.Y-1))
			messageOrigin = gfx.TopCenter
			messageVAlign = tbcfg.VAlignTop
		case MessageOnLeft:
			messageMatrix = d.matrix.Moved(pixel.V(frameMin.X-2, (frameMax.Y+frameMin.Y)/2))
			messageOrigin = gfx.RightCenter
			messageVAlign = tbcfg.VAlignMiddle
		case MessageOnRight:
			messageMatrix = d.matrix.Moved(pixel.V(frameMax.X+2, (frameMax.Y+frameMin.Y)/2))
			messageOrigin = gfx.LeftCenter
			messageVAlign = tbcfg.VAlignMiddle
		}
	}

	var badgeMatrix pixel.Matrix
	var badgeOrigin gfx.OriginLocation
	if d.target.Badge != nil {
		switch d.target.Badge.Placement {
		case BadgeInTopRight:
			badgeMatrix = d.matrix.Moved(pixel.V(frameMax.X+3, frameMax.Y-6))
			badgeOrigin = gfx.BottomRight
		case BadgeInBottomRight:
			badgeMatrix = d.matrix.Moved(pixel.V(frameMax.X+3, frameMin.Y+6))
			badgeOrigin = gfx.TopRight
		case BadgeOnBottomMiddle:
			badgeMatrix = d.matrix.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMin.Y+6))
			badgeOrigin = gfx.TopCenter
		case BadgeOnTopMiddle:
			badgeMatrix = d.matrix.Moved(pixel.V((frameMin.X+frameMax.X)/2, frameMax.Y-6))
			badgeOrigin = gfx.BottomCenter
		}
	}

	badgeStart := 0.75
	if d.targetBadge != nil && progress > badgeStart {
		badgeProgress := (progress - badgeStart) / (1.0 - badgeStart)
		badgeMask := colors.WithAlpha(colors.White.RGBA, badgeProgress)
		d.targetBadge.RenderWithColorMask(target, badgeMatrix, badgeOrigin, badgeMask)
	}

	if d.targetContent != nil {
		contentMask := colors.WithAlpha(colors.White.RGBA, progress)
		d.targetContent.Render(target, messageMatrix, append(
			d.targetContentOpts,
			tbcfg.RenderFrom(messageOrigin),
			tbcfg.VAligned(messageVAlign),
			tbcfg.ColorMask(contentMask),
		)...)
	}
}

func (d *Drawer) IsTransitioned() bool {
	return d.transitionElapsed >= d.transitionTime
}

func (d *Drawer) transitionProgress() float64 {
	if d.transitionTime <= 0 {
		return 1
	}
	progress := d.transitionElapsed / d.transitionTime
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}
	return interp.Smootherstep(progress)
}

func (d *Drawer) currentArea() pixel.Rect {
	return lerpRect(d.fromArea, d.target.Area, d.transitionProgress())
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
