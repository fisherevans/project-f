package adventure

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
