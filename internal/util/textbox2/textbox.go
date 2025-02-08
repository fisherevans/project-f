package textbox2

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"image/color"
	"math/rand"
	"strings"
)

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

type ExpandMode int

const (
	ExpandFull ExpandMode = iota
	ExpandFit
)

type Font struct {
	atlas        *text.Atlas
	letterHeight int
	lineSpacing  int
	tailHeight   int
}

var FontSmall = Font{
	atlas:        resources.Fonts.M6.Atlas,
	letterHeight: 6,
	lineSpacing:  2,
	tailHeight:   2,
}

var FontLarge = Font{
	atlas:        resources.Fonts.M7.Atlas,
	letterHeight: 7,
	lineSpacing:  3,
	tailHeight:   2,
}

type Padding struct {
	paddingX int
	paddingY int
}

var PaddingNarrow = Padding{
	paddingX: 4,
	paddingY: 2,
}

var PaddingNormal = Padding{
	paddingX: 5,
	paddingY: 3,
}

type Instance struct {
	Font
	text *text.Text
	imd  *imdraw.IMDraw

	maxWidth   int
	alignment  Alignment
	expandMode ExpandMode

	Padding

	foreground color.Color
	background color.Color
}

func NewInstance(font Font, padding Padding, maxWidth int, alignment Alignment, expandMode ExpandMode) *Instance {
	return &Instance{
		Font: font,
		text: text.New(pixel.ZV, font.atlas),
		imd:  imdraw.New(nil),

		maxWidth:   maxWidth,
		alignment:  alignment,
		expandMode: expandMode,

		Padding: padding,

		foreground: pixel.RGB(.1, .2, .2),
		background: pixel.RGB(.6, .9, .9),
	}
}

type RenderEffect interface {
	ColorOverride() *color.Color
	RenderDelta() pixel.Vec
	Update(ctx *game.Context, timeDelta float64)
}

type rumbleRenderEffect struct {
	rate    float64
	elapsed float64
	dx      int
	dy      int
}

func newRumble(rate float64) *rumbleRenderEffect {
	return &rumbleRenderEffect{
		rate: rate,
	}
}

func (r *rumbleRenderEffect) ColorOverride() *color.Color {
	return nil
}

func (r *rumbleRenderEffect) RenderDelta() pixel.Vec {
	return pixel.V(float64(r.dx), float64(r.dy))
}

func (r *rumbleRenderEffect) Update(ctx *game.Context, timeDelta float64) {
	r.elapsed += timeDelta
	for r.elapsed > r.rate {
		r.elapsed -= r.rate
		r.dy = rand.Intn(3) - 1
		r.dx = rand.Intn(3) - 1
	}
	ctx.DebugBR("delta: %d, %d", r.dx, r.dy)
}

type character struct {
	c      byte
	color  color.Color
	effect RenderEffect

	width int
}

func newCharacter(char byte, color color.Color, effect RenderEffect, text *text.Text) *character {
	return &character{
		c:      char,
		color:  color,
		effect: effect,
		width:  int(text.BoundsOf(string(char)).W()),
	}
}

type cgroup struct {
	characters []*character
	isSpacer   bool
}

func newSpaceCgroup(text *text.Text) *cgroup {
	cg := &cgroup{
		isSpacer: true,
	}
	cg.append(newCharacter(' ', color.Transparent, nil, text))
	return cg
}

func newSimpleCgroup(word string, color color.Color, text *text.Text) *cgroup {
	cg := &cgroup{}
	var effect RenderEffect
	if len(word) == 4 {
		effect = newRumble(0.1)
	}
	for _, ch := range []byte(word) {
		cg.append(newCharacter(ch, color, effect, text))
	}
	return cg
}

func (w *cgroup) append(c *character) {
	w.characters = append(w.characters, c)
}

func (w *cgroup) asString() string {
	var str []byte
	for _, c := range w.characters {
		str = append(str, c.c)
	}
	return string(str)
}

type phrase struct {
	cgroups []*cgroup
}

func (p *phrase) append(cg *cgroup) {
	p.cgroups = append(p.cgroups, cg)
}

type line struct {
	phrase *phrase
	text   string
	width  int
}

func newLine() *line {
	return &line{
		phrase: &phrase{},
	}
}

type Content struct {
	lines   []*line
	effects []RenderEffect
}

func (l *Content) Update(ctx *game.Context, timeDelta float64) {
	for _, e := range l.effects {
		e.Update(ctx, timeDelta)
	}
}

