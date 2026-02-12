package xenolog

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

// Variables declared in init_vars.go

type textRenderer struct {
	target pixel.Target
	matrix pixel.Matrix
	tb     *textbox.Instance
	opts   []tbcfg.ConfigOpt
}

func newTextRenderer(target pixel.Target, tb *textbox.Instance) *textRenderer {
	return &textRenderer{
		target: target,
		matrix: pixel.IM,
		tb:     tb,
	}
}

func (r *textRenderer) withMatrix(m pixel.Matrix) *textRenderer {
	r.matrix = m
	return r
}

func (r *textRenderer) withOpts(opts ...tbcfg.ConfigOpt) *textRenderer {
	r.opts = append(r.opts, opts...)
	return r
}

func (r *textRenderer) render(text string, x, y int, mask pixel.RGBA, opts ...tbcfg.ConfigOpt) (int, int) {
	c := r.tb.NewComplexContent(text)
	matrix := r.matrix.Moved(gfx.IVec(x, y))
	c.Render(r.target, matrix, append(r.opts, append(opts, tbcfg.Foreground(mask))...)...)
	return c.Width(), c.Height()
}

func flashingHighlight() pixel.RGBA {
	return colors.Lerp(screenColors.Text, screenColors.Highlight, game.Utils().TimeCycleSin(selectBoxSubLabelFlashSpeed))
}

func flashingHighlightSlow() pixel.RGBA {
	return colors.Lerp(screenColors.Text, screenColors.Highlight, game.Utils().TimeCycleSin(selectBoxSubLabelFlashSpeed/2))
}

func saveOrNotify() {
	if err := game.CurrentSave().Save(); err != nil {
		log.Err(err).Msg("failed to save game")
		game.DebugNotificationf("Failed to save game!!!")
	}
}
