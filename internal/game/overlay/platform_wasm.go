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

// PhysicalDPR returns the effective DPR that the WASM backend's syncCanvasSize
// chose for this device. Window.Bounds() reports physical pixels at this scale,
// so layout math that needs CSS pixels must divide by this value.
// Mirrors the effectiveDPR logic in the pixel fork so the two stay in sync.
func PhysicalDPR() float64 {
	native := js.Global().Get("devicePixelRatio").Float()
	if native < 1 {
		native = 1
	}
	max := opengl.MaxDevicePixelRatio
	if max <= 0 {
		max = 3
	}
	limit := int(math.Min(math.Floor(native), math.Floor(max)))
	for d := limit; d >= 1; d-- {
		ratio := native / float64(d)
		if math.Abs(ratio-math.Round(ratio)) < 0.1 {
			return float64(d)
		}
	}
	return 1
}

func physicalDPR() float64 { return PhysicalDPR() }

// SupportsFullscreen reports whether the browser supports the Fullscreen API.
// iOS Safari does not support it.
func SupportsFullscreen() bool {
	return js.Global().Get("document").Get("fullscreenEnabled").Bool()
}
