package game

import (
	"fmt"
	"math"
	"strings"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/rs/zerolog/log"
	"golang.design/x/clipboard"
)

type CommandConsole struct {
	input   string
	buffer  []string
	handler func(string)
	active  bool
	dx      float64

	history        []string
	historyPos     int
	scrollOffset   int
	bgOpacity      float64
	lastGlobalKeys []string

	cursorPos    int
	lineWrapping bool
}

func newConsole(handler func(string)) *CommandConsole {
	_ = clipboard.Init()
	return &CommandConsole{
		handler:      handler,
		bgOpacity:    0.2,
		lineWrapping: true,
	}
}

func (c *CommandConsole) IsActive() bool {
	return c.active
}

func (c *CommandConsole) SetLastGlobalKeys(keys []string) {
	c.lastGlobalKeys = keys
}

func (c *CommandConsole) GetLastGlobalKey(idx int) string {
	if idx < 1 || idx > len(c.lastGlobalKeys) {
		return ""
	}
	return c.lastGlobalKeys[idx-1]
}

func (c *CommandConsole) RunCapture(fn func()) []string {
	before := len(c.buffer)
	fn()
	if len(c.buffer) <= before {
		return nil
	}
	return c.buffer[before:]
}

func (c *CommandConsole) WriteLines(lines ...string) {
	for _, line := range lines {
		line = strings.ReplaceAll(line, "\t", "    ")
		splitLines := strings.Split(line, "\n")
		c.buffer = append(c.buffer, splitLines...)
		for _, sl := range splitLines {
			log.Info().Msg("console: " + sl)
		}
	}
}

