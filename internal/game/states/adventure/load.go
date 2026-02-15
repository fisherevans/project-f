package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/tiles"
)

var DefaultWaypointName = "default"

func initializeMap(a *State, m *resources.Map, waypoint string) {
	a.controls = m.Controls
	sceneClearColor := util.OrDefaultString(m.Controls.SceneClearColor, "#000")
	a.sceneClear = colors.FromString(sceneClearColor)
	lightClearValue := util.OrDefaultString(m.Controls.LightingClearColor, "#fff")
	a.lightClear = colors.FromString(lightClearValue)
	defaultPlayerMode := util.OrDefaultString(m.Controls.DefaultPlayerMode, "human")
	log.Info().
		Str("sceneClear", sceneClearColor).
		Str("lightingClear", lightClearValue).
		Str("defaultPlayerMode", defaultPlayerMode).
		Msg("loading map from config")

	playerLoaded := false
	if waypoint == "" {
		waypoint = DefaultWaypointName
	}

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
	for _, bgz := range m.BackgroundColorAreas {
		a.backgroundColorAreas = append(a.backgroundColorAreas, bgz.Moved(dx, dy))
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
			DashGapMetadataKey.Set(entity, DashGapConfig{})
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
		if templateKey := newEntityParams.Properties.GetString("template", ""); templateKey != "" {
			template, ok := entityPropertyTemplates[templateKey]
			if !ok {
				log.Fatal().Str("template", templateKey).Str("entity", entityId).Msg("template property does not exist")
			}
			newEntityParams.Properties = util.MergeProps(util.NewProps(template), mapEntity.Properties)
		}
		if a.registerParameterizedEntity(newEntityParams) {
			log.Debug().Msgf("added parameterized entity '%s'", entityId)
			continue
		}
		if mapEntity.SpriteId != nil && *mapEntity.SpriteId == tiles.Player {
			thisName := mapEntity.Properties.GetString("waypoint_name", DefaultWaypointName)
			a.teleports[TeleportReference("teleport:"+thisName)] = Teleport{
				Location: location,
			}
			if waypoint != thisName {
				continue
			}
			loadPlayerEntity(newEntityParams, a.entities, defaultPlayerMode)
			playerLoaded = true
			continue
		}
		if mapEntity.Class == "ShadowMob" {
			a.AddMob(NewShadowMob(entityId, location, NewShadowMobParamsFromProperties(newEntityParams.Properties)))
			log.Debug().Msgf("added shadow mob '%s'", entityId)
			continue
		}
		entityType := newEntityParams.Properties.GetString("type", "")
		if mapEntity.SpriteId != nil {
			switch *mapEntity.SpriteId {
			case tiles.ShadowMob:
				entityType = "shadow"
			}
		}
		switch entityType {
		case "stairs":
			ref := TeleportReference("teleport:" + newEntityParams.Properties.GetString("ref", ""))
			if _, exists := a.teleports[ref]; exists {
				log.Fatal().Msgf("stairs reference %s already exists", ref)
				break
			}
			dest := TeleportReference("teleport:" + newEntityParams.Properties.GetString("destination", ""))
			a.teleports[ref] = Teleport{
				Destination:   dest,
				Location:      location,
				ExitDirection: input.DirectionFromString(newEntityParams.Properties.GetString("exit_direction", "")),
			}
			a.zones.SetZoneId(location, string(ref))
		case "shadow":
			a.AddMob(NewShadowMob(entityId, location, NewShadowMobParamsFromProperties(newEntityParams.Properties)))
		default:
			log.Warn().Msgf("Unknown entity type: %s / %s", entityId, entityType)
		}
	}
	if !playerLoaded {
		log.Fatal().Msg("no waypoint found for player to load into, looking for " + waypoint)
	}
	a.eventDispatcher.Init()
}
