package resources

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/gopxl/pixel/v2"
	"github.com/lafriks/go-tiled"
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
		Layers:             map[MapLayerName]*Layer{},
		Entities:           map[string]*Entity{},
		SceneClearColor:    getColor(tiledMap, "scene_clear_color"),
		LightingClearColor: getColor(tiledMap, "lighting_clear_color"),
	}

	loggedTiles := map[string]bool{}
	tileToSpriteId := func(tileId uint32, tileset *tiled.Tileset) TilesheetSpriteId {
		tileSetX := int(tileId) % tileset.Columns
		tileSetY := int(tileId) / tileset.Columns
		id := TilesheetSpriteId{
			Tilesheet: tileset.Name,
			Column:    tileSetX + 1,
			Row:       tileSetY + 1,
		}
		tiledName := fmt.Sprintf("%s/%d", tileset.Name, tileId)
		if !loggedTiles[tiledName] {
			loggedTiles[tiledName] = true
			log.Debug().Msgf("Tiled %#v // GID: %d", id, tileId)
		}
		return id
	}

	for _, tiledLayer := range tiledMap.Layers {
		if !slices.Contains(MapLayers, MapLayerName(tiledLayer.Name)) {
			log.Warn().Msgf("Skipping unknown layer %s", tiledLayer.Name)
			continue
		}
		gameLayer := &Layer{}

		for tileId, tiledTile := range tiledLayer.Tiles {
			if tiledTile == nil || tiledTile.IsNil() || tiledTile.Tileset == nil {
				continue
			}
			gameTile := &Tile{
				X:        tileId % tiledMap.Width,
				Y:        tiledMap.Height - (tileId / tiledMap.Width),
				SpriteId: tileToSpriteId(tiledTile.ID, tiledTile.Tileset),
			}
			gameLayer.Tiles = append(gameLayer.Tiles, gameTile)
		}
		gameMap.Layers[MapLayerName(tiledLayer.Name)] = gameLayer
	}

	for _, objectGroup := range tiledMap.ObjectGroups {
		if objectGroup.Name == "entities" {
			for _, object := range objectGroup.Objects {
				metadata := map[string]any{}
				for _, property := range object.Properties {
					metadata[property.Name] = property.Value // TODO non string types?
				}

				e := &Entity{
					ID:         int(object.ID),
					X:          int((object.X + float64(tiledMap.TileWidth)/2) / float64(tiledMap.TileWidth)),
					Y:          tiledMap.Height - int((object.Y-float64(tiledMap.TileHeight)/2)/float64(tiledMap.TileHeight)),
					Properties: metadata,
					SpriteGID:  int(object.GID),
				}

				if tiledTile, err := tiledMap.TileGIDToTile(object.GID); err == nil {
					e.SpriteId = tileToSpriteId(tiledTile.ID, tiledTile.Tileset)
				}

				gameMap.Entities[fmt.Sprintf("tiled-%d", object.ID)] = e
			}
		} else if objectGroup.Name == "ambient_light" {
			for _, o := range objectGroup.Objects {
				cHex := o.Properties.GetString("color")
				if cHex == "" {
					log.Warn().Msgf("ambient light %d missing color", o.ID)
					continue
				}
				j, _ := json.MarshalIndent(o, "", "  ")
				fmt.Println(string(j))
				height := int(o.Height/float64(tiledMap.TileWidth)) - 1
				z := AmbientZone{
					Color: colors.FromString(cHex),
					X:     int(o.X / float64(tiledMap.TileWidth)),
					Y:     tiledMap.Height - int(o.Y/float64(tiledMap.TileHeight)) - height,
					W:     int(o.Width/float64(tiledMap.TileWidth)) - 1,
					H:     height,
				}
				glowSizeString := o.Properties.GetString("glow_size")
				if glowSizeString != "" {
					z.GlowSize, err = strconv.ParseFloat(glowSizeString, 64)
					if err != nil {
						log.Fatal().Err(err).Msgf("failed to parse glow size %s", glowSizeString)
					}
				}
				fmt.Printf("%#v\n", z)
				gameMap.AmbientZones = append(gameMap.AmbientZones, z)
			}
		} else {
			log.Warn().Msgf("Skipping unknown object group %s", objectGroup.Name)
			continue
		}
	}

	if _, exists := maps[resourceName]; exists {
		log.Fatal().Msgf("Map already exists with name %s", resourceName)
	}
	maps[resourceName] = gameMap
	return nil
}
