package runtime

import (
	"bytes"
	"fmt"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"golang.design/x/clipboard"
	"image"
	"image/png"
	"log"
)

// CanvasToImage reads a canvas's pixels into a top-down RGBA image. Must be
// called on the game (GL) thread, since it reads GPU pixels.
func CanvasToImage(c *opengl.Canvas) (*image.RGBA, error) {
	b := c.Bounds()
	w, h := int(b.W()), int(b.H())

	src := c.Pixels() // len should be w*h*4, RGBA
	if len(src) != w*h*4 {
		return nil, fmt.Errorf("unexpected pixel buffer size: got %d, want %d", len(src), w*h*4)
	}

	// Pixels() is bottom-to-top; flip to top-down for image encoding.
	row := w * 4
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * row
		dstOff := y * row
		copy(img.Pix[dstOff:dstOff+row], src[srcOff:srcOff+row])
	}
	return img, nil
}

// CanvasToPNG encodes a canvas to PNG bytes. Must be called on the game thread.
func CanvasToPNG(c *opengl.Canvas) ([]byte, error) {
	img, err := CanvasToImage(c)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return buf.Bytes(), nil
}

func CopyCanvasToClipboard(c *opengl.Canvas) {
	pngBytes, err := CanvasToPNG(c)
	if err != nil {
		log.Println("canvas capture:", err)
		return
	}
	if err := clipboard.Init(); err != nil {
		log.Println("clipboard init:", err)
		return
	}
	clipboard.Write(clipboard.FmtImage, pngBytes) // PNG bytes
}
