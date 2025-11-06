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
	Init(ctx EntityContext, gameState GameState, state any) *HandlerOutput
	HandleEvent(ctx EntityContext, gameState GameState, state any, event any) *HandlerOutput
}

// EventHandlerFunc is a function that handles a specific event type
type EventHandlerFunc func(ctx EntityContext, gameState GameState, state any, event any) *HandlerOutput

type EntityContext interface {
	EntityId() string
	GetMetadata(key string) any
	GetStringMetadata(key string) string
	GetBoolMetadata(key string) bool
}

type BasicEntityContext struct {
	id       string
	metadata map[string]func() any
}

func NewBasicEntityContext(id string) *BasicEntityContext {
	return &BasicEntityContext{
		id: id,
	}
}

func (e *BasicEntityContext) WithMetadata(key string, value func() any) *BasicEntityContext {
	if e.metadata == nil {
		e.metadata = map[string]func() any{}
	}
	e.metadata[key] = value
	return e
}

func (e *BasicEntityContext) EntityId() string {
	return e.id
}

func (e *BasicEntityContext) GetMetadata(key string) any {
	if e.metadata == nil {
		e.metadata = map[string]func() any{}
	}
	v, ok := e.metadata[key]
	if !ok || v == nil {
		return nil
	}
	return v()
}

func (e *BasicEntityContext) GetStringMetadata(key string) string {
	v := e.GetMetadata(key)
	str, _ := v.(string)
	return str
}

func (e *BasicEntityContext) GetBoolMetadata(key string) bool {
	v := e.GetMetadata(key)
	b, _ := v.(bool)
	return b
}
