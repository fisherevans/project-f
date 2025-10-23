package events

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

type gojaEventHandler struct {
	vm       *goja.Runtime
	program  *goja.Program
	fnInit   func(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput
	handlers map[reflect.Type]EventHandlerFunc // event type -> wrapped function
}

func (e *gojaEventHandler) Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {
	if e.fnInit != nil {
		return e.fnInit(ctx, world, state)
	}
	return nil
}

func (e *gojaEventHandler) HandleEvent(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput {
	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	handler, exists := e.handlers[eventType]
	if !exists {
		return nil // No handler registered for this event
	}

	return handler(ctx, world, state, event)
}

func NewGojaEventHandler(id string, program *goja.Program) (EventHandler, error) {
	handler := &gojaEventHandler{
		vm:       newGojaVM(),
		program:  program,
		handlers: make(map[reflect.Type]EventHandlerFunc),
	}
	globalVariables := map[string]any{
		"log": gojaLogs{
			caller: id,
		},
	}
	for name, v := range globalVariables {
		if err := handler.vm.Set(name, v); err != nil {
			return nil, fmt.Errorf("failed to set global function %s: %v", name, err)
		}
	}
	if _, err := handler.vm.RunProgram(program); err != nil {
		return nil, fmt.Errorf("error compiling goja script: %w", err)
	}

	// Extract handler object (const handler = {...})
	handlerObj := handler.vm.Get("handler")
	if handlerObj == nil || goja.IsUndefined(handlerObj) || goja.IsNull(handlerObj) {
		return nil, fmt.Errorf("script must export 'const handler = {...}' object")
	}

	obj := handlerObj.ToObject(handler.vm)
	
	// Validate: check for unknown methods
	validMethods := make(map[string]bool)
	validMethods["Init"] = true
	for _, jsFuncName := range registry.handlers {
		validMethods[jsFuncName] = true
	}
	
	for _, key := range obj.Keys() {
		if !validMethods[key] {
			return nil, fmt.Errorf("handler has unknown method '%s' (valid methods: Init, %v)", key, getValidMethodNames())
		}
	}
	
	// Register Init method
	if v := obj.Get("Init"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			handler.fnInit = wrapInitFunction(handler, fn)
		}
	}

	// Auto-register all event handlers from registry
	for eventType, jsFuncName := range registry.handlers {
		if v := obj.Get(jsFuncName); v != nil {
			if fn, ok := goja.AssertFunction(v); ok {
				handler.handlers[eventType] = wrapGenericHandler(handler, fn)
			}
		}
	}

	return handler, nil
}

func getValidMethodNames() []string {
	names := []string{"Init"}
	for _, jsFuncName := range registry.handlers {
		names = append(names, jsFuncName)
	}
	return names
}

func wrapInitFunction(e *gojaEventHandler, fn goja.Callable) func(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {
	return func(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {
		fnState := goja.Undefined()
		if state != nil {
			fnState = state.(goja.Value)
		}
		result, err := fn(
			goja.Undefined(),
			e.vm.ToValue(ctx),
			e.vm.ToValue(world),
			fnState,
		)
		if err != nil {
			log.Error().Err(err).Msg("Init function failed")
			return nil
		}
		return e.parseOutput(result)
	}
}

func wrapGenericHandler(e *gojaEventHandler, fn goja.Callable) EventHandlerFunc {
	return func(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput {
		fnState := goja.Undefined()
		if state != nil {
			fnState = state.(goja.Value)
		}

		result, err := fn(
			goja.Undefined(),
			e.vm.ToValue(ctx),
			e.vm.ToValue(world),
			fnState,
			e.vm.ToValue(event),
		)
		if err != nil {
			log.Error().Err(err).Msg("Event handler function failed")
			return nil
		}
		return e.parseOutput(result)
	}
}

// parseEffects converts JS return value to []Effect
// Uses JSON marshaling for fully automatic conversion
func (e *gojaEventHandler) parseOutput(result goja.Value) *HandlerOutput {
	if isNil(result) {
		return nil
	}
	obj := result.ToObject(e.vm)
	output := &HandlerOutput{
		Effects: e.parseEffects(obj.Get("effects")),
	}
	if state := obj.Get("state"); !isNil(state) {
		output.State = state
	}
	return output
}

func isNil(v goja.Value) bool {
	return v == nil || goja.IsNull(v) || goja.IsUndefined(v)
}

// camelCaseMapper converts Go TitleCase field names to JavaScript camelCase
type camelCaseMapper struct{}

func (camelCaseMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	// Use JSON tag if available
	if tag := f.Tag.Get("json"); tag != "" && tag != "-" {
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		return tag
	}

	name := f.Name
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func (camelCaseMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	// Keep methods as-is (Id(), Position(), etc.)
	return m.Name
}

func newGojaVM() *goja.Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(camelCaseMapper{})
	return vm
}

type gojaLogs struct {
	caller string
}

func (l gojaLogs) Info(fmt string, args ...any) {
	log.Info().Str("caller", l.caller).Msgf(fmt, args...)
}

func (l gojaLogs) Warn(fmt string, args ...any) {
	log.Warn().Str("caller", l.caller).Msgf(fmt, args...)
}

func (l gojaLogs) Error(fmt string, args ...any) {
	log.Error().Str("caller", l.caller).Msgf(fmt, args...)
}
