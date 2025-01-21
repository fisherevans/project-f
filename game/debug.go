package game

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

func (c *Context) DebugTL(format string, a ...any) {
	c.Debug(AreaTopLeft, format, a...)
}

func (c *Context) DebugBL(format string, a ...any) {
	c.Debug(AreaBottomLeft, format, a...)
}

func (c *Context) DebugTR(format string, a ...any) {
	c.Debug(AreaTopRight, format, a...)
}

func (c *Context) DebugBR(format string, a ...any) {
	c.Debug(AreaBottomRight, format, a...)
}

func (c *Context) Debug(area DebugArea, format string, a ...any) {
	if c.debugLines == nil {
		c.debugLines = map[DebugArea][]string{}
	}
	if _, ok := c.debugLines[area]; !ok {
		c.debugLines[area] = []string{}
	}
	c.debugLines[area] = append(c.debugLines[area], fmt.Sprintf(format, a...))
}

func (c *Context) PopDebugLines() map[DebugArea][]string {
	out := c.debugLines
	c.debugLines = nil
	return out
}

var debugText = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))
var debugPadding = 10.0

func RenderDebugLines(win *opengl.Window, areaLines map[DebugArea][]string) {
	for area, lines := range areaLines {
		debugText.Clear()
		for _, line := range lines {
			debugText.WriteString(line + "\n")
		}

		left := debugPadding
		right := win.Bounds().W() - debugText.Bounds().W() - debugPadding
		top := win.Bounds().H() - debugPadding - debugText.LineHeight
		bottom := debugText.Bounds().H() + debugPadding

		switch area {
		case AreaTopLeft:
			debugText.Draw(win, pixel.IM.Moved(pixel.V(left, top)))
		case AreaBottomLeft:
			debugText.Draw(win, pixel.IM.Moved(pixel.V(left, bottom)))
		case AreaTopRight:
			debugText.Draw(win, pixel.IM.Moved(pixel.V(right, top)))
		case AreaBottomRight:
			debugText.Draw(win, pixel.IM.Moved(pixel.V(right, bottom)))
		}
	}
}
