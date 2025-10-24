package adventure

import (
	"slices"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/tiles"
)

func initializeMap(a *State, m *resources.Map) {
	a.sceneClear = m.SceneClearColor
	var minX, maxX, minY, maxY int
	for _, layer := range m.Layers {
		for _, tile := range layer.Tiles {
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
	for _, layerName := range util.Concat(resources.MapLayersUnder, resources.MapLayersOver) {
		thisRenderLayer := renderLayer{
			tiles: make([][]pixelutil.BoundedDrawable, a.mapWidth),
		}
		for x := 0; x < a.mapWidth; x++ {
			thisRenderLayer.tiles[x] = make([]pixelutil.BoundedDrawable, a.mapHeight)
		}
		for _, tile := range m.Layers[layerName].Tiles {
			ref := atlas.GetTilesheetSpriteById(tile.SpriteId)
			thisRenderLayer.tiles[tile.X+dx][tile.Y+dy] = ref
		}
		if slices.Contains(resources.MapLayersOver, layerName) {
			a.overlayRenderLayers = append(a.overlayRenderLayers, thisRenderLayer)
		} else if slices.Contains(resources.MapLayersUnder, layerName) {
			a.baseRenderLayers = append(a.baseRenderLayers, thisRenderLayer)
		}
	}
	for _, collisionTile := range m.Layers[resources.LayerCollision].Tiles {
		location := adjustedLocation(collisionTile.X, collisionTile.Y)
		switch collisionTile.SpriteId {
		case resources.TileCollisionBlock:
			a.AddEntity(newCollisionEntity(location))
		case resources.TileCollisionJumpHorizontal,
			resources.TileCollisionJumpVertical,
			resources.TileCollisionJumpAll:
			a.AddEntity(newDashGapEntity(location))
		}
	}
	for stringEntityId, entity := range m.Entities {
		if entity == nil {
			log.Fatal().Str("entityId", stringEntityId).Msg("entity cannot be nil")
		}
		entityId := EntityId(stringEntityId)
		location := adjustedLocation(entity.X, entity.Y)
		if a.registerParameterizedEntity(entityId, location, entity) {
			log.Info().Msgf("added parameterized entity '%s'", stringEntityId)
			continue
		}
		entityType := entity.GetStringMetadata("type", "")
		if entity.SpriteId != nil {
			switch *entity.SpriteId {
			case tiles.ShadowMob:
				entityType = "shadow"
			}
		}
		switch entityType {
		case "player":
			a.player = &Player{
				BaseEntity: BaseEntity{
					id: entityId,
				},
				Passable: newPassablePreventIngress(true),
				Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
					MoveStateIdle:    anim.AshaIdle(atlas),
					MoveStateWalking: anim.AshaWalk(atlas),
					MoveStateRunning: anim.AshaRun(atlas),
					MoveStateDashing: anim.Dash(atlas),
				},
				Lights: map[MoveState]*Light{
					MoveStateIdle: {
						RenderDetails: LightRenderDetails{
							SizeScale: 1.5,
							ColorMask: pixel.RGB(1, 1, 1),
						},
					},
					MoveStateWalking: {
						RenderDetails: LightRenderDetails{
							SizeScale: 1.5,
							ColorMask: pixel.RGB(1, 1, 1),
						},
					},
					MoveStateRunning: {
						RenderDetails: LightRenderDetails{
							SizeScale: 1.5,
							ColorMask: pixel.RGB(1, 1, 1),
						},
					},
					MoveStateDashing: {
						RenderDetails: LightRenderDetails{
							SizeScale: 1.5,
							ColorMask: pixel.RGB(1, 1, 1),
						},
					},
				},
			}
			
			// Register with movement controller and behavior
			speeds := map[MoveState]float64{
				MoveStateWalking: characterSpeed,
				MoveStateRunning: characterSpeed * 1.75,
				MoveStateDashing: characterSpeed * 1.5,
			}
			movementState := a.movementController.Register(entityId, location, speeds)
			a.player.SetMovementState(movementState)
			
			behavior := NewPlayerBehavior(entityId)
			a.behaviors[entityId] = behavior
			
			a.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedPlayerDefault)
			a.AddEntity(a.player)
			a.worldState.Set("player_id", string(entityId))
		case "stairs":
			ref := TeleportReference("teleport:" + entity.GetStringMetadata("ref", ""))
			if _, exists := a.teleports[ref]; exists {
				log.Fatal().Msgf("stairs reference %s already exists", ref)
				break
			}
			dest := TeleportReference("teleport:" + entity.GetStringMetadata("destination", ""))
			a.teleports[ref] = Teleport{
				Destination:   dest,
				Location:      location,
				ExitDirection: input.DirectionFromString(entity.GetStringMetadata("exit_direction", "")),
			}
			a.zones.SetZoneId(location, string(ref))
		case "shadow":
			a.AddMob(NewShadowMob(entityId, location))
		default:
			log.Warn().Msgf("Unknown entity type: %s / %s", entityId, entityType)
		}
	}
	a.processEffects(a.eventDispatcher.Init(a.worldState))
}
