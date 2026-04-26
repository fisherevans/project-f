package adventure

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ScriptFile struct {
	Data               map[string][]string        `yaml:"data"`
	Consts             map[string]any             `yaml:"consts"`
	Handlers           map[string]*HandlerDef     `yaml:"handlers"`
	Sequences          map[string]*SequenceDef    `yaml:"sequences"`
	CustomActions      map[string]*CustomActionDef `yaml:"custom_actions"`
	PropertyTemplates  map[string]map[string]any  `yaml:"property_templates"`
}

type SequenceDef struct {
	Params []string    `yaml:"params"`
	Steps  []*StepNode `yaml:"steps"`
}

type CustomActionDef struct {
	Description string              `yaml:"description"`
	Params      []*CustomActionParam `yaml:"params"`
	Steps       []*StepNode         `yaml:"steps"`
}

type CustomActionParam struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Default     any    `yaml:"default,omitempty"`
}

type HandlerPropDef struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required,omitempty"`
	Default     any    `yaml:"default,omitempty"`
	Description string `yaml:"description"`
}

const (
	HookModeFirstMatch = ""
	HookModeAll        = "all"
)

type HookDef struct {
	Mode  string     `yaml:"mode,omitempty"`
	Rules []*RuleDef `yaml:"rules"`
}

func (h *HookDef) HasRules() bool {
	return h != nil && len(h.Rules) > 0
}

func (h *HookDef) IsAllMode() bool {
	return h != nil && h.Mode == HookModeAll
}

func (h *HookDef) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.SequenceNode {
		h.Mode = HookModeFirstMatch
		return value.Decode(&h.Rules)
	}
	if value.Kind == yaml.MappingNode {
		type hookAlias HookDef
		var alias hookAlias
		if err := value.Decode(&alias); err != nil {
			return err
		}
		*h = HookDef(alias)
		return nil
	}
	return fmt.Errorf("hook must be a list of rules or a map with mode+rules, got kind %d", value.Kind)
}

type HandlerDef struct {
	Props                []*HandlerPropDef `yaml:"props,omitempty"`
	Var                  map[string]any    `yaml:"var,omitempty"`
	OnInteractSelf       *HookDef `yaml:"on_interact_self,omitempty"`
	OnInteract           *HookDef `yaml:"on_interact,omitempty"`
	OnZoneActivity       *HookDef `yaml:"on_zone_activity,omitempty"`
	OnBroadcast          *HookDef `yaml:"on_broadcast,omitempty"`
	OnGlobalUpdated      *HookDef `yaml:"on_global_updated,omitempty"`
	OnInit               *HookDef `yaml:"on_init,omitempty"`
	OnStateEnter         *HookDef `yaml:"on_state_enter,omitempty"`
	OnCombatComplete     *HookDef `yaml:"on_combat_complete,omitempty"`
	OnTimerComplete      *HookDef `yaml:"on_timer_complete,omitempty"`
	OnMotionCompleteSelf *HookDef `yaml:"on_motion_complete_self,omitempty"`
	OnMotionComplete     *HookDef `yaml:"on_motion_complete,omitempty"`
}

type RuleDef struct {
	When   *ConditionNode `yaml:"when"`
	Filter map[string]any `yaml:"filter"`
	Steps  []*StepNode    `yaml:"steps"`
}

type StepNode struct {
	Kind   string
	Params any
}

func (s *StepNode) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return fmt.Errorf("step must be a map with exactly one key, got scalar: %s", value.Value)
	}
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("step must be a map, got kind %d", value.Kind)
	}
	if len(value.Content) != 2 {
		return fmt.Errorf("step must have exactly one key, got %d", len(value.Content)/2)
	}
	s.Kind = value.Content[0].Value
	var params any
	if err := value.Content[1].Decode(&params); err != nil {
		return fmt.Errorf("failed to decode step params for %q: %w", s.Kind, err)
	}
	s.Params = params
	return nil
}

type ConditionNode struct {
	Kind   string
	Params any
}

func (c *ConditionNode) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("condition must be a map, got kind %d", value.Kind)
	}
	if len(value.Content) != 2 {
		return fmt.Errorf("condition must have exactly one key, got %d", len(value.Content)/2)
	}
	c.Kind = value.Content[0].Value
	var params any
	if err := value.Content[1].Decode(&params); err != nil {
		return fmt.Errorf("failed to decode condition params for %q: %w", c.Kind, err)
	}
	c.Params = params
	return nil
}

func ParseScriptFile(data []byte) (*ScriptFile, error) {
	var sf ScriptFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("failed to parse script file: %w", err)
	}
	return &sf, nil
}
