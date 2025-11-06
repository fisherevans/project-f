package rpg

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type StateValue struct {
	key    string
	value  any
	exists bool
}

func newStateValue(key string, value any, exists bool) *StateValue {
	return &StateValue{
		key:    key,
		value:  value,
		exists: exists,
	}
}

func newExistingStateValue(key string, value any) *StateValue {
	return newStateValue(key, value, true)
}

func newMissingStateValue(key string) *StateValue {
	return newStateValue(key, nil, false)
}

func (v *StateValue) errorLog() *zerolog.Event {
	return log.Error().Str("key", v.key)
}

func (v *StateValue) Exists() bool {
	return v.exists
}

func (v *StateValue) Value() any {
	return v.value
}

func (v *StateValue) AsString(defaultValue string) string {
	if !v.exists {
		return defaultValue
	}
	str, ok := v.value.(string)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a string")
		return defaultValue
	}
	return str
}

func (v *StateValue) AsInt(defaultValue int) int {
	if !v.exists {
		return defaultValue
	}
	i, ok := v.value.(int)
	if !ok {
		v.errorLog().Msgf("value is set, but is not an int")
		return defaultValue
	}
	return i
}

func (v *StateValue) AsFloat(defaultValue float64) float64 {
	if !v.exists {
		return defaultValue
	}
	f, ok := v.value.(float64)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a float64")
		return defaultValue
	}
	return f
}

func (v *StateValue) AsBool(defaultValue bool) bool {
	if !v.exists {
		return defaultValue
	}
	b, ok := v.value.(bool)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a bool")
		return defaultValue
	}
	return b
}

func (v *StateValue) AsStringSlice(defaultValue []string) []string {
	if !v.exists {
		return defaultValue
	}
	slice, ok := v.value.([]string)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a []string")
		return defaultValue
	}
	return slice
}

func (v *StateValue) AsIntSlice(defaultValue []int) []int {
	if !v.exists {
		return defaultValue
	}
	slice, ok := v.value.([]int)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a []int")
		return defaultValue
	}
	return slice
}

func (v *StateValue) AsFloatSlice(defaultValue []float64) []float64 {
	if !v.exists {
		return defaultValue
	}
	slice, ok := v.value.([]float64)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a []float64")
		return defaultValue
	}
	return slice
}

func (v *StateValue) AsBoolSlice(defaultValue []bool) []bool {
	if !v.exists {
		return defaultValue
	}
	slice, ok := v.value.([]bool)
	if !ok {
		v.errorLog().Msgf("value is set, but is not a []bool")
		return defaultValue
	}
	return slice
}

func (v *StateValue) As(target any) bool {
	if !v.exists {
		return false
	}

	if target == nil {
		log.Error().Msg("target is nil")
		return false
	}

	// Ensure target is a pointer
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr {
		v.errorLog().Msgf("target must be a pointer, got %T", target)
		return false
	}

	// Ensure target doesn't point to a pointer
	targetElem := targetVal.Elem()
	if targetElem.Kind() == reflect.Ptr {
		v.errorLog().Msgf("target must not be a pointer to a pointer")
		return false
	}

	valueVal := reflect.ValueOf(v.value)
	targetType := targetElem.Type()
	valueType := valueVal.Type()

	// If types match exactly, assign directly
	if valueType.AssignableTo(targetType) {
		targetElem.Set(valueVal)
		return true
	}

	// If value is a map (likely from YAML unmarshaling), try to convert it to target type
	if valueVal.Kind() == reflect.Map {
		if err := remarshalToType(v.value, target); err != nil {
			v.errorLog().Err(err).Msgf("failed to convert map to %T", target)
			return false
		}
		return true
	}

	// Types don't match and can't convert
	v.errorLog().Msgf("type mismatch: stored type %T, requested type %T", v.value, target)
	return false
}

// remarshalToType converts a generic map (from YAML) to the target struct type
func remarshalToType(value any, target any) error {
	// Marshal the value back to YAML
	data, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Unmarshal into target with strict mode to catch unknown fields
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target type: %w", err)
	}

	return nil
}

type State struct {
	values map[string]any
}

func NewState() *State {
	return &State{
		values: make(map[string]any),
	}
}

func (s *State) Delete(key string) *StateValue {
	oldValue, exists := s.values[key]
	if !exists {
		return newMissingStateValue(key)
	}
	delete(s.values, key)
	return newExistingStateValue(key, oldValue)
}

func (s *State) Set(key string, value any) *StateValue {
	if s.values == nil {
		s.values = make(map[string]any)
	}

	// Get old value for listener notification
	oldValue, oldExists := s.values[key]

	// Normalize the value to ensure it's serializable
	normalizedValue := s.normalizeValue(value)
	if normalizedValue != nil {
		s.values[key] = normalizedValue
	} else {
		log.Error().Str("key", key).Msgf("invalid value type: %T", value)
	}
	return newStateValue(key, oldValue, oldExists)
}

// normalizeValue ensures the value is one of the allowed types or a struct
// Returns nil if the value type is not allowed
func (s *State) normalizeValue(value any) any {
	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)
	kind := val.Kind()

	// If it's a pointer, dereference it
	if kind == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		kind = val.Kind()
		value = val.Interface()
	}

	// Check for allowed primitive types
	switch kind {
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
		return value

	case reflect.Slice:
		// Check if it's a slice of allowed types
		elemKind := val.Type().Elem().Kind()
		switch elemKind {
		case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
			return value
		default:
			log.Error().Msgf("slice element type not allowed: %v", elemKind)
			return nil
		}

	case reflect.Struct:
		// Structs are allowed
		return value

	default:
		log.Error().Msgf("value type not allowed: %v", kind)
		return nil
	}
}

func (s *State) Get(key string) *StateValue {
	if s.values == nil {
		return newMissingStateValue(key)
	}
	v, exists := s.values[key]
	return newStateValue(key, v, exists)
}

// MarshalYAML makes State serialize as its inner map.
func (s *State) MarshalYAML() (any, error) {
	return s.values, nil
}

// UnmarshalYAML fills State from a YAML map.
func (s *State) UnmarshalYAML(n *yaml.Node) error {
	var m map[string]any
	if err := n.Decode(&m); err != nil {
		return err
	}
	s.values = m
	return nil
}

type MutableState interface {
	ReadableState
	Set(key string, value any) *StateValue
	Delete(key string) *StateValue
}

type ReadableState interface {
	Get(key string) *StateValue
}
