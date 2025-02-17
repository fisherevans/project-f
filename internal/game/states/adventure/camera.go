package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"math"
)

type Camera interface {
	SetLocation(location pixel.Vec)
	CurrentLocation() pixel.Vec
	Update(ctx *game.Context, s *State, timeDelta float64)
	ComputeRenderDetails(ctx *game.Context, s *State, targetBounds pixel.Rect) (MapBounds, pixel.Matrix)
}

type cameraLocation struct {
	location pixel.Vec
}

func (c cameraLocation) CurrentLocation() pixel.Vec {
	return c.location
}

func (c cameraLocation) SetLocation(location pixel.Vec) {
	c.location = location
}

func (c cameraLocation) ComputeRenderDetails(ctx *game.Context, s *State, targetBounds pixel.Rect) (MapBounds, pixel.Matrix) {
	cameraMapX := int(math.Round(c.location.X))
	cameraMapY := int(math.Round(c.location.Y))
	bounds := MapBounds{
		MinX: util.MaxInt(0, cameraMapX-cameraRenderDistanceX),
		MinY: util.MaxInt(0, cameraMapY-cameraRenderDistanceY),
		MaxX: util.MinInt(s.mapWidth-1, cameraMapX+cameraRenderDistanceX),
		MaxY: util.MinInt(s.mapHeight-1, cameraMapY+cameraRenderDistanceY),
	}
	renderMatrix := pixel.IM.
		Moved(c.location.Scaled(-1 * resources.MapTileSize.Float())).
		Moved(targetBounds.Center())
	return bounds, renderMatrix
}

type StaticCamera struct {
	cameraLocation
}

func NewStaticCamera(location pixel.Vec) *StaticCamera {
	return &StaticCamera{cameraLocation{location: location}}
}

func (c *StaticCamera) Update(ctx *game.Context, s *State, timeDelta float64) {
}

const EntityCameraSpeedNoLag float64 = 0
const EntityCameraSpeedSlow float64 = 3
const EntityCameraSpeedMedium float64 = 5
const EntityCameraSpeedFast float64 = 7

type EntityCamera struct {
	cameraLocation
	target EntityId
	speed  float64
}

func NewFollowCamera(target EntityId, initialLocation pixel.Vec, speed float64) *EntityCamera {
	return &EntityCamera{
		cameraLocation: cameraLocation{location: pixel.V(initialLocation.X, initialLocation.Y)},
		target:         target,
		speed:          speed,
	}
}

func (c *EntityCamera) Update(ctx *game.Context, s *State, timeDelta float64) {
	target, found := s.entities[c.target]
	if !found {
		return
	}
	targetLocation := target.RenderMapLocation()
	if c.speed == EntityCameraSpeedNoLag {
		c.location = targetLocation
		return
	}
	delta := targetLocation.Sub(c.location)
	c.location = c.location.Add(delta.Scaled(math.Min(timeDelta*c.speed, 1.0)))
}
