package game

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

type DebugToggles struct {
	toggles map[pixel.Button]*DebugToggle
}

type DebugToggle struct {
	toggleState bool
	pressed     bool
	justPressed bool
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

func (dt *DebugToggle) update(key pixel.Button, win *opengl.Window) {
	if win.JustPressed(key) {
		dt.justPressed = true
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

func newToggles() *DebugToggles {
	dt := &DebugToggles{
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

func (dt *DebugToggles) update(win *opengl.Window) {
	for key, toggle := range dt.toggles {
		toggle.update(key, win)
	}
}

func (dt *DebugToggles) F1() *DebugToggle {
	return dt.toggles[pixel.KeyF1]
}
func (dt *DebugToggles) F2() *DebugToggle {
	return dt.toggles[pixel.KeyF2]
}
func (dt *DebugToggles) F3() *DebugToggle {
	return dt.toggles[pixel.KeyF3]
}
func (dt *DebugToggles) F4() *DebugToggle {
	return dt.toggles[pixel.KeyF4]
}
func (dt *DebugToggles) F5() *DebugToggle {
	return dt.toggles[pixel.KeyF5]
}
func (dt *DebugToggles) F6() *DebugToggle {
	return dt.toggles[pixel.KeyF6]
}
func (dt *DebugToggles) F7() *DebugToggle {
	return dt.toggles[pixel.KeyF7]
}
func (dt *DebugToggles) F8() *DebugToggle {
	return dt.toggles[pixel.KeyF8]
}
