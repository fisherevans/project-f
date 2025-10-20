package events

import (
	"fmt"
	"reflect"
	"strings"
)

type EffectType string

const (
	EffectTypeDialogue EffectType = "dialogue"
)

type Effect struct {
	Type EffectType             `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func (e Effect) Validate() error {
	switch e.Type {
	case EffectTypeDialogue:
		return requireAll(
			hasTypedField[string](e.Data, "text"),
		)
	default:
		return fmt.Errorf("unknown effect type: %s", e.Type)
	}
}

func requireAll(errors ...error) error {
	var issues []string
	for _, err := range errors {
		if err != nil {
			issues = append(issues, err.Error())
		}
	}
	if len(issues) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(issues, ", "))
	}
	return nil
}

func hasTypedField[T any](data map[string]interface{}, key string) error {
	v, ok := data[key]
	if !ok {
		return fmt.Errorf("missing %q field", key)
	}
	if _, ok := v.(T); !ok {
		want := reflect.TypeOf((*T)(nil)).Elem() // works even if T is nilable/interface
		got := reflect.TypeOf(v)
		return fmt.Errorf("%s field must be of type %s (got %s)", key, want, got)
	}
	return nil
}
