package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"image/color"
	"math/rand"
)

type RenderEffect interface {
	ColorOverride() *color.Color
	RenderDelta() pixel.Vec
	Update(ctx *game.Context, timeDelta float64)
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

func (r *rumbleRenderEffect) ColorOverride() *color.Color {
	return nil
}

func (r *rumbleRenderEffect) RenderDelta() pixel.Vec {
	return pixel.V(float64(r.dx), float64(r.dy))
}

func (r *rumbleRenderEffect) Update(ctx *game.Context, timeDelta float64) {
	r.elapsed += timeDelta
	for r.elapsed > r.rate {
		r.elapsed -= r.rate
		r.dy = rand.Intn(3) - 1
		r.dx = rand.Intn(3) - 1
	}
	ctx.DebugBR("delta: %d, %d", r.dx, r.dy)
}
