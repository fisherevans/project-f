package input

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type ButtonState struct {
	pressed     bool
	justPressed bool
	repeated    bool
}

func (b *ButtonState) String() string {
	return fmt.Sprintf("pressed: %t, just: %t, repeated: %t", b.pressed, b.justPressed, b.repeated)
}

func (b *ButtonState) updateState(win *opengl.Window, keys ...pixel.Button) {
	b.pressed = false
	b.justPressed = false
	b.repeated = false
	for _, key := range keys {
		if win.Pressed(key) {
			b.pressed = true
		}
		if win.JustPressed(key) {
			b.justPressed = true
		}
		if win.Repeated(key) {
			b.repeated = true
		}
	}
}

func (b *ButtonState) IsPressed() bool {
	return b.pressed
}

func (b *ButtonState) JustPressed() bool {
	return b.justPressed
}

func (b *ButtonState) JustRepeated() bool {
	return b.repeated
}

func (b *ButtonState) JustPressedOrRepeated() bool {
	return b.justPressed || b.repeated
}

type Button struct {
	ButtonState
	sourceKeys []pixel.Button
}

func (b *Button) updateButton(win *opengl.Window) {
	b.ButtonState.updateState(win, b.sourceKeys...)
}
