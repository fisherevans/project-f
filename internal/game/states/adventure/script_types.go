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

type HandlerDef struct {
	Props                []*HandlerPropDef `yaml:"props,omitempty"`
	Var                  map[string]any `yaml:"var,omitempty"`
	OnInteractSelf       []*RuleDef `yaml:"on_interact_self"`
	OnInteract           []*RuleDef `yaml:"on_interact"`
	OnZoneActivity       []*RuleDef `yaml:"on_zone_activity"`
	OnBroadcast          []*RuleDef `yaml:"on_broadcast"`
	OnGlobalUpdated      []*RuleDef `yaml:"on_global_updated"`
	OnInit               []*RuleDef `yaml:"on_init"`
	OnStateEnter         []*RuleDef `yaml:"on_state_enter"`
	OnCombatComplete     []*RuleDef `yaml:"on_combat_complete"`
	OnTimerComplete      []*RuleDef `yaml:"on_timer_complete"`
	OnMotionCompleteSelf []*RuleDef `yaml:"on_motion_complete_self"`
	OnMotionComplete     []*RuleDef `yaml:"on_motion_complete"`
}

type RuleDef struct {
	When     *ConditionNode     `yaml:"when"`
	Filter   map[string]any     `yaml:"filter"`
	Steps    []*StepNode        `yaml:"steps"`
	SetState map[string]any     `yaml:"set_state"`
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
