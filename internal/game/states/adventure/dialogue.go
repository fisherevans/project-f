package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
	"github.com/gopxl/pixel/v2"
)

type DialogueSystem struct {
	queuedDialogues []Dialogue

	toAppend  []Dialogue
	toPrepend []Dialogue

	frameBatch *pixel.Batch
}

func NewDialogueSystem() *DialogueSystem {
	return &DialogueSystem{
		frameBatch: pixel.NewBatch(&pixel.TrianglesData{}, resources.SpriteAtlas),
	}
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
var dialogueFrame = resources.Frames["dialogue_frame"]
var dialogueBox = textbox.NewInstance(
	textbox.FontLarge,
	textbox.NewConfig(game.GameWidth-dialogueFrameMargin*2-dialogueFrame.HorizontalPadding()).
		Paging(2, true).
		Foreground(colors.HexColor("#00164e")))

func (ds *DialogueSystem) OnTick(ctx *game.Context, s *State, target pixel.Target, bounds MapBounds, timeDelta float64) {
	defer ds.flushPending()
	ctx.DebugBR("dialogue queue: %d", len(ds.queuedDialogues))
	if !ds.HasPriority() {
		return
	}
	dialogue := ds.queuedDialogues[0]

	dialogue.Content().Update(ctx, timeDelta)

	ds.frameBatch.Clear()
	frameBounds := pixel.R(
		float64(dialogueFrameMargin),
		float64(dialogueFrameMargin),
		float64(game.GameWidth-dialogueFrameMargin),
		float64(dialogueFrameMargin+dialogue.Content().Height()+dialogueFrame.VerticalPadding()))
	frames.Draw(ds.frameBatch, dialogueFrame, frameBounds, pixel.IM)
	ds.frameBatch.Draw(target)

	bottomLeft := pixel.V(float64(dialogueFrameMargin+dialogueFrame.LeftPadding()), float64(dialogueFrameMargin+dialogueFrame.BottomPadding()))
	dialogueBox.Render(ctx, target, pixel.IM.Moved(bottomLeft), dialogue.Content())

	a := ctx.Controls.ButtonA().JustPressed()
	bPressed := ctx.Controls.ButtonB().IsPressed()
	bJustPressed := ctx.Controls.ButtonB().JustPressed()
	down := ctx.Controls.DPad().DirectionJustPressed(input.Down)
	up := ctx.Controls.DPad().DirectionJustPressed(input.Up)
	if a || bPressed || bJustPressed || down {
		if dialogue.Content().IsContentFullyDisplayed() {
			if a || bJustPressed {
				ds.queuedDialogues = ds.queuedDialogues[1:]
				dialogue.OnDismiss(ctx, s)
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
	Message() string
	Content() *textbox.Content
	OnDismiss(*game.Context, *State)
}

type basicDialogue struct {
	message string
	content *textbox.Content
}

func NewBasicDialogue(message string) Dialogue {
	content := dialogueBox.NewComplexContent(message, textbox.WithTyping(0.0333))
	return &basicDialogue{
		message: message,
		content: content,
	}
}

func (b basicDialogue) Message() string {
	return b.message
}

func (b basicDialogue) Content() *textbox.Content {
	return b.content
}

func (b basicDialogue) OnDismiss(context *game.Context, state *State) {
}
