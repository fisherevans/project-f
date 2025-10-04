package xenolog

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	screenWidth    = 206
	screenHeight   = 128
	screenClear    = colors.HexString("#0476d0")
	uiMask         = colors.HexString("#a6d8ff")
	uiMaskSelected = colors.HexString("#ffffff")

	spriteAnimech   = atlas.GetSprite("xenolog/select_animech")
	spritePrimortal = atlas.GetSprite("xenolog/select_primortal")
)

type Screen struct {
	canvas     *shaders.Canvas
	batch      *pixel.Batch
	selectLeft bool
	left       *selectBox
	right      *selectBox
}

func NewScreen() *Screen {
	s := &Screen{
		canvas:     shaders.NewCanvas(screenWidth, screenHeight),
		batch:      atlas.NewBatch(),
		selectLeft: true,
		left:       newSelectBox(spriteAnimech, "Animech", "Upgrade Available"),
		right:      newSelectBox(spritePrimortal, "Primortals", ""),
	}
	return s
}

func (s *Screen) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(screenWidth), float64(screenHeight))
}

func (s *Screen) OnTick(state *State, ctx *game.Context, targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
	center := pixel.IM.Moved(gfx.IVec(screenWidth/2, screenHeight/2))
	s.batch.Clear()

	if ctx.Controls.DPad().JustPressedDirection() == input.Left {
		s.selectLeft = true
	} else if ctx.Controls.DPad().JustPressedDirection() == input.Right {
		s.selectLeft = false
	}

	dx := math.Floor(float64(selectBoxWidth) * 0.6)
	s.left.render(ctx, center.Moved(pixel.V(-dx, 0)), s.batch, s.selectLeft, timeDelta)
	s.right.render(ctx, center.Moved(pixel.V(dx, 0)), s.batch, !s.selectLeft, timeDelta)

	s.canvas.Clear(screenClear)
	s.batch.Draw(s.canvas)
	s.canvas.Draw(target, targetMatrix)
}

var (
	selectBoxHeight = 82
	selectBoxWidth  = 74
	selectBoxFrame  = frames.New("xenolog/wide_frame", atlas, frames.WithRenderOrigin(gfx.Centered))

	selectBoxLabelMargin = 3
	selectBoxLabelText   = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))

	selectBoxSubLabelFlashSpeed = 1.0
	selectBoxSubLabelMargin     = 3
	selectBoxSubLabelText       = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.RenderFrom(gfx.TopCenter)))

	selectBoxArrow             = atlas.GetSprite("xenolog/select_arrow_down")
	selectBoxArrowMargin       = 2.0
	selectBoxArrowBounceHeight = 3.0
	selectBoxArrowBounceSpeed  = 1.0
)

type selectBox struct {
	sprite   pixelutil.BoundedDrawable
	label    *textbox.Content
	subLabel *textbox.Content

	elapsed float64
}

func (b *selectBox) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
}

func newSelectBox(sprite pixelutil.BoundedDrawable, label, subLabel string) *selectBox {
	s := &selectBox{
		sprite: sprite,
	}
	if label != "" {
		s.label = selectBoxLabelText.NewSimpleContent(label)
	}
	if subLabel != "" {
		s.subLabel = selectBoxSubLabelText.NewSimpleContent(subLabel)
	}
	return s
}

func (b *selectBox) render(ctx *game.Context, center pixel.Matrix, target pixel.Target, selected bool, timeDelta float64) {
	b.elapsed += timeDelta

	mask := uiMask
	if selected {
		mask = uiMaskSelected
	}

	frameRect := pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
	selectBoxFrame.Draw(target, frameRect, center, frames.WithColor(mask))

	bottomCenter := center.Moved(gfx.IVec(0, -selectBoxHeight/2))
	spriteCenter := bottomCenter.Moved(gfx.BottomCenter.Align(b.sprite))
	b.sprite.DrawColorMask(target, spriteCenter, mask)

	if b.label != nil {
		selectBoxLabelText.Render(ctx, target, center.Moved(gfx.IVec(0, selectBoxHeight/2-selectBoxLabelMargin)), b.label, tbcfg.Foreground(mask))
	}

	if b.subLabel != nil {
		subLabelMaskLerp := (math.Sin(b.elapsed*selectBoxSubLabelFlashSpeed*2.0*math.Pi) + 1.0) / 2.0
		subLabelMask := colors.Lerp(uiMask, colors.White.RGBA, subLabelMaskLerp)
		selectBoxSubLabelText.Render(ctx, target,
			center.Moved(gfx.IVec(0, -selectBoxHeight/2-selectBoxSubLabelMargin)),
			b.subLabel,
			tbcfg.Foreground(subLabelMask))
	}

	if selected {
		phase := b.elapsed * selectBoxArrowBounceSpeed * 2 * math.Pi
		dy := math.Abs(math.Sin(phase)) * selectBoxArrowBounceHeight
		arrowBottomCenter := center.Moved(gfx.IVec(0, selectBoxHeight/2)).
			Moved(pixel.V(0, selectBoxArrowMargin+dy))
		selectBoxArrow.Draw(target, arrowBottomCenter.Moved(gfx.BottomCenter.Align(selectBoxArrow)))
	}
}
