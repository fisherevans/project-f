package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"math/rand"
)

type RenderEffect interface {
	Update(ctx *game.Context, timeDelta float64)
	Apply(params *characterRenderParams)
}

type rumbleRenderEffect struct {
	rate    float64
	elapsed float64
	dx      int
	dy      int
}

func newRumble(rate float64) *rumbleRenderEffect {
	return &rumbleRenderEffect{
		rate: rate,
	}
}

func (r *rumbleRenderEffect) Update(ctx *game.Context, timeDelta float64) {
	r.elapsed += timeDelta
	for r.elapsed > r.rate {
		r.elapsed -= r.rate
		r.dy = rand.Intn(3) - 1
		r.dx = rand.Intn(3) - 1
	}
}

func (r *rumbleRenderEffect) Apply(params *characterRenderParams) {
	params.drawDelta = params.drawDelta.Add(pixel.V(float64(r.dx), float64(r.dy)))
}
