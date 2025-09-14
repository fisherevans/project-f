package adventure

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type Hud struct {
	ElythiumCount     int
	elythiumCountIcon *anim.AnimatedSprite
}

func NewHud() *Hud {
	return &Hud{
		ElythiumCount:     0,
		elythiumCountIcon: anim.Load(atlas, "adventure/hud/elythium", "default"),
	}
}

func (h *Hud) OnTick(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	h.onTickElythiumCount(ctx, s, target, matrix, bounds, timeDelta)
}

var hucCountTextHeight = 10
var hudCountText = textbox.NewInstance(
	atlas.GetFont(resources.FontNameM5x7),
	tbcfg.NewConfig(game.GameWidth/3, hucCountTextHeight,
		tbcfg.HAligned(tbcfg.HAlignRight),
		tbcfg.VAligned(tbcfg.VAlignMiddle),
		tbcfg.WithExpandMode(tbcfg.ExpandFit)))

func (h *Hud) onTickElythiumCount(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	h.elythiumCountIcon.Update(timeDelta)
	sprite := h.elythiumCountIcon.Sprite()
	padding := 6.0
	topRight := pixel.IM.Moved(pixel.V(game.GameWidth-padding, game.GameHeight-padding))
	h.elythiumCountIcon.Sprite().Draw(target, topRight.Moved(gfx.TopRight.Align(sprite)))

	content := hudCountText.NewComplexContent(fmt.Sprintf("{+o:#3d1632,+c:#edb2dc}%d", h.ElythiumCount))
	txtVNudge := -1.0 // push it down or up to align with sprite
	txtVNudge -= (sprite.Bounds().H() - content.Bounds().H()) / 2
	txtHNudge := -1.0 // padding between number and sprite
	txtHNudge -= sprite.Bounds().W()
	txtM := topRight.Moved(pixel.V(txtHNudge, txtVNudge))
	hudCountText.Render(ctx, target, txtM, content, tbcfg.RenderFrom(gfx.TopRight))
}
