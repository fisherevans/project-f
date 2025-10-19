package scripting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// ScriptModule represents a compiled JavaScript module
type ScriptModule struct {
	Name         string
	Source       string
	Hash         string
	Program      *goja.Program
	Handlers     map[EventType]bool
	Subscriptions Subscription
}

// Subscription defines what events an entity listens to
type Subscription struct {
	EnterZone  []string               // Specific zone names
	LeaveZone  []string               // Specific zone names
	Trigger    []string               // Specific trigger names
	VarPrefix  []string               // Variable prefixes to watch
	Priority   int                    // Handler priority (higher = earlier)
	Metadata   map[string]interface{} // Additional subscription data
}

// ScriptRegistry manages compiled script modules
type ScriptRegistry struct {
	mu      sync.RWMutex
	modules map[string]*ScriptModule
	vm      *goja.Runtime
}

// NewScriptRegistry creates a new script registry
func NewScriptRegistry() *ScriptRegistry {
	return &ScriptRegistry{
		modules: make(map[string]*ScriptModule),
		vm:      goja.New(),
	}
}

// Register compiles and registers a script module
func (r *ScriptRegistry) Register(name, source string) (*ScriptModule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Calculate hash
	hash := r.calculateHash(source)

	// Check if already registered with same hash
	if existing, ok := r.modules[name]; ok {
		if existing.Hash == hash {
			return existing, nil
		}
	}

	// Compile the script
	program, err := goja.Compile(name, source, false)
	if err != nil {
		return nil, fmt.Errorf("failed to compile script %s: %w", name, err)
	}

	// Create a new VM instance to detect handlers
	vm := goja.New()
	r.setupVM(vm)

	// Run the script to get exports
	if _, err := vm.RunProgram(program); err != nil {
		return nil, fmt.Errorf("failed to run script %s: %w", name, err)
	}

	// Detect available handlers
	handlers := r.detectHandlers(vm)

	// Parse subscription metadata
	subscription := r.parseSubscription(vm)

	module := &ScriptModule{
		Name:          name,
		Source:        source,
		Hash:          hash,
		Program:       program,
		Handlers:      handlers,
		Subscriptions: subscription,
	}

	r.modules[name] = module
	return module, nil
}

// Get retrieves a registered module
func (r *ScriptRegistry) Get(name string) (*ScriptModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.modules[name]
	return module, ok
}

// GetAll returns all registered modules
func (r *ScriptRegistry) GetAll() map[string]*ScriptModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ScriptModule, len(r.modules))
	for k, v := range r.modules {
		result[k] = v
	}
	return result
}

// CreateExecutor creates a new executor for a module
func (r *ScriptRegistry) CreateExecutor(module *ScriptModule, worldVars *WorldVarStore) (*ScriptExecutor, error) {
	vm := goja.New()
	r.setupVM(vm)

	// Inject world API
	r.injectWorldAPI(vm, worldVars)

	// Run the module
	if _, err := vm.RunProgram(module.Program); err != nil {
		return nil, fmt.Errorf("failed to initialize module %s: %w", module.Name, err)
	}

	return &ScriptExecutor{
		vm:     vm,
		module: module,
	}, nil
}

func (r *ScriptRegistry) setupVM(vm *goja.Runtime) {
	// Disable dangerous APIs
	vm.Set("eval", goja.Undefined())

	// Provide deterministic time/random (would be injected from engine)
	vm.Set("getTime", func() int64 {
		// In production, this would come from the game engine's deterministic clock
		return 0
	})

	vm.Set("random", func() float64 {
		// In production, this would come from the game engine's deterministic RNG
		return 0.5
	})
}

func (r *ScriptRegistry) injectWorldAPI(vm *goja.Runtime, worldVars *WorldVarStore) {
	// Create world object
	world := vm.NewObject()

	// Variable access
	world.Set("getVar", func(key string) interface{} {
		return worldVars.Get(key)
	})

	world.Set("hasVar", func(key string) bool {
		return worldVars.Has(key)
	})

	vm.Set("world", world)
}

func (r *ScriptRegistry) detectHandlers(vm *goja.Runtime) map[EventType]bool {
	handlers := make(map[EventType]bool)

	handlerNames := map[string]EventType{
		"OnInit":        EventTypeInit,
		"OnInteract":    EventTypeInteract,
		"OnEnterZone":   EventTypeEnterZone,
		"OnLeaveZone":   EventTypeLeaveZone,
		"OnTimer":       EventTypeTimer,
		"OnTrigger":     EventTypeTrigger,
		"OnFlagChanged": EventTypeFlagChanged,
		"OnStep":        EventTypeStep,
	}

	for funcName, eventType := range handlerNames {
		val := vm.Get(funcName)
		if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
			if _, ok := goja.AssertFunction(val); ok {
				handlers[eventType] = true
			}
		}
	}

	return handlers
}

