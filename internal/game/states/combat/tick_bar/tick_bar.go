package tick_bar

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/sprites"
)

type Config struct {
	Width       int
	TickSpacing int
}

type Renderer struct {
	atlas  *resources.Atlas
	config Config

	activeFrame  *frames.Instance
	pendingFrame *frames.Instance

	tickBubbleDisplayVBar    pixelutil.BoundedDrawable
	tickBubbleDisplaySpecial pixelutil.BoundedDrawable
	tickBubbleDisplayNormal  pixelutil.BoundedDrawable
	tickBubbleDisplayNone    pixelutil.BoundedDrawable

	tickBubbleOverlayInterrupt pixelutil.BoundedDrawable

	tickBarStanceBoxPending pixelutil.BoundedDrawable
	tickBarStanceBoxActive  pixelutil.BoundedDrawable

	stanceIcons map[rpg.CombatStance]pixelutil.BoundedDrawable
}

func NewRenderer(atlas *resources.Atlas, config Config) *Renderer {
	r := &Renderer{
		atlas:  atlas,
		config: config,
	}

	r.activeFrame = frames.New("combat/tick_bar/skill_active_frame", r.atlas)
	r.pendingFrame = frames.New("combat/tick_bar/skill_pending_frame", r.atlas)

	r.tickBubbleDisplayNone = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 1, 1)
	r.tickBubbleDisplayNormal = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 2, 1)
	r.tickBubbleDisplaySpecial = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 3, 1)
	r.tickBubbleDisplayVBar = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 4, 1)

	r.tickBubbleOverlayInterrupt = atlas.GetTilesheetSprite("combat/tick_bar/bubbles_overlay", 1, 1)

	r.tickBarStanceBoxActive = atlas.GetSprite("combat/tick_bar/skill_active_stance_box")
	r.tickBarStanceBoxPending = atlas.GetSprite("combat/tick_bar/skill_pending_stance_box")

	r.stanceIcons = sprites.StanceIcons(atlas)

	return r
}

type DrawOptions struct {
	interruptedAt int
	mask          pixel.RGBA
	active        bool
	alpha         float64
	skillProgress float64
	flip          bool
}

func NewDrawOptions() DrawOptions {
	return DrawOptions{
		interruptedAt: -1,
		mask:          colors.White.RGBA,
		active:        true,
		alpha:         1.0,
		skillProgress: 0,
		flip:          false,
	}
}

func (o DrawOptions) InterruptedAt(i int) DrawOptions {
	o.interruptedAt = i
	return o
}

func (o DrawOptions) Mask(m pixel.RGBA) DrawOptions {
	o.mask = m
	return o
}

func (o DrawOptions) Active(a bool) DrawOptions {
	o.active = a
	return o
}

func (o DrawOptions) Alpha(a float64) DrawOptions {
	o.alpha = a
	return o
}

func (o DrawOptions) SkillProgress(p float64) DrawOptions {
	o.skillProgress = p
	return o
}

func (o DrawOptions) Flip(f bool) DrawOptions {
	o.flip = f
	return o
}

type DrawResult struct {
	TickDotCenterMatrices []pixel.Matrix
}

