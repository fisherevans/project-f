package particles

import (
	"testing"

	"fisherevans.com/project/f/internal/game/anim"
	"github.com/gopxl/pixel/v2"
)

func TestGroup(t *testing.T) {
	posFactory := func() pixel.Vec { return pixel.ZV }
	velFactory := func(pos pixel.Vec) pixel.Vec { return pixel.ZV }
	ageFactory := func() float64 { return 1.0 }
	animFactory := func() *anim.AnimatedSprite { return nil }

	g := NewGroup(posFactory, velFactory, animFactory, ageFactory, pixel.ZV)
	if len(g.particles) != 0 {
		t.Errorf("expected 0 particles, got %d", len(g.particles))
	}

	g.AddParticle()
	if len(g.particles) != 1 {
		t.Errorf("expected 1 particle, got %d", len(g.particles))
	}

	// 0.5s passed, not dead
	g.Render(nil, pixel.IM, 0.5)
	if len(g.particles) != 1 {
		t.Errorf("expected 1 particle, got %d", len(g.particles))
	}

	// another 0.6s passed, total 1.1s, should be dead
	g.Render(nil, pixel.IM, 0.6)
	if len(g.particles) != 0 {
		t.Errorf("expected 0 particles, got %d", len(g.particles))
	}
}
