package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Properties struct {
	kvs map[string]any
}

func NewProps(kvs map[string]any) *Properties {
	if kvs == nil {
		kvs = make(map[string]any)
	}
	return &Properties{
		kvs: kvs,
	}
}

func (p *Properties) Get(key string) (any, bool) {
	v, ok := p.kvs[key]
	return v, ok
}

func (p *Properties) GetString(key string, defaultValue string) string {
	v, ok := p.kvs[key]
	if !ok {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func (p *Properties) GetInt(key string, defaultValue int) int {
	v, ok := p.kvs[key]
	if !ok {
		return defaultValue
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func (p *Properties) GetFloat(key string, defaultValue float64) float64 {
	v, ok := p.kvs[key]
	if !ok {
		return defaultValue
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func (p *Properties) GetBool(key string, defaultValue bool) bool {
	v, ok := p.kvs[key]
	if !ok {
		return defaultValue
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultValue
}

func (p *Properties) LoadStruct(target any) bool {
	data, err := yaml.Marshal(p.kvs)
	if err != nil {
		return false
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return false
	}
	return true
}

func (p *Properties) LoadStructFromKey(key string, target any) bool {
	v, ok := p.kvs[key]
	if !ok {
		return false
	}

	var data []byte
	var err error

	switch val := v.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	default:
		data, err = yaml.Marshal(val)
		if err != nil {
			return false
		}
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		log.Fatal().Str("key", key).Type("target", target).Err(err).
			Msgf("failed to load struct from key")
		return false
	}
	return true
}
