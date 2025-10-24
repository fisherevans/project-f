package events

// HandlerOutput is returned by event handlers to specify state changes and effects
type HandlerOutput struct {
	State   any
	Effects []Effect
}

func NewOutput() *HandlerOutput {
	return &HandlerOutput{}
}

func (o *HandlerOutput) WithState(v any) *HandlerOutput {
	o.State = v
	return o
}

func (o *HandlerOutput) WithEffects(v ...Effect) *HandlerOutput {
	o.Effects = append(o.Effects, v...)
	return o
}

func (o *HandlerOutput) WithSerialPlan(v ...Effect) *HandlerOutput {
	return o.WithEffects(Effect{
		Plan: NewSerialPlan(v...),
	})
}

func (o *HandlerOutput) WithParallelPlan(v ...Effect) *HandlerOutput {
	return o.WithEffects(Effect{
		Plan: NewParallelPlan(v...),
	})
}

// EventHandler is the interface that all event handlers must implement
type EventHandler interface {
	Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput
	HandleEvent(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput
}

// EventHandlerFunc is a function that handles a specific event type
type EventHandlerFunc func(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput

type EntityContext interface {
	Id() string
	Mode() string
}

func NewEphemeralEntityContext(id string) EntityContext {
	return ephemeralEntityContext{
		id: id,
	}
}

type ephemeralEntityContext struct {
	id string
}

func (e ephemeralEntityContext) Id() string {
	return e.id
}

func (e ephemeralEntityContext) Mode() string {
	return ""
}
