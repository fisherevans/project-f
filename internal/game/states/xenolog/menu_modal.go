package xenolog

import (
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type menuModal struct {
	screen  *screen.Instance[*State]
	title   string
	text    string
	options []modalOption

	selection    int
	elapsed      float64
	closing      bool
	closeElapsed float64

	onClose func()
}

type modalOption struct {
	label    string
	onSelect func()
}

func newConfirmationModal(screen *screen.Instance[*State], title, text string, onSelect func()) *menuModal {
	return newMenuModal(screen, title, text, modalOption{
		label: "cancel",
	}, modalOption{
		label:    "confirm",
		onSelect: onSelect,
	})
}

func newMenuModal(screen *screen.Instance[*State], title, text string, options ...modalOption) *menuModal {
	modal := &menuModal{
		screen:    screen,
		title:     title,
		text:      text,
		options:   options,
		selection: len(options) - 1,
	}
	return modal
}

func (m *menuModal) Enter() {
}

func (m *menuModal) OnTick(target pixel.Target, timeDelta float64) {
	// Handle input only when not closing
	if !m.closing {
		controls := game.Controls[*State]()
		if controls.ButtonB().JustPressed() {
			m.startClosing(nil)
		}
		if controls.ButtonA().JustPressed() {
			selectedOption := m.options[m.selection].onSelect
			m.startClosing(selectedOption)
		}
		if controls.DPad().JustPressedDirection() == input.Left {
			m.selection--
		}
		if controls.DPad().JustPressedDirection() == input.Right {
			m.selection++
		}
		m.selection = max(0, min(m.selection, len(m.options)-1))
	}

	animationTime := 0.3
	var progress float64
	var dy int
	var overlayAlpha float64

	if m.closing {
		// Animate out
		m.closeElapsed += timeDelta
		progress = interp.Smootherstep(min(m.closeElapsed/animationTime, 1.0))
		dy = -int(float64(screenHeight) * progress)
		overlayAlpha = 0.75 * (1.0 - progress)

		// Complete the close when animation finishes
		if progress >= 1.0 {
			m.screen.PopMenu(false)
			if m.onClose != nil {
				m.onClose()
			}
			return
		}
	} else {
		// Animate in
		m.elapsed += timeDelta
		progress = interp.Smootherstep(min(m.elapsed/animationTime, 1.0))
		dy = -int(float64(screenHeight) * (1.0 - progress))
		overlayAlpha = 0.75 * progress
	}

	m.screen.DrawLastScreen()
	gfx.DrawRect(atlas, target, gfx.Moved(screenWidth/2, screenHeight/2), gfx.Centered, screenWidth, screenHeight, colors.WithAlpha(screenColors.Dark, overlayAlpha))

	modalCenter := gfx.Moved(screenWidth/2, screenHeight/5*3+dy)
	titleContent := titleTextbox.NewSimpleContent(m.title)
	textContent := smallTextbox.NewComplexContent(m.text)

	type button struct {
		content *textbox.Content
	}
	buttonTextXPadding := 5
	buttonTextYPadding := 3
	buttonMargins := 5
	totalButtonWidth := 0
	var buttons []button
	for id, option := range m.options {
		if id > 0 {
			totalButtonWidth += buttonMargins
		}
		b := button{
			content: smallTextbox.NewSimpleContent(option.label),
		}
		totalButtonWidth += b.content.Width() + buttonTextXPadding*2
		buttons = append(buttons, b)
	}

	framePadding := 6
	frameWidth := max(titleContent.Width(), textContent.Width(), totalButtonWidth) + framePadding*2
	frameHeight := titleContent.Height() + textContent.Height() + buttons[0].content.Height() + buttonTextYPadding*2 + framePadding*4

	frameR := pixel.R(0, 0, float64(frameWidth), float64(frameHeight))
	frame4px.Draw(target, frameR, modalCenter, frames.WithColor(screenColors.Dark))
	frame4pxBorder.Draw(target, frameR, modalCenter, frames.WithColor(screenColors.Text))

	topCenter := modalCenter.Moved(gfx.IVec(0, frameHeight/2-framePadding))
	titleContent.Render(target, topCenter, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.Foreground(screenColors.Highlight))
	topCenter = topCenter.Moved(gfx.IVec(0, -titleContent.Height()-framePadding))
	textContent.Render(target, topCenter, tbcfg.RenderFrom(gfx.TopCenter), tbcfg.Foreground(screenColors.Text))
	topCenter = topCenter.Moved(gfx.IVec(0, -textContent.Height()-framePadding))

	buttonTopLeft := topCenter.Moved(gfx.IVec(-totalButtonWidth/2, 0))
	for id, b := range buttons {
		borderMask, bgMask, fgMask := screenColors.Clear, screenColors.Dark, screenColors.Text
		if id == m.selection {
			borderMask, bgMask, fgMask = screenColors.Dark, flashingHighlight(), screenColors.Dark
		}
		buttonFrameR := pixel.R(0, 0, float64(b.content.Width()+buttonTextXPadding*2), float64(b.content.Height()+buttonTextYPadding*2))
		frame2px.Draw(target, buttonFrameR, buttonTopLeft, frames.WithColor(bgMask), frames.WithRenderOrigin(gfx.TopLeft))
		frame2pxBorder.Draw(target, buttonFrameR, buttonTopLeft, frames.WithColor(borderMask), frames.WithRenderOrigin(gfx.TopLeft))
		b.content.Render(target, buttonTopLeft.Moved(gfx.IVec(buttonTextXPadding+b.content.Width()/2, -buttonTextYPadding)),
			tbcfg.RenderFrom(gfx.TopCenter), tbcfg.Foreground(fgMask))
		buttonTopLeft = buttonTopLeft.Moved(gfx.IVec(b.content.Width()+buttonTextXPadding*2+buttonMargins, 0))
	}
}

func (m *menuModal) startClosing(onClose func()) {
	m.closing = true
	m.onClose = onClose
}
