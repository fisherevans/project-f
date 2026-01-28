package adventure

import (
	"math"
	"sort"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
)

type ChatterState int

const (
	ChatterVisible ChatterState = iota
	ChatterHidden
	ChatterComplete
)

type Chatter interface {
	ChatterId() string
	CompletionId() string
	Content() *textbox.Content
	State() ChatterState
	RenderAbove() pixel.Vec
	Update(s *State, timeDelta float64)
	EntityId() string
}

type ChatterSystem struct {
	chatters []Chatter
	toAdd    []Chatter
}

func NewChatterSystem() *ChatterSystem {
	return &ChatterSystem{}
}

func (c *ChatterSystem) Add(chatter Chatter) {
	c.toAdd = append(c.toAdd, chatter)
}

var chatterArrow pixelutil.BoundedDrawable
var chatterFrame *frames.Instance
var chatterBox *textbox.Instance

func (c *ChatterSystem) OnTick(s *State, target pixel.Target, cameraDelta pixel.Vec, bounds MapBounds, timeDelta float64) {
	if len(c.toAdd) > 0 {
		for _, add := range c.toAdd {
			c.chatters = append(c.chatters, add)
			entity, exist := s.entities.GetEntity(add.EntityId())
			distance := math.MaxFloat64
			if exist {
				distance = s.camera.CurrentLocation().Sub(entity.GetPreciseLocation()).Len()
			}
			volume := ChatterFalloff.AttenuationVolume(distance) * 0.75
			log.Info().Str("entity_id", add.EntityId()).Float64("volume", volume).Float64("distance", distance).Msg("playing chatter")
			game.GetAudioSystem().PlaySFX("speech/chatter", volume)
		}
		c.toAdd = nil
	}
	c.sortChatters()
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	for _, chatter := range c.chatters {
		chatter.Update(s, timeDelta)
		if chatter.State() == ChatterComplete {
			e := EventChatterComplete{
				ChatterId: chatter.CompletionId(),
				EntityId:  chatter.EntityId(),
			}
			s.eventDispatcher.Dispatch(e)
			// Mark chatter complete for plan tracking
			s.planExecutor.MarkComplete(chatter.CompletionId())
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

		chatter.Content().Update(timeDelta, nil)
		chatter.Content().Render(target, renderMatrix.Moved(pixel.V(float64(-1*chatter.Content().Width()/2), float64(chatterFrame.BottomPadding()))))
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
	chatterId      string
	completionId   string
	content        *textbox.Content
	target         string
	renderLocation pixel.Vec
	displayTime    float64
	elapsedTime    float64
}

func (b *basicEntityChatter) ChatterId() string {
	return b.chatterId
}

func (b *basicEntityChatter) CompletionId() string {
	return b.completionId
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

func newBasicEntityChatter(target string, displayTime float64, message string, id string, completionId string) Chatter {
	content := chatterBox.NewSimpleContent(message)
	//content.SetTypingSpeed(0.01)
	return &basicEntityChatter{
		chatterId:    id,
		completionId: completionId,
		displayTime:  displayTime,
		content:      content,
		target:       target,
	}
}
