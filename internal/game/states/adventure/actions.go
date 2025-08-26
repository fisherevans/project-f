package adventure

import (
	"fisherevans.com/project/f/internal/game"
)

type ActionState int

const (
	ActionComplete ActionState = iota
	ActionIncomplete
)

type ActionFunction func(*game.Context, *State, float64) ActionState

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
	ctx.DebugBR("action queue: %d", len(q.actions))
	remaining := q.actions[:0] // Reuse the same slice memory
	for _, action := range q.actions {
		if action.Execute(ctx, s, timeDelta) == ActionIncomplete {
			remaining = append(remaining, action)
		}
	}
	q.actions = remaining
}

type baseAction struct {
	action ActionFunction
}

func (b *baseAction) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	return b.action(ctx, s, timeDelta)
}

func NewSimpleAction(action func(*game.Context, *State)) *baseAction {
	return &baseAction{
		action: func(ctx *game.Context, s *State, _ float64) ActionState {
			action(ctx, s)
			return ActionComplete
		},
	}
}

func NewWaitForAction(f func() bool) *baseAction {
	return &baseAction{
		action: func(ctx *game.Context, s *State, _ float64) ActionState {
			if f() {
				return ActionComplete
			}
			return ActionIncomplete
		},
	}
}

func NewAction(fn ActionFunction) Action {
	return &baseAction{fn}
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

type sleepAction struct {
	sleepSeconds   float64
	elapsedSeconds float64
}

func (d *sleepAction) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	d.elapsedSeconds += timeDelta
	if d.elapsedSeconds < d.sleepSeconds {
		return ActionIncomplete
	}
	return ActionComplete
}

func NewSleepAction(sleepSeconds float64) Action {
	return &sleepAction{
		sleepSeconds: sleepSeconds,
	}
}

type serialActions struct {
	actions []Action
}

func (c *serialActions) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	for len(c.actions) > 0 {
		if c.actions[0].Execute(ctx, s, timeDelta) == ActionIncomplete {
			return ActionIncomplete
		}
		c.actions = c.actions[1:]
	}
	return ActionIncomplete
}

func NewSerialActions(actions ...Action) Action {
	return &serialActions{
		actions: actions,
	}
}

type parallelActions struct {
	actions []Action
}

func (c *parallelActions) Execute(ctx *game.Context, s *State, timeDelta float64) ActionState {
	if len(c.actions) == 0 {
		return ActionComplete
	}
	var remaining []Action
	for _, action := range c.actions {
		if action.Execute(ctx, s, timeDelta) == ActionIncomplete {
			remaining = append(remaining, action)
		}
	}
	c.actions = remaining
	if len(c.actions) == 0 {
		return ActionComplete
	}
	return ActionIncomplete
}

func NewParallelActions(actions ...Action) Action {
	return &parallelActions{
		actions: actions,
	}
}
