package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
)

type Instance struct {
	Font
	text *text.Text
	imd  *imdraw.IMDraw
	cfg  *Config
}

func NewInstance(font Font, cfg Config) *Instance {
	return &Instance{
		Font: font,
		text: text.New(pixel.ZV, font.atlas),
		imd:  imdraw.New(nil),
		cfg:  &cfg,
	}
}

type characterRenderParams struct {
	drawDelta  pixel.Vec
	foreground pixel.RGBA
}

func (tb *Instance) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, content *Content) {
	tb.text.Clear()

	pageLines := content.pageLines()
	renderLineCount := len(pageLines)
	if tb.cfg.linesPerPage != 0 {
		renderLineCount = tb.cfg.linesPerPage
	}

	//matrix = matrix.Moved(pixel.V(float64(0), float64(tb.tailHeight)).Floor())

	tb.imd.Clear()

	scrollDy := int((content.scrollPosition - float64(content.startLine)) * float64(tb.letterHeight+tb.lineSpacing))

	for lineId, line := range pageLines {
		lineTypingProgress := 0
		y := float64(((renderLineCount - 1 - lineId) * (tb.letterHeight + tb.lineSpacing)) + tb.tailHeight + scrollDy)
		if tb.cfg.origin == TopLeft {
			y -= float64(content.height)
		}
		var x int
		alignment := tb.cfg.alignment
		if content.alignmentOverride != nil {
			alignment = *content.alignmentOverride
		}
		switch alignment {
		case AlignLeft:
			x = 0
		case AlignCenter:
			x = (content.width - line.width) / 2
		case AlignRight:
			x = content.width - line.width
		}
		tb.text.Dot = pixel.V(float64(x), y)
		for _, c := range line.characters {
			if lineTypingProgress >= line.typingDone {
				break
			}
			// effect
			renderParams := &characterRenderParams{
				foreground: tb.cfg.foreground,
			}
			if c.style.color != nil {
				renderParams.foreground = c.style.color.foreground
			}
			for _, effect := range c.style.effects {
				effect.Apply(renderParams)
			}
			tb.text.Dot = tb.text.Dot.Add(renderParams.drawDelta)
			// underline
			if c.style.underline != nil {
				start := matrix.Project(tb.text.Dot).Add(pixel.V(-1, -1))
				extraLength := 0.0
				if c.style.shadow != nil {
					extraLength = 1
				}
				end := start.Add(pixel.V(float64(c.width)+2+extraLength, 0))
				tb.imd.Color = c.style.underline.color
				tb.imd.Push(start, end)
				tb.imd.Line(1)
			}
			// shadow
			if c.style.shadow != nil {
				origDot := tb.text.Dot
				for dx := 0; dx <= 1; dx++ {
					for dy := 0; dy >= -1; dy-- {
						if dx == 0 && dy == 0 {
							continue
						}
						delta := pixel.V(float64(dx), float64(dy))
						tb.text.Dot = origDot.Add(delta)
						tb.text.Color = c.style.shadow.color
						tb.text.WriteByte(c.c)
					}
				}
				tb.text.Dot = origDot
			}
			// text
			tb.text.Color = renderParams.foreground
			tb.text.WriteByte(c.c)
			// undo effect delta
			tb.text.Dot = tb.text.Dot.Sub(renderParams.drawDelta)
			x += c.width
			lineTypingProgress += c.typingWeight
		}
	}

	tb.imd.Draw(target)
	tb.text.Draw(target, matrix)
}
