package runtime

import (
	"bytes"
	"image"
	"image/png"
	"log"

	"github.com/gopxl/pixel/v2/backends/opengl"
	"golang.design/x/clipboard"
)

func CopyCanvasToClipboard(c *opengl.Canvas) {
	b := c.Bounds()
	w, h := int(b.W()), int(b.H())

	src := c.Pixels() // len should be w*h*4, RGBA
	if len(src) != w*h*4 {
		log.Printf("unexpected pixel buffer size: got %d, want %d", len(src), w*h*4)
		return
	}

	// If your Pixels() comes bottom-to-top, flip; otherwise you can skip this block.
	row := w * 4
	dst := make([]byte, len(src))
	for y := 0; y < h; y++ {
		srcOff := (h - 1 - y) * row // flip vertically
		dstOff := y * row
		copy(dst[dstOff:dstOff+row], src[srcOff:srcOff+row])
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, dst) // or copy(img.Pix, src) if no flip needed

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Println("png encode:", err)
		return
	}

	if err := clipboard.Init(); err != nil {
		log.Println("clipboard init:", err)
		return
	}
	clipboard.Write(clipboard.FmtImage, buf.Bytes()) // PNG bytes
}
