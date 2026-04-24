package schema

import (
	_ "embed"
	"encoding/json"
)

//go:embed script_schema.json
var scriptSchemaJSON []byte

func LoadScriptSchema() (*ScriptSchema, error) {
	var s ScriptSchema
	if err := json.Unmarshal(scriptSchemaJSON, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

type ScriptSchema struct {
	StepKinds         map[string]StepKindDef  `json:"stepKinds"`
	Actions           map[string]CallableDef  `json:"actions"`
	Conditions        map[string]CallableDef  `json:"conditions"`
	BuiltinConditions map[string]ConditionDef `json:"builtinConditions"`
	EventHooks        map[string]EventHookDef `json:"eventHooks"`
	TemplateVars      []TemplateVarDef        `json:"templateVars"`
}

type ParamDef struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type StepKindDef struct {
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Category        string     `json:"category"`
	ParamStyle      string     `json:"paramStyle"`
	Params          []ParamDef `json:"params,omitempty"`
	AcceptsSubSteps bool       `json:"acceptsSubSteps,omitempty"`
}

type CallableDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Params      []ParamDef `json:"params,omitempty"`
}

type ConditionDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Params      []ParamDef `json:"params,omitempty"`
	IsComposite bool       `json:"isComposite,omitempty"`
}

type EventHookDef struct {
	Name         string     `json:"name"`
	YAMLKey      string     `json:"yamlKey"`
	Description  string     `json:"description"`
	EventType    string     `json:"eventType"`
	FilterFields []ParamDef `json:"filterFields,omitempty"`
}

type TemplateVarDef struct {
	Pattern     string `json:"pattern"`
	Description string `json:"description"`
}
