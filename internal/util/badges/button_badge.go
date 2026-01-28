package badges

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type ButtonColorStyle struct {
	Action    pixel.RGBA
	Button    pixel.RGBA
	Highlight pixel.RGBA
}

var (
	ButtonStyleStandard = ButtonColorStyle{
		Action:    colors.White.RGBA,
		Button:    colors.Black.RGBA,
		Highlight: colors.ButtonHighlight.RGBA,
	}

	buttonBadgeHeight  = 9
	buttonBadgeSpacing = 2
	buttonTextPadding  = 1
)

type ButtonAction struct {
	frame                        *frames.Instance
	textbox                      *textbox.Instance
	buttonContent, actionContent *textbox.Content
	style                        ButtonColorStyle
	highlighted                  bool
	flipped                      bool
}

func (s *ButtonAction) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(s.textbox.GetConfig().BoxWidth), float64(skillTypeHeight))
}

func (s *ButtonAction) Render(target pixel.Target, matrix pixel.Matrix, origin gfx.OriginLocation) {
	buttonContentW := s.buttonContent.Width()
	actionContentW := s.actionContent.Width()
	buttonFrameW := buttonBadgeSpacing + buttonTextPadding + buttonContentW + buttonTextPadding + buttonBadgeSpacing
	actionFrameW := buttonBadgeSpacing + actionContentW + buttonTextPadding + buttonBadgeSpacing
	fullFrameW := buttonFrameW + actionFrameW
	matrix = matrix.Moved(origin.AlignFrom(gfx.BottomLeft, float64(fullFrameW), float64(buttonBadgeHeight)))

	var buttonFrameX, buttonTextX, actionTextX int
	if s.flipped {
		// Action on left, button on right
		buttonFrameX = actionFrameW
		actionTextX = buttonBadgeSpacing + buttonTextPadding
		buttonTextX = actionFrameW + buttonBadgeSpacing + buttonTextPadding
	} else {
		// Button on left, action on right (default)
		buttonFrameX = 0
		buttonTextX = buttonBadgeSpacing + buttonTextPadding
		actionTextX = buttonFrameW + buttonBadgeSpacing + buttonTextPadding - 1
	}

	// Draw action background (full width, slightly lower)
	s.frame.Draw(target, pixel.R(0, 0, float64(fullFrameW), float64(buttonBadgeHeight-2)),
		matrix.Moved(gfx.IVec(0, 1)),
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(s.style.Action))

	// Draw highlight if needed
	if s.highlighted {
		s.frame.Draw(target, pixel.R(0, 0, float64(buttonFrameW+2), float64(buttonBadgeHeight+2)),
			matrix.Moved(pixel.V(float64(buttonFrameX)-1, -1)),
			frames.WithRenderOrigin(gfx.BottomLeft),
			frames.WithColor(s.style.Highlight))
	}

	// Draw button frame
	s.frame.Draw(target, pixel.R(0, 0, float64(buttonFrameW), float64(buttonBadgeHeight)),
		matrix.Moved(gfx.IVec(buttonFrameX, 0)),
		frames.WithRenderOrigin(gfx.BottomLeft),
		frames.WithColor(s.style.Button))

	// Render button text
	s.buttonContent.Render(target,
		matrix.Moved(gfx.IVec(buttonTextX, buttonBadgeSpacing)),
		tbcfg.Foreground(s.style.Action))

	// Render action text
	s.actionContent.Render(target,
		matrix.Moved(gfx.IVec(actionTextX, buttonBadgeSpacing)),
		tbcfg.Foreground(s.style.Button))
}

func (b *Builder) ButtonAction(button, action string, style ButtonColorStyle) *ButtonAction {
	tb := textbox.NewInstance(
		b.atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(0, buttonBadgeHeight,
			tbcfg.HAligned(tbcfg.HAlignLeft),
			tbcfg.VAligned(tbcfg.VAlignBottom),
			tbcfg.RenderFrom(gfx.BottomLeft),
		))
	return &ButtonAction{
		frame:         frames.New("common/rounded_frame_2px", b.atlas),
		textbox:       tb,
		buttonContent: tb.NewSimpleContent(button),
		actionContent: tb.NewSimpleContent(action),
		style:         style,
	}
}

func (b *ButtonAction) Highlighted(highlighted bool) *ButtonAction {
	b.highlighted = highlighted
	return b
}

func (b *ButtonAction) Flipped() *ButtonAction {
	b.flipped = !b.flipped
	return b
}
