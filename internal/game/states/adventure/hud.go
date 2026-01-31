package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type Hud struct {
	state             *State
	elythiumCountIcon *anim.AnimatedSprite

	researchIcon *anim.AnimatedSprite
}

func NewHud(state *State) *Hud {
	return &Hud{
		state:             state,
		elythiumCountIcon: anim.LoadTilesheetAnimation(atlas, "adventure/hud/elythium", "default"),
		researchIcon:      anim.LoadTilesheetAnimation(atlas, "adventure/hud/research", "default"),
	}
}

func (h *Hud) OnTick(s *State, target pixel.Target, cameraDelta pixel.Vec, bounds MapBounds, timeDelta float64) {
	h.onTickElythiumCount(s, target, pixel.IM.Moved(cameraDelta), bounds, timeDelta)
}

var hucCountTextHeight = 10
var hudCountText *textbox.Instance

func (h *Hud) onTickElythiumCount(s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	if !s.EvaluateControlToggles(s.controls.Elythium.Default, s.controls.Elythium.ToggledBy) {
		return
	}

	h.elythiumCountIcon.Update(timeDelta)
	sprite := h.elythiumCountIcon.Sprite()
	padding := 6.0
	topRight := pixel.IM.Moved(pixel.V(game.GameWidth-padding, game.GameHeight-padding))
	h.elythiumCountIcon.Sprite().Draw(target, topRight.Moved(gfx.TopRight.Align(sprite)))

	content := hudCountText.NewComplexContent(fmt.Sprintf("{+o:#3d1632,+c:#edb2dc}%d", h.state.globals.Get(rpg.GlobalKeyElythium).AsInt(0)))
	txtVNudge := -1.0 // push it down or up to align with sprite
	txtVNudge -= (sprite.Bounds().H() - content.Bounds().H()) / 2
	txtHNudge := -1.0 // padding between number and sprite
	txtM := topRight.Moved(pixel.V(txtHNudge-float64(content.Width()), txtVNudge))
	content.Render(target, txtM, tbcfg.RenderFrom(gfx.TopRight))
}
