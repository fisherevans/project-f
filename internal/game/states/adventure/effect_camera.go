package adventure

func init() {
	registerStepConverter("override_camera", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		e := NewOverrideCameraEffect()
		if followMap, ok := m["follow"].(map[string]any); ok {
			fc := FollowCamera{ResetPosition: mapBool(followMap, "reset_position")}
			if eid := mapStr(followMap, "entity"); eid != "" {
				fc.EntityId = &eid
			}
			e = e.WithFollow(fc)
		}
		return []Effect{e}
	})
	registerStepConverter("pop_camera", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		maintain := false
		if b, ok := step.Params.(bool); ok {
			maintain = b
		} else {
			m := resolveMap(step.Params, tc)
			maintain = mapBool(m, "maintain_location")
		}
		return []Effect{NewPopCameraOverrideEffect(maintain)}
	})
	registerStepConverter("mutate_camera", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		e := NewMutateFollowCameraEffect()
		if eid := mapStr(m, "follow_entity"); eid != "" {
			e = e.WithFollowEntityId(eid)
		}
		if mapHas(m, "reset_position") {
			e = e.WithResetPosition(mapBool(m, "reset_position"))
		}
		return []Effect{e}
	})
}

type EffectOverrideCamera struct {
	instantEffect
	Follow *FollowCamera `one_of:"type"`
}

func (e *EffectOverrideCamera) Process(source EntityReader, s *State) bool {
	if e.Follow == nil {
		logEffectWarnf(source, e, "follow camera override requires a follow camera")
		return false
	}
	if e.Follow.EntityId == nil {
		logEffectWarnf(source, e, "currently Id is required to set follow camera")
	}
	target, ok := s.entities.GetEntity(*e.Follow.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to follow")
		return false
	}
	location := s.camera.CurrentLocation()
	if e.Follow.ResetPosition {
		location = target.GetPreciseLocation()
	}
	camera := NewSimpleEntityCamera(target.GetId(), location, EntityCameraSpeedMedium, true)
	camera.isSnapped = false
	s.OverrideCamera(camera)
	return true
}

type EffectPopCameraOverride struct {
	instantEffect
	MaintainCurrentLocation bool
}

func (e *EffectPopCameraOverride) Process(source EntityReader, s *State) bool {
	s.PopOverrideCamera(e.MaintainCurrentLocation)
	return true
}

type EffectMutateFollowCamera struct {
	instantEffect
	FollowEntityId *string
	ResetPosition  *bool
}

func (e *EffectMutateFollowCamera) Process(source EntityReader, s *State) bool {
	camera := s.camera
	if override, ok := camera.(*CameraOverride); ok {
		camera = override.newCamera
	}
	if e.FollowEntityId != nil {
		if ts, ok := camera.(TargetingCamera); ok {
			ts.SetTarget(*e.FollowEntityId)
		} else {
			logEffectWarnf(source, e, "camera does not support setting target")
		}
	}
	if e.ResetPosition != nil && *e.ResetPosition {
		camera.ResetPosition(s)
	}
	return true
}

type FollowCamera struct {
	EntityId      *string `one_of:"target"`
	ResetPosition bool
}
