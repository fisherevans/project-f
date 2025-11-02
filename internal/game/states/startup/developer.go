package startup

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
)

var developerHighlight = colors.FromString("#0cd1e3")

type phase struct {
	from, until float64
	render      func(target pixel.Target, progression, timeDelta float64)
	fn          interp.Function
}

type DeveloperState struct {
	game.BaseState
	elapsed        float64
	phases         []phase
	effects        []*effect
	random         *rand.Rand
	effectTriggers []int
}

func NewDeveloper(_ game.StartupDeveloperIntent) game.State {
	s := &DeveloperState{
		random: rand.New(rand.NewSource(12333)),
	}
	s.phases = []phase{
		{render: s.renderMainIcon, from: 3, until: 5},
		{render: s.renderName, from: 2, until: 4},
		{render: s.renderPresents, from: 4, until: 6},
		{render: s.renderEffects, from: 0, until: 10},
		{render: s.renderShootingIcon, from: 0.1, until: 2.5, fn: interp.Linear},
		{render: s.renderBars, from: 0, until: 0.45},
	}
	effectDelta := game.GameWidth / 30
	for x := game.GameWidth - effectDelta; x > 0; x -= effectDelta {
		s.effectTriggers = append(s.effectTriggers, x)
	}
	return s
}

func (s *DeveloperState) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	if game.Controls[*DeveloperState]().ButtonB().JustPressed() {
		game.SetActiveStateIntent(game.StartupDeviceIntent{})
	}
	s.elapsed += timeDelta
	target.Clear(colors.FromString("#563da6"))

	for _, p := range s.phases {
		phaseElapsed := s.elapsed - p.from
		phaseDuration := p.until - p.from
		progression := min(max(phaseElapsed/phaseDuration, 0), 1)
		if p.fn == nil {
			progression = interp.Smootherstep(progression)
		} else {
			progression = p.fn(progression)
		}
		p.render(target, progression, timeDelta)
	}
}

func (s *DeveloperState) renderBars(target pixel.Target, progression, _ float64) {
	bannerHeight := 24
	dy := int((float64(game.GameHeight/2) - float64(bannerHeight) + 2) * (1.0 - progression))
	game.DebugBL("prog: %f", progression)
	game.DebugBL("dy: %d", dy)
	gfx.DrawRect(atlas, target, gfx.Moved(0, +dy+bannerHeight), gfx.TopLeft, game.GameWidth, game.GameHeight, colors.Black.RGBA)
	gfx.DrawRect(atlas, target, gfx.Moved(0, game.GameHeight-dy-bannerHeight), gfx.BottomLeft, game.GameWidth, game.GameHeight, colors.Black.RGBA)
}

type effect struct {
	s                  *DeveloperState
	sprite             pixelutil.BoundedDrawable
	velocity, position pixel.Vec
	rotation           float64
	rotationSpeed      float64
	age                float64
	maxAge             float64
	mask               pixel.RGBA
}

var effectGravity = pixel.V(0, 0)

var effectAlphaKey = interp.NewKeys().
	WithFunction(interp.Smootherstep).
	WithKey(0, 0).
	WithKey(0.1, 1).
	WithKey(0.5, 1).
	WithKey(1, 0)

func (e *effect) render(target pixel.Target, globlaEffectProgress, timeDelta float64) bool {
	e.age += timeDelta
	progress := min(e.age/e.maxAge, 1.0)
	mask := colors.WithAlpha(e.mask, effectAlphaKey.Interpolate(progress))
	e.rotation = math.Mod(e.rotation+timeDelta*e.rotationSpeed, 1.0)
	e.velocity = e.velocity.Add(effectGravity.Scaled(timeDelta))
	e.position = e.position.Add(e.velocity.Scaled(timeDelta))
	rotation := e.rotation
	//rotation = math.Mod(math.Round(rotation*8.0), 8.0) / 8.0
	m := pixel.IM.Rotated(pixel.ZV, rotation*math.Pi*2).Moved(e.position)
	e.sprite.DrawColorMask(target, m, mask)
	return e.age > e.maxAge // not done yet
}

