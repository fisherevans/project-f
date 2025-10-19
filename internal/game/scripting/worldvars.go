package scripting

import (
	"encoding/json"
	"strings"
	"sync"
)

// WorldVarStore manages the global namespaced key-value store
type WorldVarStore struct {
	mu        sync.RWMutex
	vars      map[string]interface{}
	listeners map[string][]VarChangeListener
}

// VarChangeListener is called when a variable changes
type VarChangeListener func(key string, oldValue, newValue interface{})

// NewWorldVarStore creates a new world variable store
func NewWorldVarStore() *WorldVarStore {
	return &WorldVarStore{
		vars:      make(map[string]interface{}),
		listeners: make(map[string][]VarChangeListener),
	}
}

// Get retrieves a variable value
func (w *WorldVarStore) Get(key string) interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.vars[key]
}

// GetString retrieves a variable as a string
func (w *WorldVarStore) GetString(key string) string {
	val := w.Get(key)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetBool retrieves a variable as a bool
func (w *WorldVarStore) GetBool(key string) bool {
	val := w.Get(key)
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetInt retrieves a variable as an int
func (w *WorldVarStore) GetInt(key string) int {
	val := w.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

// GetFloat retrieves a variable as a float64
func (w *WorldVarStore) GetFloat(key string) float64 {
	val := w.Get(key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0
}

// Set sets a variable value and notifies listeners
func (w *WorldVarStore) Set(key string, value interface{}) {
	w.mu.Lock()
	oldValue := w.vars[key]
	w.vars[key] = value
	w.mu.Unlock()

	w.notifyListeners(key, oldValue, value)
}

// Clear removes a variable
func (w *WorldVarStore) Clear(key string) {
	w.mu.Lock()
	oldValue := w.vars[key]
	delete(w.vars, key)
	w.mu.Unlock()

	w.notifyListeners(key, oldValue, nil)
}

// Has checks if a variable exists
func (w *WorldVarStore) Has(key string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, exists := w.vars[key]
	return exists
}

// GetAll returns a copy of all variables
func (w *WorldVarStore) GetAll() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[string]interface{}, len(w.vars))
	for k, v := range w.vars {
		result[k] = v
	}
	return result
}

// GetByPrefix returns all variables matching a prefix
func (w *WorldVarStore) GetByPrefix(prefix string) map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range w.vars {
		if strings.HasPrefix(k, prefix) {
			result[k] = v
		}
	}
	return result
}

// AddListener registers a listener for variable changes
// The pattern can be:
// - Exact key: "map.meadow.gate"
// - Prefix pattern: "map.meadow.*"
func (w *WorldVarStore) AddListener(pattern string, listener VarChangeListener) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.listeners[pattern] = append(w.listeners[pattern], listener)
}

// RemoveListener removes all listeners for a pattern
func (w *WorldVarStore) RemoveListener(pattern string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.listeners, pattern)
}

func (w *WorldVarStore) notifyListeners(key string, oldValue, newValue interface{}) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for pattern, listeners := range w.listeners {
		if w.matchesPattern(key, pattern) {
			for _, listener := range listeners {
				listener(key, oldValue, newValue)
			}
		}
	}
}

func (w *WorldVarStore) matchesPattern(key, pattern string) bool {
	// Exact match
	if key == pattern {
		return true
	}

	// Prefix match (pattern ends with .*)
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix)
	}

	return false
}

// Serialize converts the store to JSON
func (w *WorldVarStore) Serialize() ([]byte, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return json.Marshal(w.vars)
}

// Deserialize loads the store from JSON
func (w *WorldVarStore) Deserialize(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return json.Unmarshal(data, &w.vars)
}

// ValidateKey checks if a key follows the naming convention
// Valid formats: global.*, map.*, topic.*
func ValidateKey(key string) bool {
	if key == "" {
		return false
	}

	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return false
	}

	scope := parts[0]
	return scope == "global" || scope == "map" || scope == "topic"
}
