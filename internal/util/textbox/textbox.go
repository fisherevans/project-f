package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"image/color"
	"math"
	"strings"
)

/*
TODO
- organize functions and files
- support new line characters
- support underlined words
- support text shadows
-
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

type character struct {
	c            byte
	color        color.Color
	effect       RenderEffect
	typingWeight int
	width        int
}

func newCharacter(char byte, color color.Color, effect RenderEffect, typingWeight int, text *text.Text) *character {
	return &character{
		c:            char,
		color:        color,
		effect:       effect,
		typingWeight: typingWeight,
		width:        int(text.BoundsOf(string(char)).W()),
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
	cg.append(newCharacter(' ', color.Transparent, nil, 1, text))
	return cg
}

func newSimpleCgroup(word string, color color.Color, text *text.Text) *cgroup {
	cg := &cgroup{}
	var effect RenderEffect
	if len(word) == 4 {
		effect = newRumble(0.1)
	}
	for _, ch := range []byte(word) {
		weight := 1
		switch ch {
		case ' ':
			weight = 0
		case '.':
			weight = 12
		case ',':
			weight = 6
		}
		cg.append(newCharacter(ch, color, effect, weight, text))
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

type line struct {
	cgroups []*cgroup
	text    string
	width   int
}

func newLine() *line {
	return &line{}
}

func (l *line) append(cg *cgroup) {
	l.cgroups = append(l.cgroups, cg)
}

type Content struct {
	tb               *Instance
	lines            []*line
	effects          []RenderEffect
	maxLineWidth     int
	typingController TypingController

	currentPage                  int
	currentPageTypeFaster        bool
	currentPageTypingDone        int
	currentPageTotalTypingWeight int
}

type defaultTypingController struct {
	elapsed          float64
	timePerCharacter float64
	fasterScale      float64
}

func (c *defaultTypingController) UpdateTypingProgress(timeDelta float64, typeFaster bool) int {
	if typeFaster {
		timeDelta = timeDelta * c.fasterScale
	}
	c.elapsed += timeDelta
	return int(c.elapsed/c.timePerCharacter) + 1
}

func (c *defaultTypingController) Reset() {
	c.elapsed = 0
}

func (c *Content) TypeFaster() {
	c.currentPageTypeFaster = true
}

func WithTyping(timePerCharacter float64) ContentOpt {
	return func(c *Content) {
		c.typingController = &defaultTypingController{
			timePerCharacter: timePerCharacter,
			fasterScale:      4,
		}
	}
}

func (c *Content) ContentFullyDisplayed() bool {
	if c.tb.cfg.lineCount != 0 && c.currentPage != c.lastPage() {
		return false
	}
	return c.PageFullyDisplayed()
}

func (c *Content) PageFullyDisplayed() bool {
	if c.typingController == nil {
		return true
	}
	return c.currentPageTypingDone >= c.currentPageTotalTypingWeight
}

type TypingController interface {
	UpdateTypingProgress(elapsed float64, typeFaster bool) int
	Reset()
}

func (l *Content) Update(ctx *game.Context, timeDelta float64) {
	if l.typingController == nil {
		l.currentPageTypingDone = l.currentPageTotalTypingWeight
	} else {
		l.currentPageTypingDone = util.MinInt(l.typingController.UpdateTypingProgress(timeDelta, l.currentPageTypeFaster), l.currentPageTotalTypingWeight)
	}
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
	for _, cg := range l.cgroups {
		for _, ch := range cg.characters {
			if ch.effect == nil {
				continue
			}
			c.registerEffect(ch.effect)
		}
	}
	if l.width > c.maxLineWidth {
		c.maxLineWidth = l.width
	}
}

func (c *Content) NextPage() {
	var page int
	if c.tb.cfg.lineCount != 0 {
		page = util.MinInt(c.lastPage(), c.currentPage+1)
	}
	c.setPage(page)
}

func (c *Content) lastPage() int {
	if c.tb.cfg.lineCount == 0 {
		return 0
	}
	return int(math.Ceil(float64(len(c.lines))/float64(c.tb.cfg.lineCount))) - 1
}

func (c *Content) setPage(page int) {
	if c.tb.cfg.lineCount == 0 {
		page = 0
	}
	if c.typingController != nil {
		c.typingController.Reset()
	}
	c.currentPage = page
	c.currentPageTotalTypingWeight = 0
	c.currentPageTypeFaster = false
	for _, line := range c.pageLines() {
		for _, cg := range line.cgroups {
			for _, ch := range cg.characters {
				c.currentPageTotalTypingWeight += ch.typingWeight
			}
		}
	}
}

func (c *Content) pageLineIdRange() (int, int) {
	if c.tb.cfg.lineCount == 0 {
		return 0, len(c.lines) - 1
	}
	lastIndex := len(c.lines) - 1
	startLine := util.MinInt(c.currentPage*c.tb.cfg.lineCount, lastIndex)
	endLine := util.MinInt(startLine+c.tb.cfg.lineCount, lastIndex+1)
	return startLine, endLine
}

func (c *Content) pageLines() []*line {
	startLine, endLine := c.pageLineIdRange()
	return c.lines[startLine:endLine]
}

type ContentOpt func(content *Content)

func (tb *Instance) newContent(cgroups []*cgroup, opts ...ContentOpt) *Content {
	content := &Content{
		tb: tb,
	}
	var pendingSpaces []*cgroup
	pendingSpacesText := ""
	currentLine := newLine()
	for _, cg := range cgroups {
		cgString := cg.asString()
		if cg.isSpacer {
			pendingSpaces = append(pendingSpaces, cg)
			pendingSpacesText += cgString
			continue
		}
		pendingLineText := currentLine.text + pendingSpacesText + cgString
		if int(tb.text.BoundsOf(pendingLineText).W()) > tb.maxTextWidth() {
			pendingSpaces = nil
			pendingSpacesText = ""
			pendingLineText = ""
			currentLine.width = int(tb.text.BoundsOf(currentLine.text).W())
			content.appendLine(currentLine)
			currentLine = newLine()
		}
		if len(pendingSpaces) > 0 {
			for _, space := range pendingSpaces {
				currentLine.append(space)
			}
			currentLine.text += pendingSpacesText
			pendingSpaces = nil
			pendingSpacesText = ""
		}
		currentLine.append(cg)
		currentLine.text += cgString
	}
	currentLine.width = int(tb.text.BoundsOf(currentLine.text).W())
	content.appendLine(currentLine)

	content.setPage(0)

	for _, opt := range opts {
		opt(content)
	}
	return content
}

func (tb *Instance) NewRainbowContent(msg string, opts ...ContentOpt) *Content {
	var cgroups []*cgroup
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			cgroups = append(cgroups, newSpaceCgroup(tb.text))
		}
		cgroups = append(cgroups, newSimpleCgroup(part, util.StringToColor(part, 0.95, 0.35), tb.text))
	}
	return tb.newContent(cgroups, opts...)
}

func (tb *Instance) NewSimpleContent(msg string, color color.Color, opts ...ContentOpt) *Content {
	var cgroups []*cgroup
	parts := strings.Split(msg, " ")
	for id, part := range parts {
		if id > 0 {
			cgroups = append(cgroups, newSpaceCgroup(tb.text))
		}
		cgroups = append(cgroups, newSimpleCgroup(part, color, tb.text))
	}
	return tb.newContent(cgroups, opts...)
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
	if tb.cfg.lineCount != 0 {
		renderLineCount = tb.cfg.lineCount
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

	tb.imd.Draw(target)
	typingProgress := 0
line_loop:
	for lineId, line := range pageLines {
		y := float64(((renderLineCount - 1 - lineId) * (tb.letterHeight + tb.lineSpacing)) + tb.tailHeight)
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
				if typingProgress > content.currentPageTypingDone {
					break line_loop
				}
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
				typingProgress += c.typingWeight
			}
		}
	}

	tb.text.Draw(target, matrix)

	start, end := content.pageLineIdRange()
	ctx.DebugBR("current page: %d", content.currentPage)
	ctx.DebugBR("start/end line ids: %d/%d", start, end)
	ctx.DebugBR("last page: %d", content.lastPage())
	ctx.DebugBR("lines: %d", len(content.lines))

}
