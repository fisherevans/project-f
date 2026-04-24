package util

import (
	"fmt"
	"io"
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

func MergeProps(props ...*Properties) *Properties {
	kvs := make(map[string]any)
	for _, prop := range props {
		for k, v := range prop.kvs {
			kvs[k] = v
		}
	}
	return &Properties{
		kvs: kvs,
	}
}

func (p *Properties) All() map[string]any {
	if p == nil {
		return nil
	}
	return p.kvs
}

func (p *Properties) Get(key string) (any, bool) {
	if p == nil {
		return nil, false
	}
	v, ok := p.kvs[key]
	return v, ok
}

func (p *Properties) GetString(key string, defaultValue string) string {
	v, ok := p.Get(key)
	if !ok {
		return defaultValue
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func (p *Properties) GetInt(key string, defaultValue int) int {
	v, ok := p.Get(key)
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
	v, ok := p.Get(key)
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
	v, ok := p.Get(key)
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

func (p *Properties) LoadStructFromKey(key string, target any) bool {
	v, ok := p.Get(key)
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

	value := string(data)
	if strings.Trim(value, " ") == "" {
		return false
	}
	decoder := yaml.NewDecoder(strings.NewReader(value))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		if err == io.EOF {
			return false
		}
		log.Fatal().Any("props", p).Str("key", key).Type("target", target).Str("data", value).Err(err).
			Msgf("failed to load struct from key")
		return false
	}
	return true
}
