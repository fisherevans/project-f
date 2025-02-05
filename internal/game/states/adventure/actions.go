package adventure

import (
	"fisherevans.com/project/f/internal/game"
)

type ActionState bool

const (
	ActionIncomplete ActionState = false
	ActionComplete               = true
)

type Action interface {
	Execute(ctx *game.Context, s *State, timeDelta float64) ActionState
}

type ActionQueue struct {
	actions []Action
}

func NewActionQueue() *ActionQueue {
	return &ActionQueue{}
}

func (q *ActionQueue) Add(a Action) {
	q.actions = append(q.actions, a)
}

func (q *ActionQueue) Depth() int {
	return len(q.actions)
}

func (q *ActionQueue) ExecuteActions(ctx *game.Context, s *State, timeDelta float64) {
	remaining := q.actions[:0] // Reuse the same slice memory
	for _, action := range q.actions {
		if action.Execute(ctx, s, timeDelta) == ActionIncomplete {
			remaining = append(remaining, action)
		}
	}
	q.actions = remaining
}

type baseAction struct {
	action func(*game.Context, *State, float64) ActionState
}

func (b *baseAction) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	return b.action(ctx, s, timeDelta)
}

func NewChangeCameraAction(camera func(ctx *game.Context, s *State) Camera) Action {
	return &baseAction{
		action: func(ctx *game.Context, s *State, timeDelta float64) ActionState {
			s.camera = camera(ctx, s)
			ctx.Notify("changing camera to %s", s.camera.CurrentLocation().String())
			return ActionComplete
		},
	}
}

type delayAction struct {
	action         Action
	delaySeconds   float64
	elapsedSeconds float64
}

func (d *delayAction) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	d.elapsedSeconds += timeDelta
	if d.elapsedSeconds < d.delaySeconds {
		return ActionIncomplete
	}
	return d.action.Execute(ctx, s, timeDelta)
}

func NewDelayAction(action Action, delaySeconds float64) Action {
	return &delayAction{
		action:       action,
		delaySeconds: delaySeconds,
	}
}

type chainedActions struct {
	actions []Action
}

func (c *chainedActions) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	if len(c.actions) == 0 {
		return ActionComplete
	}
	state := c.actions[0].Execute(ctx, s, timeDelta)
	if state == ActionIncomplete {
		return ActionIncomplete
	}
	c.actions = c.actions[1:]
	if len(c.actions) == 0 {
		return ActionComplete
	}
	return ActionIncomplete
}

func NewChainedActions(actions ...Action) Action {
	return &chainedActions{
		actions: actions,
	}
}
