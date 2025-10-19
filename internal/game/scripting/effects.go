package scripting

import (
	"encoding/json"
	"fmt"
)

// EffectType represents the type of effect to apply
type EffectType string

const (
	EffectTypeSetVar      EffectType = "SetVar"
	EffectTypeIncVar      EffectType = "IncVar"
	EffectTypeClearVar    EffectType = "ClearVar"
	EffectTypeMoveTo      EffectType = "MoveTo"
	EffectTypeOpenDoor    EffectType = "OpenDoor"
	EffectTypeCloseDoor   EffectType = "CloseDoor"
	EffectTypeTrigger     EffectType = "Trigger"
	EffectTypeStartBattle EffectType = "StartBattle"
	EffectTypeSetCamera   EffectType = "SetCamera"
	EffectTypePlayMusic   EffectType = "PlayMusic"
	EffectTypePlaySound   EffectType = "PlaySound"
	EffectTypeShowDialog  EffectType = "ShowDialog"
	EffectTypeStartTimer  EffectType = "StartTimer"
	EffectTypeStopTimer   EffectType = "StopTimer"
	EffectTypeSpawnEntity EffectType = "SpawnEntity"
	EffectTypeRemoveEntity EffectType = "RemoveEntity"
	EffectTypeTransaction EffectType = "Transaction"
)

