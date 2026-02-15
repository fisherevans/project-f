package adventure

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
)

// todo camera follow equlibrium causes jitter between pixels in a stable state due to slight variations in time delta
// setting speed to 0 (instant follow), show none of this issue
// unclear how to avoid this problem with a camera with "momentum"...

type TargetingCamera interface {
	Camera
	SetTarget(target string)
}

type Camera interface {
	SetLocation(location pixel.Vec)
	CurrentLocation() pixel.Vec
	TargetLocation(s *State) pixel.Vec
	Update(s *State, timeDelta float64)
	ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Vec)
	ResetPosition(s *State)
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

func (c *cameraLocation) ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Vec) {
	cameraMapX := int(math.Round(c.location.X))
	cameraMapY := int(math.Round(c.location.Y))
	bounds := MapBounds{
		MinX: util.MaxInt(0, cameraMapX-cameraRenderDistanceX),
		MinY: util.MaxInt(0, cameraMapY-cameraRenderDistanceY),
		MaxX: util.MinInt(s.mapWidth-1, cameraMapX+cameraRenderDistanceX),
		MaxY: util.MinInt(s.mapHeight-1, cameraMapY+cameraRenderDistanceY),
	}
	// avoid screen tearing by moving the camera only by full pixels
	moveDelta := c.location.Scaled(-resources.MapTileSize.Float())
	return bounds, moveDelta.Add(targetBounds.Center())
}

type StaticCamera struct {
	cameraLocation
}

func NewStaticCamera(location pixel.Vec) *StaticCamera {
	return &StaticCamera{cameraLocation{location: location}}
}

func (c *StaticCamera) Update(s *State, timeDelta float64) {
}

func (c *StaticCamera) ResetPosition(s *State) {

}

func (c *StaticCamera) TargetLocation(s *State) pixel.Vec {
	return c.location
}

const EntityCameraSpeedNoLag float64 = 0
const EntityCameraSpeedSlow float64 = 3
const EntityCameraSpeedMedium float64 = 5
const EntityCameraSpeedFast float64 = 7

const EntityCameraSpeedPlayerDefault = EntityCameraSpeedMedium

type EntityCamera struct {
	cameraLocation
	ghostLocation pixel.Vec
	target        string
	speed         float64

	lastDeltas     []pixel.Vec
	nextDeltaIndex int
	fullDeltas     bool
	activeMean     pixel.Vec
}

func NewComplexEntityCamera(target string, initialLocation pixel.Vec, speed float64) *EntityCamera {
	return &EntityCamera{
		cameraLocation: cameraLocation{location: initialLocation},
		ghostLocation:  initialLocation,
		target:         target,
		speed:          speed,
		lastDeltas:     make([]pixel.Vec, 16),
	}
}

const (
	entityCameraEpsilon = 1.0 / float64(resources.MapTileSize) / 2.0 // stddev is less than 1 half a pixel
)

func (c *EntityCamera) Update(s *State, timeDelta float64) {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return
	}
	targetLocation := target.GetPreciseLocation()
	if c.speed == EntityCameraSpeedNoLag {
		c.location = targetLocation
		return
	}

	t := 1.0 - math.Exp(-c.speed*timeDelta)
	c.ghostLocation = c.ghostLocation.Add(targetLocation.Sub(c.ghostLocation).Scaled(t))

	// if camera is directly on top, reset smoothing logic
	if c.ghostLocation.Sub(targetLocation).Len() < entityCameraEpsilon && c.location.Sub(targetLocation).Len() < entityCameraEpsilon {
		c.location = targetLocation
		c.nextDeltaIndex = 0
		c.fullDeltas = false
		return
	}

	currentDelta := targetLocation.Sub(c.ghostLocation)
	c.lastDeltas[c.nextDeltaIndex] = currentDelta
	c.nextDeltaIndex++
	if c.nextDeltaIndex >= len(c.lastDeltas) {
		c.nextDeltaIndex = 0
		c.fullDeltas = true
	}

	newLocation := c.ghostLocation
	if c.fullDeltas {
		mean, stdDev := StdDev(c.lastDeltas)
		normalizeX := stdDev.X < entityCameraEpsilon
		normalizeY := stdDev.Y < entityCameraEpsilon
		if normalizeX {
			newLocation.X = targetLocation.X - mean.X
		}
		if normalizeY {
			newLocation.Y = targetLocation.Y - mean.Y
		}
		game.DebugBRf("camera stddev: %.3f / %.3f", stdDev.X, stdDev.Y)
		game.DebugBRf("camera mean: %.3f / %.3f", mean.X, mean.Y)
		game.DebugBRf("camera normalize: %-5v / %-5v", normalizeX, normalizeY)
	}
	c.location = newLocation
}

