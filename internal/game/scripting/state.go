package scripting

import (
	"encoding/json"
	"fmt"
	"sync"
)

// ScriptState represents the persisted state for an entity's script
type ScriptState struct {
	Module  string                 `json:"module"`
	Version int                    `json:"version"`
	Data    map[string]interface{} `json:"data"`
}

// StateStore manages entity script states
type StateStore struct {
	mu     sync.RWMutex
	states map[string]*ScriptState
}

// NewStateStore creates a new state store
func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]*ScriptState),
	}
}

// Get retrieves the state for an entity
func (s *StateStore) Get(entityID string) (*ScriptState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.states[entityID]
	return state, ok
}

// Set stores the state for an entity
func (s *StateStore) Set(entityID string, state *ScriptState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[entityID] = state
}

// Delete removes the state for an entity
func (s *StateStore) Delete(entityID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, entityID)
}

// GetAll returns all entity states
func (s *StateStore) GetAll() map[string]*ScriptState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*ScriptState, len(s.states))
	for k, v := range s.states {
		result[k] = v
	}
	return result
}

// Serialize converts all states to JSON
func (s *StateStore) Serialize() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(s.states)
}

// Deserialize loads states from JSON
func (s *StateStore) Deserialize(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return json.Unmarshal(data, &s.states)
}

// MigrateState handles state version migrations
func MigrateState(state *ScriptState, targetVersion int, migrator StateMigrator) error {
	if state.Version == targetVersion {
		return nil
	}

	if state.Version > targetVersion {
		return fmt.Errorf("cannot downgrade state from version %d to %d", state.Version, targetVersion)
	}

	// Migrate step by step
	for state.Version < targetVersion {
		nextVersion := state.Version + 1
		newData, err := migrator.Migrate(state.Data, state.Version, nextVersion)
		if err != nil {
			return fmt.Errorf("failed to migrate from version %d to %d: %w", state.Version, nextVersion, err)
		}
		state.Data = newData
		state.Version = nextVersion
	}

	return nil
}

// StateMigrator handles state schema migrations
type StateMigrator interface {
	Migrate(data map[string]interface{}, fromVersion, toVersion int) (map[string]interface{}, error)
}

// DefaultMigrator provides a simple migration implementation
type DefaultMigrator struct {
	migrations map[int]MigrationFunc
}

// MigrationFunc transforms state from one version to the next
type MigrationFunc func(data map[string]interface{}) (map[string]interface{}, error)

// NewDefaultMigrator creates a new default migrator
func NewDefaultMigrator() *DefaultMigrator {
	return &DefaultMigrator{
		migrations: make(map[int]MigrationFunc),
	}
}

// RegisterMigration registers a migration function for a specific version
func (m *DefaultMigrator) RegisterMigration(toVersion int, fn MigrationFunc) {
	m.migrations[toVersion] = fn
}

// Migrate performs the migration
func (m *DefaultMigrator) Migrate(data map[string]interface{}, fromVersion, toVersion int) (map[string]interface{}, error) {
	fn, ok := m.migrations[toVersion]
	if !ok {
		// No migration needed, return as-is
		return data, nil
	}

	return fn(data)
}

// SaveFormat represents the complete save file structure
type SaveFormat struct {
	World    WorldSave              `json:"world"`
	Entities []EntitySave           `json:"entities"`
	Plans    []PlanSave             `json:"plans"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// WorldSave represents the world state in a save file
type WorldSave struct {
	Vars map[string]interface{} `json:"vars"`
}

// EntitySave represents an entity's script state in a save file
type EntitySave struct {
	ID     string       `json:"id"`
	Script *ScriptState `json:"script"`
}

// PlanSave represents a plan's execution state in a save file
type PlanSave struct {
	EntityID string      `json:"entity"`
	Cursor   *PlanCursor `json:"cursor"`
}

// SaveSystem handles serialization of the entire scripting system
type SaveSystem struct {
	dispatcher *EventDispatcher
	worldVars  *WorldVarStore
}

// NewSaveSystem creates a new save system
func NewSaveSystem(dispatcher *EventDispatcher, worldVars *WorldVarStore) *SaveSystem {
	return &SaveSystem{
		dispatcher: dispatcher,
		worldVars:  worldVars,
	}
}

// Save creates a save file
func (s *SaveSystem) Save() (*SaveFormat, error) {
	save := &SaveFormat{
		World: WorldSave{
			Vars: s.worldVars.GetAll(),
		},
		Entities: make([]EntitySave, 0),
		Plans:    make([]PlanSave, 0),
		Metadata: make(map[string]interface{}),
	}

	// Save entity states
	states := s.dispatcher.stateStore.GetAll()
	for entityID, state := range states {
		save.Entities = append(save.Entities, EntitySave{
			ID:     entityID,
			Script: state,
		})
	}

	// Save active plans
	plans := s.dispatcher.planRunner.GetActivePlans()
	for _, cursor := range plans {
		save.Plans = append(save.Plans, PlanSave{
			EntityID: cursor.EntityID,
			Cursor:   cursor,
		})
	}

	return save, nil
}

// Load restores from a save file
func (s *SaveSystem) Load(save *SaveFormat) error {
	// Restore world vars
	for key, value := range save.World.Vars {
		s.worldVars.Set(key, value)
	}

	// Restore entity states
	for _, entitySave := range save.Entities {
		if err := s.dispatcher.AttachScript(entitySave.ID, entitySave.Script.Module, entitySave.Script.Data); err != nil {
			return fmt.Errorf("failed to restore entity %s: %w", entitySave.ID, err)
		}
	}

	// Restore plans (simplified - would need full plan structure)
	// for _, planSave := range save.Plans {
	// 	s.dispatcher.planRunner.activePlans[planSave.Cursor.PlanID] = planSave.Cursor
	// }

	return nil
}

// SerializeToJSON converts a save to JSON
func (s *SaveSystem) SerializeToJSON(save *SaveFormat) ([]byte, error) {
	return json.MarshalIndent(save, "", "  ")
}

// DeserializeFromJSON loads a save from JSON
func (s *SaveSystem) DeserializeFromJSON(data []byte) (*SaveFormat, error) {
	var save SaveFormat
	if err := json.Unmarshal(data, &save); err != nil {
		return nil, err
	}
	return &save, nil
}
