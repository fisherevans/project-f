package badges

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

var (
	buttonBadgeHeight  = 9
	buttonBadgeSpacing = 2
	buttonTextPadding  = 1
)

type ButtonAction struct {
	frame                                    *frames.Instance
	textbox                                  *textbox.Instance
	buttonContent, actionContent             *textbox.Content
	actionColor, buttonColor, highlightColor pixel.RGBA
	highlighted                              bool
}

func (s *ButtonAction) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight))
}

func (s *ButtonAction) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation) {
	buttonContentW := s.buttonContent.Width()
	actionContentW := s.actionContent.Width()
	buttonFrameW := buttonBadgeSpacing + buttonTextPadding + buttonContentW + buttonTextPadding + buttonBadgeSpacing
	fullFrameW := buttonFrameW + buttonBadgeSpacing + actionContentW + buttonTextPadding + buttonBadgeSpacing
	matrix = alignFromBottomLeft(matrix, origin, fullFrameW, buttonBadgeHeight)
	s.frame.Draw(target, pixel.R(0, 0, float64(fullFrameW), float64(buttonBadgeHeight-2)),
		matrix.Moved(gfx.IVec(0, 1)),
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(s.actionColor))
	if s.highlighted {
		s.frame.Draw(target, pixel.R(0, 0, float64(buttonFrameW+2), float64(buttonBadgeHeight+2)),
			matrix.Moved(pixel.V(-1, -1)),
			frames.WithRenderOrigin(gfx.BottomLeft),
			frames.WithColor(s.highlightColor))
	}
	s.frame.Draw(target, pixel.R(0, 0, float64(buttonFrameW), float64(buttonBadgeHeight)), matrix,
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(s.buttonColor))
	s.textbox.Render(ctx, target, matrix.Moved(gfx.IVec(buttonBadgeSpacing+buttonTextPadding, buttonBadgeSpacing)), s.buttonContent, tbcfg.Foreground(s.actionColor))
	s.textbox.Render(ctx, target, matrix.Moved(gfx.IVec(buttonFrameW+buttonBadgeSpacing, buttonBadgeSpacing)), s.actionContent, tbcfg.Foreground(s.buttonColor))
}

func (b *Builder) ButtonAction(button, action string) *ButtonAction {
	tb := textbox.NewInstance(
		b.atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(0, buttonBadgeHeight,
			tbcfg.HAligned(tbcfg.AlignLeft),
			tbcfg.VAligned(tbcfg.AlignBottom),
			tbcfg.RenderFrom(gfx.BottomLeft),
		))
	return &ButtonAction{
		frame:          frames.New("common/rounded_frame_2px", b.atlas),
		textbox:        tb,
		buttonContent:  tb.NewSimpleContent(button),
		actionContent:  tb.NewSimpleContent(action),
		actionColor:    colors.White.RGBA,
		buttonColor:    colors.Black.RGBA,
		highlightColor: colors.ButtonHighlight.RGBA,
	}
}

func (b *ButtonAction) Highlighted(highlighted bool) *ButtonAction {
	b.highlighted = highlighted
	return b
}
