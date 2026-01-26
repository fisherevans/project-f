package adventure

// HandlerOutput is returned by event handlers to specify state changes and effects
type HandlerOutput struct {
	State   any
	Effects []Effect // Effect is now an interface
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
	return o.WithEffects(NewSerialPlan(v...))
}

func (o *HandlerOutput) WithParallelPlan(v ...Effect) *HandlerOutput {
	return o.WithEffects(NewParallelPlan(v...))
}

// EventHandler is the interface that all event handlers must implement
type EventHandler interface {
	Init(thisEntity EntityReader, globals StateGlobalsReader, state any) *HandlerOutput
	HandleEvent(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput
}

// EventHandlerFunc is a function that handles a specific event type
type EventHandlerFunc func(thisEntity EntityReader, globals StateGlobalsReader, state any, event any) *HandlerOutput