func (c *Content) registerEffect(effect RenderEffect) {
	for _, e := range c.effects {
		if e == effect {
			return
		}
	}
	c.effects = append(c.effects, effect)
}

func (c *Content) appendLine(l *line) {
	c.lines = append(c.lines, l)
	for _, cg := range l.phrase.cgroups {
		for _, ch := range cg.characters {
			if ch.effect == nil {
				continue
			}
			c.registerEffect(ch.effect)
		}
	}
}

func newContent(p *phrase, text *text.Text, maxWidth int) *Content {
	content := &Content{}
	var pendingSpaces []*cgroup
	pendingSpacesText := ""
	currentLine := newLine()
	for _, cg := range p.cgroups {
		if cg.isSpacer {
			pendingSpaces = append(pendingSpaces, cg)
			pendingSpacesText += cg.asString()
			continue
		}
		cgString := cg.asString()
		pendingLineText := currentLine.text + pendingSpacesText + cgString
		if int(text.BoundsOf(pendingLineText).W()) > maxWidth {
			pendingSpaces = nil
			pendingLineText = ""
			currentLine.width = int(text.BoundsOf(currentLine.text).W())
			content.appendLine(currentLine)
			currentLine = newLine()
		}
		if len(pendingSpaces) > 0 {
			for _, space := range pendingSpaces {
				currentLine.phrase.append(space)
			}
			currentLine.text += pendingSpacesText
			pendingSpaces = nil
			pendingSpacesText = ""
		}
		currentLine.phrase.append(cg)
		currentLine.text += cgString
	}
	currentLine.width = int(text.BoundsOf(currentLine.text).W())
	content.appendLine(currentLine)
	return content
}

func (tb *Instance) NewRainbowPhrase(msg string) *Content {
	p := &phrase{}
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			p.append(newSpaceCgroup(tb.text))
		}
		p.append(newSimpleCgroup(part, util.StringToColor(part, 0.95, 0.35), tb.text))
	}
	return newContent(p, tb.text, tb.maxWidth)
}

func (tb *Instance) NewSimplePhrase(msg string, color color.Color) *Content {
	p := &phrase{}
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			p.append(newSpaceCgroup(tb.text))
		}
		p.append(newSimpleCgroup(part, color, tb.text))
	}
	return newContent(p, tb.text, tb.maxWidth)
}

var borderWidth = 2

func (tb *Instance) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, content *Content) {
	tb.text.Clear()

	maxTextWidth := tb.maxWidth - tb.paddingX*2 - borderWidth*2/2

	maxLineWidth := maxTextWidth
	if tb.expandMode == ExpandFit {
		maxLineWidth = 0
		for _, line := range content.lines {
			if line.width > maxLineWidth {
				maxLineWidth = line.width
			}
		}
	}

	boxHeight := (len(content.lines) * tb.letterHeight) + ((len(content.lines) - 1) * tb.lineSpacing) + tb.tailHeight*2
	boxPadding := pixel.V(float64(tb.paddingX), float64(tb.paddingY))
	boxBounds := pixel.Rect{
		Min: pixel.ZV.Sub(boxPadding),
		Max: pixel.V(float64(maxLineWidth), float64(boxHeight)).Add(boxPadding),
	}

	matrix = matrix.Moved(pixel.V(float64(-maxLineWidth/2), float64(borderWidth+tb.tailHeight)).Floor())

	tb.imd.Clear()

	tb.imd.Color = tb.background
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(0)

	tb.imd.Color = tb.foreground
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(float64(borderWidth))

	tb.imd.Draw(target)

	for lineId, line := range content.lines {
		y := float64(((len(content.lines) - 1 - lineId) * (tb.letterHeight + tb.lineSpacing)) + tb.tailHeight)
		var x int
		switch tb.alignment {
		case AlignLeft:
			x = 0
		case AlignCenter:
			x = (maxLineWidth - line.width) / 2
		case AlignRight:
			x = line.width - maxLineWidth
		}
		tb.text.Dot = pixel.V(float64(x), y)
		for _, cg := range line.phrase.cgroups {
			for _, c := range cg.characters {
				tb.text.Color = c.color
				if c.effect != nil {
					tb.text.Dot = tb.text.Dot.Add(c.effect.RenderDelta())
					if colorOverride := c.effect.ColorOverride(); colorOverride != nil {
						tb.text.Color = *colorOverride
					}
				}
				tb.text.WriteByte(c.c)
				if c.effect != nil {
					tb.text.Dot = tb.text.Dot.Sub(c.effect.RenderDelta())
				}
				x += c.width
			}
		}
		tb.text.Draw(target, matrix)
	}

}
