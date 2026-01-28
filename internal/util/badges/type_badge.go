package badges

import (
	"strings"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillTypeWidthLong  = 33
	skillTypeWidthShort = 19
	skillTypeHeight     = 7
)

type skillType struct {
	frame   *frames.Instance
	textbox *textbox.Instance
	content *textbox.Content
	bg, fg  pixel.RGBA
}

func (s *skillType) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight))
}

func (s *skillType) Render(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation) {
	s.RenderMask(target, matrix, origin, s.bg, s.fg)
}

func (s *skillType) RenderMask(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation, fg, bg pixel.RGBA) {
	matrix = matrix.Moved(origin.AlignFrom(gfx.BottomLeft, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight)))
	s.frame.Draw(target, pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight)), matrix,
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(bg))
	s.content.Render(target, matrix, tbcfg.Foreground(fg))
}

func (b *Builder) Of(name string, bg, fg pixel.RGBA, width int) Instance {
	tb := textbox.NewInstance(
		b.atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(width, skillTypeHeight,
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
			tbcfg.RenderFrom(gfx.BottomLeft),
		))
	return &skillType{
		textbox: tb,
		content: tb.NewSimpleContent(strings.Replace(name, "k", "K", -1)),
		frame:   frames.New("common/rounded_frame_1px", b.atlas),
		fg:      fg,
		bg:      bg,
	}
}
