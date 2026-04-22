//go:build js && wasm

package overlay

import (
	"math"
	"syscall/js"

	"github.com/gopxl/pixel/v2/backends/opengl"
)

func setCursorVisible(_ *opengl.Window, visible bool) {
	fn := js.Global().Get("primortalSetCursorVisible")
	if fn.Truthy() {
		fn.Invoke(visible)
	}
}

// ApplyFullscreen calls the JS bridge registered by web/index.html with the
// desired fullscreen state. The JS side handles the actual API call.
func ApplyFullscreen(win *opengl.Window, fullscreen bool) {
	fn := js.Global().Get("primortalToggleFullscreen")
	if fn.Truthy() {
		fn.Invoke(fullscreen)
	}
}

// PhysicalDPR returns the DPR the WASM backend applied when sizing the canvas
// backing store: min(devicePixelRatio, MaxDevicePixelRatio). Window.Bounds()
// reports physical pixels at this scale, so layout math that needs CSS pixels
// must divide by this value.
func PhysicalDPR() float64 {
	native := js.Global().Get("devicePixelRatio").Float()
	if native < 1 {
		native = 1
	}
	max := opengl.MaxDevicePixelRatio
	if max <= 0 {
		max = 3
	}
	return math.Min(native, max)
}

func physicalDPR() float64 { return PhysicalDPR() }

// SupportsFullscreen reports whether the browser supports the Fullscreen API.
// iOS Safari does not support it.
func SupportsFullscreen() bool {
	return js.Global().Get("document").Get("fullscreenEnabled").Truthy()
}
