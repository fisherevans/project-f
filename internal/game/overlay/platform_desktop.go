//go:build !js

package overlay

import "github.com/gopxl/pixel/v2/backends/opengl"

func setCursorVisible(win *opengl.Window, visible bool) {
	win.SetCursorVisible(visible)
}

// ApplyFullscreen switches the window to/from the primary monitor.
func ApplyFullscreen(win *opengl.Window, fullscreen bool) {
	if fullscreen {
		win.SetMonitor(opengl.PrimaryMonitor())
	} else {
		win.SetMonitor(nil)
	}
}
