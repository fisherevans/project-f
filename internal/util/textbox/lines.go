package textbox

import (
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
	"strings"
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

type character struct {
	c            byte
	width        int
	typingWeight int

	effect RenderEffect

	// styling
	color     *cColor
	shadow    *cShadow
	underline *cUnderline
}

func newCharacter(char byte, colorOverride *cColor, effect RenderEffect, shadow *cShadow, underline *cUnderline, typingWeight int, text *text.Text) *character {
	return &character{
		c:            char,
		color:        colorOverride,
		effect:       effect,
		typingWeight: typingWeight,
		shadow:       shadow,
		underline:    underline,
		width:        int(text.BoundsOf(string(char)).W()),
	}
}

type cgroup struct {
	characters   []*character
	isWhitespace bool
}

func newSpaceCgroup(text *text.Text) *cgroup {
	cg := &cgroup{
		isWhitespace: true,
	}
	cg.append(newCharacter(' ', &cColor{pixel.RGBA{}}, nil, nil, nil, 1, text))
	return cg
}

func newTestFeaturesCGroup(word string, colorOverride *cColor, text *text.Text) *cgroup {
	cg := &cgroup{}
	var effect RenderEffect
	if strings.Contains(word, "e") {
		effect = newRumble(0.1)
	}
	var shadow *cShadow
	if len(word) == 5 {
		shadow = &cShadow{
			color: pixel.RGB(0.1, 0.1, 0.1),
		}
	}
	var underline *cUnderline
	if strings.Contains(word, "r") {
		uColor := pixel.RGB(0.1, 0.1, 0.1)
		if colorOverride != nil {
			uColor = util.ScaleColor(colorOverride.foreground, 0.75)
		}
		underline = &cUnderline{
			color: uColor,
		}
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
		cg.append(newCharacter(ch, colorOverride, effect, shadow, underline, weight, text))
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
	for _, cg := range l.cgroups {
		for _, c := range cg.characters {
			l.typingTotal += c.typingWeight
		}
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

func (l *line) append(cg *cgroup) {
	l.cgroups = append(l.cgroups, cg)
}
