package title

import (
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var atlas = resources.CreateAtlas(resources.AtlasFilter{
	DoIncludeSprite: resources.RequireSpritePrefix("title/"),
	FontNames: []string{
		resources.FontNameFF,
	},
})

func init() {
	game.RegisterStateFactory(New)
}

type phase struct {
	from, until float64
	fn          func(float64) float64
	render      func(target pixel.Target, center pixel.Matrix, progression, elapsed float64)
}

type State struct {
	game.BaseState
	elapsed    float64
	maxElapsed float64
	phases     []phase
	exiting    bool
}

func New(intent game.TitleIntent) game.State {
	phases := []phase{
		{render: renderBackground, from: 1, until: 9},
		{render: renderStars, from: 0, until: 6},
		{render: renderClouds, from: 3, until: 5},
		{render: renderMountainBack, from: 2, until: 4},
		{render: renderCity, from: 3, until: 7},
		{render: renderMountainFront, from: 0, until: 2},
		{render: renderMoon, from: 4, until: 6},
		{render: renderTitle, from: 4, until: 6},
		{render: renderSubtitle, from: 5, until: 7},
		{render: renderShip, from: 4, until: 10},
		{render: renderStar, from: 6, until: 8},
		{render: renderReady, from: 6, until: 10},
	}
	maxUntil := 0.0
	for _, p := range phases {
		maxUntil = max(maxUntil, p.until)
	}
	return &State{
		phases:     phases,
		maxElapsed: maxUntil,
	}
}

var text = textbox.NewInstance(
	atlas.GetFont(resources.FontNameFF),
	tbcfg.NewConfig(0, 0, tbcfg.WithExpandMode(tbcfg.ExpandFit), tbcfg.Foreground(colors.White.RGBA)))

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	target.Clear(colors.FromString("#0a0510"))
	if s.exiting {
		s.elapsed -= timeDelta * 18
	} else {
		s.elapsed += timeDelta
	}
	s.elapsed = max(0, min(s.elapsed, s.maxElapsed))
	center := gfx.Moved(game.GameWidth/2, game.GameHeight/2)
	for _, p := range s.phases {
		phaseElapsed := s.elapsed - p.from
		phaseDuration := p.until - p.from
		progression := phaseElapsed / phaseDuration
		if p.fn == nil {
			progression = interp.Smootherstep(progression)
		} else {
			progression = p.fn(progression)
		}
		progression = max(0, min(progression, 1))
		p.render(target, center, progression, s.elapsed)
	}
	ctrls := game.Controls[*State]()
	if ctrls.ButtonA().JustPressed() || ctrls.ButtonStart().JustPressed() || ctrls.ButtonStart().JustPressed() || ctrls.ButtonSelect().JustPressed() {
		s.exiting = true
	}
	if s.exiting && s.elapsed <= 0 {
		game.SetActiveStateIntent(game.DefaultSelectorState())
	}
}

func renderBackground(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/bg")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderStars(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/stars")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderMountainFront(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/mtn_front")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderClouds(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/clouds")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, moved(center, 40, 30, progression), mask)
}

func renderCity(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/city")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, moved(center, 0, 10, progression), mask)
}

func renderMountainBack(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/mtn_back")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderTitle(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/title")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, moved(center, 20, 0, progression), mask)
}

func renderMoon(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/moon")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderSubtitle(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/subtitle")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	sprite.DrawColorMask(target, moved(center, 20, 0, progression), mask)
}

func renderShip(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/ship")
	mask := colors.WithAlpha(colors.White.RGBA, progression)
	//rotation := -0.1 * math.Pi * 2.0
	//rotation = rotation - rotation*progression
	//center = center.Rotated(gfx.IVec(202, 90), rotation)
	//center = moved(center, 30, -20, progression)
	sprite.DrawColorMask(target, center, mask)
}

func renderStar(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	sprite := atlas.GetSprite("title/star")
	speed := 0.5
	amount := 0.5
	mask := colors.WithAlpha(colors.White.RGBA, progression*(1-amount+amount*game.Utils().TimeCycleSin(speed)))
	sprite.DrawColorMask(target, center, mask)
}

func renderReady(target pixel.Target, center pixel.Matrix, progression, elapsed float64) {
	content := text.NewComplexContent("{+o:black}[press any button]")
	fg := colors.Lerp(colors.FromString("#e69f1a"), colors.FromString("#facc77"), game.Utils().TimeCycleSin(1))
	fg = colors.WithAlpha(fg, progression)
	topCenter := center.Moved(gfx.IVec(0, -52))
	topCenter = moved(topCenter, 0, 35, progression)
	text.Render(target, topCenter, content, tbcfg.Foreground(fg), tbcfg.RenderFrom(gfx.TopCenter))
}

func moved(m pixel.Matrix, x, y int, progression float64) pixel.Matrix {
	dx := float64(-x) + float64(x)*progression
	dy := float64(-y) + float64(y)*progression
	return m.Moved(pixel.V(dx, dy))
}
