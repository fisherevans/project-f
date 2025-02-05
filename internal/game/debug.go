package game

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

type Notification struct {
	text        string
	elapsed     float64
	displayTime float64
}

type DebugInfo struct {
	lines         map[DebugArea][]string
	notifications []*Notification
}

func (d *DebugInfo) DebugTL(format string, a ...any) {
	d.Debug(AreaTopLeft, format, a...)
}

func (d *DebugInfo) DebugBL(format string, a ...any) {
	d.Debug(AreaBottomLeft, format, a...)
}

func (d *DebugInfo) DebugTR(format string, a ...any) {
	d.Debug(AreaTopRight, format, a...)
}

func (d *DebugInfo) DebugBR(format string, a ...any) {
	d.Debug(AreaBottomRight, format, a...)
}

func (d *DebugInfo) Debug(area DebugArea, format string, a ...any) {
	if d.lines == nil {
		d.lines = map[DebugArea][]string{}
	}
	if _, ok := d.lines[area]; !ok {
		d.lines[area] = []string{}
	}
	d.lines[area] = append(d.lines[area], fmt.Sprintf(format, a...))
}

func (d *DebugInfo) PopDebugLines() map[DebugArea][]string {
	out := d.lines
	d.lines = nil
	return out
}

func (d *DebugInfo) Notify(format string, args ...any) {
	d.notifications = append(d.notifications, &Notification{
		text:        fmt.Sprintf(format, args...),
		elapsed:     0,
		displayTime: 5,
	})
}

func (d *DebugInfo) PopNotifications(elapsed float64) []string {
	var texts []string
	filtered := d.notifications[:0] // Reuse the same slice memory
	for _, notification := range d.notifications {
		texts = append(texts, notification.text)
		notification.elapsed += elapsed
		if notification.elapsed < notification.displayTime {
			filtered = append(filtered, notification)
		}
	}
	d.notifications = filtered
	return texts
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

func RenderNotifications(win *opengl.Window, notifications []string) {
	if len(notifications) == 0 {
		return
	}
	for id, notification := range notifications {
		debugText.Clear()
		debugText.WriteString(notification + "\n")
		dy := (debugPadding + debugText.LineHeight) * (float64(id) + 1)
		debugText.Draw(win, pixel.IM.Moved(pixel.V((win.Bounds().W()-debugText.Bounds().W())/2, win.Bounds().H()-dy)))
		dy += debugText.LineHeight * 1.5
	}
}
