package resources

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

var (
	scriptResources = map[string]*compiledScript{}
)

type compiledScript struct {
	name string
	prog *goja.Program
}

// GetScriptNames returns all loaded script names
func GetScriptNames() []string {
	names := make([]string, 0, len(scriptResources))
	for name := range scriptResources {
		names = append(names, name)
	}
	return names
}

// loadScriptResource is called during init() for each .js file
func loadScriptResource(path string, name string, data []byte) error {
	prog, err := goja.Compile(name, string(data), true) // strict=true
	if err != nil {
		return fmt.Errorf("failed to compile script %s: %w", name, err)
	}
	scriptResources[name] = &compiledScript{
		name: name,
		prog: prog,
	}
	return nil
}

// ScriptEnvironment manages script instances and provides the engine API
type ScriptEnvironment struct {
	vm        *goja.Runtime
	instances map[string]*ScriptInstance
}

// NewScriptEnvironment creates a new script environment with optional custom API binding
func NewScriptEnvironment(bindAPI func(*goja.Runtime)) *ScriptEnvironment {
	if bindAPI == nil {
		// Default binding with just log
		bindAPI = func(vm *goja.Runtime) {
			_ = vm.Set("log", func(call goja.FunctionCall) goja.Value {
				for _, arg := range call.Arguments {
					log.Info().Msgf("[JS] %s", arg.String())
				}
				return goja.Undefined()
			})
		}
	}
	vm := goja.New()
	bindAPI(vm)
	return &ScriptEnvironment{
		vm:        vm,
		instances: make(map[string]*ScriptInstance),
	}
}

// ScriptReference gets or creates a script instance by name
func (e *ScriptEnvironment) ScriptReference(name string) (*ScriptInstance, error) {
	// Check if already instantiated
	if inst, exists := e.instances[name]; exists {
		return inst, nil
	}

	// Look up compiled script
	compiled, exists := scriptResources[name]
	if !exists {
		return nil, fmt.Errorf("script not found: %s", name)
	}

	// Run the program to define exports
	if _, err := e.vm.RunProgram(compiled.prog); err != nil {
		return nil, fmt.Errorf("failed to run script %s: %w", name, err)
	}

	// Cache callable functions
	inst := &ScriptInstance{
		name: name,
		env:  e,
	}

	// Try to cache common lifecycle functions
	if v := e.vm.Get("onSpawn"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			inst.onSpawn = fn
		}
	}
	if v := e.vm.Get("onUpdate"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			inst.onUpdate = fn
		}
	}
	if v := e.vm.Get("onDestroy"); v != nil {
		if fn, ok := goja.AssertFunction(v); ok {
			inst.onDestroy = fn
		}
	}

	e.instances[name] = inst
	return inst, nil
}

// ScriptInstance represents a running instance of a script
type ScriptInstance struct {
	env       *ScriptEnvironment
	name      string
	onSpawn   goja.Callable
	onUpdate  goja.Callable
	onDestroy goja.Callable
}

// OnSpawn calls the script's onSpawn function if it exists
func (si *ScriptInstance) OnSpawn(id int) error {
	if si.onSpawn == nil {
		return nil
	}
	_, err := si.onSpawn(goja.Undefined(), si.env.vm.ToValue(id))
	if err != nil {
		return fmt.Errorf("onSpawn failed in %s: %w", si.name, err)
	}
	return nil
}

// OnUpdate calls the script's onUpdate function if it exists
func (si *ScriptInstance) OnUpdate(id int, dt float64) error {
	if si.onUpdate == nil {
		return nil
	}
	_, err := si.onUpdate(
		goja.Undefined(),
		si.env.vm.ToValue(id),
		si.env.vm.ToValue(dt),
	)
	if err != nil {
		return fmt.Errorf("onUpdate failed in %s: %w", si.name, err)
	}
	return nil
}

// OnDestroy calls the script's onDestroy function if it exists
func (si *ScriptInstance) OnDestroy(id int) error {
	if si.onDestroy == nil {
		return nil
	}
	_, err := si.onDestroy(goja.Undefined(), si.env.vm.ToValue(id))
	if err != nil {
		return fmt.Errorf("onDestroy failed in %s: %w", si.name, err)
	}
	return nil
}

// Call invokes any exported function by name
func (si *ScriptInstance) Call(fnName string, args ...interface{}) (goja.Value, error) {
	v := si.env.vm.Get(fnName)
	if v == nil {
		return nil, fmt.Errorf("function %s not found in script %s", fnName, si.name)
	}
	fn, ok := goja.AssertFunction(v)
	if !ok {
		return nil, fmt.Errorf("%s is not a function in script %s", fnName, si.name)
	}

	jsArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		jsArgs[i] = si.env.vm.ToValue(arg)
	}

	result, err := fn(goja.Undefined(), jsArgs...)
	if err != nil {
		return nil, fmt.Errorf("call to %s failed in %s: %w", fnName, si.name, err)
	}
	return result, nil
}
