package adventure

type EffectOverrideCamera struct {
	instantEffect
	Follow *FollowCamera `one_of:"type"`
}

func (e *EffectOverrideCamera) Process(source EntityContext, s *State) bool {
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
	camera := NewFollowCamera(target.GetId(), location, EntityCameraSpeedMedium)
	s.OverrideCamera(camera)
	logEffectInfof(source, e, "camera overriden")
	return true
}

type EffectPopCameraOverride struct {
	instantEffect
	MaintainCurrentLocation bool
}

func (e *EffectPopCameraOverride) Process(source EntityContext, s *State) bool {
	s.PopOverrideCamera(e.MaintainCurrentLocation)
	logEffectInfof(source, e, "camera popped")
	return true
}

type EffectMutateFollowCamera struct {
	instantEffect
	FollowEntityId *string
	ResetPosition  *bool
}

func (e *EffectMutateFollowCamera) Process(source EntityContext, s *State) bool {
	camera := s.camera
	if override, ok := camera.(*CameraOverride); ok {
		camera = override.newCamera
	}
	followCamera, ok := camera.(*EntityCamera)
	if !ok {
		logEffectWarnf(source, e, "camera is not a follow camera")
		return false
	}
	if e.FollowEntityId != nil {
		followCamera.target = *e.FollowEntityId
	}
	if e.ResetPosition != nil && *e.ResetPosition {
		followCamera.ResetPosition(s)
	}
	logEffectInfof(source, e, "follow camera mutated")
	return true
}

type FollowCamera struct {
	EntityId      *string `one_of:"target"`
	ResetPosition bool
}
