package events

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unicode"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

func (e *gojaEventHandler) parseEffects(effectsVal goja.Value) []Effect {
	// Handle effects with validation
	if isNil(effectsVal) {
		return nil
	}

	effectsObj := effectsVal.ToObject(e.vm)
	if effectsObj == nil {
		return nil
	}

	// Get array length
	lengthVal := effectsObj.Get("length")
	if isNil(lengthVal) {
		return nil
	}

	length := int(lengthVal.ToInteger())
	outputEffects := make([]Effect, 0, length)

	metadata := getEffectMetadata()

	for i := 0; i < length; i++ {
		effectVal := effectsObj.Get(strconv.Itoa(i))
		if isNil(effectVal) {
			continue
		}

		effect, err := e.parseEffect(effectVal, metadata)
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Invalid effect")
			continue
		}

		outputEffects = append(outputEffects, effect)
	}
	
	return outputEffects
}

// Cached struct metadata for fast validation
type effectFieldMap struct {
	validFields map[string]string // jsName -> goFieldName
	fieldTypes  map[string]reflect.Type
}

var (
	effectMetadata     *effectFieldMap
	effectMetadataOnce sync.Once
)

func getEffectMetadata() *effectFieldMap {
	effectMetadataOnce.Do(func() {
		effectMetadata = buildFieldMap(reflect.TypeOf(Effect{}))
	})
	return effectMetadata
}

func buildFieldMap(t reflect.Type) *effectFieldMap {
	fm := &effectFieldMap{
		validFields: make(map[string]string),
		fieldTypes:  make(map[string]reflect.Type),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		goName := field.Name

		// Get JS name (camelCase)
		jsName := toCamelCase(goName)

		fm.validFields[jsName] = goName
		fm.validFields[goName] = goName // Also accept TitleCase
		fm.fieldTypes[goName] = field.Type
	}

	return fm
}

func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// validateAndExport recursively validates all struct fields at any depth
func (e *gojaEventHandler) validateAndExport(jsVal goja.Value, target interface{}, fieldPath string) error {
	obj := jsVal.ToObject(e.vm)
	if obj == nil {
		return fmt.Errorf("%s is not an object", fieldPath)
	}

	// Get the target type
	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	// Only validate structs
	if targetType.Kind() != reflect.Struct {
		// For non-structs (maps, slices, primitives), just export directly
		return e.vm.ExportTo(jsVal, target)
	}

	// Build field map for this struct
	metadata := buildFieldMap(targetType)

	// Validate all keys in the JS object
	for _, key := range obj.Keys() {
		goFieldName, ok := metadata.validFields[key]
		if !ok {
			return fmt.Errorf("%s has unknown field '%s' (valid: %v)",
				fieldPath, key, getValidFieldNames(metadata))
		}

		// Recursively validate nested structs
		nestedVal := obj.Get(key)
		if isNil(nestedVal) {
			continue
		}

		// Get the Go field type
		fieldType, found := targetType.FieldByName(goFieldName)
		if !found {
			continue
		}

		// If it's a pointer to a struct, recursively validate
		if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			nestedTarget := reflect.New(fieldType.Type.Elem()).Interface()
			if err := e.validateAndExport(nestedVal, nestedTarget, fieldPath+"."+key); err != nil {
				return err
			}
		}
		// For maps and slices, we could add validation here too if needed
	}

	// All fields validated, now export
	return e.vm.ExportTo(jsVal, target)
}

func (e *gojaEventHandler) parseEffect(val goja.Value, metadata *effectFieldMap) (Effect, error) {
	obj := val.ToObject(e.vm)
	if obj == nil {
		return Effect{}, fmt.Errorf("effect is not an object")
	}

	var effect Effect
	effectValue := reflect.ValueOf(&effect).Elem()

	// Iterate over JS object keys
	for _, key := range obj.Keys() {
		// Validate field exists
		goFieldName, ok := metadata.validFields[key]
		if !ok {
			return effect, fmt.Errorf("unknown field '%s' (valid: %v)", key, getValidFieldNames(metadata))
		}

		// Get the value
		jsVal := obj.Get(key)
		if isNil(jsVal) {
			continue
		}

		// Set the field using reflection
		field := effectValue.FieldByName(goFieldName)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// Convert based on field type
		if field.Kind() == reflect.Ptr {
			// Handle pointer fields (e.g., *EffectDialogue)
			elemType := field.Type().Elem()
			newVal := reflect.New(elemType)

			// Recursively validate nested struct
			if err := e.validateAndExport(jsVal, newVal.Interface(), key); err != nil {
				return effect, err
			}

			field.Set(newVal)
		} else {
			// Direct export for non-pointer fields
			if err := e.vm.ExportTo(jsVal, field.Addr().Interface()); err != nil {
				return effect, fmt.Errorf("failed to convert %s: %w", key, err)
			}
		}
	}

	return effect, nil
}

func getValidFieldNames(metadata *effectFieldMap) []string {
	seen := make(map[string]bool)
	names := make([]string, 0, len(metadata.validFields))

	for jsName := range metadata.validFields {
		if !seen[jsName] {
			names = append(names, jsName)
			seen[jsName] = true
		}
	}

	return names
}