func (r *Renderer) Draw(target pixel.Target, matrixTopMiddle pixel.Matrix, skill *rpg.Skill, opt DrawOptions) DrawResult {
	result := DrawResult{}
	if skill == nil {
		return result
	}
	mask := colors.WithAlpha(opt.mask, opt.alpha)
	if opt.interruptedAt >= 0 {
		mask = colors.ScaleColor(mask, 0.5)
	}
	rect := pixel.R(0, 0,
		float64(r.config.Width),
		float64(r.config.TickSpacing*(skill.Duration()+1)))
	matrixBottomLeft := matrixTopMiddle.Moved(pixel.V(-rect.W()/2, -rect.H()))

	frame := r.pendingFrame
	if opt.active {
		frame = r.activeFrame
	}
	frame.Draw(target, rect, matrixBottomLeft, frames.WithColor(mask))

	var postRenders []func() // used to render on top of the tick bar, after the dots/lines are rendered - mostly for the stance box + icons
	for i := 0; i <= skill.Duration(); i++ {
		lastStance := rpg.TickStanceNone
		if i > 0 {
			lastStance = skill.Ticks[i-1].StanceType
		}
		tick := skill.Ticks[i]
		inStance := tick.StanceType != rpg.TickStanceNone

		tickSpriteCenterMatrix := matrixBottomLeft.Moved(pixel.V(float64(r.config.Width/2), float64(r.config.TickSpacing/2+r.config.TickSpacing*(skill.Duration()-i))))
		result.TickDotCenterMatrices = append(result.TickDotCenterMatrices, tickSpriteCenterMatrix)

		// if we're entering a stance, draw the full VBar, box, and icon
		if tick.StanceType != lastStance && tick.StanceType != rpg.TickStanceNone {
			// draw the VBar
			stanceDuration := 0
			for j := i + 1; j <= skill.Duration(); j++ {
				if skill.Ticks[j].StanceType != tick.StanceType {
					break
				}
				stanceDuration++
			}
			vbarSprite := r.tickBubbleDisplayVBar
			vbarVerticalMargin := 1
			height := (r.config.TickSpacing * stanceDuration) + int(vbarSprite.Bounds().H()) - vbarVerticalMargin*2
			scale := float64(height) / vbarSprite.Bounds().H()
			vbarSprite.DrawColorMask(target, pixel.IM.ScaledXY(pixel.V(0, vbarSprite.Bounds().H()/2), pixel.V(1, scale)).Moved(pixel.V(0, float64(-vbarVerticalMargin))).Chained(tickSpriteCenterMatrix), mask)

			// get box sprite
			boxSprite := r.tickBarStanceBoxPending
			if opt.active {
				boxSprite = r.tickBarStanceBoxActive
			}

			stanceDelta := pixel.ZV
			artificialSkillProgress := opt.skillProgress + 0.25 // artificial progress to preempt overlay sprites above
			if artificialSkillProgress > float64(i) {
				stanceProgress := math.Min(artificialSkillProgress-float64(i), float64(stanceDuration))
				dy := float64(r.config.TickSpacing) * math.Max(0, stanceProgress-1)
				dy = math.Min(dy, float64(height))
				stanceDelta = pixel.V(0, -dy)
			}

			// draw box sprite
			stanceBoxMatrix := pixel.IM
			if opt.flip {
				stanceBoxMatrix = stanceBoxMatrix.ScaledXY(pixel.V(0, 0), pixel.V(-1, 1))
			}
			stanceBoxMatrix = stanceBoxMatrix.Moved(pixel.V(0, -5.5))
			stanceBoxMatrix = stanceBoxMatrix.Moved(stanceDelta)
			stanceBoxMatrix = stanceBoxMatrix.Chained(tickSpriteCenterMatrix)
			{
				thisMatrix := stanceBoxMatrix
				thisMask := mask
				postRenders = append(postRenders, func() {
					boxSprite.DrawColorMask(target, thisMatrix, thisMask)
				})
			}

			// draw the icon
			if stanceIconSprite, exists := r.stanceIcons[tick.StanceType]; exists {
				iconMatrix := pixel.IM
				if opt.flip {
					iconMatrix = iconMatrix.ScaledXY(pixel.V(0, 0), pixel.V(-1, 1))
					iconMatrix = iconMatrix.Moved(pixel.V(1.5, -5.5))
				} else {
					iconMatrix = iconMatrix.Moved(pixel.V(-1.5, -5.5))
				}
				iconMatrix = iconMatrix.Moved(stanceDelta)
				stanceMask := colors.Black.RGBA
				if opt.interruptedAt >= 0 {
					stanceMask = colors.WithAlpha(mask, 0.2)
				}
				{
					thisMatrix := iconMatrix.Chained(tickSpriteCenterMatrix)
					thisMask := stanceMask
					postRenders = append(postRenders, func() {
						stanceIconSprite.DrawColorMask(target, thisMatrix, thisMask)
					})
				}
			}
		}

		// draw dots for each tick
		var sprite pixelutil.BoundedDrawable
		bubbleMask := colors.White.RGBA
		var hasDamage, hasStatus bool
		for _, e := range tick.Effects {
			if e.Damage != nil {
				hasDamage = true
			}
			if e.Status != nil {
				hasStatus = true
				if statusColor, exists := colors.StatusColors[e.Status.Status]; exists {
					bubbleMask = statusColor
				}
			}
		}
		if hasDamage || hasStatus {
			if opt.interruptedAt >= 0 && i >= opt.interruptedAt {
				sprite = r.tickBubbleOverlayInterrupt
			} else {
				sprite = r.tickBubbleDisplayNormal
			}
		} else if !inStance {
			// only render "none" dots if we're not in a stance
			sprite = r.tickBubbleDisplayNone
		}
		bubbleMask = colors.Lerp(bubbleMask, colors.White.RGBA, 0.6)
		if sprite != nil {
			sprite.DrawColorMask(target, tickSpriteCenterMatrix, colors.MixColor(mask, bubbleMask))
		}
	}

	for _, postRender := range postRenders {
		postRender()
	}

	return result
}

func (r *Renderer) HeightOf(skill rpg.Skill) int {
	return r.config.TickSpacing * (skill.Duration() + 1)
}