func (s *DeveloperState) renderEffects(target pixel.Target, progression, timeDelta float64) {
	var remaining []*effect
	for _, e := range s.effects {
		if !e.render(target, progression, timeDelta) {
			remaining = append(remaining, e)
		}
	}
	s.effects = remaining
}

func (s *DeveloperState) renderMainIcon(target pixel.Target, progression float64, delta float64) {
	mask := colors.WithAlpha(developerHighlight, progression)
	m := centerMatrix.Moved(gfx.IVec(0, 1))
	atlas.GetSprite("startup/developer_logo").DrawColorMask(target, m, mask)
}

func (s *DeveloperState) renderName(target pixel.Target, progression float64, delta float64) {
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	m := centerMatrix.Moved(gfx.IVec(0, 6))
	atlas.GetSprite("startup/developer_name").DrawColorMask(target, m, mask)

}

func (s *DeveloperState) renderPresents(target pixel.Target, progression float64, delta float64) {
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	m := centerMatrix.Moved(pixel.V(0, -13.0))
	atlas.GetSprite("startup/developer_presents").DrawColorMask(target, m, mask)
}

func (s *DeveloperState) renderShootingIcon(target pixel.Target, progression float64, delta float64) {
	startPos := pixel.V(float64(game.GameWidth)+20, float64(game.GameHeight/2)-20)
	peakPos := pixel.V(float64(game.GameWidth)/2, float64(game.GameHeight/2)+10)

	// Calculate the origin and radius based on knobs
	endPos := pixel.V(peakPos.X-(startPos.X-peakPos.X), startPos.Y)
	originX := peakPos.X
	dx := startPos.X - originX
	dyPeak := peakPos.Y - startPos.Y // vertical distance from start to peak
	h := (dx*dx - dyPeak*dyPeak) / (2 * dyPeak)
	originY := startPos.Y - h
	radius := math.Abs(peakPos.Y - originY)

	// Calculate angles for start, peak, and end
	startAngle := math.Atan2(startPos.Y-originY, startPos.X-originX)
	endAngle := math.Atan2(endPos.Y-originY, endPos.X-originX)

	// Interpolate angle along the arc
	angle := startAngle + (endAngle-startAngle)*progression

	// Calculate position
	x := originX + math.Cos(angle)*radius
	y := originY + math.Sin(angle)*radius

	// Draw the sprite
	m := pixel.IM.Moved(pixel.V(x, y))
	sprite := atlas.GetTilesheetSprite("startup/developer_effects", 1, 1)
	sprite.DrawColorMask(target, m, developerHighlight)

	if len(s.effectTriggers) == 0 || int(x) > s.effectTriggers[0] {
		return
	}

	s.effectTriggers = s.effectTriggers[1:]
	spriteId := s.random.Intn(3)
	ySpread := 20
	dy := s.random.Intn(ySpread*2) - ySpread
	maxRotationSpeed := 0.5
	maxAge := 3.0
	maxSpeed := 15.0
	maxAlphaCull := 0.3
	baseColor := colors.Lerp(colors.White.RGBA, developerHighlight, s.random.Float64())
	s.effects = append(s.effects, &effect{
		s:             s,
		sprite:        atlas.GetTilesheetSprite("startup/developer_effects", 2+spriteId, 1),
		position:      pixel.V(x, y+float64(dy)),
		velocity:      pixel.V(s.random.Float64()*maxSpeed/2, s.random.Float64()*maxSpeed*2-maxSpeed),
		rotation:      s.random.Float64(),
		rotationSpeed: s.random.Float64()*maxRotationSpeed*2 - maxRotationSpeed,
		maxAge:        s.random.Float64()*maxAge/2.0 + maxAge/2.0,
		mask:          colors.WithAlpha(baseColor, 1.0-maxAlphaCull*s.random.Float64()),
	})
}