func (r *ScriptRegistry) parseSubscription(vm *goja.Runtime) Subscription {
	sub := Subscription{
		Priority: 0,
		Metadata: make(map[string]interface{}),
	}

	// Check for SUBSCRIPTION export
	subscriptionVal := vm.Get("SUBSCRIPTION")
	if subscriptionVal == nil || goja.IsUndefined(subscriptionVal) || goja.IsNull(subscriptionVal) {
		return sub
	}

	obj := subscriptionVal.ToObject(vm)
	if obj == nil {
		return sub
	}

	// Parse subscription fields
	if val := obj.Get("EnterZone"); val != nil && !goja.IsUndefined(val) {
		if arr, ok := val.Export().([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					sub.EnterZone = append(sub.EnterZone, s)
				}
			}
		} else if s, ok := val.Export().(string); ok {
			sub.EnterZone = []string{s}
		}
	}

	if val := obj.Get("LeaveZone"); val != nil && !goja.IsUndefined(val) {
		if arr, ok := val.Export().([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					sub.LeaveZone = append(sub.LeaveZone, s)
				}
			}
		} else if s, ok := val.Export().(string); ok {
			sub.LeaveZone = []string{s}
		}
	}

	if val := obj.Get("Trigger"); val != nil && !goja.IsUndefined(val) {
		if arr, ok := val.Export().([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					sub.Trigger = append(sub.Trigger, s)
				}
			}
		} else if s, ok := val.Export().(string); ok {
			sub.Trigger = []string{s}
		}
	}

	if val := obj.Get("VarPrefix"); val != nil && !goja.IsUndefined(val) {
		if arr, ok := val.Export().([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					sub.VarPrefix = append(sub.VarPrefix, s)
				}
			}
		} else if s, ok := val.Export().(string); ok {
			sub.VarPrefix = []string{s}
		}
	}

	if val := obj.Get("Priority"); val != nil && !goja.IsUndefined(val) {
		if i, ok := val.Export().(int64); ok {
			sub.Priority = int(i)
		} else if f, ok := val.Export().(float64); ok {
			sub.Priority = int(f)
		}
	}

	return sub
}

func (r *ScriptRegistry) calculateHash(source string) string {
	h := sha256.New()
	h.Write([]byte(source))
	return hex.EncodeToString(h.Sum(nil))
}

// ScriptExecutor executes event handlers for a specific module instance
type ScriptExecutor struct {
	vm     *goja.Runtime
	module *ScriptModule
}

// CallHandler invokes an event handler
func (e *ScriptExecutor) CallHandler(eventType EventType, self map[string]interface{}, event Event, state map[string]interface{}) (*HandlerResult, error) {
	handlerName := e.getHandlerName(eventType)
	if handlerName == "" {
		return nil, fmt.Errorf("no handler for event type %s", eventType)
	}

	handler := e.vm.Get(handlerName)
	if handler == nil || goja.IsUndefined(handler) || goja.IsNull(handler) {
		return nil, fmt.Errorf("handler %s not found", handlerName)
	}

	handlerFunc, ok := goja.AssertFunction(handler)
	if !ok {
		return nil, fmt.Errorf("handler %s is not a function", handlerName)
	}

	// Prepare arguments
	selfObj := e.vm.ToValue(self)
	worldObj := e.vm.Get("world")
	stateObj := e.vm.ToValue(state)
	eventObj := e.vm.ToValue(event.ToMap())

	// Call handler
	result, err := handlerFunc(goja.Undefined(), selfObj, worldObj, stateObj, eventObj)
	if err != nil {
		return nil, fmt.Errorf("handler %s failed: %w", handlerName, err)
	}

	// Parse result
	if result == nil || goja.IsUndefined(result) || goja.IsNull(result) {
		return &HandlerResult{}, nil
	}

	return ParseHandlerResult(result.Export())
}

func (e *ScriptExecutor) getHandlerName(eventType EventType) string {
	switch eventType {
	case EventTypeInit:
		return "OnInit"
	case EventTypeInteract:
		return "OnInteract"
	case EventTypeEnterZone:
		return "OnEnterZone"
	case EventTypeLeaveZone:
		return "OnLeaveZone"
	case EventTypeTimer:
		return "OnTimer"
	case EventTypeTrigger:
		return "OnTrigger"
	case EventTypeFlagChanged:
		return "OnFlagChanged"
	case EventTypeStep:
		return "OnStep"
	default:
		return ""
	}
}

// HasHandler checks if the module has a specific handler
func (e *ScriptExecutor) HasHandler(eventType EventType) bool {
	return e.module.Handlers[eventType]
}
