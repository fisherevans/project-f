package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
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
	Text() string
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

var chatterText = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))
var chatterImd = imdraw.New(nil)

func (c *ChatterSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(ctx, s, timeDelta)
		if chatter.State() == ChatterComplete {
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)
		chatterText.Clear()
		chatterText.WriteString(chatter.Text())
		dx := -chatterText.Bounds().W() / 2

		renderMatrix := matrix.Moved(chatter.RenderAbove().Scaled(resources.TileSizeF64).Add(pixel.V(dx, resources.TileSize/2)))

		bgBounds := chatterText.Bounds()
		matrix.Project(bgBounds.Min)
		chatterImd.Clear()
		chatterImd.Color = chatter.Style().BackgroundColor
		chatterImd.Push(renderMatrix.Project(bgBounds.Min), renderMatrix.Project(bgBounds.Max))
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
	text           string
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

func (b *basicEntityChatter) Text() string {
	return b.text
}

func (b *basicEntityChatter) Update(ctx *game.Context, s *State, timeDelta float64) {
	b.elapsedTime += timeDelta
	if entity, ok := s.entities[b.target]; ok {
		b.renderLocation = entity.RenderMapLocation()
	}
}

func (b *basicEntityChatter) Style() ChatterStyle {
	return ChatterStyle{
		ForegroundColor: pixel.RGB(.1, .2, .2),
		BackgroundColor: pixel.RGB(.6, .9, .9),
	}
}

func newBasicEntityChatter(target EntityId, text string, displayTime float64) Chatter {
	return &basicEntityChatter{
		text:        text,
		target:      target,
		displayTime: displayTime,
	}
}
