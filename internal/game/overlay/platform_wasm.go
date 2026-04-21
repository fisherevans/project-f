//go:build js && wasm

package overlay

import (
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
