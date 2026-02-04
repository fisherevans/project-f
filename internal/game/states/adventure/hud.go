package adventure

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/particles"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util/gfx"
)

var (
	elythiumBarFrame *frames.Instance
)

type Hud struct {
	state             *State
	elythiumCountIcon *anim.AnimatedSprite

	researchIcon *anim.AnimatedSprite

	sparkleParticles *particles.Group
	timeUntilSparkle float64
	sparkling        bool
}

func NewHud(state *State) *Hud {
	return &Hud{
		state:             state,
		elythiumCountIcon: anim.LoadTilesheetAnimation(atlas, "adventure/hud/elythium", "default"),
		researchIcon:      anim.LoadTilesheetAnimation(atlas, "adventure/hud/research", "default"),
		sparkleParticles: particles.NewGroup(
			particles.RandomPositionFactory(gfx.Centered, float64(elythiumBarMaxWidth+6), float64(elythiumBarHeight+6)),
			particles.RandomVelocityFactory(0.1, 0.5),
			func() *anim.AnimatedSprite {
				return anim.LoadTilesheetAnimation(atlas, "adventure/hud/elythium_sparkle", "random")
			},
			particles.RandomAgeFactory(2.5, 4),
			pixel.V(0, 0),
			particles.WithFading(0.4, 0.4),
			particles.WithColorMaskSupplier(func() pixel.RGBA {
				mask := colors.Lerp(elythiumBarColorA, elythiumBarColorB, rand.Float64())
				mask = colors.Lerp(colors.White.RGBA, mask, 0.2)
				return mask
			}),
		),
	}
}

func (h *Hud) OnTick(s *State, target pixel.Target, cameraDelta pixel.Vec, bounds MapBounds, timeDelta float64) {
	h.onTickElythiumCount(s, target, pixel.IM.Moved(cameraDelta), bounds, timeDelta)
}

var elythiumBarColorA = colors.FromString("#e735b5")
var elythiumBarColorB = colors.FromString("#e83a7e")
var elythiumBarLine = colors.FromString("#c1014e")
var elythiumBarMaxWidth = 20
var elythiumBarHeight = 3
var elythiumSparkleDelay = 0.75

func (h *Hud) onTickElythiumCount(s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	if !s.EvaluateControlToggles(s.controls.Elythium.Default, s.controls.Elythium.ToggledBy) {
		return
	}
	count := h.state.globals.Get(rpg.GlobalKeyElythium).AsInt(0)
	goal := h.state.globals.Get(rpg.GlobalKeyElythiumGoal).AsInt(10)

	h.elythiumCountIcon.Update(timeDelta)
	crystalSprite := h.elythiumCountIcon.Sprite()
	padding := 3.0
	crystalTopRight := pixel.IM.Moved(pixel.V(game.GameWidth-padding, game.GameHeight-padding))

	barWidth := min(int(math.Floor(float64(count)/float64(goal)*float64(elythiumBarMaxWidth))), elythiumBarMaxWidth)
	frameR := pixel.R(0, 0, float64(elythiumBarMaxWidth+2*2), float64(elythiumBarHeight+2*2))

	frameTopRight := crystalTopRight.Moved(pixel.V(-11, -crystalSprite.Bounds().H()/2+2))
	elythiumBarFrame.Draw(target, frameR, frameTopRight, frames.WithRenderOrigin(gfx.TopRight))

	barTopRight := frameTopRight.Moved(pixel.V(-2, -2))
	elythiumBarColor := colors.Lerp(elythiumBarColorA, elythiumBarColorB, game.Utils().TimeCycleSin(0.3))
	gfx.DrawRect(atlas, target, barTopRight, gfx.TopRight, barWidth, elythiumBarHeight, elythiumBarColor)

	if barWidth < elythiumBarMaxWidth {
		lineTopRight := barTopRight.Moved(pixel.V(float64(-barWidth+1), 0))
		gfx.DrawRect(atlas, target, lineTopRight, gfx.TopRight, 1, elythiumBarHeight, elythiumBarLine)
		h.sparkling = false
		h.timeUntilSparkle = 0
	} else {
		if !h.sparkling {
			h.sparkling = true
			h.timeUntilSparkle = elythiumSparkleDelay
			for i := 0; i < 12; i++ {
				h.sparkleParticles.AddParticle()
			}
		}
		h.timeUntilSparkle -= timeDelta
		if h.timeUntilSparkle <= 0 {
			h.sparkleParticles.AddParticle()
			h.timeUntilSparkle = elythiumSparkleDelay
		}
	}

	crystalSprite.Draw(target, crystalTopRight.Moved(gfx.TopRight.Align(crystalSprite)))

	h.sparkleParticles.Render(target, barTopRight.Moved(gfx.IVec(-elythiumBarMaxWidth/2, -elythiumBarHeight/2)), timeDelta)
}
