package game

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type DebugToggleSystem struct {
	toggles map[pixel.Button]*DebugToggle
}

type DebugToggle struct {
	toggleState bool
	pressed     bool
	justPressed bool
	presses     int
}

func (dt *DebugToggle) ToggleState() bool {
	return dt.toggleState
}

func (dt *DebugToggle) Pressed() bool {
	return dt.pressed
}

func (dt *DebugToggle) JustPressed() bool {
	return dt.justPressed
}

func (dt *DebugToggle) Presses() int {
	return dt.presses
}

func (dt *DebugToggle) update(key pixel.Button, win *opengl.Window) {
	if win.JustPressed(key) {
		dt.justPressed = true
		dt.presses++
		dt.toggleState = !dt.toggleState
	} else {
		dt.justPressed = false
	}
	if win.Pressed(key) {
		dt.pressed = true
	} else {
		dt.pressed = false
	}
}

func newToggles() *DebugToggleSystem {
	dt := &DebugToggleSystem{
		toggles: map[pixel.Button]*DebugToggle{},
	}
	keys := []pixel.Button{
		pixel.KeyF1,
		pixel.KeyF2,
		pixel.KeyF3,
		pixel.KeyF4,
		pixel.KeyF5,
		pixel.KeyF6,
		pixel.KeyF7,
		pixel.KeyF8,
	}
	for _, key := range keys {
		dt.toggles[key] = &DebugToggle{}
	}
	return dt
}

func (dt *DebugToggleSystem) update(win *opengl.Window) {
	for key, toggle := range dt.toggles {
		toggle.update(key, win)
	}
}

func (dt *DebugToggleSystem) F1() *DebugToggle {
	return dt.toggles[pixel.KeyF1]
}
func (dt *DebugToggleSystem) F2() *DebugToggle {
	return dt.toggles[pixel.KeyF2]
}
func (dt *DebugToggleSystem) F3() *DebugToggle {
	return dt.toggles[pixel.KeyF3]
}
func (dt *DebugToggleSystem) F4() *DebugToggle {
	return dt.toggles[pixel.KeyF4]
}
func (dt *DebugToggleSystem) F5() *DebugToggle {
	return dt.toggles[pixel.KeyF5]
}
func (dt *DebugToggleSystem) F6() *DebugToggle {
	return dt.toggles[pixel.KeyF6]
}
func (dt *DebugToggleSystem) F7() *DebugToggle {
	return dt.toggles[pixel.KeyF7]
}
func (dt *DebugToggleSystem) F8() *DebugToggle {
	return dt.toggles[pixel.KeyF8]
}
