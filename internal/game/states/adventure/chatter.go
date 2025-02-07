package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"strings"
)

type ChatterState int

const (
	ChatterVisible ChatterState = iota
	ChatterHidden
	ChatterComplete
)

type ChatterStyle struct {
	ForegroundColor pixel.RGBA
	BackgroundColor pixel.RGBA
}

type Chatter interface {
	Message() string
	Style() ChatterStyle
	State() ChatterState
	RenderAbove() pixel.Vec
	Update(ctx *game.Context, s *State, timeDelta float64)
}

type ChatterSystem struct {
	chatters []Chatter
}

func NewChatterSystem() *ChatterSystem {
	return &ChatterSystem{}
}

func (c *ChatterSystem) Add(chatter Chatter) {
	c.chatters = append(c.chatters, chatter)
}

var chatterFont = resources.Fonts.M6
var chatterText = text.New(pixel.ZV, chatterFont.Atlas)
var chatterImd = imdraw.New(nil)

type line struct {
	text  string
	width int
}

func (c *ChatterSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(ctx, s, timeDelta)
		if chatter.State() == ChatterComplete {
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)
		chatterText.Clear()

		maxWidth := game.GameWidth / 3
		//trimUpperY := 5
		//trimLowerY := 2

		words := strings.Split(chatter.Message(), " ")
		var lines []*line
		currentLine := &line{
			text: words[0],
		}
		for _, word := range words[1:] {
			potentialLineText := currentLine.text + " " + word
			potentialLineTextWidth := int(chatterText.BoundsOf(potentialLineText).W())
			if potentialLineTextWidth > maxWidth {
				lines = append(lines, currentLine)
				currentLine = &line{
					text: word,
				}
				continue
			}
			currentLine.text = potentialLineText
		}
		lines = append(lines, currentLine)

		chatterWidth := 0
		for _, line := range lines {
			line.width = int(chatterText.BoundsOf(line.text).W())
			if line.width > chatterWidth {
				chatterWidth = line.width
			}
		}

		letterHeight := 6
		lineSpacing := 2
		tailHeight := 2
		ctx.DebugBL("chatter width: %d", chatterWidth)
		for index, line := range lines {
			dx := (chatterWidth - line.width) / 2
			ctx.DebugBL("line %d width: %d - dx: %d = %s", index, line.width, dx, line.text)
			ctx.DebugBL("dot %d: %v", index, chatterText.Dot)
			chatterText.Dot = pixel.V(float64(dx), float64(((len(lines)-1-index)*(letterHeight+lineSpacing))+tailHeight))
			chatterText.WriteString(line.text)
		}
		chatterHeight := (len(lines) * letterHeight) + ((len(lines) - 1) * lineSpacing) + tailHeight
		chatterBounds := pixel.Rect{
			Min: pixel.ZV,
			Max: pixel.V(float64(chatterWidth), float64(chatterHeight)),
		}

		renderMatrix := matrix.Moved(chatter.RenderAbove().Scaled(resources.TileSizeF64).Add(pixel.V(float64(-chatterWidth/2), 0).Floor()))

		ctx.DebugBL("chat bounds %v", chatterBounds)
		matrix.Project(chatterBounds.Min)
		chatterImd.Clear()
		chatterImd.Color = chatter.Style().BackgroundColor
		chatterImd.Push(renderMatrix.Project(chatterBounds.Min), renderMatrix.Project(chatterBounds.Max))
		chatterImd.Rectangle(0)
		chatterImd.Draw(target)

		chatterText.DrawColorMask(
			target,
			renderMatrix,
			chatter.Style().ForegroundColor)
	}
	c.chatters = incompleteChatters
}

type basicEntityChatter struct {
	message        string
	style          ChatterStyle
	target         EntityId
	renderLocation pixel.Vec
	displayTime    float64
	elapsedTime    float64
}

func (b *basicEntityChatter) State() ChatterState {
	if b.elapsedTime > b.displayTime {
		return ChatterComplete
	}
	return ChatterVisible
}

func (b *basicEntityChatter) RenderAbove() pixel.Vec {
	return b.renderLocation
}

func (b *basicEntityChatter) Message() string {
	return b.message
}

func (b *basicEntityChatter) Update(ctx *game.Context, s *State, timeDelta float64) {
	b.elapsedTime += timeDelta
	if entity, ok := s.entities[b.target]; ok {
		b.renderLocation = entity.RenderMapLocation()
	}
}

func (b *basicEntityChatter) Style() ChatterStyle {
	return b.style
}

func newBasicEntityChatter(target EntityId, displayTime float64, message string) Chatter {
	return &basicEntityChatter{
		displayTime: displayTime,
		message:     message,
		style: ChatterStyle{
			ForegroundColor: pixel.RGB(.1, .2, .2),
			BackgroundColor: pixel.RGB(.6, .9, .9),
		},
		target: target,
	}
}
