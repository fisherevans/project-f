package combat

import (
	"fmt"
	"math/rand"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
)

func (s *State) AddFX(fx ...FX) {
	s.fx = append(s.fx, fx...)
}

var fxMaxAge = 4.0
var fxGravity = -100.0

type FX interface {
	Update(ctx *game.Context, s *State, timeDelta float64) bool
	Render(ctx *game.Context, target pixel.Target)
}

type baseFx struct {
	Position   pixel.Vec
	Velocity   pixel.Vec
	Age        float64
	SpeedScale float64
}

func newBaseFx(combatant Combatant) *baseFx {
	var p, v pixel.Vec
	if combatant.IsPlayer() {
		p = pixel.V(game.GameWidth*0.15, game.GameHeight*0.5)
		v = pixel.V(20, rand.Float64()*50+50)
	} else {
		p = pixel.V(game.GameWidth*0.85, game.GameHeight*0.5)
		v = pixel.V(-20, rand.Float64()*50+50)
	}
	return &baseFx{
		Position:   p,
		Velocity:   v,
		Age:        0,
		SpeedScale: 2,
	}
}

func (fx *baseFx) Update(ctx *game.Context, s *State, timeDelta float64) bool {
	if fx.SpeedScale > 0 {
		timeDelta = timeDelta * fx.SpeedScale
	}
	fx.Age += timeDelta
	fx.Position = fx.Position.Add(fx.Velocity.Scaled(timeDelta))
	fx.Velocity = pixel.V(fx.Velocity.X, fx.Velocity.Y+fxGravity*timeDelta)
	return fx.Age > fxMaxAge
}

var baseFxText = text.New(pixel.ZV, atlas.GetFont(resources.FontNameM3x6).Atlas).
	AlignedTo(pixel.Center)

func (fx *baseFx) renderFx(ctx *game.Context, color pixel.RGBA, text string, target pixel.Target) {
	color = colors.WithAlpha(color, 1.0-(fx.Age/fxMaxAge))

	baseFxText.Clear()
	baseFxText.Dot = pixel.ZV
	baseFxText.Color = colors.ScaleColor(color, 0.1)
	baseFxText.WriteString(text)
	baseFxText.Draw(target, pixel.IM.Moved(fx.Position))

	baseFxText.Dot = pixel.ZV.Add(pixel.V(-1, 1))
	baseFxText.Color = color
	baseFxText.WriteString(text)
	baseFxText.Draw(target, pixel.IM.Moved(fx.Position))
}

type HealthAdjustFX struct {
	*baseFx
	DisplayAmount int
	Color         pixel.RGBA
}

func NewHealthAdjustFX(amount int, color pixel.RGBA, target Combatant) *HealthAdjustFX {
	return &HealthAdjustFX{
		DisplayAmount: amount,
		Color:         color,
		baseFx:        newBaseFx(target),
	}
}

func (fx *HealthAdjustFX) Render(ctx *game.Context, target pixel.Target) {
	fx.baseFx.renderFx(ctx, fx.Color, fmt.Sprintf("%+d", fx.DisplayAmount), target)
}

type WordFX struct {
	*baseFx
	Word  string
	Color pixel.RGBA
}

func NewWordFX(word string, color pixel.RGBA, target Combatant) *WordFX {
	b := newBaseFx(target)
	b.Velocity = pixel.ZV
	b.Position.Y -= 5
	b.SpeedScale = 0.5
	return &WordFX{
		Word:   word,
		Color:  color,
		baseFx: b,
	}
}

func (fx *WordFX) Render(ctx *game.Context, target pixel.Target) {
	fx.baseFx.renderFx(ctx, fx.Color, fx.Word, target)
}
