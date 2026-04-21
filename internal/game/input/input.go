package input

import (
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

// VirtualState carries one frame of on-screen gamepad input. Set via
// Controls.SetVirtual before each Controls.Update call.
type VirtualState struct {
	A, B, Start, Select bool
	Dir                 Direction
}

type Controls struct {
	buttonA      *Button
	buttonB      *Button
	buttonStart  *Button
	buttonSelect *Button
	dpad         *DirectionalButton

	virtual     VirtualState
	prevVirtual VirtualState
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

// SetVirtual stores virtual gamepad state for the next Update call. Call this
// every frame before Update; the zero value clears all virtual input.
func (c *Controls) SetVirtual(v VirtualState) {
	c.virtual = v
}

func (c *Controls) Update(win *opengl.Window) {
	c.buttonA.updateButton(win)
	c.buttonB.updateButton(win)
	c.buttonStart.updateButton(win)
	c.buttonSelect.updateButton(win)
	c.dpad.updateDirectional(win)

	c.buttonA.applyVirtual(c.virtual.A, c.prevVirtual.A)
	c.buttonB.applyVirtual(c.virtual.B, c.prevVirtual.B)
	c.buttonStart.applyVirtual(c.virtual.Start, c.prevVirtual.Start)
	c.buttonSelect.applyVirtual(c.virtual.Select, c.prevVirtual.Select)
	c.dpad.applyVirtualDir(c.virtual.Dir, c.prevVirtual.Dir)

	c.prevVirtual = c.virtual
	c.virtual = VirtualState{} // cleared each frame; caller must re-set
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
