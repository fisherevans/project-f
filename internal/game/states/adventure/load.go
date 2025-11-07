package adventure

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/tiles"
)

func initializeMap(a *State, m *resources.Map) {
	a.sceneClear = m.SceneClearColor
	a.lightClear = m.LightingClearColor

	// get world bounds first, in order to adjust movement of other object
	var minX, maxX, minY, maxY int
	for _, g := range m.TileLayerGroups {
		for _, l := range g.Layers {
			for _, tile := range l.Tiles {
				if tile.X < minX {
					minX = tile.X
				}
				if tile.X > maxX {
					maxX = tile.X
				}
				if tile.Y < minY {
					minY = tile.Y
				}
				if tile.Y > maxY {
					maxY = tile.Y
				}
			}
		}
	}
	dx, dy := 0-minX, 0-minY
	adjustedLocation := func(x, y int) MapLocation {
		return MapLocation{
			X: x + dx,
			Y: y + dy,
		}
	}
	for _, az := range m.AmbientLightAreas {
		a.ambientLightAreas = append(a.ambientLightAreas, az.Moved(dx, dy))
	}
	for _, z := range m.Zones {
		a.zones.RegisterZone(z.Moved(dx, dy))
	}
	a.mapWidth, a.mapHeight = maxX-minX+1, maxY-minY+1
	for _, group := range m.TileLayerGroups {
		var targetLayers *[]renderLayer
		if group.GroupName == "under" {
			targetLayers = &a.underRenderLayers
		} else if group.GroupName == "over" {
			targetLayers = &a.overRenderLayers
		} else {
			log.Warn().Msgf("unknown tile layer group: %s", group.GroupName)
			continue
		}
		for _, mapLayer := range group.Layers {
			thisRenderLayer := renderLayer{
				tiles: make([][]pixelutil.BoundedDrawable, a.mapWidth),
			}
			for x := 0; x < a.mapWidth; x++ {
				thisRenderLayer.tiles[x] = make([]pixelutil.BoundedDrawable, a.mapHeight)
			}
			for _, tile := range mapLayer.Tiles {
				ref := atlas.GetTilesheetSpriteById(tile.SpriteId)
				thisRenderLayer.tiles[tile.X+dx][tile.Y+dy] = ref
			}
			*targetLayers = append(*targetLayers, thisRenderLayer)
		}
	}
	for _, collisionTile := range m.CollisionTiles {
		location := adjustedLocation(collisionTile.X, collisionTile.Y)
		id := fmt.Sprintf("collision-%d-%d", location.X, location.Y)
		switch collisionTile.SpriteId {
		case resources.TileNPCCollisionBlock:
			entity := a.entities.RegisterEntity(id, location)
			targetNonPlayers := func(id string) bool {
				return id != a.player
			}
			impedance := NewConditionalImpedance(ImpedanceImpassable, targetNonPlayers)
			AttachConditionalBlockIngressPresence(entity, false, impedance, targetNonPlayers)
		case resources.TilePlayerCollisionBlock:
			entity := a.entities.RegisterEntity(id, location)
			targetPlayer := func(id string) bool {
				return id == a.player
			}
			impedance := NewConditionalImpedance(ImpedanceImpassable, targetPlayer)
			AttachConditionalBlockIngressPresence(entity, false, impedance, targetPlayer)
		case resources.TileCollisionBlock:
			entity := a.entities.RegisterEntity(id, location)
			AttachBlockIngressPresence(entity, false, NewImpassableImpedance())
		case resources.TileCollisionJumpHorizontal,
			resources.TileCollisionJumpVertical,
			resources.TileCollisionJumpAll:
			entity := a.entities.RegisterEntity(id, location)
			AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
			entity.SetState(DashGapState{})
		}
	}
	for entityId, mapEntity := range m.Entities {
		if mapEntity == nil {
			log.Fatal().Str("entityId", entityId).Msg("entity cannot be nil")
			continue
		}
		location := adjustedLocation(mapEntity.X, mapEntity.Y)
		newEntityParams := NewEntityParams{
			EntityId:   entityId,
			Class:      mapEntity.Class,
			SpriteId:   mapEntity.SpriteId,
			Location:   location,
			Properties: mapEntity.Properties,
		}
		if a.registerParameterizedEntity(newEntityParams) {
			log.Debug().Msgf("added parameterized entity '%s'", entityId)
			continue
		}
		entityType := mapEntity.Properties.GetString("type", "")
		if mapEntity.SpriteId != nil {
			switch *mapEntity.SpriteId {
			case tiles.ShadowMob:
				entityType = "shadow"
			}
		}
		switch entityType {
		case "stairs":
			ref := TeleportReference("teleport:" + mapEntity.Properties.GetString("ref", ""))
			if _, exists := a.teleports[ref]; exists {
				log.Fatal().Msgf("stairs reference %s already exists", ref)
				break
			}
			dest := TeleportReference("teleport:" + mapEntity.Properties.GetString("destination", ""))
			a.teleports[ref] = Teleport{
				Destination:   dest,
				Location:      location,
				ExitDirection: input.DirectionFromString(mapEntity.Properties.GetString("exit_direction", "")),
			}
			a.zones.SetZoneId(location, string(ref))
		case "shadow":
			a.AddMob(NewShadowMob(entityId, location))
		default:
			log.Warn().Msgf("Unknown entity type: %s / %s", entityId, entityType)
		}
	}
	a.eventDispatcher.Init()
}
