package edit_obj

import "fisherevans.com/project/f/internal/game"

type Action interface {
	Label() string
	Execute(ctx *game.Context, s *Editor)
}

type basicAction struct {
	label    string
	doAction func(ctx *game.Context, s *Editor)
}

func (b basicAction) Label() string {
	return b.label
}

func (b basicAction) Execute(ctx *game.Context, s *Editor) {
	b.doAction(ctx, s)
}

func newBasicAction(label string, doAction func(ctx *game.Context, s *Editor)) Action {
	return basicAction{label: label, doAction: doAction}
}
