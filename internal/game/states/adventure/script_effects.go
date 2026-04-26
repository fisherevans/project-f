package adventure

import (
	"fmt"
	"math/rand/v2"

	"fisherevans.com/project/f/internal/game/input"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func convertSteps(steps []*StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	var effects []Effect
	for _, step := range steps {
		e := convertStep(step, tc, sequences)
		if e != nil {
			effects = append(effects, e...)
		}
	}
	return effects
}

func init() {
	registerStepConverter("action", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		switch v := step.Params.(type) {
		case string:
			return convertAction(v, nil, tc)
		default:
			m := resolveMap(step.Params, tc)
			name := mapStr(m, "name")
			if name == "" {
				name = resolveString(step.Params, tc)
			}
			if name == "" {
				log.Warn().Interface("params", step.Params).Msg("action step missing 'name'")
				return nil
			}
			return convertAction(name, m, tc)
		}
	})
	registerStepConverter("parallel", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		items, ok := step.Params.([]any)
		if !ok {
			log.Warn().Msg("parallel step requires a list of steps")
			return nil
		}
		var subSteps []*StepNode
		for _, item := range items {
			stepMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			for k, v := range stepMap {
				subSteps = append(subSteps, &StepNode{Kind: k, Params: v})
			}
		}
		return []Effect{NewParallelPlan(convertSteps(subSteps, tc, sequences)...)}
	})
	registerStepConverter("focused_sequence", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		return convertFocusedSequence(step.Params, tc, sequences)
	})
	registerStepConverter("ref", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		seqName := mapStr(m, "name")
		seq, ok := sequences[seqName]
		if !ok {
			log.Warn().Str("name", seqName).Msg("unknown sequence ref")
			return nil
		}
		childTC := *tc
		childTC.Params = make(map[string]string)
		for k, v := range tc.Params {
			childTC.Params[k] = v
		}
		if paramMap, ok := m["params"].(map[string]any); ok {
			for k, v := range paramMap {
				childTC.Params[k] = fmt.Sprintf("%v", v)
			}
		}
		return convertSteps(seq.Steps, &childTC, sequences)
	})
}

func convertAction(name string, params map[string]any, tc *TemplateContext) []Effect {
	fn, ok := getScriptAction(name)
	if !ok {
		log.Warn().Str("name", name).Msg("unknown script action")
		return nil
	}
	return []Effect{NewFunctionEffect(func(source EntityReader, s *State) {
		if source == nil {
			if e, ok := s.entities.GetEntity(tc.SelfId); ok {
				source = e
			} else {
				source = &NullEntity{id: tc.SelfId}
			}
		}
		fn(s, source, params)
	})}
}

func convertFocusedSequence(params any, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	m := resolveMap(params, tc)
	focusedEntity := tc.SelfId
	if eid := mapStr(m, "entity"); eid != "" {
		focusedEntity = eid
	}
	target := tc.PlayerId
	if t := mapStr(m, "target"); t != "" {
		target = t
	}
	moveCamera := mapBool(m, "move_camera")
	facePlayer := mapBool(m, "face_player")

	extractSteps := func(key string) []Effect {
		raw, ok := m[key].([]any)
		if !ok {
			return nil
		}
		var effects []Effect
		for _, rawStep := range raw {
			stepMap, ok := rawStep.(map[string]any)
			if !ok {
				continue
			}
			for k, v := range stepMap {
				subStep := &StepNode{Kind: k, Params: v}
				effects = append(effects, convertStep(subStep, tc, sequences)...)
			}
		}
		return effects
	}

	builder := NewFocusedSequenceBuilder(focusedEntity, target)
	if moveCamera {
		builder = builder.WithMoveCamera(true)
	}
	if facePlayer {
		builder = builder.WithFacePlayer(true)
	}
	builder = builder.WithPreEffects(extractSteps("pre_effects")...)
	builder = builder.WithMiddleEffects(extractSteps("effects")...)
	builder = builder.WithPostEffects(extractSteps("post_effects")...)
	return builder.BuildEffects()
}

// NullEntity is a minimal EntityReader for use when no real entity is available
type NullEntity struct {
	id string
}