func (c *EntityCamera) ResetPosition(s *State) {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return
	}
	c.SetLocation(target.GetPreciseLocation())
}

func (c *EntityCamera) SetLocation(location pixel.Vec) {
	c.fullDeltas = false
	c.nextDeltaIndex = 0
	c.ghostLocation = location
	c.location = c.ghostLocation
}

func (c *EntityCamera) SetTarget(target string) {
	c.target = target
}

func (c *EntityCamera) TargetLocation(s *State) pixel.Vec {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return c.location
	}
	return target.GetPreciseLocation()
}

func StdDev(vs []pixel.Vec) (pixel.Vec, pixel.Vec) {
	if len(vs) == 0 {
		return pixel.ZV, pixel.ZV
	}
	var mean, sum pixel.Vec
	for _, n := range vs {
		mean = mean.Add(n)
	}
	mean = mean.Scaled(1.0 / float64(len(vs)))
	for _, v := range vs {
		diff := v.Sub(mean)
		sum = sum.Add(pixel.V(diff.X*diff.X, diff.Y*diff.Y))
	}
	variance := sum.Scaled(1.0 / float64(len(vs)))
	return mean, pixel.V(math.Sqrt(variance.X), math.Sqrt(variance.Y))
}

type SimpleEntityCamera struct {
	cameraLocation
	target       string
	speed        float64
	snapToTarget bool

	isSnapped bool
}

func NewSimpleEntityCamera(target string, initialLocation pixel.Vec, speed float64, snapToTarget bool) *SimpleEntityCamera {
	return &SimpleEntityCamera{
		cameraLocation: cameraLocation{location: initialLocation},
		target:         target,
		speed:          speed,
		snapToTarget:   snapToTarget,
		isSnapped:      true,
	}
}

var cameraSnapDistanceThreshold = 1.0 / float64(resources.MapTileSize)
var cameraUnsnapDistanceThreshold = float64(resources.MapTileSize) / 2.0

func (c *SimpleEntityCamera) Update(s *State, timeDelta float64) {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return
	}
	targetLocation := target.GetPreciseLocation()
	if c.speed == EntityCameraSpeedNoLag {
		c.location = targetLocation
		return
	}

	delta := targetLocation.Sub(c.location)
	if c.snapToTarget && delta.Len() < cameraSnapDistanceThreshold && !c.isSnapped {
		c.isSnapped = true
	} else if c.isSnapped && delta.Len() > cameraUnsnapDistanceThreshold {
		c.isSnapped = false
	}

	if c.isSnapped {
		c.location = targetLocation
	} else {
		t := 1.0 - math.Exp(-c.speed*timeDelta)
		c.location = c.location.Add(delta.Scaled(t))
	}
}

func (c *SimpleEntityCamera) ResetPosition(s *State) {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return
	}
	c.SetLocation(target.GetPreciseLocation())
	c.isSnapped = false
}

func (c *SimpleEntityCamera) SetLocation(location pixel.Vec) {
	c.location = location
	c.isSnapped = false
}

func (c *SimpleEntityCamera) SetTarget(target string) {
	c.target = target
	c.isSnapped = false
}

func (c *SimpleEntityCamera) TargetLocation(s *State) pixel.Vec {
	target, found := s.entities.GetEntity(c.target)
	if !found {
		return c.location
	}
	return target.GetPreciseLocation()
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

func (c *CameraOverride) ComputeRenderDetails(s *State, targetBounds pixel.Rect) (MapBounds, pixel.Vec) {
	return c.newCamera.ComputeRenderDetails(s, targetBounds)
}

func (c *CameraOverride) ResetPosition(s *State) {
	c.newCamera.ResetPosition(s)
	c.replacedCamera.ResetPosition(s)
}

func (c *CameraOverride) TargetLocation(s *State) pixel.Vec {
	return c.newCamera.TargetLocation(s)
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
		replacedFollow.SetLocation(newFollow.CurrentLocation())
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
