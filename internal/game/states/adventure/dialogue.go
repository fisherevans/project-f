package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/util/textbox2"
	"github.com/gopxl/pixel/v2"
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

var dialogueBoxMargin = 4
var dialogueBox = textbox2.NewInstance(textbox2.FontLarge, textbox2.PaddingNormal, game.GameWidth-dialogueBoxMargin*2, textbox2.AlignLeft, textbox2.ExpandFull)

func (ds *DialogueSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, bounds MapBounds, timeDelta float64) {
	defer ds.flushPending()
	ctx.DebugBR("dialogue queue: %d", len(ds.queuedDialogues))
	if !ds.HasPriority() {
		return
	}
	dialogue := ds.queuedDialogues[0]

	dialogue.Content().Update(ctx, timeDelta)
	dialogueBox.Render(ctx, target, pixel.IM.Moved(pixel.V(game.GameWidth/2, float64(dialogueBoxMargin))), dialogue.Content())
	if ctx.Controls.ButtonA().JustPressed() {
		ds.queuedDialogues = ds.queuedDialogues[1:]
		dialogue.OnDismiss(ctx, s)
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
	Message() string
	Content() *textbox2.Content
	OnDismiss(*game.Context, *State)
}

type basicDialogue struct {
	message string
	content *textbox2.Content
}

func NewBasicDialogue(message string) Dialogue {
	content := dialogueBox.NewRainbowPhrase(message)
	return &basicDialogue{
		message: message,
		content: content,
	}
}

func (b basicDialogue) Message() string {
	return b.message
}

func (b basicDialogue) Content() *textbox2.Content {
	return b.content
}

func (b basicDialogue) OnDismiss(context *game.Context, state *State) {
}
