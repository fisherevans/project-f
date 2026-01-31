package resources

import (
	"encoding/json"
	"fmt"

	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
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

var TileLayerGroupNames = []string{
	"under",
	"over",
}

var ControlLayerGroupName = "control"

type ControlToggle struct {
	GlobalKey string `yaml:"global_key"`
	ZoneId    string `yaml:"zone_id"`
}

type ElythiumControls struct {
	Default   bool            `yaml:"default"`
	ToggledBy []ControlToggle `yaml:"toggled_by"`
}

type MapControls struct {
	Elythium           ElythiumControls `yaml:"elythium"`
	SceneClearColor    string           `yaml:"scene_clear_color"`
	LightingClearColor string           `yaml:"lighting_clear_color"`
}

type Map struct {
	TileLayerGroups      []*TileLayerGroup
	Entities             map[string]*Entity
	AmbientLightAreas    []*ColoredArea
	BackgroundColorAreas []*ColoredArea
	Zones                []*Zone
	CollisionTiles       []*Tile
	Controls             *MapControls
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

type TileLayerGroup struct {
	GroupName string       `json:"groupName"`
	Layers    []*TileLayer `json:"layers"`
}

type TileLayer struct {
	LayerName string  `json:"layerName"`
	Tiles     []*Tile `json:"tiles"`
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
	ID         int              `json:"id"`
	X          int              `json:"x"`
	Y          int              `json:"y"`
	Properties *util.Properties `json:"properties"`
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

type ColoredArea struct {
	RectangleEntity
	Color    pixel.RGBA
	GlowSize float64
}

func (z ColoredArea) Moved(dx int, dy int) ColoredArea {
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
