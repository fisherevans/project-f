package scripting

import (
	"fmt"
	"os"
	"path/filepath"
)

// ScriptLoader provides utilities for loading scripts from files
type ScriptLoader struct {
	registry  *ScriptRegistry
	baseDir   string
	loadedFiles map[string]bool
}

// NewScriptLoader creates a new script loader
func NewScriptLoader(registry *ScriptRegistry, baseDir string) *ScriptLoader {
	return &ScriptLoader{
		registry:    registry,
		baseDir:     baseDir,
		loadedFiles: make(map[string]bool),
	}
}

// LoadFile loads a single script file
func (l *ScriptLoader) LoadFile(filename string) error {
	path := filepath.Join(l.baseDir, filename)
	
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read script file %s: %w", path, err)
	}

	// Use filename without extension as module name
	moduleName := filepath.Base(filename)
	if ext := filepath.Ext(moduleName); ext != "" {
		moduleName = moduleName[:len(moduleName)-len(ext)]
	}

	_, err = l.registry.Register(moduleName, string(data))
	if err != nil {
		return fmt.Errorf("failed to register script %s: %w", moduleName, err)
	}

	l.loadedFiles[filename] = true
	return nil
}

// LoadDirectory loads all .js files from a directory
func (l *ScriptLoader) LoadDirectory(dir string) error {
	fullPath := filepath.Join(l.baseDir, dir)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".js" {
			relPath := filepath.Join(dir, entry.Name())
			if err := l.LoadFile(relPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetLoadedFiles returns a list of loaded file paths
func (l *ScriptLoader) GetLoadedFiles() []string {
	files := make([]string, 0, len(l.loadedFiles))
	for file := range l.loadedFiles {
		files = append(files, file)
	}
	return files
}

// EffectBuilder provides a fluent API for building effects
type EffectBuilder struct {
	effect Effect
}

// NewEffect creates a new effect builder
func NewEffect(effectType EffectType) *EffectBuilder {
	return &EffectBuilder{
		effect: Effect{
			Type: effectType,
			Data: make(map[string]interface{}),
		},
	}
}

// WithEntityID sets the entity ID
func (b *EffectBuilder) WithEntityID(entityID string) *EffectBuilder {
	b.effect.EntityID = entityID
	return b
}

// WithPriority sets the priority
func (b *EffectBuilder) WithPriority(priority int) *EffectBuilder {
	b.effect.Priority = priority
	return b
}

// WithData sets a data field
func (b *EffectBuilder) WithData(key string, value interface{}) *EffectBuilder {
	b.effect.Data[key] = value
	return b
}

// Build returns the constructed effect
func (b *EffectBuilder) Build() Effect {
	return b.effect
}

// Common effect builders

// SetVar creates a SetVar effect
func SetVar(key string, value interface{}) Effect {
	return NewEffect(EffectTypeSetVar).
		WithData("key", key).
		WithData("value", value).
		Build()
}

// IncVar creates an IncVar effect
func IncVar(key string, delta interface{}) Effect {
	return NewEffect(EffectTypeIncVar).
		WithData("key", key).
		WithData("value", delta).
		Build()
}

// ClearVar creates a ClearVar effect
func ClearVar(key string) Effect {
	return NewEffect(EffectTypeClearVar).
		WithData("key", key).
		Build()
}

// ShowDialog creates a ShowDialog effect
func ShowDialog(text string, speaker string) Effect {
	builder := NewEffect(EffectTypeShowDialog).
		WithData("text", text)
	
	if speaker != "" {
		builder.WithData("speaker", speaker)
	}
	
	return builder.Build()
}

// MoveTo creates a MoveTo effect
func MoveTo(x, y float64) Effect {
	return NewEffect(EffectTypeMoveTo).
		WithData("x", x).
		WithData("y", y).
		Build()
}

// OpenDoor creates an OpenDoor effect
func OpenDoor(doorID string) Effect {
	return NewEffect(EffectTypeOpenDoor).
		WithData("doorId", doorID).
		Build()
}

// CloseDoor creates a CloseDoor effect
func CloseDoor(doorID string) Effect {
	return NewEffect(EffectTypeCloseDoor).
		WithData("doorId", doorID).
		Build()
}

// Trigger creates a Trigger effect
func Trigger(triggerName string, data map[string]interface{}) Effect {
	builder := NewEffect(EffectTypeTrigger).
		WithData("triggerName", triggerName)
	
	if data != nil {
		builder.WithData("data", data)
	}
	
	return builder.Build()
}

// StartTimer creates a StartTimer effect
func StartTimer(timerName string, duration float64) Effect {
	return NewEffect(EffectTypeStartTimer).
		WithData("timerName", timerName).
		WithData("duration", duration).
		Build()
}

// StopTimer creates a StopTimer effect
func StopTimer(timerName string) Effect {
	return NewEffect(EffectTypeStopTimer).
		WithData("timerName", timerName).
		Build()
}

// PlayMusic creates a PlayMusic effect
func PlayMusic(musicID string, fadeIn float64) Effect {
	builder := NewEffect(EffectTypePlayMusic).
		WithData("musicId", musicID)
	
	if fadeIn > 0 {
		builder.WithData("fadeIn", fadeIn)
	}
	
	return builder.Build()
}

// PlaySound creates a PlaySound effect
func PlaySound(soundID string, volume float64) Effect {
	builder := NewEffect(EffectTypePlaySound).
		WithData("soundId", soundID)
	
	if volume > 0 {
		builder.WithData("volume", volume)
	}
	
	return builder.Build()
}

// PlanBuilder provides a fluent API for building plans
type PlanBuilder struct {
	plan Plan
}

// NewPlan creates a new plan builder
func NewPlan(planType PlanType) *PlanBuilder {
	return &PlanBuilder{
		plan: Plan{
			Type:     planType,
			Data:     make(map[string]interface{}),
			Children: make([]*Plan, 0),
		},
	}
}

// WithData sets a data field
func (b *PlanBuilder) WithData(key string, value interface{}) *PlanBuilder {
	b.plan.Data[key] = value
	return b
}

// AddChild adds a child plan
func (b *PlanBuilder) AddChild(child *Plan) *PlanBuilder {
	b.plan.Children = append(b.plan.Children, child)
	return b
}

// Build returns the constructed plan
func (b *PlanBuilder) Build() *Plan {
	return &b.plan
}

// Common plan builders

// Seq creates a sequential plan
func Seq(children ...*Plan) *Plan {
	builder := NewPlan(PlanTypeSeq)
	for _, child := range children {
		builder.AddChild(child)
	}
	return builder.Build()
}

// Par creates a parallel plan
func Par(children ...*Plan) *Plan {
	builder := NewPlan(PlanTypePar)
	for _, child := range children {
		builder.AddChild(child)
	}
	return builder.Build()
}

// Wait creates a wait plan
func Wait(duration float64) *Plan {
	return NewPlan(PlanTypeWait).
		WithData("duration", duration).
		Build()
}

// If creates a conditional plan
func If(condition bool, thenPlan *Plan, elsePlan *Plan) *Plan {
	builder := NewPlan(PlanTypeIf).
		WithData("condition", condition).
		AddChild(thenPlan)
	
	if elsePlan != nil {
		builder.AddChild(elsePlan)
	}
	
	return builder.Build()
}

// EffectPlan creates an effect plan node
func EffectPlan(effect Effect) *Plan {
	return &Plan{
		Type: PlanTypeEffect,
		Data: map[string]interface{}{
			"type": effect.Type,
			"data": effect.Data,
		},
	}
}

// BatchDispatcher allows dispatching multiple events efficiently
type BatchDispatcher struct {
	dispatcher *EventDispatcher
	events     []Event
}

// NewBatchDispatcher creates a new batch dispatcher
func NewBatchDispatcher(dispatcher *EventDispatcher) *BatchDispatcher {
	return &BatchDispatcher{
		dispatcher: dispatcher,
		events:     make([]Event, 0),
	}
}

// Add adds an event to the batch
func (b *BatchDispatcher) Add(event Event) {
	b.events = append(b.events, event)
}

// Dispatch dispatches all batched events
func (b *BatchDispatcher) Dispatch() error {
	for _, event := range b.events {
		if err := b.dispatcher.Dispatch(event); err != nil {
			return err
		}
	}
	b.events = b.events[:0] // Clear
	return nil
}

// Count returns the number of batched events
func (b *BatchDispatcher) Count() int {
	return len(b.events)
}

// Clear clears the batch
func (b *BatchDispatcher) Clear() {
	b.events = b.events[:0]
}
