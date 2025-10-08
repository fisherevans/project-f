package xenolog

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/game/shaders/bloom"
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

var (
	smallTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))

	mediumTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))
)

type Screen struct {
	canvas     *shaders.Canvas
	batch      *pixel.Batch
	bloom      *bloom.Helper
	selectLeft bool
	left       *selectBox
	right      *selectBox
	isSelected bool

	animechMenu    *animechMenu
	primortalsMenu *primortalsMenu
}

func NewScreen(ctx *game.Context) *Screen {
	s := &Screen{
		canvas: shaders.NewCanvas(screenWidth, screenHeight),
		batch:  atlas.NewBatch(),
		bloom: bloom.NewHelper(
			screenWidth, screenHeight,
			bloom.DefaultBrightnessConfig(),
			bloom.DefaultBlurConfig(),
			bloom.DefaultBlendConfig().WithIntensity(0.5)),
		selectLeft:     true,
		left:           newSelectBox(spriteAnimech, "Animech", ""),
		right:          newSelectBox(spritePrimortal, "Primortals", ""),
		animechMenu:    newAnimechMenu(ctx),
		primortalsMenu: newPrimortalsMenu(ctx),
	}
	s.refreshState(ctx)
	return s
}

func (s *Screen) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(screenWidth), float64(screenHeight))
}

func (s *Screen) OnTick(state *State, ctx *game.Context, targetMatrix pixel.Matrix, target pixel.Target, timeDelta float64) {
	s.batch.Clear()

	if s.isSelected {
		if s.selectLeft {
			s.animechMenu.RenderAnimech(s, ctx, s.batch, timeDelta)
		} else {
			s.primortalsMenu.RenderPrimortals(s, ctx, s.batch, timeDelta)
		}
	} else {
		s.RenderMenuSelection(ctx, s.batch, timeDelta)
	}

	s.canvas.Clear(screenClear)
	s.batch.Draw(s.canvas)
	s.canvas.Draw(target, targetMatrix)

	bloomed := s.bloom.ApplyBloom(s.canvas)
	bloomed.Draw(target, targetMatrix)
}

func (s *Screen) RenderMenuSelection(ctx *game.Context, target *pixel.Batch, timeDelta float64) {
	center := pixel.IM.Moved(gfx.IVec(screenWidth/2, screenHeight/2))
	if ctx.Controls.DPad().JustPressedDirection() == input.Left {
		s.selectLeft = true
	} else if ctx.Controls.DPad().JustPressedDirection() == input.Right {
		s.selectLeft = false
	}
	if ctx.Controls.ButtonA().JustPressed() {
		s.isSelected = true
	}

	dx := math.Floor(float64(selectBoxWidth) * 0.6)
	s.left.render(ctx, center.Moved(pixel.V(-dx, 0)), target, s.selectLeft, timeDelta)
	s.right.render(ctx, center.Moved(pixel.V(dx, 0)), target, !s.selectLeft, timeDelta)
}

func (s *Screen) returnToMenu(ctx *game.Context) {
	s.isSelected = false
	s.refreshState(ctx)
}

func (s *Screen) refreshState(ctx *game.Context) {
	s.left.subLabel = ""
	if s.animechMenu.isUpgradeAvailable(ctx) {
		s.left.subLabel = "Upgrade Available"
	}

	s.right.subLabel = ""
	if s.primortalsMenu.isUpgradeAvailable(ctx) {
		s.right.subLabel = "Upgrade Available"
	}
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
	label    string
	subLabel string

	elapsed float64
}

func (b *selectBox) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
}

func newSelectBox(sprite pixelutil.BoundedDrawable, label, subLabel string) *selectBox {
	s := &selectBox{
		sprite:   sprite,
		label:    label,
		subLabel: subLabel,
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

	if b.label != "" {
		c := selectBoxLabelText.NewSimpleContent(b.label)
		selectBoxLabelText.Render(ctx, target, center.Moved(gfx.IVec(0, selectBoxHeight/2-selectBoxLabelMargin)), c, tbcfg.Foreground(mask))
	}

	if b.subLabel != "" {
		subLabelMask := flash(b.elapsed, selectBoxSubLabelFlashSpeed, uiMask, colors.White.RGBA)
		c := selectBoxSubLabelText.NewSimpleContent(b.subLabel)
		selectBoxSubLabelText.Render(ctx, target,
			center.Moved(gfx.IVec(0, -selectBoxHeight/2-selectBoxSubLabelMargin)),
			c,
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

func flash(elapsed, speed float64, from, to pixel.RGBA) pixel.RGBA {
	lerp := (math.Sin(elapsed*speed*2.0*math.Pi) + 1.0) / 2.0
	mask := colors.Lerp(from, to, lerp)
	return mask
}
