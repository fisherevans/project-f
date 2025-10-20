package events

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

func NewGojaEventHandler(program *goja.Program) (EventHandler, error) {
	handler := &gojaEventHandler{
		vm:      goja.New(),
		program: program,
	}
	globalFunctions := map[string]func(call goja.FunctionCall) goja.Value{
		"log": func(call goja.FunctionCall) goja.Value {
			for _, arg := range call.Arguments {
				log.Info().Msgf("[JS] %s", arg.String())
			}
			return goja.Undefined()
		},
	}
	for name, fn := range globalFunctions {
		if err := handler.vm.Set(name, fn); err != nil {
			return nil, fmt.Errorf("failed to set global function %s: %v", name, err)
		}
	}
	if _, err := handler.vm.RunProgram(program); err != nil {
		return nil, fmt.Errorf("error compiling goja script: %w", err)
	}
	if v := handler.vm.Get("Init"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			handler.fnInit = handler.wrapInitFunction(fn)
		}
	}
	eventHandlerFns := map[string]*eventHandlerFn{
		"OnInteract": &handler.fnOnInteract,
	}
	for jsFnName, ptr := range eventHandlerFns {
		if v := handler.vm.Get(jsFnName); v != nil {
			if fn, ok := goja.AssertFunction(v); ok {
				*ptr = handler.wrapEventHandlerFunction(fn)
			}
		}
	}
	return handler, nil
}

type eventHandlerFn func(self, world, state, event ReadableObject) *HandlerOutput

type gojaEventHandler struct {
	vm      *goja.Runtime
	program *goja.Program

	fnInit       func(self, world, state ReadableObject) *HandlerOutput
	fnOnInteract eventHandlerFn
}

func (e *gojaEventHandler) Init(self, world, state ReadableObject) *HandlerOutput {
	if e.fnInit == nil {
		return nil
	}
	return e.fnInit(self, world, state)
}

func (e *gojaEventHandler) OnInteract(self, world, state, event ReadableObject) *HandlerOutput {
	if e.fnOnInteract == nil {
		return nil
	}
	return e.fnOnInteract(self, world, state, event)
}

func (e *gojaEventHandler) wrapInitFunction(fn goja.Callable) func(self, world, state ReadableObject) *HandlerOutput {
	return func(self, world, state ReadableObject) *HandlerOutput {
		// Convert Go objects to goja values
		selfVal := e.toGojaValue(self)
		worldVal := e.toGojaValue(world)
		stateVal := e.toGojaValue(state)

		// Call the JavaScript function
		result, err := fn(goja.Undefined(), selfVal, worldVal, stateVal)
		if err != nil {
			log.Error().Err(err).Msg("Init function failed")
			return nil
		}

		// Convert result back to HandlerOutput
		return e.parseHandlerOutput(result)
	}
}

func (e *gojaEventHandler) wrapEventHandlerFunction(fn goja.Callable) eventHandlerFn {
	return func(self, world, state, event ReadableObject) *HandlerOutput {
		// Convert Go objects to goja values
		selfVal := e.toGojaValue(self)
		worldVal := e.toGojaValue(world)
		stateVal := e.toGojaValue(state)
		eventVal := e.toGojaValue(event)

		// Call the JavaScript function
		result, err := fn(goja.Undefined(), selfVal, worldVal, stateVal, eventVal)
		if err != nil {
			log.Error().Err(err).Msg("Event handler function failed")
			return nil
		}

		// Convert result back to HandlerOutput
		return e.parseHandlerOutput(result)
	}
}

func (e *gojaEventHandler) parseHandlerOutput(result goja.Value) *HandlerOutput {
	if result == nil || goja.IsUndefined(result) || goja.IsNull(result) {
		return nil
	}

	obj := result.ToObject(e.vm)
	if obj == nil {
		log.Warn().Msg("Handler output is not an object")
		return nil
	}

	output := &HandlerOutput{}

	// Parse state
	if stateVal := obj.Get("state"); stateVal != nil && !goja.IsUndefined(stateVal) && !goja.IsNull(stateVal) {
		exported := stateVal.Export()
		if stateMap, ok := exported.(map[string]interface{}); ok {
			newState := NewObject()
			for k, v := range stateMap {
				newState.Set(k, v)
			}
			output.State = newState
		} else {
			log.Warn().Msgf("Handler output 'state' field is not an object: %v", stateVal)
		}
	}

	// Parse effects
	if effectsVal := obj.Get("effects"); effectsVal != nil && !goja.IsUndefined(effectsVal) && !goja.IsNull(effectsVal) {
		if effectsObj := effectsVal.ToObject(e.vm); effectsObj != nil {
			if exported := effectsVal.Export(); exported != nil {
				if effectsSlice, ok := exported.([]interface{}); ok {
					for i, effectInterface := range effectsSlice {
						if effectMap, ok := effectInterface.(map[string]interface{}); ok {
							effect := Effect{}
							if typeVal, ok := effectMap["type"].(string); ok {
								effect.Type = EffectType(typeVal)
							} else {
								log.Warn().Msgf("Effect[%d] 'type' field is not a string or missing: %v", i, effectMap["type"])
								continue
							}
							if dataVal, ok := effectMap["data"]; ok {
								if dataObj, ok := dataVal.(map[string]interface{}); ok {
									effect.Data = dataObj
								} else {
									log.Warn().Msgf("Effect[%d] 'data' field is not an object: %v", i, dataVal)
								}
							}
							output.Effects = append(output.Effects, effect)
						} else {
							log.Warn().Msgf("Effect[%d] is not an object: %v", i, effectInterface)
						}
					}
				} else {
					log.Warn().Msgf("Handler output 'effects' field is not an array: %v", exported)
				}
			}
		} else {
			log.Warn().Msgf("Handler output 'effects' field could not be converted to object: %v", effectsVal)
		}
	}

	return output
}

// toGojaValue converts a ReadableObject to a goja.Value
func (e *gojaEventHandler) toGojaValue(obj ReadableObject) goja.Value {
	if obj == nil {
		return goja.Undefined()
	}
	// Check if it's a mapState and convert to JS object
	if ms, ok := obj.(*mapState); ok {
		return e.vm.ToValue(ms.ToMap())
	}
	// Otherwise, convert it normally
	return e.vm.ToValue(obj)
}
