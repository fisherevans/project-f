package rpg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	GlobalKeyElythium = "elythium"
)

type GlobalValue struct {
	key    string
	value  any
	exists bool
}

func newGlobalValue(key string, value any, exists bool) *GlobalValue {
	return &GlobalValue{
		key:    key,
		value:  value,
		exists: exists,
	}
}

func newExistingGlobalValue(key string, value any) *GlobalValue {
	return newGlobalValue(key, value, true)
}

func newMissingGlobalValue(key string) *GlobalValue {
	return newGlobalValue(key, nil, false)
}

func (v *GlobalValue) errorLog() *zerolog.Event {
	return log.Error().Str("key", v.key)
}

func (v *GlobalValue) Exists() bool {
	return v.exists
}

func (v *GlobalValue) Value() any {
	return v.value
}

func (v *GlobalValue) AsString(defaultValue string) string {
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

func (v *GlobalValue) AsInt(defaultValue int) int {
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

func (v *GlobalValue) AsFloat(defaultValue float64) float64 {
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

func (v *GlobalValue) AsBool(defaultValue bool) bool {
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

func (v *GlobalValue) AsStringSlice(defaultValue []string) []string {
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

func (v *GlobalValue) AsIntSlice(defaultValue []int) []int {
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

func (v *GlobalValue) AsFloatSlice(defaultValue []float64) []float64 {
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

func (v *GlobalValue) AsBoolSlice(defaultValue []bool) []bool {
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

func (v *GlobalValue) As(target any) bool {
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

type defaultGlobals struct {
	values map[string]any
}

func (s *defaultGlobals) Delete(key string) *GlobalValue {
	oldValue, exists := s.values[key]
	if !exists {
		return newMissingGlobalValue(key)
	}
	delete(s.values, key)
	return newExistingGlobalValue(key, oldValue)
}

func (s *defaultGlobals) Set(key string, value any) *GlobalValue {
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
		s.Delete(key)
	}
	return newGlobalValue(key, oldValue, oldExists)
}

// normalizeValue ensures the value is one of the allowed types or a struct
// Returns nil if the value type is not allowed
func (s *defaultGlobals) normalizeValue(value any) any {
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

func (s *defaultGlobals) Get(key string) *GlobalValue {
	if s.values == nil {
		return newMissingGlobalValue(key)
	}
	v, exists := s.values[key]
	return newGlobalValue(key, v, exists)
}

func (s *defaultGlobals) KeysWithPrefix(prefix string) []string {
	var keys []string
	for k := range s.values {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys
}

// MarshalYAML makes defaultGlobals serialize as its inner map.
func (s *defaultGlobals) MarshalYAML() (any, error) {
	return s.values, nil
}

// UnmarshalYAML fills defaultGlobals from a YAML map.
func (s *defaultGlobals) UnmarshalYAML(n *yaml.Node) error {
	var m map[string]any
	if err := n.Decode(&m); err != nil {
		return err
	}
	s.values = m
	return nil
}

type Globals interface {
	GlobalsReader
	Set(key string, value any) *GlobalValue
	Delete(key string) *GlobalValue
	KeysWithPrefix(prefix string) []string
}

type GlobalsReader interface {
	Get(key string) *GlobalValue
}
