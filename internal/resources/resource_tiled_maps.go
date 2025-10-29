package resources

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"github.com/lafriks/go-tiled"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/internal/util/colors"
)

func getColor(tiledMap *tiled.Map, key string) pixel.RGBA {
	if tiledMap == nil || tiledMap.Properties == nil {
		return colors.Black.RGBA
	}
	return colors.FromString(tiledMap.Properties.GetString(key))
}

func loadTiledMap(path string, resourceName string, _ []byte) error {
	var tiledMap *tiled.Map
	var err error
	tiledMap, err = tiled.LoadFile(path, tiled.WithFileSystem(assets.FS))
	if err != nil {
		fmt.Printf("error parsing tiledMap: %s", err.Error())
		os.Exit(2)
	}

	log := log.With().Str("path", path).Logger()

	gameMap := &Map{
		Entities:           map[string]*Entity{},
		SceneClearColor:    getColor(tiledMap, "scene_clear_color"),
		LightingClearColor: getColor(tiledMap, "lighting_clear_color"),
	}

	for _, tiledGroup := range tiledMap.Groups {
		if slices.Contains(TileLayerGroupNames, tiledGroup.Name) {
			layerGroup := parseTileLayers(tiledMap, tiledGroup, log)
			gameMap.TileLayerGroups = append(gameMap.TileLayerGroups, layerGroup)
		} else if tiledGroup.Name == ControlLayerGroupName {
			populateControlObjects(tiledMap, tiledGroup, log, gameMap)
		} else {
			log.Error().Msgf("Unknown tiledGroup: %s", tiledGroup.Name)
		}
	}

	for _, layer := range tiledMap.Layers {
		log.Error().Msgf("Unknown ungrouped tiledLayer: %s", layer.Name)
	}

	if _, exists := maps[resourceName]; exists {
		log.Fatal().Msgf("Map already exists with name %s", resourceName)
	}
	maps[resourceName] = gameMap
	return nil
}

func populateControlObjects(tiledMap *tiled.Map, tiledGroup *tiled.Group, log zerolog.Logger, gameMap *Map) {
	for _, objectGroup := range tiledGroup.ObjectGroups {
		if objectGroup.Name == "entities" {
			for _, object := range objectGroup.Objects {
				metadata := map[string]any{}
				for _, property := range object.Properties {
					metadata[property.Name] = property.Value // TODO non string types?
				}
				class := object.Class
				if class == "" {
					class = object.Type
				}
				// For tile entities, center them by adding half their size
				offsetX := object.Width / 2
				offsetY := object.Height / 2

				e := &Entity{
					ID:         int(object.ID),
					X:          int((object.X + offsetX) / float64(tiledMap.TileWidth)),
					Y:          tiledMap.Height - int((object.Y-offsetY)/float64(tiledMap.TileHeight)),
					Properties: util.NewProps(metadata),
					SpriteGID:  int(object.GID),
					Class:      class,
				}

				entityId := object.Properties.GetString("entity_id")
				if entityId == "" {
					entityId = fmt.Sprintf("tiled-%d", object.ID)
				}
				if tiledTile, err := tiledMap.TileGIDToTile(object.GID); err == nil {
					e.SpriteId = tileToSpriteId(tiledTile.ID, tiledTile.Tileset)
				}

				gameMap.Entities[entityId] = e
			}
		} else if objectGroup.Name == "ambient_light" {
			for _, o := range objectGroup.Objects {
				cHex := o.Properties.GetString("color")
				if cHex == "" {
					log.Warn().Msgf("ambient light %d missing color", o.ID)
					continue
				}
				area := &AmbientLightArea{
					Color:           colors.FromString(cHex),
					RectangleEntity: newRectangleEntity(tiledMap, o),
				}
				glowSizeString := o.Properties.GetString("glow_size")
				if glowSizeString != "" {
					var err error
					area.GlowSize, err = strconv.ParseFloat(glowSizeString, 64)
					if err != nil {
						log.Fatal().Err(err).Msgf("failed to parse glow size %s", glowSizeString)
					}
				}
				gameMap.AmbientLightAreas = append(gameMap.AmbientLightAreas, area)
			}
		} else if objectGroup.Name == "zones" {
			for _, o := range objectGroup.Objects {
				z := &Zone{
					RectangleEntity: newRectangleEntity(tiledMap, o),
					ZoneId:          o.Properties.GetString("zone_id"),
				}
				if z.ZoneId == "" {
					log.Fatal().Msgf("zone missing zone_id: %#v", o)
				}
				gameMap.Zones = append(gameMap.Zones, z)
			}
		} else {
			log.Warn().Msgf("Skipping unknown control object group %s", objectGroup.Name)
			continue
		}
	}
	controlTileLayers := parseTileLayers(tiledMap, tiledGroup, log)
	for _, layer := range controlTileLayers.Layers {
		if layer.LayerName != "collision" {
			log.Error().Msgf("Unknown ungrouped tiledLayer: %s", layer.LayerName)
			continue
		}
		gameMap.CollisionTiles = layer.Tiles
	}
}

func parseTileLayers(tiledMap *tiled.Map, tiledGroup *tiled.Group, log zerolog.Logger) *TileLayerGroup {
	group := &TileLayerGroup{
		GroupName: tiledGroup.Name,
	}
	for _, tiledLayer := range tiledGroup.Layers {
		gameLayer := &TileLayer{
			LayerName: tiledLayer.Name,
		}

		for tileId, tiledTile := range tiledLayer.Tiles {
			if tiledTile == nil || tiledTile.IsNil() || tiledTile.Tileset == nil {
				continue
			}
			gameTile := &Tile{
				X:        tileId % tiledMap.Width,
				Y:        tiledMap.Height - (tileId / tiledMap.Width),
				SpriteId: *tileToSpriteId(tiledTile.ID, tiledTile.Tileset),
			}
			gameLayer.Tiles = append(gameLayer.Tiles, gameTile)
		}
		group.Layers = append(group.Layers, gameLayer)
	}
	return group
}

func newRectangleEntity(tiledMap *tiled.Map, o *tiled.Object) RectangleEntity {
	height := int(o.Height/float64(tiledMap.TileWidth)) - 1
	return RectangleEntity{
		X: int(o.X / float64(tiledMap.TileWidth)),
		Y: tiledMap.Height - int(o.Y/float64(tiledMap.TileHeight)) - height,
		W: int(o.Width/float64(tiledMap.TileWidth)) - 1,
		H: height,
	}
}

func tileToSpriteId(tileId uint32, tileset *tiled.Tileset) *TilesheetSpriteId {
	if tileset == nil {
		return nil
	}
	tileSetX := int(tileId) % tileset.Columns
	tileSetY := int(tileId) / tileset.Columns
	id := TilesheetSpriteId{
		Tilesheet: tileset.Name,
		Column:    tileSetX + 1,
		Row:       tileSetY + 1,
	}
	return &id
}
