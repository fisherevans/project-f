package menu

import (
	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
)

var (
	mainFrame *frames.Instance
	mainText  *textbox.Instance
)

type MenuItem interface {
	Height() int
	Render(leftCenter pixel.Matrix, target pixel.Target, isHighlighted bool)
	OnSelect()
	OnDPadLeft()
	OnDPadRight()
}

type Menu struct {
	state       *State
	items       []MenuItem
	highlighted int
}

func (m *Menu) Render(s *State, topLeft pixel.Matrix, target pixel.Target, targetBounds pixel.Rect) {
	totalHeight := menuPadding * 2
	maxWidth := menuContentWidth + menuPadding*2
	for index, i := range m.items {
		if index > 0 {
			totalHeight += itemPadding // item padding
		}
		totalHeight += i.Height()
		// todo consider changing max width
	}
	frameBounds := pixel.R(0, 0, float64(maxWidth+menuPadding*2), float64(totalHeight+menuPadding*2))
	mainFrame.Draw(target, frameBounds, topLeft)

	topLeft = topLeft.Moved(gfx.IVec(menuPadding, -menuPadding))
	for index, item := range m.items {
		itemLeftCenter := topLeft.Moved(gfx.IVec(0, -item.Height()/2))
		item.Render(itemLeftCenter, target, index == m.highlighted)
		topLeft = topLeft.Moved(gfx.IVec(0, -item.Height()-itemPadding))
	}
}

func (m *Menu) HandleInputs() {
	controls := game.Controls[*State]()
	if controls.ButtonB().JustPressed() {
		m.state.PopMenu()
	}
	if controls.ButtonA().JustPressed() {
		m.items[m.highlighted].OnSelect()
	}
	switch controls.DPad().JustPressedOrRepeatedDirection() {
	case input.Up:
		m.highlighted = max(0, m.highlighted-1)
	case input.Down:
		m.highlighted = min(len(m.items)-1, m.highlighted+1)
	case input.Left:
		m.items[m.highlighted].OnDPadLeft()
	case input.Right:
		m.items[m.highlighted].OnDPadRight()
	}
}
