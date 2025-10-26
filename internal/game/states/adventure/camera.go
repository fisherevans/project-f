package adventure

import (
	"math"

	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
)

type Camera interface {
	SetLocation(location pixel.Vec)
	CurrentLocation() pixel.Vec
	Update(s *State, timeDelta float64)
	ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Matrix)
}

type cameraLocation struct {
	location pixel.Vec
}

func (c *cameraLocation) CurrentLocation() pixel.Vec {
	return c.location
}

func (c *cameraLocation) SetLocation(location pixel.Vec) {
	c.location = location
}

func (c *cameraLocation) ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Matrix) {
	cameraMapX := int(math.Round(c.location.X))
	cameraMapY := int(math.Round(c.location.Y))
	bounds := MapBounds{
		MinX: util.MaxInt(0, cameraMapX-cameraRenderDistanceX),
		MinY: util.MaxInt(0, cameraMapY-cameraRenderDistanceY),
		MaxX: util.MinInt(s.mapWidth-1, cameraMapX+cameraRenderDistanceX),
		MaxY: util.MinInt(s.mapHeight-1, cameraMapY+cameraRenderDistanceY),
	}
	// avoid screen tearing by moving the camera only by full pixels
	moveDeltaX := -int(math.Round(c.location.X * resources.MapTileSize.Float()))
	moveDeltaY := -int(math.Round(c.location.Y * resources.MapTileSize.Float()))
	moveDelta := gfx.IVec(moveDeltaX, moveDeltaY)
	renderMatrix := pixel.IM.
		Moved(moveDelta).
		Moved(targetBounds.Center())
	return bounds, renderMatrix
}

type StaticCamera struct {
	cameraLocation
}

func NewStaticCamera(location pixel.Vec) *StaticCamera {
	return &StaticCamera{cameraLocation{location: location}}
}

func (c *StaticCamera) Update(s *State, timeDelta float64) {
}

const EntityCameraSpeedNoLag float64 = 0
const EntityCameraSpeedSlow float64 = 3
const EntityCameraSpeedMedium float64 = 5
const EntityCameraSpeedFast float64 = 7

const EntityCameraSpeedPlayerDefault = EntityCameraSpeedMedium

type EntityCamera struct {
	cameraLocation
	target string
	speed  float64
}

func NewFollowCamera(target string, initialLocation pixel.Vec, speed float64) *EntityCamera {
	return &EntityCamera{
		cameraLocation: cameraLocation{location: pixel.V(initialLocation.X, initialLocation.Y)},
		target:         target,
		speed:          speed,
	}
}

func (c *EntityCamera) Update(s *State, timeDelta float64) {
	target, found := s.entities.positions[c.target]
	if !found {
		return
	}
	targetLocation := target.PreciseLocation()
	if c.speed == EntityCameraSpeedNoLag {
		c.location = targetLocation
		return
	}
	delta := targetLocation.Sub(c.location)
	c.location = c.location.Add(delta.Scaled(math.Min(timeDelta*c.speed, 1.0)))
}

type CameraOverride struct {
	replacedCamera Camera
	newCamera      Camera
}

func NewCameraOverride(replacedCamera Camera, newCamera Camera) *CameraOverride {
	return &CameraOverride{
		replacedCamera: replacedCamera,
		newCamera:      newCamera,
	}
}

func (c *CameraOverride) SetLocation(location pixel.Vec) {
	c.newCamera.SetLocation(location)
}

func (c *CameraOverride) CurrentLocation() pixel.Vec {
	return c.newCamera.CurrentLocation()
}

func (c *CameraOverride) Update(s *State, timeDelta float64) {
	c.newCamera.Update(s, timeDelta)
}

func (c *CameraOverride) ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Matrix) {
	return c.newCamera.ComputeRenderDetails(s, targetBounds)
}

func (s *State) OverrideCamera(newCamera Camera) {
	s.camera = NewCameraOverride(s.camera, newCamera)
}

func (s *State) PopOverrideCamera(maintainCurrentLocation bool) {
	override, ok := s.camera.(*CameraOverride)
	if !ok {
		log.Warn().Msg("camera isn't overridden, nothing to do")
		return
	}
	// todo consider making single highly customizable camera
	newFollow, newOk := getCameraToMutate(override.newCamera).(*EntityCamera)
	replacedFollow, replacedOk := getCameraToMutate(override.replacedCamera).(*EntityCamera)
	if newOk && replacedOk && maintainCurrentLocation {
		replacedFollow.location = newFollow.location
	}
	s.camera = override.replacedCamera
}

func getCameraToMutate(c Camera) Camera {
	override, ok := c.(*CameraOverride)
	if ok {
		return override.newCamera
	}
	return c
}
