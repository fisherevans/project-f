//go:build js && wasm

package overlay

import (
	"syscall/js"

	"github.com/gopxl/pixel/v2/backends/opengl"
)

// ApplyFullscreen calls the JS bridge registered by web/index.html.
func ApplyFullscreen(win *opengl.Window, fullscreen bool) {
	fn := js.Global().Get("primortalToggleFullscreen")
	if fn.Truthy() {
		fn.Invoke(fullscreen)
	}
}

// installScaleBridge registers primortalSetScale so the overlay can drive the
// canvas size. Called from audio_wasm.go's init via resources.RunOnceInitialized.
func installScaleBridge() {
	js.Global().Set("primortalOnScaleChanged", js.FuncOf(func(_ js.Value, args []js.Value) any {
		// Go → JS direction: overlay calls JS to set canvas size
		return nil
	}))
}

// applyScale drives the canvas clientWidth/Height via the JS bridge.
func applyScale(n int, baseW, baseH int) {
	fn := js.Global().Get("primortalSetScale")
	if fn.Truthy() {
		fn.Invoke(n, baseW, baseH)
	}
}
