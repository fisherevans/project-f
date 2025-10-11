package xenolog

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	smallTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))

	mediumTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))
)

type textRenderer struct {
	target pixel.Target
	tb     *textbox.Instance
}

func newTextRenderer(target pixel.Target, tb *textbox.Instance) *textRenderer {
	return &textRenderer{
		target: target,
		tb:     tb,
	}
}

func (r *textRenderer) render(text string, x, y int, mask pixel.RGBA, opts ...tbcfg.ConfigOpt) (int, int) {
	c := r.tb.NewComplexContent(text)
	matrix := pixel.IM.Moved(gfx.IVec(x, y))
	r.tb.Render(r.target, matrix, c, append(opts, tbcfg.Foreground(mask))...)
	return c.Width(), c.Height()
}

func flashingHighlight() pixel.RGBA {
	return colors.Lerp(uiMask, uiMaskSelected, game.Utils().TimeCycleSin(selectBoxSubLabelFlashSpeed))
}

func saveOrNotify() {
	if err := game.CurrentSave().Save(); err != nil {
		log.Err(err).Msg("failed to save game")
		game.DebugNotification("Failed to save game!!!")
	}
}
