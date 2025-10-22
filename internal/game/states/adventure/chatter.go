package adventure

import (
	"sort"

	"fisherevans.com/project/f/internal/game/events"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type ChatterState int

const (
	ChatterVisible ChatterState = iota
	ChatterHidden
	ChatterComplete
)

type Chatter interface {
	Content() *textbox.Content
	State() ChatterState
	RenderAbove() pixel.Vec
	Update(s *State, timeDelta float64)
	Id() string
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

var chatterArrow = atlas.GetSprite("chatter/chatter_box_arrow")
var chatterFrame = frames.New("chatter/chatter_box", atlas)
var chatterBox = textbox.NewInstance(
	atlas.GetFont(resources.FontNameM3x6),
	tbcfg.NewConfig(game.GameWidth/3, 0,
		tbcfg.HAligned(tbcfg.HAlignCenter),
		tbcfg.WithExpandMode(tbcfg.ExpandFit)))

func (c *ChatterSystem) OnTick(s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	c.sortChatters()
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(s, timeDelta)
		if chatter.State() == ChatterComplete {
			e := events.EventChatterComplete{
				ChatterId: chatter.Id(),
			}
			if basic, ok := chatter.(*basicEntityChatter); ok { // todo messy
				e.EntityId = string(basic.target)
			}
			s.eventDispatcher.Dispatch(e)
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)

		renderMatrix := matrix.Moved(chatter.RenderAbove().
			Scaled(resources.MapTileSize.Float()).
			Add(pixel.V(0, resources.MapTileSize.Float())))

		frameWidth := chatter.Content().Width() + chatterFrame.HorizontalPadding()
		frameHeight := chatter.Content().Height() + chatterFrame.VerticalPadding()
		frameRect := pixel.R(0, 0, float64(frameWidth), float64(frameHeight))
		chatterFrame.Draw(target, frameRect, renderMatrix.Moved(pixel.V(float64(-1*frameWidth/2), 0)))

		chatterArrow.Draw(target, renderMatrix)

		chatter.Content().Update(timeDelta)
		chatterBox.Render(target, renderMatrix.Moved(pixel.V(float64(-1*chatter.Content().Width()/2), float64(chatterFrame.BottomPadding()))), chatter.Content())
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
	id             string
	content        *textbox.Content
	target         EntityId
	renderLocation pixel.Vec
	displayTime    float64
	elapsedTime    float64
}

func (b *basicEntityChatter) Id() string {
	return b.id
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

func (b *basicEntityChatter) Content() *textbox.Content {
	return b.content
}

func (b *basicEntityChatter) Update(s *State, timeDelta float64) {
	b.elapsedTime += timeDelta
	if entity, ok := s.entities[b.target]; ok {
		b.renderLocation = entity.RenderMapLocation()
	}
}

func newBasicEntityChatter(target EntityId, displayTime float64, message string, id string) Chatter {
	content := chatterBox.NewSimpleContent(message)
	//content.SetTypingSpeed(0.01)
	return &basicEntityChatter{
		id:          id,
		displayTime: displayTime,
		content:     content,
		target:      target,
	}
}
