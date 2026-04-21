package overlay

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

// Pointer is a normalised view of mouse / single-touch input at window resolution.
type Pointer struct {
	Pos      pixel.Vec
	Down     bool // left button / finger currently held
	JustDown bool // became down this frame
	JustUp   bool // released this frame
}

func pollPointer(win *opengl.Window, prev Pointer) Pointer {
	wasDown := prev.Down
	isDown := win.Pressed(pixel.MouseButtonLeft)
	return Pointer{
		Pos:      win.MousePosition(),
		Down:     isDown,
		JustDown: !wasDown && isDown,
		JustUp:   wasDown && !isDown,
	}
}
