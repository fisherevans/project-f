package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
)

/*
TODO
- support new line characters - when we do parsing
  - Hello {+s}World{-s}!
  - Hello {+u}World{-u}!
  - Hello {+c:red}World{-c}!
  - Hello {+g}World{-g}! - glossary short hand for multiple effects
  - Hello {+w:10}World{-w}!
  - Hel{+u}lo {+c:red}{+r}World{-} - reset all
  - Hel{+u}lo {+c:red,+r}Wor{-u}ld{-c,-r} - multi
- support underlined words
*/

type Instance struct {
	Font
	text *text.Text
	imd  *imdraw.IMDraw

	cfg *Config
}

func NewInstance(font Font, cfg Config) *Instance {
	return &Instance{
		Font: font,
		text: text.New(pixel.ZV, font.atlas),
		imd:  imdraw.New(nil),
		cfg:  &cfg,
	}
}

func (tb *Instance) maxTextWidth() int {
	return tb.cfg.maxWidth - tb.cfg.padding.x*2 - borderWidth*2/2
}

var borderWidth = 2

func (tb *Instance) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, content *Content) {
	tb.text.Clear()

	textWidth := content.maxLineWidth
	if tb.cfg.expandMode == ExpandFull {
		textWidth = tb.maxTextWidth()
	}

	pageLines := content.pageLines()
	renderLineCount := len(pageLines)
	if tb.cfg.linesPerPage != 0 {
		renderLineCount = tb.cfg.linesPerPage
	}

	boxHeight := (renderLineCount * tb.letterHeight) + ((renderLineCount - 1) * tb.lineSpacing) + tb.tailHeight*2
	boxPadding := pixel.V(float64(tb.cfg.padding.x), float64(tb.cfg.padding.y))
	boxBounds := pixel.Rect{
		Min: pixel.ZV.Sub(boxPadding),
		Max: pixel.V(float64(textWidth), float64(boxHeight)).Add(boxPadding),
	}

	matrix = matrix.Moved(pixel.V(float64(-textWidth/2), float64(borderWidth+tb.tailHeight)).Floor())

	tb.imd.Clear()

	tb.imd.Color = tb.cfg.background
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(0)

	tb.imd.Color = tb.cfg.foreground
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(float64(borderWidth))
	scrollDy := int((content.scrollPosition - float64(content.startLine)) * float64(tb.letterHeight+tb.lineSpacing))

	for lineId, line := range pageLines {
		lineTypingProgress := 0
		y := float64(((renderLineCount - 1 - lineId) * (tb.letterHeight + tb.lineSpacing)) + tb.tailHeight + scrollDy)
		var x int
		switch tb.cfg.alignment {
		case AlignLeft:
			x = 0
		case AlignCenter:
			x = (textWidth - line.width) / 2
		case AlignRight:
			x = line.width - textWidth
		}
		tb.text.Dot = pixel.V(float64(x), y)
		for _, cg := range line.cgroups {
			for _, c := range cg.characters {
				if lineTypingProgress >= line.typingDone {
					break
				}
				// effect
				foreground := tb.cfg.foreground
				if c.color != nil {
					foreground = c.color.foreground
				}
				if c.effect != nil {
					tb.text.Dot = tb.text.Dot.Add(c.effect.RenderDelta())
					if colorOverride := c.effect.ColorOverride(); colorOverride != nil {
						foreground = *colorOverride
					}
				}
				// underline
				if c.underline != nil {
					start := matrix.Project(tb.text.Dot).Add(pixel.V(-1, -1))
					extraLength := 0.0
					if c.shadow != nil {
						extraLength = 1
					}
					end := start.Add(pixel.V(float64(c.width)+2+extraLength, 0))
					tb.imd.Color = c.underline.color
					tb.imd.Push(start, end)
					tb.imd.Line(1)
				}
				// shadow
				if c.shadow != nil {
					origDot := tb.text.Dot
					for dx := 0; dx <= 1; dx++ {
						for dy := 0; dy >= -1; dy-- {
							if dx == 0 && dy == 0 {
								continue
							}
							delta := pixel.V(float64(dx), float64(dy))
							tb.text.Dot = origDot.Add(delta)
							tb.text.Color = c.shadow.color
							tb.text.WriteByte(c.c)
						}
					}
					tb.text.Dot = origDot
				}
				// text
				tb.text.Color = foreground
				tb.text.WriteByte(c.c)
				// undo effect delta
				if c.effect != nil {
					tb.text.Dot = tb.text.Dot.Sub(c.effect.RenderDelta())
				}
				x += c.width
				lineTypingProgress += c.typingWeight
			}
		}
	}

	tb.imd.Draw(target)
	tb.text.Draw(target, matrix)

	ctx.DebugBR("current page: %d (%.1f)", content.startLine, content.scrollPosition)
	ctx.DebugBR("last page: %d", content.lastStartLine())
	ctx.DebugBR("lines: %d", len(content.lines))

}