func (c *CommandConsole) Write(p []byte) (n int, err error) {
	c.WriteLines(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

func (c *CommandConsole) Writef(format string, args ...any) {
	c.WriteLines(fmt.Sprintf(format, args...))
}

func (c *CommandConsole) OnTick(win *opengl.Window) {
	pressedOrRepeased := func(button pixel.Button) bool {
		return win.JustPressed(button) || win.Repeated(button)
	}
	if !c.active && win.JustPressed(pixel.KeySlash) {
		c.active = true
		return
	}
	if !c.active {
		return
	}
	if win.JustPressed(pixel.KeyEscape) {
		c.active = false
		c.input = ""
		c.cursorPos = 0
		return
	}

	runes := []rune(c.input)
	if c.cursorPos > len(runes) {
		c.cursorPos = len(runes)
	}
	if c.cursorPos < 0 {
		c.cursorPos = 0
	}

	shift := win.Pressed(pixel.KeyLeftShift) || win.Pressed(pixel.KeyRightShift)
	ctrl := win.Pressed(pixel.KeyLeftControl) || win.Pressed(pixel.KeyRightControl)
	cmd := win.Pressed(pixel.KeyLeftSuper) || win.Pressed(pixel.KeyRightSuper)

	// Clipboard paste
	if (ctrl || cmd) && win.JustPressed(pixel.KeyV) {
		pasted := string(clipboard.Read(clipboard.FmtText))
		runes := []rune(c.input)
		c.input = string(append(runes[:c.cursorPos], append([]rune(pasted), runes[c.cursorPos:]...)...))
		c.cursorPos += len([]rune(pasted))
	}

	// Clipboard copy
	if (ctrl || cmd) && win.JustPressed(pixel.KeyC) {
		fullBuffer := strings.Join(c.buffer, "\n")
		clipboard.Write(clipboard.FmtText, []byte(fullBuffer))
	}

	// Line wrapping toggle
	if (ctrl || cmd) && win.JustPressed(pixel.KeyW) {
		c.lineWrapping = !c.lineWrapping
	}

	typed := win.Typed()
	if typed != "" {
		runes := []rune(c.input)
		c.input = string(append(runes[:c.cursorPos], append([]rune(typed), runes[c.cursorPos:]...)...))
		c.cursorPos += len([]rune(typed))
	}

	if pressedOrRepeased(pixel.KeyBackspace) && c.cursorPos > 0 {
		runes := []rune(c.input)
		c.input = string(append(runes[:c.cursorPos-1], runes[c.cursorPos:]...))
		c.cursorPos--
	}

	if pressedOrRepeased(pixel.KeyDelete) && c.cursorPos < len([]rune(c.input)) {
		runes := []rune(c.input)
		c.input = string(append(runes[:c.cursorPos], runes[c.cursorPos+1:]...))
	}

	if win.JustPressed(pixel.KeyEnter) {
		if c.input != "" {
			c.history = append(c.history, c.input)
			c.historyPos = len(c.history)
		}
		c.buffer = append(c.buffer, "> "+c.input)
		log.Info().Msg("command: " + c.input)
		c.handler(c.input)
		c.input = ""
		c.cursorPos = 0
	}

	if pressedOrRepeased(pixel.KeyUp) {
		if shift {
			c.scrollOffset++
		} else {
			if c.historyPos > 0 {
				c.historyPos--
				c.input = c.history[c.historyPos]
				c.cursorPos = len([]rune(c.input))
			}
		}
	}
	if pressedOrRepeased(pixel.KeyDown) {
		if shift {
			if c.scrollOffset > 0 {
				c.scrollOffset--
			}
		} else {
			if c.historyPos < len(c.history)-1 {
				c.historyPos++
				c.input = c.history[c.historyPos]
				c.cursorPos = len([]rune(c.input))
			} else {
				c.historyPos = len(c.history)
				c.input = ""
				c.cursorPos = 0
			}
		}
	}

	if pressedOrRepeased(pixel.KeyRight) {
		if shift {
			c.bgOpacity = math.Min(1, c.bgOpacity+0.05)
		} else if cmd {
			c.cursorPos = len([]rune(c.input))
		} else if ctrl {
			runes := []rune(c.input)
			for c.cursorPos < len(runes) && runes[c.cursorPos] == ' ' {
				c.cursorPos++
			}
			for c.cursorPos < len(runes) && runes[c.cursorPos] != ' ' {
				c.cursorPos++
			}
		} else if c.cursorPos < len([]rune(c.input)) {
			c.cursorPos++
		}
	}
	if pressedOrRepeased(pixel.KeyLeft) {
		if shift {
			c.bgOpacity = math.Max(0, c.bgOpacity-0.05)
		} else if cmd {
			c.cursorPos = 0
		} else if ctrl {
			runes := []rune(c.input)
			for c.cursorPos > 0 && runes[c.cursorPos-1] == ' ' {
				c.cursorPos--
			}
			for c.cursorPos > 0 && runes[c.cursorPos-1] != ' ' {
				c.cursorPos--
			}
		} else if c.cursorPos > 0 {
			c.cursorPos--
		}
	}

	imd := imdraw.New(nil)
	imd.Color = colors.WithAlpha(colors.Black.RGBA, c.bgOpacity)
	imd.Push(pixel.V(0, 0), pixel.V(win.Bounds().W(), win.Bounds().H()))
	imd.Rectangle(0)
	imd.Draw(win)

	grey := pixel.RGB(0.75, 0.75, 0.75)
	white := pixel.RGB(1, 1, 1)
	y, lineHeight := 20.0, 18.0
	drawLine := func(s string, color pixel.RGBA) {
		x := 20.0
		debugText.Clear()
		debugText.WriteString(s)
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				debugText.DrawColorMask(win, pixel.IM.Moved(pixel.V(x+float64(dx), y+float64(dy))), pixel.RGB(0, 0, 0))
			}
		}
		debugText.DrawColorMask(win, pixel.IM.Moved(pixel.V(x, y)), color)
		y += lineHeight
	}

	// Render input with cursor
	debugText.Clear()
	debugText.WriteString("$ " + string([]rune(c.input)[:c.cursorPos]))
	cursorX := 20.0 + debugText.Bounds().W()

	drawLine("$ "+c.input, white)

	// Draw cursor
	if int(TimeElapsed()*2)%2 == 0 {
		cursorImd := imdraw.New(nil)
		cursorImd.Color = white
		cursorImd.Push(pixel.V(cursorX, y-lineHeight+2), pixel.V(cursorX+2, y-2))
		cursorImd.Rectangle(0)
		cursorImd.Draw(win)
	}

	drawLine("", grey)

	var displayBuffer []string
	if c.lineWrapping {
		maxWidth := win.Bounds().W() - 40
		charWidth := 7.0 // approximate for basicfont 7x13
		for _, line := range c.buffer {
			if line == "" {
				displayBuffer = append(displayBuffer, "")
				continue
			}
			wrapped := wrapText(line, maxWidth, charWidth)
			displayBuffer = append(displayBuffer, wrapped...)
		}
	} else {
		displayBuffer = c.buffer
	}

	startLine := c.scrollOffset
	for i := 0; i < len(displayBuffer)-startLine; i++ {
		idx := len(displayBuffer) - 1 - startLine - i
		if idx < 0 {
			break
		}
		drawLine(displayBuffer[idx], white)
		if y > win.Bounds().H() {
			break
		}
	}
}

func wrapText(s string, maxWidth, charWidth float64) []string {
	charsPerLine := int(maxWidth / charWidth)
	if charsPerLine <= 0 {
		return []string{s}
	}
	var lines []string
	runes := []rune(s)
	for len(runes) > charsPerLine {
		lines = append(lines, string(runes[:charsPerLine]))
		runes = runes[charsPerLine:]
	}
	lines = append(lines, string(runes))
	return lines
}
