package textbox

import (
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
)

type cShadow struct {
	color pixel.RGBA
}

type cColor struct {
	foreground pixel.RGBA
}

type cUnderline struct {
	color pixel.RGBA
}

type cOutline struct {
	color pixel.RGBA
}

type cStyle struct {
	effects   []RenderEffect
	color     *cColor
	shadow    *cShadow
	underline *cUnderline
	outline   *cOutline
}

func (s cStyle) String() string {
	return fmt.Sprintf("cStyle{color=%v, shadow=%v, underline=%v, effects=%d}", s.color, s.shadow, s.underline, len(s.effects))
}

type character struct {
	c            byte
	typingWeight int
	width        int
	style        cStyle
}

func newCharacter(char byte, typingWeight int, text *text.Text, style cStyle) *character {
	return &character{
		c:            char,
		typingWeight: typingWeight,
		width:        int(text.BoundsOf(string(char)).W()),
		style:        style,
	}
}

func (c *character) String() string {
	return fmt.Sprintf("'%b' - %s", c.c, c.style)
}

func (c *character) isWhitespace() bool {
	return c.c == ' ' || c.c == '\t'
}

func asString(characters []*character) string {
	var str []byte
	for _, c := range characters {
		str = append(str, c.c)
	}
	return string(str)
}

type line struct {
	characters []*character
	text       string

	// set on commit
	width       int
	typingTotal int
	typingDone  int
}

func newLine() *line {
	return &line{}
}

func (l *line) commit(text *text.Text) {
	l.width = int(text.BoundsOf(l.text).W())
	l.typingTotal = 0
	for _, c := range l.characters {
		l.typingTotal += c.typingWeight
	}
	l.typingDone = 0
}

func (l *line) doTyping(todo int) int {
	if l.typingDone+todo > l.typingTotal {
		remaining := l.typingDone + todo - l.typingTotal
		l.typingDone = l.typingTotal
		return remaining
	}
	l.typingDone += todo
	return 0
}

func (l *line) typeAll() {
	l.typingDone = l.typingTotal
}

func (l *line) append(cs ...*character) {
	l.characters = append(l.characters, cs...)
	l.text += asString(cs)
}
