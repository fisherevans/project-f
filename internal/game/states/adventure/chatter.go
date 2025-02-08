package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/textbox2"
	"github.com/gopxl/pixel/v2"
	"sort"
)

type ChatterState int

const (
	ChatterVisible ChatterState = iota
	ChatterHidden
	ChatterComplete
)

type Chatter interface {
	Content() *textbox2.Content
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

var chatterBox = textbox2.NewInstance(textbox2.FontSmall, textbox2.PaddingNarrow, game.GameWidth/3, textbox2.AlignCenter, textbox2.ExpandFit)

func (c *ChatterSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	c.sortChatters()
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(ctx, s, timeDelta)
		if chatter.State() == ChatterComplete {
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)

		renderMatrix := matrix.Moved(chatter.RenderAbove().Scaled(resources.TileSizeF64).Add(pixel.V(0, resources.TileSizeF64*0.75)))

		chatter.Content().Update(ctx, timeDelta)
		chatterBox.Render(ctx, target, renderMatrix, chatter.Content())
	}
	c.chatters = incompleteChatters
}

func (c *ChatterSystem) sortChatters() {
	sortedChatters := make([]Chatter, 0, len(c.chatters))
	for _, chatter := range c.chatters {
		sortedChatters = append(sortedChatters, chatter)
	}
	sort.Slice(sortedChatters, func(i, j int) bool {
		iL, jL := sortedChatters[i].RenderAbove(), sortedChatters[j].RenderAbove()
		if iL.Y != jL.Y {
			return iL.Y > jL.Y
		}
		return iL.X < jL.X
	})
	c.chatters = sortedChatters
}

type basicEntityChatter struct {
	content        *textbox2.Content
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

func (b *basicEntityChatter) Content() *textbox2.Content {
	return b.content
}

func (b *basicEntityChatter) Update(ctx *game.Context, s *State, timeDelta float64) {
	b.elapsedTime += timeDelta
	if entity, ok := s.entities[b.target]; ok {
		b.renderLocation = entity.RenderMapLocation()
	}
}

func newBasicEntityChatter(target EntityId, displayTime float64, message string) Chatter {
	return &basicEntityChatter{
		displayTime: displayTime,
		content:     chatterBox.NewRainbowPhrase(message),
		target:      target,
	}
}
