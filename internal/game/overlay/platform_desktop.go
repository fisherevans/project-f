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

// PhysicalDPR returns 1.0 on desktop. The pixel desktop backend reports logical
// pixels in Window.Bounds(), so no DPR correction is needed.
func PhysicalDPR() float64 { return 1.0 }

func physicalDPR() float64 { return 1.0 }

func SupportsFullscreen() bool { return true }