func (e *NullEntity) GetId() string                               { return e.id }
func (e *NullEntity) GetLocation() MapLocation                    { return MapLocation{} }
func (e *NullEntity) GetPreciseLocation() pixel.Vec               { return pixel.ZV }
func (e *NullEntity) GetFacingDirection() input.Direction          { return input.NotPressed }
func (e *NullEntity) IsMoving() bool                              { return false }
func (e *NullEntity) GetMovementState() MoveState                 { return MoveStateIdle }
func (e *NullEntity) GetMovementSpeed(_ MoveState) float64        { return 0 }
func (e *NullEntity) GetMetadata() *EntityMetadata                { return nil }
func (e *NullEntity) GetBehavior() (EntityBehavior, bool)         { return nil, false }
func (e *NullEntity) HasPushedBehavior() bool                     { return false }
func (e *NullEntity) IsBehaviorEnabled() bool                     { return false }
func (e *NullEntity) IsSoundEnabled() bool                        { return false }
func (e *NullEntity) GetPresence() (EntityPresence, bool)         { return nil, false }
func (e *NullEntity) GetRenderer() (EntityRenderer, bool)         { return nil, false }

// helpers

func resolveString(v any, tc *TemplateContext) string {
	s, _ := v.(string)
	return tc.Resolve(s)
}

func resolveMap(v any, tc *TemplateContext) map[string]any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	resolved := tc.ResolveAny(m)
	rm, _ := resolved.(map[string]any)
	return rm
}

func mapStr(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func mapFloat(m map[string]any, key string, def float64) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return def
	}
}

func mapInt(m map[string]any, key string, def int) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return def
	}
}

func mapBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func mapHas(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}


func selectFromList(list []string, mode string, counter int) string {
	if len(list) == 0 {
		return ""
	}
	switch mode {
	case "random":
		return list[rand.IntN(len(list))]
	case "capped":
		idx := min(counter, len(list)-1)
		return list[idx]
	default: // "cycling"
		return list[counter%len(list)]
	}
}

func extractListSelection(m map[string]any, tc *TemplateContext) (message string, effects []Effect, err error) {
	listName := resolveString(m["list"], tc)
	list, ok := getDataList(listName)
	if !ok {
		return "", nil, fmt.Errorf("unknown data list %q", listName)
	}
	mode := "cycling"
	if modeStr, ok := m["select"].(string); ok {
		mode = modeStr
	}
	counterKey := ""
	if ck, ok := m["counter_key"].(string); ok {
		counterKey = tc.Resolve(ck)
	}
	increment := true
	if inc, ok := m["increment"].(bool); ok {
		increment = inc
	}

	counter := 0
	if counterKey != "" && tc.Globals != nil {
		counter = tc.Globals.Get(counterKey).AsInt(0)
	}

	message = selectFromList(list, mode, counter)

	if increment && counterKey != "" {
		effects = append(effects, NewSetGlobalEffect(counterKey, counter+1))
	}
	return message, effects, nil
}

func convertPickDialogue(step *StepNode, tc *TemplateContext) []Effect {
	m := resolveMap(step.Params, tc)
	if m == nil {
		log.Warn().Msg("pick_dialogue: expected map")
		return nil
	}
	message, counterEffects, err := extractListSelection(m, tc)
	if err != nil {
		log.Warn().Err(err).Msg("pick_dialogue")
		return nil
	}
	var effects []Effect
	effects = append(effects, NewDialogueEffect(message))
	effects = append(effects, counterEffects...)
	return effects
}

func convertPickSelfDialogue(step *StepNode, tc *TemplateContext) []Effect {
	m := resolveMap(step.Params, tc)
	if m == nil {
		log.Warn().Msg("pick_self_dialogue: expected map")
		return nil
	}
	message, counterEffects, err := extractListSelection(m, tc)
	if err != nil {
		log.Warn().Err(err).Msg("pick_self_dialogue")
		return nil
	}
	var effects []Effect
	effects = append(effects, NewSelfDialogueEffect(message))
	effects = append(effects, counterEffects...)
	return effects
}

func convertPickChatter(step *StepNode, tc *TemplateContext) []Effect {
	m := resolveMap(step.Params, tc)
	if m == nil {
		log.Warn().Msg("pick_chatter: expected map")
		return nil
	}
	message, counterEffects, err := extractListSelection(m, tc)
	if err != nil {
		log.Warn().Err(err).Msg("pick_chatter")
		return nil
	}
	entity := mapStr(m, "entity")
	duration := mapFloat(m, "duration", 3)
	var effects []Effect
	effects = append(effects, NewChatterEffect(entity, duration, message))
	effects = append(effects, counterEffects...)
	return effects
}

