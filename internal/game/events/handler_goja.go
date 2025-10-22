package events

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

type handlerFn[T any] func(ctx EntityContext, world WorldStateReader, state any, event *T) *HandlerOutput

type gojaEventHandler struct {
	vm      *goja.Runtime
	program *goja.Program

	fnInit                 func(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput
	fnOnInteract           handlerFn[EventOnInteract]
	fnOnDialogueComplete   handlerFn[EventDialogueComplete]
	fnOnChatterComplete    handlerFn[EventChatterComplete]
	fnOnTimerComplete      handlerFn[EventTimerComplete]
	fnOnEntityZoneActivity handlerFn[EventEntityZoneActivity]
}

func (e *gojaEventHandler) Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {
	if e.fnInit != nil {
		return e.fnInit(ctx, world, state)
	}
	return nil
}

func (e *gojaEventHandler) OnInteract(ctx EntityContext, world WorldStateReader, state any, event *EventOnInteract) *HandlerOutput {
	if e.fnOnInteract == nil {
		return nil
	}
	return e.fnOnInteract(ctx, world, state, event)
}

func (e *gojaEventHandler) OnDialogueComplete(ctx EntityContext, world WorldStateReader, state any, event *EventDialogueComplete) *HandlerOutput {
	if e.fnOnDialogueComplete == nil {
		return nil
	}
	return e.fnOnDialogueComplete(ctx, world, state, event)
}

func (e *gojaEventHandler) OnChatterComplete(ctx EntityContext, world WorldStateReader, state any, event *EventChatterComplete) *HandlerOutput {
	if e.fnOnChatterComplete == nil {
		return nil
	}
	return e.fnOnChatterComplete(ctx, world, state, event)
}

func (e *gojaEventHandler) OnTimerComplete(ctx EntityContext, world WorldStateReader, state any, event *EventTimerComplete) *HandlerOutput {
	if e.fnOnTimerComplete == nil {
		return nil
	}
	return e.fnOnTimerComplete(ctx, world, state, event)
}

func (e *gojaEventHandler) OnEntityZoneActivity(ctx EntityContext, world WorldStateReader, state any, event *EventEntityZoneActivity) *HandlerOutput {
	if e.fnOnEntityZoneActivity == nil {
		return nil
	}
	return e.fnOnEntityZoneActivity(ctx, world, state, event)
}

func NewGojaEventHandler(id string, program *goja.Program) (EventHandler, error) {
	handler := &gojaEventHandler{
		vm:      newGojaVM(),
		program: program,
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

	// Register Init function
	if v := handler.vm.Get("Init"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			handler.fnInit = wrapInitFunction(handler, fn)
		}
	}

	for name, register := range gojaHandlerMapping {
		if v := handler.vm.Get(name); v != nil {
			if fn, ok := goja.AssertFunction(v); ok {
				register(handler, fn)
			}
		}
	}

	return handler, nil
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

func wrapHandlerFn[T any](e *gojaEventHandler, fn goja.Callable) handlerFn[T] {
	return func(ctx EntityContext, world WorldStateReader, state any, event *T) *HandlerOutput {
		fnState := goja.Undefined()
		if state != nil {
			fnState = state.(goja.Value)
		}
		// State is now mutable - JS can modify it directly
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
