package adventure

import (
	"sort"

	"fisherevans.com/project/f/internal/util/gfx"
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
	Id() string
	Content() *textbox.Content
	State() ChatterState
	RenderAbove() pixel.Vec
	Update(s *State, timeDelta float64)
	EntityId() string
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

func (c *ChatterSystem) OnTick(s *State, target pixel.Target, cameraDelta pixel.Vec, bounds MapBounds, timeDelta float64) {
	c.sortChatters()
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(s, timeDelta)
		if chatter.State() == ChatterComplete {
			e := EventChatterComplete{
				ChatterId: chatter.Id(),
				EntityId:  chatter.EntityId(),
			}
			s.eventDispatcher.Dispatch(e)
			// Mark chatter complete for plan tracking
			s.planExecutor.MarkChatterComplete(chatter.Id())
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)

		moveDelta := chatter.RenderAbove().Scaled(resources.MapTileSize.Float())
		renderMatrix := pixel.IM.Moved(cameraDelta).Moved(moveDelta.Add(gfx.IVec(0, resources.MapTileSize.Int())))

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
	target         string
	renderLocation pixel.Vec
	displayTime    float64
	elapsedTime    float64
}

func (b *basicEntityChatter) Id() string {
	return b.id
}

func (b *basicEntityChatter) EntityId() string {
	return b.target
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
	entity, ok := s.entities.GetEntity(b.target)
	if !ok {
		return
	}
	b.renderLocation = entity.GetPreciseLocation()
}

func newBasicEntityChatter(target string, displayTime float64, message string, id string) Chatter {
	content := chatterBox.NewSimpleContent(message)
	//content.SetTypingSpeed(0.01)
	return &basicEntityChatter{
		id:          id,
		displayTime: displayTime,
		content:     content,
		target:      target,
	}
}
