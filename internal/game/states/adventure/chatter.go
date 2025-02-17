package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
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
	Content() *textbox.Content
	State() ChatterState
	RenderAbove() pixel.Vec
	Update(ctx *game.Context, s *State, timeDelta float64)
}

type ChatterSystem struct {
	chatters    []Chatter
	spriteBatch *pixel.Batch
	textBatch   *pixel.Batch
}

func NewChatterSystem() *ChatterSystem {
	return &ChatterSystem{
		spriteBatch: pixel.NewBatch(&pixel.TrianglesData{}, resources.SpriteAtlas),
		textBatch:   pixel.NewBatch(&pixel.TrianglesData{}, textbox.FontSmall.GetAtlas().Picture()),
	}
}

func (c *ChatterSystem) Add(chatter Chatter) {
	c.chatters = append(c.chatters, chatter)
}

var chatterArrow = resources.Sprites["chatter_arrow"]
var chatterFrame = resources.Frames["chatter_frame"]
var chatterBox = textbox.NewInstance(
	textbox.FontSmall,
	textbox.NewConfig(game.GameWidth/3).
		Aligned(textbox.AlignCenter).
		ExpandMode(textbox.ExpandFit))

func (c *ChatterSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, matrix pixel.Matrix, bounds MapBounds, timeDelta float64) {
	c.sortChatters()
	incompleteChatters := c.chatters[:0] // Reuse the same slice memory
	c.spriteBatch.Clear()
	c.textBatch.Clear()
	for _, chatter := range c.chatters {
		chatter.Update(ctx, s, timeDelta)
		if chatter.State() == ChatterComplete {
			continue
		}
		incompleteChatters = append(incompleteChatters, chatter)

		renderMatrix := matrix.Moved(chatter.RenderAbove().
			Scaled(resources.MapTileSize.Float()).
			Add(pixel.V(0, resources.MapTileSize.Float())))

		frameWidth := chatter.Content().Width() + chatterFrame.HorizontalPadding()
		frameHeight := chatter.Content().Height() + chatterFrame.VerticalPadding()
		frameRect := pixel.R(0, 0, float64(frameWidth), float64(frameHeight))
		frames.Draw(c.spriteBatch, chatterFrame, frameRect, renderMatrix.Moved(pixel.V(float64(-1*frameWidth/2), 0)))

		chatterArrow.Sprite.Draw(c.spriteBatch, renderMatrix)

		chatter.Content().Update(ctx, timeDelta)
		chatterBox.Render(ctx, c.textBatch, renderMatrix.Moved(pixel.V(float64(-1*chatter.Content().Width()/2), float64(chatterFrame.BottomPadding()))), chatter.Content())
	}
	c.chatters = incompleteChatters
	c.spriteBatch.Draw(target)
	c.textBatch.Draw(target)
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
	content        *textbox.Content
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

func (b *basicEntityChatter) Content() *textbox.Content {
	return b.content
}

func (b *basicEntityChatter) Update(ctx *game.Context, s *State, timeDelta float64) {
	b.elapsedTime += timeDelta
	if entity, ok := s.entities[b.target]; ok {
		b.renderLocation = entity.RenderMapLocation()
	}
}

func newBasicEntityChatter(target EntityId, displayTime float64, message string) Chatter {
	content := chatterBox.NewSimpleContent(message)
	//content.SetTypingSpeed(0.01)
	return &basicEntityChatter{
		displayTime: displayTime,
		content:     content,
		target:      target,
	}
}