// Effect represents an action to be applied to the game world
type Effect struct {
	Type     EffectType             `json:"type"`
	EntityID string                 `json:"entityId,omitempty"`
	Priority int                    `json:"priority,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

// Validate checks if the effect has all required fields
func (e *Effect) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("effect type is required")
	}
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}

	switch e.Type {
	case EffectTypeSetVar, EffectTypeIncVar, EffectTypeClearVar:
		if _, ok := e.Data["key"]; !ok {
			return fmt.Errorf("%s effect requires 'key' field", e.Type)
		}
		if e.Type == EffectTypeSetVar || e.Type == EffectTypeIncVar {
			if _, ok := e.Data["value"]; !ok {
				return fmt.Errorf("%s effect requires 'value' field", e.Type)
			}
		}
	case EffectTypeMoveTo:
		if _, ok := e.Data["x"]; !ok {
			return fmt.Errorf("MoveTo effect requires 'x' field")
		}
		if _, ok := e.Data["y"]; !ok {
			return fmt.Errorf("MoveTo effect requires 'y' field")
		}
	case EffectTypeOpenDoor, EffectTypeCloseDoor:
		if _, ok := e.Data["doorId"]; !ok {
			return fmt.Errorf("%s effect requires 'doorId' field", e.Type)
		}
	case EffectTypeTrigger:
		if _, ok := e.Data["triggerName"]; !ok {
			return fmt.Errorf("Trigger effect requires 'triggerName' field")
		}
	case EffectTypeStartBattle:
		if _, ok := e.Data["battleId"]; !ok {
			return fmt.Errorf("StartBattle effect requires 'battleId' field")
		}
	case EffectTypeSetCamera:
		if _, ok := e.Data["target"]; !ok {
			return fmt.Errorf("SetCamera effect requires 'target' field")
		}
	case EffectTypePlayMusic:
		if _, ok := e.Data["musicId"]; !ok {
			return fmt.Errorf("PlayMusic effect requires 'musicId' field")
		}
	case EffectTypePlaySound:
		if _, ok := e.Data["soundId"]; !ok {
			return fmt.Errorf("PlaySound effect requires 'soundId' field")
		}
	case EffectTypeShowDialog:
		if _, ok := e.Data["text"]; !ok {
			return fmt.Errorf("ShowDialog effect requires 'text' field")
		}
	case EffectTypeStartTimer:
		if _, ok := e.Data["timerName"]; !ok {
			return fmt.Errorf("StartTimer effect requires 'timerName' field")
		}
		if _, ok := e.Data["duration"]; !ok {
			return fmt.Errorf("StartTimer effect requires 'duration' field")
		}
	case EffectTypeStopTimer:
		if _, ok := e.Data["timerName"]; !ok {
			return fmt.Errorf("StopTimer effect requires 'timerName' field")
		}
	case EffectTypeSpawnEntity:
		if _, ok := e.Data["entityType"]; !ok {
			return fmt.Errorf("SpawnEntity effect requires 'entityType' field")
		}
	case EffectTypeRemoveEntity:
		if e.EntityID == "" {
			if _, ok := e.Data["entityId"]; !ok {
				return fmt.Errorf("RemoveEntity effect requires 'entityId' field")
			}
		}
	case EffectTypeTransaction:
		if _, ok := e.Data["effects"]; !ok {
			return fmt.Errorf("Transaction effect requires 'effects' field")
		}
	}

	return nil
}

// EffectApplier handles the application of effects to the game world
type EffectApplier struct {
	worldVars *WorldVarStore
	conflicts []ConflictReport
}

// NewEffectApplier creates a new effect applier
func NewEffectApplier(worldVars *WorldVarStore) *EffectApplier {
	return &EffectApplier{
		worldVars: worldVars,
		conflicts: make([]ConflictReport, 0),
	}
}

// ConflictReport describes a conflict between effects
type ConflictReport struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Effects     []Effect `json:"effects"`
}

// ApplyEffects validates and applies a batch of effects
func (a *EffectApplier) ApplyEffects(effects []Effect) error {
	a.conflicts = make([]ConflictReport, 0)

	// Validate all effects first
	for i, effect := range effects {
		if err := effect.Validate(); err != nil {
			return fmt.Errorf("effect %d validation failed: %w", i, err)
		}
	}

	// Group effects by type for conflict detection
	varEffects := make([]Effect, 0)
	doorEffects := make([]Effect, 0)
	cameraEffects := make([]Effect, 0)
	musicEffects := make([]Effect, 0)
	transactionEffects := make([]Effect, 0)
	otherEffects := make([]Effect, 0)

	for _, effect := range effects {
		switch effect.Type {
		case EffectTypeSetVar, EffectTypeIncVar, EffectTypeClearVar:
			varEffects = append(varEffects, effect)
		case EffectTypeOpenDoor, EffectTypeCloseDoor:
			doorEffects = append(doorEffects, effect)
		case EffectTypeSetCamera:
			cameraEffects = append(cameraEffects, effect)
		case EffectTypePlayMusic:
			musicEffects = append(musicEffects, effect)
		case EffectTypeTransaction:
			transactionEffects = append(transactionEffects, effect)
		default:
			otherEffects = append(otherEffects, effect)
		}
	}

	// Apply transactions first (all-or-nothing)
	for _, txEffect := range transactionEffects {
		if err := a.applyTransaction(txEffect); err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}
	}

	// Apply variable effects (last-writer-wins by priority)
	if err := a.applyVarEffects(varEffects); err != nil {
		return err
	}

	// Apply door effects (detect conflicts)
	if err := a.applyDoorEffects(doorEffects); err != nil {
		return err
	}

	// Apply camera effects (highest priority wins)
	if err := a.applyCameraEffects(cameraEffects); err != nil {
		return err
	}

	// Apply music effects (highest priority wins)
	if err := a.applyMusicEffects(musicEffects); err != nil {
		return err
	}

	// Apply other effects
	for _, effect := range otherEffects {
		if err := a.applySingleEffect(effect); err != nil {
			return err
		}
	}

	return nil
}

// GetConflicts returns all detected conflicts
func (a *EffectApplier) GetConflicts() []ConflictReport {
	return a.conflicts
}

func (a *EffectApplier) applyVarEffects(effects []Effect) error {
	// Group by key
	byKey := make(map[string][]Effect)
	for _, effect := range effects {
		key := effect.Data["key"].(string)
		byKey[key] = append(byKey[key], effect)
	}

	// Apply last-writer-wins per key (by priority, then order)
	for key, keyEffects := range byKey {
		if len(keyEffects) > 1 {
			// Sort by priority (descending), then by order
			// For now, just take the last one
			a.conflicts = append(a.conflicts, ConflictReport{
				Type:        "VarConflict",
				Description: fmt.Sprintf("Multiple effects targeting variable '%s'", key),
				Effects:     keyEffects,
			})
		}

		// Apply the last effect
		effect := keyEffects[len(keyEffects)-1]
		if err := a.applySingleEffect(effect); err != nil {
			return err
		}
	}

	return nil
}

func (a *EffectApplier) applyDoorEffects(effects []Effect) error {
	// Group by doorId
	byDoor := make(map[string][]Effect)
	for _, effect := range effects {
		doorId := effect.Data["doorId"].(string)
		byDoor[doorId] = append(byDoor[doorId], effect)
	}

	// Detect conflicts
	for doorId, doorEffects := range byDoor {
		if len(doorEffects) > 1 {
			a.conflicts = append(a.conflicts, ConflictReport{
				Type:        "DoorConflict",
				Description: fmt.Sprintf("Multiple effects targeting door '%s'", doorId),
				Effects:     doorEffects,
			})
		}

		// Apply the last effect
		effect := doorEffects[len(doorEffects)-1]
		if err := a.applySingleEffect(effect); err != nil {
			return err
		}
	}

	return nil
}

func (a *EffectApplier) applyCameraEffects(effects []Effect) error {
	if len(effects) == 0 {
		return nil
	}

	if len(effects) > 1 {
		a.conflicts = append(a.conflicts, ConflictReport{
			Type:        "CameraConflict",
			Description: "Multiple camera effects in same tick",
			Effects:     effects,
		})
	}

	// Apply highest priority (last in list for now)
	return a.applySingleEffect(effects[len(effects)-1])
}

func (a *EffectApplier) applyMusicEffects(effects []Effect) error {
	if len(effects) == 0 {
		return nil
	}

	if len(effects) > 1 {
		a.conflicts = append(a.conflicts, ConflictReport{
			Type:        "MusicConflict",
			Description: "Multiple music effects in same tick",
			Effects:     effects,
		})
	}

	// Apply highest priority (last in list for now)
	return a.applySingleEffect(effects[len(effects)-1])
}

func (a *EffectApplier) applyTransaction(effect Effect) error {
	// Extract nested effects
	effectsData, ok := effect.Data["effects"]
	if !ok {
		return fmt.Errorf("transaction missing effects")
	}

	// Convert to Effect slice
	jsonBytes, err := json.Marshal(effectsData)
	if err != nil {
		return err
	}

	var nestedEffects []Effect
	if err := json.Unmarshal(jsonBytes, &nestedEffects); err != nil {
		return err
	}

	// Validate all nested effects first
	for i, nested := range nestedEffects {
		if err := nested.Validate(); err != nil {
			return fmt.Errorf("transaction effect %d invalid: %w", i, err)
		}
	}

	// Apply all (in a real implementation, this would be atomic)
	for _, nested := range nestedEffects {
		if err := a.applySingleEffect(nested); err != nil {
			return fmt.Errorf("transaction failed at effect: %w", err)
		}
	}

	return nil
}

func (a *EffectApplier) applySingleEffect(effect Effect) error {
	switch effect.Type {
	case EffectTypeSetVar:
		key := effect.Data["key"].(string)
		value := effect.Data["value"]
		a.worldVars.Set(key, value)

	case EffectTypeIncVar:
		key := effect.Data["key"].(string)
		delta := effect.Data["value"]
		current := a.worldVars.Get(key)
		
		// Handle numeric increment
		var newValue interface{}
		switch v := current.(type) {
		case int:
			newValue = v + int(delta.(float64))
		case float64:
			newValue = v + delta.(float64)
		default:
			newValue = delta
		}
		a.worldVars.Set(key, newValue)

	case EffectTypeClearVar:
		key := effect.Data["key"].(string)
		a.worldVars.Clear(key)

	// Other effect types would be implemented here
	// For now, they're no-ops (would integrate with actual game systems)
	case EffectTypeMoveTo, EffectTypeOpenDoor, EffectTypeCloseDoor,
		EffectTypeTrigger, EffectTypeStartBattle, EffectTypeSetCamera,
		EffectTypePlayMusic, EffectTypePlaySound, EffectTypeShowDialog,
		EffectTypeStartTimer, EffectTypeStopTimer, EffectTypeSpawnEntity,
		EffectTypeRemoveEntity:
		// These would integrate with the actual game systems
		// For now, just log them
		// log.Debug().Str("type", string(effect.Type)).Interface("data", effect.Data).Msg("Effect applied")

	default:
		return fmt.Errorf("unknown effect type: %s", effect.Type)
	}

	return nil
}
