package input

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"time"
)

type Controls struct {
	buttonA      *Button
	buttonB      *Button
	buttonStart  *Button
	buttonSelect *Button
	dpad         *DirectionalButton
}

func NewControls() *Controls {
	return &Controls{
		buttonA: &Button{
			sourceKeys: []pixel.Button{pixel.KeySpace, pixel.KeyEnter},
		},
		buttonB: &Button{
			sourceKeys: []pixel.Button{pixel.KeyLeftShift, pixel.KeyRightShift, pixel.KeyLeftControl},
		},
		buttonStart: &Button{
			sourceKeys: []pixel.Button{pixel.KeyEscape},
		},
		buttonSelect: &Button{
			sourceKeys: []pixel.Button{pixel.KeyTab, pixel.KeyLeftAlt},
		},
		dpad: &DirectionalButton{
			sourceKeys: map[pixel.Button]Direction{
				pixel.KeyUp:    Up,
				pixel.KeyDown:  Down,
				pixel.KeyLeft:  Left,
				pixel.KeyRight: Right,

				pixel.KeyW: Up,
				pixel.KeyA: Left,
				pixel.KeyS: Down,
				pixel.KeyD: Right,
			},
			lastJustPressed: map[pixel.Button]time.Time{},
		},
	}
}

func (c *Controls) Update(win *opengl.Window) {
	c.buttonA.updateButton(win)
	c.buttonB.updateButton(win)
	c.buttonStart.updateButton(win)
	c.buttonSelect.updateButton(win)
	c.dpad.updateDirectional(win)
}

func (c *Controls) ButtonA() *Button {
	return c.buttonA
}
func (c *Controls) ButtonB() *Button {
	return c.buttonB
}
func (c *Controls) ButtonStart() *Button {
	return c.buttonStart
}
func (c *Controls) ButtonSelect() *Button {
	return c.buttonSelect
}
func (c *Controls) DPad() *DirectionalButton {
	return c.dpad
}
