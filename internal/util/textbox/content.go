package textbox

import (
	"math"
	"strings"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"

	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type OnTypeListener func(s string, speed float64)

type Content struct {
	tb               *Instance
	lines            []*line
	effects          []RenderEffect
	typingController TypingController
	onTypeListener   OnTypeListener

	scrollPosition float64
	startLine      int

	progressFaster bool

	alignmentOverride *tbcfg.HAlignment

	autoWrapWidth int

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

func WithAutoWrapWidth(width int) ContentOpt {
	return func(c *Content) {
		c.autoWrapWidth = width
	}
}

func WithOnTypeListener(listener OnTypeListener) ContentOpt {
	return func(c *Content) {
		c.onTypeListener = listener
	}
}

func WithAlignment(a tbcfg.HAlignment) ContentOpt {
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

func (c *Content) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(c.width), float64(c.height))
}

func (c *Content) Update(timeDelta float64, listener OnTypeListener) {
	// update character effects
	for _, e := range c.effects {
		e.Update(timeDelta)
	}
	if c.scrollPosition != float64(c.startLine) {
		toScroll := timeDelta / c.tb.cfg.ScrollTimePerLine
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
		var typed string
		speed := 1.0
		if c.progressFaster {
			speed = 2.0
		}
		toType := c.typingController.TypeSome(timeDelta, c.progressFaster)
		for _, line := range c.pageLines() {
			toType, typed = line.doTyping(toType)
			if typed != "" {
				listener(typed, speed)
			}
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
	advance := c.tb.cfg.LinesPerPage
	if c.tb.cfg.LineByLine {
		advance = 1
	}
	c.setPage(c.startLine + advance)
}

func (c *Content) PreviousPage() {
	regress := c.tb.cfg.LinesPerPage
	if c.tb.cfg.LineByLine {
		regress = 1
	}
	c.setPage(c.startLine - regress)
}

func (c *Content) onLastPage() bool {
	return c.tb.cfg.LinesPerPage == 0 || c.startLine == c.lastStartLine()
}

func (c *Content) lastStartLine() int {
	return util.MaxInt(len(c.lines)-c.tb.cfg.LinesPerPage, 0)
}

func (c *Content) setPage(newStartLine int) {
	if c.tb.cfg.LinesPerPage == 0 {
		newStartLine = 0
	}
	newStartLine = util.Clamp(0, newStartLine, c.lastStartLine())
	c.startLine = newStartLine
	c.progressFaster = false
}

func (c *Content) pageLines() []*line {
	if c.tb.cfg.LinesPerPage == 0 {
		return c.lines
	}
	scrollDelta := c.scrollPosition - float64(c.startLine)
	if scrollDelta < 0 {
		return []*line{c.lines[c.startLine]}
	}
	if scrollDelta > 0 {
		return []*line{&line{}, c.lines[c.startLine+1]}
	}
	return c.lines[c.startLine:util.MinInt(c.startLine+c.tb.cfg.LinesPerPage, len(c.lines))]
}

func (tb *Instance) newContent(paragraphs [][]*character, opts ...ContentOpt) *Content {
	content := &Content{
		tb:            tb,
		autoWrapWidth: tb.cfg.BoxWidth,
	}
	for _, opt := range opts {
		opt(content)
	}

	currentLine := newLine()

	var pendingWhitespace, pendingText []*character

	flushPendingText := func() {
		if len(pendingText) > 0 {
			potentialLineText := currentLine.text + asString(pendingWhitespace) + asString(pendingText)
			if content.autoWrapWidth > 0 && int(tb.text.BoundsOf(potentialLineText).W()) > content.autoWrapWidth {
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

	// pre compute details
	lineCount := tb.cfg.LinesPerPage
	if lineCount == 0 {
		lineCount = len(content.lines)
	}
	content.height = (lineCount * tb.Metadata.LetterHeight) + ((lineCount - 1) * tb.effectiveLineSpacing()) + tb.Metadata.TailHeight*2

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
	if l.width > c.width {
		c.width = l.width
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

// Render is a convenience method that renders this content using its associated textbox.
// This eliminates the need to manually pair content with the correct textbox instance.
func (c *Content) Render(target pixel.Target, matrix pixel.Matrix, opts ...tbcfg.ConfigOpt) pixel.Vec {
	return c.tb.render(target, matrix, c, opts...)
}

func (tb *Instance) NewSimpleContent(msg string, opts ...ContentOpt) *Content {
	msg = messageSanitizer.Replace(msg)
	var characters []*character
	for _, c := range []byte(msg) {
		characters = append(characters, newCharacter(c, 1, tb.text, cStyle{}))
	}
	return tb.newContent([][]*character{characters}, opts...)
}
