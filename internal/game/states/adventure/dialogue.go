package adventure

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type DialogueSystem struct {
	queuedDialogues []Dialogue

	toAppend  []Dialogue
	toPrepend []Dialogue
}

func NewDialogueSystem() *DialogueSystem {
	return &DialogueSystem{}
}

func (ds *DialogueSystem) Append(d Dialogue) {
	ds.toAppend = append(ds.toAppend, d)
}

func (ds *DialogueSystem) Prepend(d Dialogue) {
	ds.toPrepend = append([]Dialogue{d}, ds.toPrepend...)
}

func (ds *DialogueSystem) HasPriority() bool {
	return len(ds.queuedDialogues) > 0
}

var dialogueFrameMargin = 4
var dialogueFrame = frames.New("dialogue/dialogue_frame", atlas)
var dialogueBox = textbox.NewInstance(
	atlas.GetFont(resources.FontNameM5x7),
	tbcfg.NewConfig(game.GameWidth-dialogueFrameMargin*2-dialogueFrame.HorizontalPadding(), 0,
		tbcfg.Paging(2, true),
		tbcfg.Foreground(colors.HexString("#00164e")),
		tbcfg.ExtraLineSpacing(4)))

func (ds *DialogueSystem) OnTick(s *State, target pixel.Target, bounds MapBounds, timeDelta float64) {
	defer ds.flushPending()
	game.DebugBR("dialogue queue: %d", len(ds.queuedDialogues))
	entity, ok := s.entities.GetEntity(s.player)
	if !ok {
		log.Error().Str("player", string(s.player)).Msg("player not found in state")
		return
	}
	if !ds.HasPriority() {
		entity.EnableBehavior("dialogue")
		return
	}
	entity.DisableBehavior("dialogue")

	dialogue := ds.queuedDialogues[0]

	dialogue.Content().Update(timeDelta)

	frameBounds := pixel.R(
		float64(dialogueFrameMargin),
		float64(dialogueFrameMargin),
		float64(game.GameWidth-dialogueFrameMargin),
		float64(dialogueFrameMargin+dialogue.Content().Height()+dialogueFrame.VerticalPadding()))
	dialogueFrame.Draw(target, frameBounds, pixel.IM)

	bottomLeft := gfx.IVec(dialogueFrameMargin+dialogueFrame.LeftPadding(), dialogueFrameMargin+dialogueFrame.BottomPadding())

	dialogueBox.Render(target, pixel.IM.Moved(bottomLeft), dialogue.Content())

	a := game.Controls[*State]().
		ButtonA().JustPressed()
	bPressed := game.Controls[*State]().
		ButtonB().IsPressed()
	bJustPressed := game.Controls[*State]().
		ButtonB().JustPressed()
	down := game.Controls[*State]().
		DPad().DirectionJustPressed(input.Down)
	up := game.Controls[*State]().
		DPad().DirectionJustPressed(input.Up)
	if a || bPressed || bJustPressed || down {
		if dialogue.Content().IsContentFullyDisplayed() {
			if a || bJustPressed {
				ds.queuedDialogues = ds.queuedDialogues[1:]
				s.eventDispatcher.Dispatch(EventDialogueComplete{
					DialogueId: dialogue.EntityId(),
				})
				s.planExecutor.MarkDialogueComplete(dialogue.EntityId())
			}
		} else if dialogue.Content().IsPageFullyDisplayed() {
			dialogue.Content().NextPage()
		} else {
			dialogue.Content().ProgressFaster()
		}
	} else if up {
		dialogue.Content().PreviousPage()
	}
}

func (ds *DialogueSystem) flushPending() {
	if len(ds.toPrepend) > 0 {
		ds.queuedDialogues = append(ds.toPrepend, ds.queuedDialogues...)
		ds.toPrepend = ds.toPrepend[:0]
	}
	if len(ds.toAppend) > 0 {
		ds.queuedDialogues = append(ds.queuedDialogues, ds.toAppend...)
		ds.toAppend = ds.toAppend[:0]
	}
}

type Dialogue interface {
	EntityId() string
	Message() string
	Content() *textbox.Content
}

type basicDialogue struct {
	id      string
	message string
	content *textbox.Content
}

func NewBasicDialogue(message string, id string) Dialogue {
	content := dialogueBox.NewComplexContent(message, textbox.WithTyping(0.0333))
	return &basicDialogue{
		id:      id,
		message: message,
		content: content,
	}
}

func (b basicDialogue) EntityId() string {
	return b.id
}

func (b basicDialogue) Message() string {
	return b.message
}

func (b basicDialogue) Content() *textbox.Content {
	return b.content
}