func convertConfigureModeEntity(step *StepNode, tc *TemplateContext) []Effect {
	m := resolveMap(step.Params, tc)
	if m == nil {
		log.Warn().Msg("configure_mode_entity: expected map")
		return nil
	}
	entityId := mapStr(m, "entity")
	modesRaw, ok := m["modes"].(map[string]any)
	if !ok {
		log.Warn().Msg("configure_mode_entity: modes must be a map")
		return nil
	}

	allAnimations := map[string][]AnimationReference{}
	allLights := map[string][]LightConfig{}
	soundConfig := NewModeBaseSoundProviderConfig()
	hasSounds := false

	for modeName, modeVal := range modesRaw {
		modeMap, ok := modeVal.(map[string]any)
		if !ok {
			continue
		}

		if animsRaw, ok := modeMap["animations"].([]any); ok {
			var refs []AnimationReference
			for _, animRaw := range animsRaw {
				animMap, ok := animRaw.(map[string]any)
				if !ok {
					continue
				}
				ref := AnimationReference{
					Name: resolveString(animMap["name"], tc),
				}
				if cm, ok := animMap["color_mask"].(string); ok {
					resolved := tc.Resolve(cm)
					ref.ColorMask = &resolved
				}
				if offset, ok := animMap["offset"].(map[string]any); ok {
					ref.Offset = &OffsetConfig{
						X: toFloat64(offset["x"]),
						Y: toFloat64(offset["y"]),
					}
				}
				refs = append(refs, ref)
			}
			if len(refs) > 0 {
				allAnimations[modeName] = refs
			}
		}

		if lightsRaw, ok := modeMap["lights"].([]any); ok {
			var configs []LightConfig
			for _, lightRaw := range lightsRaw {
				lightMap, ok := lightRaw.(map[string]any)
				if !ok {
					continue
				}
				lc := LightConfig{
					Color: resolveString(lightMap["color"], tc),
					Size:  toFloat64(lightMap["size"]),
				}
				if mod, ok := lightMap["modifier"].(string); ok {
					resolved := tc.Resolve(mod)
					lc.Modifier = &resolved
				}
				configs = append(configs, lc)
			}
			if len(configs) > 0 {
				allLights[modeName] = configs
			}
		}

		if soundsRaw, ok := modeMap["sounds_on_enter"].([]any); ok {
			for _, soundRaw := range soundsRaw {
				soundMap, ok := soundRaw.(map[string]any)
				if !ok {
					continue
				}
				se := parseSoundEffect(soundMap, tc)
				soundConfig = soundConfig.WithSoundOnEnter(modeName, se)
				hasSounds = true
			}
		}
	}

	var effects []Effect

	if len(allAnimations) > 0 || len(allLights) > 0 {
		e := NewMutateModeBasedEntityEffect(entityId)
		if len(allAnimations) > 0 {
			e = e.WithAnimations(allAnimations)
		}
		if len(allLights) > 0 {
			e = e.WithLights(allLights)
		}
		effects = append(effects, e)
	}

	if hasSounds {
		effects = append(effects, NewAddSoundProviderEffect(entityId).WithModeBase(soundConfig))
	}

	return effects
}

func parseSoundEffect(m map[string]any, tc *TemplateContext) SoundEffect {
	se := SoundEffect{
		Name:   resolveString(m["name"], tc),
		Volume: 1.0,
	}
	if loop, ok := m["loop"].(bool); ok {
		se.Loop = loop
	}
	if vol, ok := m["volume"]; ok {
		se.Volume = toFloat64(vol)
	}
	if fadeIn, ok := m["fade_in"]; ok {
		se.FadeInSeconds = toFloat64(fadeIn)
	}
	if falloffStr, ok := m["falloff"].(string); ok {
		se.Falloff = parseFalloff(falloffStr)
	}
	return se
}

func parseFalloff(name string) Falloff {
	switch name {
	case "footstep":
		return FootstepFalloff
	case "explosion":
		return ExplosionFalloff
	case "chatter":
		return ChatterFalloff
	case "ambient":
		return AmbientFalloff
	case "standard":
		return StandardFalloff
	default:
		return StandardFalloff
	}
}
