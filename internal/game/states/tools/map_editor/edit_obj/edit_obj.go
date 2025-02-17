package edit_obj

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font/basicfont"
)

type OnApplied func()

type Editor struct {
	game.BaseState
	win                  *opengl.Window
	backState, nextState game.State
	onApplied            OnApplied
	header               string
	fields               []Field
	actions              []Action
	selected             int
}

func New(win *opengl.Window, header string, fields []Field, actions []Action, backState, nextState game.State, onApplied OnApplied) game.State {
	editor := &Editor{
		win:       win,
		header:    header,
		fields:    fields,
		backState: backState,
		nextState: nextState,
		onApplied: onApplied,
		actions:   actions,
	}
	editor.actions = append(editor.actions, newBasicAction("Save Object", func(ctx *game.Context, s *Editor) {
		for _, field := range s.fields {
			field.Apply()
		}
		if s.onApplied != nil {
			s.onApplied()
		}
		ctx.SwapActiveState(s.nextState)
	}))
	return editor
}

var textDrawer = text.New(pixel.ZV, text.NewAtlas(basicfont.Face7x13, text.ASCII))

func (s *Editor) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	if s.win.JustPressed(pixel.KeyEscape) {
		ctx.SwapActiveState(s.backState)
	}

	switch ctx.Controls.DPad().JustPressedOrRepeatedDirection() {
	case input.Up:
		s.selected--
	case input.Down:
		s.selected++
	}
	if s.selected < 0 {
		s.selected = len(s.fields)
	}
	if s.selected >= len(s.fields)+len(s.actions) {
		s.selected = 0
	}

	if s.win.JustPressed(pixel.KeyEnter) {
		// edit specific field
		if s.selected >= 0 && s.selected < len(s.fields) {
			ctx.SwapActiveState(s.fields[s.selected].Edit(s.win, s))
			return
		}
		if s.selected >= len(s.fields) && s.selected < len(s.fields)+len(s.actions) {
			s.actions[s.selected-len(s.fields)].Execute(ctx, s)
			return
		}
		log.Error().Msg("selected object field is invalid!")
	}

	textDrawer.Clear()
	var lines = 0
	if s.header != "" {
		textDrawer.WriteString(s.header + "\n")
		lines++
	}
	for id, field := range s.fields {
		left := "  "
		if id == s.selected {
			left = "> "
		}
		textDrawer.WriteString(fmt.Sprintf("%s%s: %s\n", left, field.Name(), field.Value()))
		lines++
	}
	if len(s.actions) > 0 {
		textDrawer.WriteString("\nActions:\n")
		lines += 2
	}
	for id, action := range s.actions {
		left := "  "
		if len(s.fields)+id == s.selected {
			left = "> "
		}
		textDrawer.WriteString(fmt.Sprintf("%s%s\n", left, action.Label()))
		lines++
	}
	textDrawer.Draw(s.win, pixel.IM.Moved(pixel.V(10, textDrawer.LineHeight*float64(lines)+5)))

	ctx.DebugBR("enter: confirm")
	ctx.DebugBR("esc: cancel")
}
