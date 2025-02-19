package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2/ext/text"
	"math"
	"strings"
)

type Content struct {
	tb               *Instance
	lines            []*line
	effects          []RenderEffect
	maxLineWidth     int
	typingController TypingController

	scrollPosition float64
	startLine      int

	progressFaster bool

	alignmentOverride *Alignment

	// pre-computed
	width  int
	height int
}

type ContentOpt func(content *Content)

func WithTyping(timePerCharacter float64) ContentOpt {
	return func(c *Content) {
		c.typingController = &defaultTypingController{
			timePerCharacter: timePerCharacter,
			fasterScale:      4,
		}
	}
}

func WithAlignment(a Alignment) ContentOpt {
	return func(c *Content) {
		c.alignmentOverride = &a
	}
}

func (c *Content) Width() int {
	return c.width
}

func (c *Content) Height() int {
	return c.height
}

func (c *Content) Update(ctx *game.Context, timeDelta float64) {
	// update character effects
	for _, e := range c.effects {
		e.Update(ctx, timeDelta)
	}
	if c.scrollPosition != float64(c.startLine) {
		toScroll := timeDelta / c.tb.cfg.scrollTimePerLine
		if c.progressFaster {
			toScroll *= 2
		}
		maxToScroll := math.Abs(c.scrollPosition - float64(c.startLine))
		toScroll = math.Min(toScroll, maxToScroll)
		sign := 1.0
		if float64(c.startLine) < c.scrollPosition {
			sign = -1
		}
		c.scrollPosition += toScroll * sign
	}
	if c.scrollPosition != float64(c.startLine) {
		return // don't type if we're still scrolling
	}
	// type lines
	if c.typingController == nil {
		for _, line := range c.pageLines() {
			line.typeAll()
		}
	} else {
		toType := c.typingController.TypeSome(timeDelta, c.progressFaster)
		for _, line := range c.pageLines() {
			toType = line.doTyping(toType)
			if toType == 0 {
				break
			}
		}
	}
}

func (c *Content) ProgressFaster() {
	c.progressFaster = true
}

func (c *Content) IsContentFullyDisplayed() bool {
	if !c.onLastPage() {
		return false
	}
	return c.IsPageFullyDisplayed()
}

func (c *Content) IsPageFullyDisplayed() bool {
	if c.scrollPosition != float64(c.startLine) {
		return false
	}
	for _, line := range c.pageLines() {
		if line.typingDone < line.typingTotal {
			return false
		}
	}
	return true
}

func (c *Content) NextPage() {
	advance := c.tb.cfg.linesPerPage
	if c.tb.cfg.lineByLine {
		advance = 1
	}
	c.setPage(c.startLine + advance)
}

func (c *Content) PreviousPage() {
	regress := c.tb.cfg.linesPerPage
	if c.tb.cfg.lineByLine {
		regress = 1
	}
	c.setPage(c.startLine - regress)
}

func (c *Content) onLastPage() bool {
	return c.tb.cfg.linesPerPage == 0 || c.startLine == c.lastStartLine()
}

func (c *Content) lastStartLine() int {
	return util.MaxInt(len(c.lines)-c.tb.cfg.linesPerPage, 0)
}

func (c *Content) setPage(newStartLine int) {
	if c.tb.cfg.linesPerPage == 0 {
		newStartLine = 0
	}
	newStartLine = util.Clamp(0, newStartLine, c.lastStartLine())
	c.startLine = newStartLine
	c.progressFaster = false
}

func (c *Content) pageLines() []*line {
	if c.tb.cfg.linesPerPage == 0 {
		return c.lines
	}
	scrollDelta := c.scrollPosition - float64(c.startLine)
	if scrollDelta < 0 {
		return []*line{c.lines[c.startLine]}
	}
	if scrollDelta > 0 {
		return []*line{&line{}, c.lines[c.startLine+1]}
	}
	return c.lines[c.startLine:util.MinInt(c.startLine+c.tb.cfg.linesPerPage, len(c.lines))]
}

func (tb *Instance) newContent(paragraphs [][]*character, opts ...ContentOpt) *Content {
	content := &Content{
		tb: tb,
	}
	currentLine := newLine()

	var pendingWhitespace, pendingText []*character

	flushPendingText := func() {
		if len(pendingText) > 0 {
			potentialLineText := currentLine.text + asString(pendingWhitespace) + asString(pendingText)
			if int(tb.text.BoundsOf(potentialLineText).W()) > tb.cfg.maxWidth {
				content.appendLine(currentLine, tb.text)
				currentLine = newLine()
				pendingWhitespace = nil // drop pending space on new lines
			} else {
				currentLine.append(pendingWhitespace...)
				pendingWhitespace = nil
			}
			currentLine.append(pendingText...)
			pendingText = nil
		}
	}

	for _, characters := range paragraphs {
		for _, c := range characters {
			if c.isWhitespace() {
				flushPendingText()
				pendingWhitespace = append(pendingWhitespace, c)
				continue
			}
			pendingText = append(pendingText, c)
		}
		flushPendingText()
		if len(currentLine.characters) > 0 {
			content.appendLine(currentLine, tb.text)
			currentLine = newLine()
		}
		pendingText, pendingWhitespace = nil, nil
	}

	content.setPage(0)

	for _, opt := range opts {
		opt(content)
	}

	// pre compute details
	if tb.cfg.expandMode == ExpandFull {
		content.width = tb.cfg.maxWidth
	} else {
		content.width = content.maxLineWidth
	}

	lineCount := tb.cfg.linesPerPage
	if lineCount == 0 {
		lineCount = len(content.lines)
	}
	content.height = (lineCount * tb.capHeight) + ((lineCount - 1) * tb.lineSpacing) + tb.tailHeight*2

	if content.typingController == nil {
		for _, line := range content.lines {
			line.typingDone = line.typingTotal
		}
	}

	return content
}

func (c *Content) appendLine(l *line, text *text.Text) {
	c.lines = append(c.lines, l)
	for _, ch := range l.characters {
		for _, e := range ch.style.effects {
			c.registerEffect(e)
		}
	}
	l.commit(text)
	if l.width > c.maxLineWidth {
		c.maxLineWidth = l.width
	}
}

func (c *Content) registerEffect(effect RenderEffect) {
	if effect == nil {
		return
	}
	for _, e := range c.effects {
		if e == effect {
			return
		}
	}
	c.effects = append(c.effects, effect)
}

func (c *Content) String() string {
	var lines []string
	for _, l := range c.lines {
		lines = append(lines, l.text)
	}
	return strings.Join(lines, "\n")
}

func (tb *Instance) NewSimpleContent(msg string, opts ...ContentOpt) *Content {
	var characters []*character
	for _, c := range []byte(msg) {
		characters = append(characters, newCharacter(c, 1, tb.text, cStyle{}))
	}
	return tb.newContent([][]*character{characters}, opts...)
}
