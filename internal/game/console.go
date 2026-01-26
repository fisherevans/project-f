package game

import (
	"fmt"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/rs/zerolog/log"
)

type CommandConsole struct {
	input   string
	buffer  []string
	handler func(string)
	active  bool
	dx      float64
}

func newConsole(handler func(string)) *CommandConsole {
	return &CommandConsole{
		handler: handler,
	}
}

func (c *CommandConsole) IsActive() bool {
	return c.active
}

func (c *CommandConsole) Write(lines ...string) {
	c.buffer = append(c.buffer, lines...)
	for _, line := range lines {
		log.Info().Msg("console: " + line)
	}
}

func (c *CommandConsole) Writef(format string, args ...any) {
	c.Write(fmt.Sprintf(format, args...))
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
		return
	}
	c.input += win.Typed()
	if pressedOrRepeased(pixel.KeyBackspace) && len(c.input) > 0 {
		c.input = c.input[:len(c.input)-1]
	}
	if win.JustPressed(pixel.KeyEnter) {
		c.buffer = append(c.buffer, "")
		c.buffer = append(c.buffer, "> "+c.input)
		log.Info().Msg("command: " + c.input)
		c.handler(c.input)
		c.input = ""
	}

	if pressedOrRepeased(pixel.KeyRight) {
		c.dx -= 20
	}
	if pressedOrRepeased(pixel.KeyLeft) {
		c.dx += 20
	}

	imd := imdraw.New(nil)
	imd.Color = colors.WithAlpha(colors.Black.RGBA, 0.2)
	imd.Push(pixel.V(0, 0), pixel.V(win.Bounds().W(), win.Bounds().H()))
	imd.Rectangle(0)

	grey := pixel.RGB(0.75, 0.75, 0.75)
	white := pixel.RGB(1, 1, 1)
	y, lineHeight := 20, 18
	drawLine := func(s string, color pixel.RGBA) {
		x := 20 + c.dx
		debugText.Clear()
		debugText.WriteString(s)
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				debugText.DrawColorMask(win, pixel.IM.Moved(pixel.V(x+float64(dx), float64(y+dy))), pixel.RGB(0, 0, 0))
			}
		}
		debugText.DrawColorMask(win, pixel.IM.Moved(pixel.V(x, float64(y))), color)
		y += lineHeight
	}

	drawLine("$ "+c.input, white)
	drawLine("", grey)
	for x := len(c.buffer) - 1; x >= 0; x-- {
		drawLine(c.buffer[x], white)
	}
}
