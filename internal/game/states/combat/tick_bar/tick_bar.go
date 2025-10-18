package tick_bar

import (
	"math"
	"runtime"
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/sprites"
)

type Renderer struct {
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

func NewRenderer(atlas *resources.Atlas) *Renderer {
	r := &Renderer{}

	r.activeFrame = frames.New("combat/tick_bar/skill_active_frame", atlas)
	r.pendingFrame = frames.New("combat/tick_bar/skill_pending_frame", atlas)

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
	width       int
	tickSpacing int

	interruptedAt int
	mask          pixel.RGBA
	active        bool
	alpha         float64
	skillProgress float64
	flip          bool
}

func NewDrawOptions(width, tickSpacing int) DrawOptions {
	return DrawOptions{
		width:         width,
		tickSpacing:   tickSpacing,
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
	// Diagnostic: Track memory and time for first draw
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	startTime := time.Now()
	defer func() {
		runtime.ReadMemStats(&m2)
		elapsed := time.Since(startTime)
		allocDiff := int64(m2.Alloc) - int64(m1.Alloc)
		heapDiff := int64(m2.HeapAlloc) - int64(m1.HeapAlloc)
		if elapsed > 100*time.Millisecond || allocDiff > 10*1024*1024 {
			log.Warn().
				Dur("elapsed", elapsed).
				Int64("alloc_mb", allocDiff/(1024*1024)).
				Int64("heap_mb", heapDiff/(1024*1024)).
				Uint64("total_alloc_mb", m2.TotalAlloc/(1024*1024)).
				Uint32("num_gc", m2.NumGC-m1.NumGC).
				Uint64("sys_mb", m2.Sys/(1024*1024)).
				Msg("TICK_BAR_DRAW: Slow or large allocation detected")
			
			// Force GC and check if memory is released
			runtime.GC()
			var m3 runtime.MemStats
			runtime.ReadMemStats(&m3)
			log.Info().
				Int64("after_gc_alloc_mb", int64(m3.Alloc)/(1024*1024)).
				Int64("freed_mb", (int64(m2.Alloc)-int64(m3.Alloc))/(1024*1024)).
				Msg("TICK_BAR_DRAW: Memory after forced GC")
		}
	}()
	
	result := DrawResult{}
	if skill == nil {
		return result
	}
	mask := colors.WithAlphaTodoFix(opt.mask, opt.alpha)
	if opt.interruptedAt >= 0 {
		mask = colors.ScaleColor(mask, 0.5)
	}
	rect := pixel.R(0, 0,
		float64(opt.width),
		float64(opt.tickSpacing*(skill.Duration()+1)))
	matrixBottomLeft := matrixTopMiddle.Moved(pixel.V(-rect.W()/2, -rect.H()))

	frame := r.pendingFrame
	if opt.active {
		frame = r.activeFrame
	}
	
	// Diagnostic: Time the frame draw specifically
	frameStart := time.Now()
	frame.Draw(target, rect, matrixBottomLeft, frames.WithColor(mask))
	frameElapsed := time.Since(frameStart)
	if frameElapsed > 50*time.Millisecond {
		log.Warn().
			Dur("frame_draw_ms", frameElapsed).
			Float64("rect_w", rect.W()).
			Float64("rect_h", rect.H()).
			Int("skill_duration", skill.Duration()).
			Msg("TICK_BAR_DRAW: Frame draw was slow")
	}

	var postRenders []func() // used to render on top of the tick bar, after the dots/lines are rendered - mostly for the stance box + icons
	for i := 0; i <= skill.Duration(); i++ {
		lastStance := rpg.TickStanceNone
		if i > 0 {
			lastStance = skill.Ticks[i-1].StanceType
		}
		tick := skill.Ticks[i]
		inStance := tick.StanceType != rpg.TickStanceNone

		tickSpriteCenterMatrix := matrixBottomLeft.Moved(pixel.V(float64(opt.width/2), float64(opt.tickSpacing/2+opt.tickSpacing*(skill.Duration()-i))))
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
			height := (opt.tickSpacing * stanceDuration) + int(vbarSprite.Bounds().H()) - vbarVerticalMargin*2
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
				dy := float64(opt.tickSpacing) * math.Max(0, stanceProgress-1)
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
				thisSprite := boxSprite
				postRenders = append(postRenders, func() {
					thisSprite.DrawColorMask(target, thisMatrix, thisMask)
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
					stanceMask = colors.WithAlphaTodoFix(mask, 0.2)
				}
				{
					thisMatrix := iconMatrix.Chained(tickSpriteCenterMatrix)
					thisMask := stanceMask
					thisSprite := stanceIconSprite
					postRenders = append(postRenders, func() {
						thisSprite.DrawColorMask(target, thisMatrix, thisMask)
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

func (r *Renderer) HeightOf(skill rpg.Skill, opt DrawOptions) int {
	return opt.tickSpacing * (skill.Duration() + 1)
}
