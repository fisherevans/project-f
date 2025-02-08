package textbox

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"image/color"
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

type line struct {
	text  string
	width int
}

var borderWidth = 2

func (tb *Instance) Render(target pixel.Target, matrix pixel.Matrix, message string) {
	tb.text.Clear()

	maxTextWidth := tb.maxWidth - tb.paddingX*2 - borderWidth*2/2

	words := strings.Split(message, " ")
	var lines []*line
	currentLine := &line{
		text: words[0],
	}
	for _, word := range words[1:] {
		potentialLineText := currentLine.text + " " + word
		potentialLineTextWidth := int(tb.text.BoundsOf(potentialLineText).W())
		if potentialLineTextWidth > maxTextWidth {
			lines = append(lines, currentLine)
			currentLine = &line{
				text: word,
			}
			continue
		}
		currentLine.text = potentialLineText
	}
	lines = append(lines, currentLine)

	chatterWidth := maxTextWidth
	if tb.expandMode == ExpandFit {
		chatterWidth = 0
		for _, line := range lines {
			line.width = int(tb.text.BoundsOf(line.text).W())
			if line.width > chatterWidth {
				chatterWidth = line.width
			}
		}
	}

	for index, line := range lines {
		var dx int
		switch tb.alignment {
		case AlignLeft:
			dx = 0
		case AlignCenter:
			dx = (chatterWidth - line.width) / 2
		case AlignRight:
			dx = line.width - chatterWidth
		}

		tb.text.Dot = pixel.V(float64(dx), float64(((len(lines)-1-index)*(tb.letterHeight+tb.lineSpacing))+tb.tailHeight))
		tb.text.WriteString(line.text)
	}
	chatterHeight := (len(lines) * tb.letterHeight) + ((len(lines) - 1) * tb.lineSpacing) + tb.tailHeight*2

	boxPadding := pixel.V(float64(tb.paddingX), float64(tb.paddingY))
	boxBounds := pixel.Rect{
		Min: pixel.ZV.Sub(boxPadding),
		Max: pixel.V(float64(chatterWidth), float64(chatterHeight)).Add(boxPadding),
	}

	matrix = matrix.Moved(pixel.V(float64(-chatterWidth/2), float64(borderWidth+tb.tailHeight)).Floor())

	tb.imd.Clear()

	tb.imd.Color = tb.background
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(0)

	tb.imd.Color = tb.foreground
	tb.imd.Push(matrix.Project(boxBounds.Min), matrix.Project(boxBounds.Max))
	tb.imd.Rectangle(float64(borderWidth))

	tb.imd.Draw(target)

	tb.text.DrawColorMask(
		target,
		matrix,
		tb.foreground)
}
