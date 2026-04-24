package adventure

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ScriptFile struct {
	Handlers  map[string]*HandlerDef  `yaml:"handlers"`
	Sequences map[string]*SequenceDef `yaml:"sequences"`
}

type SequenceDef struct {
	Params []string    `yaml:"params"`
	Steps  []*StepNode `yaml:"steps"`
}

type HandlerDef struct {
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
