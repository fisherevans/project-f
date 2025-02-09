package textbox

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util"
	"math"
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

func (c *Content) ContentFullyDisplayed() bool {
	if !c.onLastPage() {
		return false
	}
	return c.PageFullyDisplayed()
}

func (c *Content) PageFullyDisplayed() bool {
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

func (tb *Instance) newContent(cgroups []*cgroup, opts ...ContentOpt) *Content {
	content := &Content{
		tb: tb,
	}
	var pendingWhitespace []*cgroup
	pendingWhitespaceText := ""
	currentLine := newLine()
	for _, cg := range cgroups {
		cgString := cg.asString()
		if cg.isWhitespace {
			pendingWhitespace = append(pendingWhitespace, cg)
			pendingWhitespaceText += cgString
			continue
		}
		pendingLineText := currentLine.text + pendingWhitespaceText + cgString
		if int(tb.text.BoundsOf(pendingLineText).W()) > tb.maxTextWidth() {
			pendingWhitespace = nil
			pendingWhitespaceText = ""
			pendingLineText = ""
			currentLine.commit(tb.text)
			content.appendLine(currentLine)
			currentLine = newLine()
		}
		if len(pendingWhitespace) > 0 {
			for _, space := range pendingWhitespace {
				currentLine.append(space)
			}
			currentLine.text += pendingWhitespaceText
			pendingWhitespace = nil
			pendingWhitespaceText = ""
		}
		currentLine.append(cg)
		currentLine.text += cgString
	}
	currentLine.commit(tb.text)
	content.appendLine(currentLine)

	content.setPage(0)

	for _, opt := range opts {
		opt(content)
	}
	return content
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

func (c *Content) registerEffect(effect RenderEffect) {
	for _, e := range c.effects {
		if e == effect {
			return
		}
	}
	c.effects = append(c.effects, effect)
}
