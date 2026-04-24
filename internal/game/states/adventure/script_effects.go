package adventure

import (
	"fmt"

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

func convertStep(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	switch step.Kind {
	case "dialogue":
		text := resolveString(step.Params, tc)
		return []Effect{NewDialogueEffect(text)}

	case "self_dialogue":
		text := resolveString(step.Params, tc)
		return []Effect{NewSelfDialogueEffect(text)}

	case "chatter":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		duration := mapFloat(m, "duration", 3)
		message := mapStr(m, "message")
		return []Effect{NewChatterEffect(entityId, duration, message)}

	case "timer":
		switch v := step.Params.(type) {
		case float64:
			return []Effect{NewTimerEffect(v)}
		case int:
			return []Effect{NewTimerEffect(float64(v))}
		default:
			m := resolveMap(step.Params, tc)
			duration := mapFloat(m, "duration", 1)
			e := NewTimerEffect(duration)
			if id := mapStr(m, "id"); id != "" {
				e = e.WithTimerId(id)
			}
			return []Effect{e}
		}

	case "play_sound":
		sound := resolveString(step.Params, tc)
		return []Effect{NewPlaySoundEffect(sound)}

	case "set_world_state":
		m := resolveMap(step.Params, tc)
		key := mapStr(m, "key")
		value := m["value"]
		return []Effect{NewSetWorldStateEffect(key, value)}

	case "set_run_state":
		m := resolveMap(step.Params, tc)
		key := mapStr(m, "key")
		value := m["value"]
		return []Effect{NewSetRunStateEffect(key, value)}

	case "delete_entity":
		entityId := resolveString(step.Params, tc)
		return []Effect{NewDeleteEntityEffect(entityId)}

	case "face_direction":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		e := NewEntityFaceDirectionEffect(entityId)
		if dir := mapStr(m, "direction"); dir != "" {
			d := input.DirectionFromString(dir)
			e = e.WithDirection(d)
		}
		return []Effect{e}

	case "face_entity":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		targetEntity := mapStr(m, "target")
		return []Effect{NewEntityFaceDirectionEffect(entityId).WithTargetEntity(targetEntity)}

	case "scripted_motion":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		e := NewStartScriptedMotionEffect(entityId)
		if toEntity := mapStr(m, "to_entity"); toEntity != "" {
			e = e.WithToEntityId(toEntity)
		}
		return []Effect{e}

	case "push_behavior":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		e := NewPushEntityBehaviorEffect(entityId)
		if mapBool(m, "scripted_motion") {
			zone := mapStr(m, "active_player_zone")
			e = e.WithScriptedMotion(EntityBehaviorScriptedMotion{ActivePlayerZone: zone})
		}
		if facingEntity := mapStr(m, "facing_entity"); facingEntity != "" {
			e = e.WithFacingEntityId(facingEntity)
		}
		return []Effect{e}

	case "pop_behavior":
		entityId := resolveString(step.Params, tc)
		return []Effect{NewPopEntityBehaviorEffect(entityId)}

	case "disable_behavior":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		by := mapStr(m, "by")
		return []Effect{NewMutateEntityBehaviorEffect(entityId).WithDisableBy(by)}

	case "enable_behavior":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		by := mapStr(m, "by")
		return []Effect{NewMutateEntityBehaviorEffect(entityId).WithEnableBy(by)}

	case "reset_movement":
		entityId := resolveString(step.Params, tc)
		return []Effect{NewResetMovementEffect(entityId)}

	case "trigger_movement":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		e := NewTriggerMovementEffect(entityId)
		if dir := mapStr(m, "direction"); dir != "" {
			d := input.DirectionFromString(dir)
			e = e.WithDirection(d)
		}
		return []Effect{e}

	case "set_mode":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		mode := mapStr(m, "mode")
		return []Effect{NewMutateModeBasedEntityEffect(entityId).WithMode(mode)}

	case "set_blocking":
		m := resolveMap(step.Params, tc)
		entityId := mapStr(m, "entity")
		blocking := mapBool(m, "is_blocking")
		return []Effect{NewMutateBlockingPresenceEffect(entityId).WithIsBlockingIngress(blocking)}

	case "reset_animation":
		entityId := resolveString(step.Params, tc)
		return []Effect{NewResetModeBasedEntityAnimationEffect(entityId)}

	case "change_player_renderer":
		style := resolveString(step.Params, tc)
		return []Effect{NewChangePlayerRendererEffect(style)}

	case "override_camera":
		m := resolveMap(step.Params, tc)
		e := NewOverrideCameraEffect()
		if followMap, ok := m["follow"].(map[string]any); ok {
			fc := FollowCamera{
				ResetPosition: mapBool(followMap, "reset_position"),
			}
			if eid := mapStr(followMap, "entity"); eid != "" {
				fc.EntityId = &eid
			}
			e = e.WithFollow(fc)
		}
		return []Effect{e}

	case "pop_camera":
		maintain := false
		if b, ok := step.Params.(bool); ok {
			maintain = b
		} else {
			m := resolveMap(step.Params, tc)
			maintain = mapBool(m, "maintain_location")
		}
		return []Effect{NewPopCameraOverrideEffect(maintain)}

	case "mutate_camera":
		m := resolveMap(step.Params, tc)
		e := NewMutateFollowCameraEffect()
		if eid := mapStr(m, "follow_entity"); eid != "" {
			e = e.WithFollowEntityId(eid)
		}
		if mapHas(m, "reset_position") {
			e = e.WithResetPosition(mapBool(m, "reset_position"))
		}
		return []Effect{e}

	case "teleport_player":
		m := resolveMap(step.Params, tc)
		e := NewTeleportPlayerEffect()
		if ref := mapStr(m, "to_reference"); ref != "" {
			e = e.WithToReference(ref)
		}
		if toEntity := mapStr(m, "to_entity"); toEntity != "" {
			e = e.WithToEntityId(toEntity)
		}
		if style := mapStr(m, "transition_style"); style != "" {
			e = e.WithTransitionStyle(style)
		}
		if interstitialSteps, ok := m["interstitial"].([]any); ok {
			var interstitialEffects []Effect
			for _, rawStep := range interstitialSteps {
				stepMap, ok := rawStep.(map[string]any)
				if !ok {
					continue
				}
				for k, v := range stepMap {
					subStep := &StepNode{Kind: k, Params: v}
					interstitialEffects = append(interstitialEffects, convertStep(subStep, tc, sequences)...)
				}
			}
			e = e.WithInterstitialEffects(interstitialEffects)
		}
		return []Effect{e}

	case "load_map":
		m := resolveMap(step.Params, tc)
		mapName := mapStr(m, "map")
		e := NewLoadMapEffect(mapName)
		if wp := mapStr(m, "waypoint"); wp != "" {
			e = e.WithWaypoint(wp)
		}
		return []Effect{e}

	case "fade":
		m := resolveMap(step.Params, tc)
		duration := mapFloat(m, "duration", 0.33)
		transitions := mapInt(m, "transitions", 1)
		e := NewFadeEffect(duration, transitions)
		if from := mapStr(m, "from_color"); from != "" {
			e = e.WithFromColor(from)
		}
		if to := mapStr(m, "to_color"); to != "" {
			e = e.WithToColor(to)
		}
		if mapHas(m, "auto_deactivate") {
			e = e.WithAutoDeactivate(mapBool(m, "auto_deactivate"))
		}
		return []Effect{e}

	case "deactivate_fade":
		fadeId := resolveString(step.Params, tc)
		return []Effect{NewDeactivateFadeEffect(fadeId)}

	case "yield_elythium":
		amount := 0
		switch v := step.Params.(type) {
		case int:
			amount = v
		case float64:
			amount = int(v)
		default:
			m := resolveMap(step.Params, tc)
			amount = mapInt(m, "amount", 0)
		}
		return []Effect{NewYieldElythiumEffect(amount)}

	case "tooltip":
		message := resolveString(step.Params, tc)
		return []Effect{NewPushTooltipEffect(message)}

	case "broadcast":
		m := resolveMap(step.Params, tc)
		id := mapStr(m, "id")
		var data any
		if d, ok := m["data"]; ok {
			data = d
		}
		return []Effect{NewSendBroadcastEffect(id, data)}

	case "wait_for":
		m := resolveMap(step.Params, tc)
		condName := mapStr(m, "condition")
		fn, ok := getScriptConditionFactory(condName)
		if !ok {
			log.Warn().Str("condition", condName).Msg("unknown script condition in wait_for")
			return nil
		}
		check := fn(m)
		return []Effect{NewWaitForConditionEffect(check)}

	case "wait_for_animation":
		entityId := resolveString(step.Params, tc)
		return []Effect{NewWaitForAnimationComplete(entityId)}

	case "action":
		switch v := step.Params.(type) {
		case string:
			return convertAction(v, nil, tc)
		default:
			m := resolveMap(step.Params, tc)
			name := mapStr(m, "name")
			if name == "" {
				name = resolveString(step.Params, tc)
			}
			return convertAction(name, m, tc)
		}

	case "parallel":
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
		children := convertSteps(subSteps, tc, sequences)
		return []Effect{NewParallelPlan(children...)}

	case "focused_sequence":
		return convertFocusedSequence(step.Params, tc, sequences)

	case "ref":
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
		if withMap, ok := m["with"].(map[string]any); ok {
			for k, v := range withMap {
				childTC.Params[k] = fmt.Sprintf("%v", v)
			}
		}
		return convertSteps(seq.Steps, &childTC, sequences)

	default:
		log.Warn().Str("kind", step.Kind).Msg("unknown step kind")
		return nil
	}
}

func convertAction(name string, params map[string]any, tc *TemplateContext) []Effect {
	fn, ok := getScriptAction(name)
	if !ok {
		log.Warn().Str("name", name).Msg("unknown script action")
		return nil
	}
	return []Effect{NewFunctionEffect(func(s *State) {
		var source EntityReader
		if e, ok := s.entities.GetEntity(tc.SelfId); ok {
			source = e
		} else {
			source = &NullEntity{id: tc.SelfId}
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
