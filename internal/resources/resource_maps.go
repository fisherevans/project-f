package resources

import (
	"encoding/json"
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/util"
)

var (
	maps = map[string]*Map{}
)

func GetMap(name string) *Map {
	m := maps[name]
	if m == nil {
		log.Error().Msgf("missing map: %s", name)
	}
	return m
}

type MapLayerName string

const (
	LayerUnder1 MapLayerName = "under_1"
	LayerUnder2 MapLayerName = "under_2"
	LayerUnder3 MapLayerName = "under_3"

	LayerOver1 MapLayerName = "over_1"
	LayerOver2 MapLayerName = "over_2"
	LayerOver3 MapLayerName = "over_3"

	LayerCollision = "collision"
)

var MapLayersUnder = []MapLayerName{
	LayerUnder1,
	LayerUnder2,
	LayerUnder3,
}

var MapLayersOver = []MapLayerName{
	LayerOver1,
	LayerOver2,
	LayerOver3,
}

var MapLayersUtility = []MapLayerName{
	LayerCollision,
}

var MapLayers = util.Concat(
	MapLayersUnder,
	MapLayersOver,
	MapLayersUtility,
)

type Map struct {
	Layers             map[MapLayerName]*Layer
	Entities           map[string]*Entity
	AmbientLightAreas  []AmbientLightArea
	Zones              []Zone
	SceneClearColor    pixel.RGBA
	LightingClearColor pixel.RGBA
}

func (m *Map) AddEntity(id string, e *Entity) {
	if m.Entities == nil {
		m.Entities = map[string]*Entity{}
	}
	m.Entities[id] = e
}

func (m *Map) RemoveEntity(id string) {
	delete(m.Entities, id)
}

type Layer struct {
	Tiles []*Tile `json:"tiles"`
}

type Tile struct {
	X        int               `json:"x"`
	Y        int               `json:"y"`
	SpriteId TilesheetSpriteId `json:"sprite_id"`
}

func (t Tile) String() string {
	return fmt.Sprintf("Tile{x:%d,y:%d,sprite:[%s]}", t.X, t.Y, t.SpriteId.String())
}

type Entity struct {
	ID         int            `json:"id"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	Properties map[string]any `json:"properties"`
	SpriteId   *TilesheetSpriteId
	Class      string
	SpriteGID  int `json:"sprite_gid,omitempty"`
}

func (e *Entity) Copy() *Entity {
	next := &Entity{}
	encoded, err := json.Marshal(e)
	if err != nil {
		log.Error().Msgf("Error marshalling Entity: %v", err)
		return next
	}
	if err := json.Unmarshal(encoded, next); err != nil {
		log.Error().Msgf("Error unmarshalling Entity: %v", err)
	}
	return next
}

func (e *Entity) GetStringMetadata(key, defaultValue string) string {
	if e.Properties == nil {
		return defaultValue
	}
	value, ok := e.Properties[key]
	if !ok {
		return defaultValue
	}
	str, ok := value.(string)
	if !ok {
		return defaultValue
	}
	return str
}

func (e *Entity) String() string {
	return fmt.Sprintf("Entity{id:%d,x:%d,y:%d,sprite_gid:%x,sprite:[%s],props:%v}", e.ID, e.X, e.Y, e.SpriteGID, e.SpriteId.String(), e.Properties)
}

type RectangleEntity struct {
	X, Y int // bottom-left in TILES
	W, H int
}

func (z RectangleEntity) Moved(dx int, dy int) RectangleEntity {
	z.X += dx
	z.Y += dy
	return z
}

type AmbientLightArea struct {
	RectangleEntity
	Color    pixel.RGBA
	GlowSize float64
}

func (z AmbientLightArea) Moved(dx int, dy int) AmbientLightArea {
	z.RectangleEntity = z.RectangleEntity.Moved(dx, dy)
	return z
}

type Zone struct {
	RectangleEntity
	ZoneId string
}

func (z Zone) Moved(dx int, dy int) Zone {
	z.RectangleEntity = z.RectangleEntity.Moved(dx, dy)
	return z
}
