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
	ElythiumCount     func() int
	elythiumCountIcon *anim.AnimatedSprite

	researchIcon *anim.AnimatedSprite
}

func NewHud(elythiumCount func() int) *Hud {
	return &Hud{
		ElythiumCount:     elythiumCount,
		elythiumCountIcon: anim.Load(atlas, "adventure/hud/elythium", "default"),
		researchIcon:      anim.Load(atlas, "adventure/hud/research", "default"),
	}
}

func (h *Hud) OnTick(s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	h.onTickElythiumCount(s, target, matrix, bounds, timeDelta)

	type topRightCount struct {
		icon   *anim.AnimatedSprite
		count  int
		fg     string
		stroke string
	}

	// todo display elythium as a guage
	counts := []topRightCount{
		{
			icon:   h.elythiumCountIcon,
			count:  h.ElythiumCount(),
			stroke: "#3d1632",
			fg:     "#edb2dc",
		},
		//{
		//	icon:   h.researchIcon,
		//	count:  game.CurrentSave().Animech.AnimechExperience, // todo not from a run
		//	stroke: "#162d3d",
		//	fg:     "#b2d4ed",
		//},
	}

	rowPadding := 2.0
	padding := 6.0
	for id, count := range counts {
		count.icon.Update(timeDelta)
		sprite := count.icon.Sprite()
		topRight := pixel.IM.
			Moved(pixel.V(game.GameWidth-padding, game.GameHeight-padding)).
			Moved(pixel.V(0, -float64(id)*(sprite.Bounds().H()+rowPadding)))
		count.icon.Sprite().Draw(target, topRight.Moved(gfx.TopRight.Align(sprite)))

		content := hudCountText.NewComplexContent(fmt.Sprintf("{+o:%s,+c:%s}%d", count.stroke, count.fg, count.count))
		txtVNudge := -1.0 // push it down or up to align with sprite
		txtVNudge -= (sprite.Bounds().H() - content.Bounds().H()) / 2
		txtHNudge := -1.0 // padding between number and sprite
		txtHNudge -= sprite.Bounds().W()
		txtM := topRight.Moved(pixel.V(txtHNudge, txtVNudge))
		hudCountText.Render(target, txtM, content, tbcfg.RenderFrom(gfx.TopRight))
	}
}

var hucCountTextHeight = 10
var hudCountText = textbox.NewInstance(
	atlas.GetFont(resources.FontNameM5x7),
	tbcfg.NewConfig(game.GameWidth/3, hucCountTextHeight,
		tbcfg.HAligned(tbcfg.HAlignRight),
		tbcfg.VAligned(tbcfg.VAlignMiddle),
		tbcfg.WithExpandMode(tbcfg.ExpandFit)))

func (h *Hud) onTickElythiumCount(s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	h.elythiumCountIcon.Update(timeDelta)
	sprite := h.elythiumCountIcon.Sprite()
	padding := 6.0
	topRight := pixel.IM.Moved(pixel.V(game.GameWidth-padding, game.GameHeight-padding))
	h.elythiumCountIcon.Sprite().Draw(target, topRight.Moved(gfx.TopRight.Align(sprite)))

	content := hudCountText.NewComplexContent(fmt.Sprintf("{+o:#3d1632,+c:#edb2dc}%d", h.ElythiumCount()))
	txtVNudge := -1.0 // push it down or up to align with sprite
	txtVNudge -= (sprite.Bounds().H() - content.Bounds().H()) / 2
	txtHNudge := -1.0 // padding between number and sprite
	txtHNudge -= sprite.Bounds().W()
	txtM := topRight.Moved(pixel.V(txtHNudge, txtVNudge))
	hudCountText.Render(target, txtM, content, tbcfg.RenderFrom(gfx.TopRight))
}
