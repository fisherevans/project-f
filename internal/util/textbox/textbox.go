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

	scrollDy := int((content.scrollPosition - float64(content.startLine)) * float64(tb.capHeight+tb.lineSpacing))

	for lineId, line := range pageLines {
		lineTypingProgress := 0
		y := float64(((renderLineCount - 1 - lineId) * (tb.capHeight + tb.lineSpacing)) + tb.tailHeight + scrollDy)
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
		var underlineColor pixel.RGBA
		var underlineStart, underlineEnd *pixel.Vec
		drawUnderline := func() {
			if underlineStart == nil || underlineEnd == nil {
				return
			}
			tb.imd.Color = underlineColor
			tb.imd.Push(*underlineStart, *underlineEnd)
			tb.imd.Line(1)

		}
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
				// todo if outline, add 1px left and right to underline
				listCharStart := matrix.Project(tb.text.Dot).Add(pixel.V(-1, -1))
				extraLength := 0.0
				if c.style.shadow != nil {
					extraLength = 1
				}
				listCharEnd := listCharStart.Add(pixel.V(float64(c.width)+2+extraLength, 0))
				if underlineStart == nil {
					underlineStart = &listCharStart
					underlineColor = c.style.underline.color
				}
				underlineEnd = &listCharEnd
			} else if underlineStart != nil {
				drawUnderline()
			}
			type render struct {
				dx, dy int
				color  pixel.RGBA
			}
			var renders []render
			// outline
			if c.style.outline != nil {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						renders = append(renders, render{
							dx:    dx,
							dy:    dy,
							color: c.style.outline.color,
						})
					}
				}
			}
			// shadow
			if c.style.shadow != nil {
				for dx := 0; dx <= 1; dx++ {
					for dy := 0; dy >= -1; dy-- {
						if dx == 0 && dy == 0 {
							continue
						}
						renders = append(renders, render{
							dx:    dx,
							dy:    dy,
							color: c.style.shadow.color,
						})
					}
				}
			}
			// actual text
			renders = append(renders, render{
				color: renderParams.foreground,
			})
			origDot := tb.text.Dot
			for _, r := range renders {
				tb.text.Dot = origDot.Add(pixel.V(float64(r.dx), float64(r.dy)))
				tb.text.Color = r.color
				tb.text.WriteByte(c.c)
			}
			// todo do we need to reset the dot, or does adding the actual text at the end result in the desired result?
			// undo effect delta
			tb.text.Dot = tb.text.Dot.Sub(renderParams.drawDelta)
			x += c.width
			lineTypingProgress += c.typingWeight
		}
		drawUnderline()
	}

	tb.imd.Draw(target)
	tb.text.Draw(target, matrix)
}
