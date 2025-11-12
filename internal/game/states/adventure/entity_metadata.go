package adventure

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type MetadataKey[T any] struct {
	name           string
	defaultFactory func() T
}

// Exists returns if a value is set for this key with the correct type
func (k MetadataKey[T]) Exists(e EntityReader) bool {
	if e == nil {
		return false
	}
	val, ok := e.GetMetadata().get(k.name)
	if !ok {
		return false
	}
	typed, ok := val.(T)
	if !ok {
		log.Error().Type("actual", val).Str("name", k.name).Type("wanted", typed).Msg("metadata value is not of type")
		return false
	}
	return true
}

// Get retrieves the typed value from the entity, or returns the default if not set
func (k MetadataKey[T]) Get(e EntityReader) T {
	if e == nil {
		log.Error().Msg("entity is nil")
		return k.defaultFactory()
	}
	val, ok := e.GetMetadata().get(k.name)
	if !ok {
		newVal := k.defaultFactory()
		e.GetMetadata().set(k.name, newVal)
		return newVal
	}
	typed, ok := val.(T)
	if !ok {
		log.Error().Type("actual", val).Str("name", k.name).Type("wanted", typed).Msg("metadata value is not of type")
		return k.defaultFactory()
	}
	return typed
}

func (k MetadataKey[T]) loadMetadata(e Entity, metadata any) {
	v, ok := k.constructFromMetadataBlob(metadata)
	if !ok {
		return
	}
	log.Debug().Str("entityId", e.GetId()).Interface("value", v).Str("name", k.name).Msg("loaded metadata")
	k.Set(e, v)
}

func (k MetadataKey[T]) constructFromMetadataBlob(metadata any) (T, bool) {
	var metadataMap map[string]any

	// Check if metadata is a string that needs to be unmarshalled
	if metadataStr, ok := metadata.(string); ok {
		err := yaml.Unmarshal([]byte(metadataStr), &metadataMap)
		if err != nil {
			log.Error().Err(err).Msgf("failed to unmarshal metadata string to map")
			return k.defaultFactory(), false
		}
	} else {
		// Try to cast metadata to map[string]any directly
		var ok bool
		metadataMap, ok = metadata.(map[string]any)
		if !ok {
			log.Error().Msgf("metadata is not a map[string]any or string, got %T", metadata)
			return k.defaultFactory(), false
		}
	}

	// Check if the key exists in the metadata
	rawValue, exists := metadataMap[k.name]
	if !exists {
		// Key not present, return default without error
		return k.defaultFactory(), false
	}

	// Marshal the raw value to YAML bytes, then unmarshal into T
	yamlBytes, err := yaml.Marshal(rawValue)
	if err != nil {
		log.Error().Err(err).Str("key", k.name).Msgf("failed to marshal metadata value")
		return k.defaultFactory(), false
	}

	// Create a new instance using the default factory
	result := k.defaultFactory()

	// Unmarshal into the result
	err = yaml.Unmarshal(yamlBytes, &result)
	if err != nil {
		log.Error().Err(err).Str("key", k.name).Msgf("failed to unmarshal metadata value into type")
		return k.defaultFactory(), false
	}

	return result, true
}

// Set stores the typed value in the entity
func (k MetadataKey[T]) Set(e Entity, value T) {
	e.GetMetadata().set(k.name, value)
}

type EntityMetadata struct {
	values map[string]any
}

func (m *EntityMetadata) set(key string, value any) {
	if m.values == nil {
		m.values = make(map[string]any)
	}
	m.values[key] = value
}

func (m *EntityMetadata) get(key string) (any, bool) {
	if m.values == nil {
		m.values = make(map[string]any)
	}
	val, ok := m.values[key]
	return val, ok
}
