package xenolog

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

// Variables declared in init_vars.go
var badgeAContinue *badges.ButtonAction

type animechLevelUpAnimation struct {
	screen             *screen.Instance
	fromLevel, toLevel int
	elapsed            float64
	totalDuration      float64
}

func newAnimechLevelUpAnimation(screen *screen.Instance, fromLevel, toLevel int) *animechLevelUpAnimation {
	animation := &animechLevelUpAnimation{
		screen:        screen,
		fromLevel:     fromLevel,
		toLevel:       toLevel,
		totalDuration: 3.5,
	}
	return animation
}

func (s *animechLevelUpAnimation) Enter() {
}

func (s *animechLevelUpAnimation) OnTick(target pixel.Target, timeDelta float64) {
	s.elapsed += timeDelta
	if game.Controls[*State]().ButtonB().JustPressed() || game.Controls[*State]().ButtonA().JustPressed() {
		s.screen.PopMenuAnimated()
	}
	smallText := newTextRenderer(target, smallTextbox)
	titleText := newTextRenderer(target, titleTextbox)

	titleText.render("Leveled Up!", screenWidth/2, screenHeight-12, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))
	fromToText := fmt.Sprintf("From level {+c:xenolog_highlight}%d{-*} to {+c:xenolog_highlight}%d{-*}", s.fromLevel, s.toLevel)
	smallText.render(fromToText, screenWidth/2, 12, screenColors.Text, tbcfg.RenderFrom(gfx.BottomCenter))

	// Base position
	center := gfx.Moved(screenWidth/2, screenHeight/2)

	// Animation phases as ratios of total duration
	startDelay := s.totalDuration * 0.1
	transitionDur := s.totalDuration * 0.7
	endStabilize := s.totalDuration * 0.2

	var scaleMultiplier float64
	var verticalOffset float64
	var rotation float64
	var spriteColor pixel.RGBA
	if s.elapsed < startDelay {
		// Phase 1: Static at start - half scale, colorText
		scaleMultiplier = 0.5
		verticalOffset = 0
		rotation = 0
		spriteColor = screenColors.Text

	} else if s.elapsed < startDelay+transitionDur {
		// Phase 2: Active evolution animation
		transitionElapsed := s.elapsed - startDelay
		transitionProgress := transitionElapsed / transitionDur

		// Intensity ramps up then down (ease in/out)
		intensity := math.Sin(transitionProgress * math.Pi) // 0 -> 1 -> 0

		// Scale grows from 0.5 to 1.0
		baseScale := 0.5 + 0.5*transitionProgress
		pulseSpeed := 8.0
		pulsePhase := transitionElapsed * pulseSpeed * 2 * math.Pi
		pulseAmount := 0.12 * intensity
		scaleMultiplier = baseScale * (1.0 + pulseAmount*math.Sin(pulsePhase))

		// Vertical floating
		floatSpeed := 2.5
		floatPhase := transitionElapsed * floatSpeed * 2 * math.Pi
		floatHeight := 6.0 * intensity
		verticalOffset = floatHeight * math.Sin(floatPhase)

		// Rotation
		rotationSpeed := 3.0
		rotationPhase := transitionElapsed * rotationSpeed * 2 * math.Pi
		maxRotation := 0.05 * intensity
		rotation = maxRotation * math.Sin(rotationPhase)

		// Color flashing - slower than before
		flashSpeed := 5.0
		flashPhase := transitionElapsed * flashSpeed * 2 * math.Pi
		flashT := (math.Sin(flashPhase) + 1.0) / 2.0
		baseColor := colors.Lerp(screenColors.Dark, screenColors.Highlight, transitionProgress)
		spriteColor = colors.Lerp(baseColor, screenColors.Highlight, flashT*intensity*0.7)

		// Opacity pulsing
		opacitySpeed := 2.0
		opacityPhase := transitionElapsed * opacitySpeed * 2 * math.Pi
		minOpacity := 0.7 + 0.2*(1.0-intensity)
		maxOpacity := 1.0
		opacity := minOpacity + (maxOpacity-minOpacity)*(math.Sin(opacityPhase)+1.0)/2.0
		spriteColor.A = spriteColor.A * opacity

	} else if s.elapsed < startDelay+transitionDur+endStabilize {
		// Phase 3: Stabilizing - transition to final state
		stabilizeElapsed := s.elapsed - (startDelay + transitionDur)
		stabilizeProgress := stabilizeElapsed / endStabilize

		// Dampen remaining oscillations while staying at full scale
		scaleMultiplier = 1.0 + (1.0-stabilizeProgress)*0.05*math.Sin(stabilizeElapsed*8.0*2*math.Pi)
		verticalOffset = (1.0 - stabilizeProgress) * 3.0 * math.Sin(stabilizeElapsed*2.5*2*math.Pi)
		rotation = (1.0 - stabilizeProgress) * 0.05 * math.Sin(stabilizeElapsed*3.0*2*math.Pi)

		// Color transitions to highlight
		spriteColor = colors.Lerp(screenColors.Text, screenColors.Highlight, stabilizeProgress)

	} else {
		// Phase 4: Final stable state - full scale, slow pulse
		scaleMultiplier = 1.0
		verticalOffset = 0
		rotation = 0

		// Slight color pulse
		spriteColor = flashingHighlightSlow()

		badgeAContinue.Render(target, gfx.Moved(screenWidth-2, 2), gfx.BottomRight)
	}

	// Build transformation matrix
	matrix := pixel.IM.
		Scaled(pixel.ZV, scaleMultiplier).
		Rotated(pixel.ZV, rotation).
		Moved(pixel.V(0, verticalOffset)).
		Chained(center)

	// Draw the animech sprite with all effects
	spriteAnimech.DrawColorMask(target, matrix, spriteColor)
}
