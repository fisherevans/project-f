package badges

import (
	"strings"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
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
	color   pixel.RGBA
}

func (s *skillType) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight))
}

func (s *skillType) Render(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation) {
	matrix = matrix.Moved(origin.AlignFrom(gfx.BottomLeft, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight)))
	s.frame.Draw(target, pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight)), matrix,
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(s.color))
	s.textbox.Render(target, matrix, s.content)
}

func (b *Builder) Of(name string, c pixel.RGBA, width int) Instance {
	tb := textbox.NewInstance(
		b.atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(width, skillTypeHeight,
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
			tbcfg.Foreground(colors.ScaleColor(c, 0.25)),
			tbcfg.RenderFrom(gfx.BottomLeft),
		))
	return &skillType{
		textbox: tb,
		content: tb.NewSimpleContent(strings.Replace(name, "k", "K", -1)),
		frame:   frames.New("common/rounded_frame_1px", b.atlas),
		color:   c,
	}
}
