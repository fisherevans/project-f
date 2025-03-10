package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	resources "fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
	"math/rand"
)

var npcRandomSpriteId = resources.TilesheetSpriteId{
	Tilesheet: "ui",
	Column:    2,
	Row:       2,
}

func initializeMap(a *State, m *resources.Map) {
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
	a.mapWidth, a.mapHeight = maxX-minX+1, maxY-minY+1
	for _, layerName := range []resources.MapLayerName{resources.LayerBase, resources.LayerDecor, resources.LayerOverlay} {
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
		if layerName == resources.LayerOverlay {
			a.overlayRenderLayers = append(a.overlayRenderLayers, thisRenderLayer)
		} else {
			a.baseRenderLayers = append(a.baseRenderLayers, thisRenderLayer)
		}
	}
	a.occupiedLocations = map[MapLocation]EntityId{}
	for _, collisionTile := range m.Layers[resources.LayerCollision].Tiles {
		location := adjustedLocation(collisionTile.X, collisionTile.Y)
		switch collisionTile.SpriteId {
		case resources.TileCollisionBlock:
			a.movementRestrictions[location] = MovementNotAllowed{}
		case resources.TileCollisionJumpHorizontal:
			a.movementRestrictions[location] = MovementJumpTile{}
		case resources.TileCollisionJumpVertical:
			a.movementRestrictions[location] = MovementJumpTile{}
		case resources.TileCollisionJumpAll:
			a.movementRestrictions[location] = MovementJumpTile{}
		}
	}
	for stringEntityId, entity := range m.Entities {
		entityId := EntityId(stringEntityId)
		location := adjustedLocation(entity.X, entity.Y)
		switch entity.Type {
		case "player":
			a.player = &Player{
				AnimatedMoveableEntity: AnimatedMoveableEntity{
					MoveableEntity: MoveableEntity{
						EntityId:        entityId,
						CurrentLocation: location,
						MoveSpeeds: map[MoveState]float64{
							MoveStateWalking: characterSpeed,
							MoveStateRunning: characterSpeed * 1.75,
							MoveStateDashing: characterSpeed * 1.5,
						},
					},
					Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
						MoveStateIdle:    anim.AshaIdle(atlas),
						MoveStateWalking: anim.AshaWalk(atlas),
						MoveStateRunning: anim.AshaRun(atlas),
						MoveStateDashing: anim.Dash(atlas),
					},
				},
			}
			a.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedMedium)
			a.AddEntity(a.player)
		case "blob":
			npc := &NPC{
				AnimatedMoveableEntity: AnimatedMoveableEntity{
					MoveableEntity: MoveableEntity{
						EntityId:        entityId,
						CurrentLocation: location,
						MoveSpeeds: map[MoveState]float64{
							MoveStateWalking: 2,
						},
					},
					Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
						MoveStateIdle:    anim.AshaIdle(atlas),
						MoveStateWalking: anim.AshaWalk(atlas),
						MoveStateRunning: anim.AshaRun(atlas),
					},
					ColorMask: pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()),
				},
				DoesMove:        true,
				IdleChance:      0.05,
				MaxIdleDuration: 6,
			}
			switch entity.GetStringMetadata("movement", "") {
			case "static":
				npc.DoesMove = false
			case "horiz":
				npc.HorizOnly = true
			}
			switch entity.GetStringMetadata("speed", "") {
			case "fast":
				npc.MoveableEntity.MoveSpeeds[MoveStateWalking] = 4
			}
			a.AddEntity(npc)
		case "chest":
			a.AddEntity(&EntityChest{
				InnateEntity: InnateEntity{
					EntityId:    entityId,
					MapLocation: location,
				},
				hasItem: true,
				item:    entity.GetStringMetadata("item", "a sock"),
			})
		case "interest":
			a.AddEntity(&EntityInterest{
				InnateEntity: InnateEntity{
					EntityId:    entityId,
					MapLocation: location,
				},
				topic: entity.GetStringMetadata("topic", ""),
			})
		case "combat":
			a.AddEntity(&EntityCombat{
				InnateEntity: InnateEntity{
					EntityId:    entityId,
					MapLocation: location,
				},
			})
		}
	}
}
