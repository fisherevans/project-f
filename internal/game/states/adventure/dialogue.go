package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/audio"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
)

type dialogueState int

const (
	dialogStateEntering dialogueState = iota
	dialogStateVisible
	dialogStateExiting
)

var dialogueDoneAnimation *anim.AnimatedSprite

type DialogueSystem struct {
	queuedDialogues []Dialogue

	toAppend  []Dialogue
	toPrepend []Dialogue

	renderStateElapsed float64
	renderState        dialogueState
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
var dialogueFrame *frames.Instance
var dialogueBox *textbox.Instance

var dialogueTransitionTime = 0.3

func (ds *DialogueSystem) OnTick(s *State, target pixel.Target, bounds MapBounds, timeDelta float64) {
	defer ds.flushPending()
	game.DebugBRf("dialogue queue: %d", len(ds.queuedDialogues))
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

	frameBounds := pixel.R(
		float64(dialogueFrameMargin),
		float64(dialogueFrameMargin),
		float64(game.GameWidth-dialogueFrameMargin),
		float64(dialogueFrameMargin+dialogue.Content().Height()+dialogueFrame.VerticalPadding()))

	textBoxBottomLeft := gfx.IVec(dialogueFrameMargin+dialogueFrame.LeftPadding(), dialogueFrameMargin+dialogueFrame.BottomPadding())

	ds.renderStateElapsed += timeDelta
	renderStateProgress := interp.Smootherstep(min(ds.renderStateElapsed/dialogueTransitionTime, 1.0))
	transitionHeight := frameBounds.H() + textBoxBottomLeft.Y
	frameBottomLeft := pixel.ZV
	switch ds.renderState {
	case dialogStateEntering:
		if ds.renderStateElapsed >= dialogueTransitionTime {
			ds.setRenderState(dialogStateVisible)
		} else {
			frameBottomLeft.Y = -(1.0 - renderStateProgress) * transitionHeight
		}
	case dialogStateExiting:
		if ds.renderStateElapsed >= dialogueTransitionTime {
			ds.popQueue(s)
			return
		}
		frameBottomLeft.Y = -(renderStateProgress) * transitionHeight
	}

	dialogueFrame.Draw(target, frameBounds, pixel.IM.Moved(frameBottomLeft))
	dotPosition := dialogue.Content().Render(target, pixel.IM.Moved(textBoxBottomLeft).Moved(frameBottomLeft))
	if ds.renderState != dialogStateVisible {
		return
	}

	var typeListener textbox.OnTypeListener
	if dialogue.Speech() != nil {
		typeListener = func(s string, speed float64) {
			dialogue.Talker().Speak(s, speed)
			//dialogue.Speech().Play(s)
		}
	}
	dialogue.Content().Update(timeDelta, typeListener)

	a := s.Controls().ButtonA().JustPressed()
	bPressed := s.Controls().ButtonB().IsPressed()
	bJustPressed := s.Controls().ButtonB().JustPressed()
	down := s.Controls().DPad().DirectionJustPressed(input.Down)
	up := s.Controls().DPad().DirectionJustPressed(input.Up)
	if a || bPressed || bJustPressed || down {
		if dialogue.Content().IsContentFullyDisplayed() {
			if a || bJustPressed {
				ds.setRenderState(dialogStateExiting)
			}
		} else if dialogue.Content().IsPageFullyDisplayed() {
			dialogue.Content().NextPage()
		} else {
			dialogue.Content().ProgressFaster()
		}
	} else if up {
		dialogue.Content().PreviousPage()
	}

	if dialogue.Content().IsContentFullyDisplayed() {
		dialogueDoneAnimation.Update(timeDelta)
		doneSprite := dialogueDoneAnimation.Sprite()
		doneMatrix := pixel.IM.Moved(textBoxBottomLeft).Moved(dotPosition).Moved(gfx.BottomLeft.Align(doneSprite))
		//doneMatrix := pixel.IM.Moved(frameBottomLeft).Moved(pixel.V(frameBounds.W()+2.0, 6.0)).Moved(gfx.BottomRight.Align(doneSprite))
		doneSprite.Draw(target, doneMatrix)
	}
}

func (ds *DialogueSystem) setRenderState(renderState dialogueState) {
	ds.renderState = renderState
	ds.renderStateElapsed = 0
	dialogueDoneAnimation.Reset()
	if renderState == dialogStateExiting {
		game.GetAudioSystem().PlaySFX("dialogue_complete", 0.2)
	}
}

func (ds *DialogueSystem) popQueue(s *State) {
	if len(ds.queuedDialogues) == 0 {
		return
	}
	dialogue := ds.queuedDialogues[0]
	ds.queuedDialogues = ds.queuedDialogues[1:]
	s.eventDispatcher.Dispatch(EventDialogueComplete{
		DialogueId: dialogue.DialogueId(),
	})
	s.planExecutor.MarkComplete(dialogue.CompletionId())
	ds.setRenderState(dialogStateEntering)
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
	DialogueId() string
	CompletionId() string
	Message() string
	Content() *textbox.Content
	Speech() *audio.SpeechGenerator
	Talker() *audio.Talker
}

type basicDialogue struct {
	dialogueId   string
	completionId string
	message      string
	content      *textbox.Content
	speech       *audio.SpeechGenerator
	talker       *audio.Talker
}

func NewBasicDialogue(message string, dialogueId, completionId string, cfg TalkerConfig) Dialogue {
	content := dialogueBox.NewComplexContent(message, textbox.WithTyping(0.0333))
	return &basicDialogue{
		dialogueId:   dialogueId,
		completionId: completionId,
		message:      message,
		content:      content,
		speech:       game.GetAudioSystem().CreateSpeechGenerator(),
		talker:       game.GetAudioSystem().NewTalker(cfg.ToParams()),
	}
}

func (b basicDialogue) DialogueId() string {
	return b.dialogueId
}

func (b basicDialogue) CompletionId() string {
	return b.completionId
}

func (b basicDialogue) Message() string {
	return b.message
}

func (b basicDialogue) Content() *textbox.Content {
	return b.content
}

func (b basicDialogue) Speech() *audio.SpeechGenerator {
	return b.speech
}

func (b basicDialogue) Talker() *audio.Talker {
	return b.talker
}
