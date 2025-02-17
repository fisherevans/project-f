package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	resources "fisherevans.com/project/f/internal/resources"
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
			tiles: make([][]*resources.SpriteReference, a.mapWidth),
		}
		for x := 0; x < a.mapWidth; x++ {
			thisRenderLayer.tiles[x] = make([]*resources.SpriteReference, a.mapHeight)
		}
		for _, tile := range m.Layers[layerName].Tiles {
			ref := resources.TilesheetSprites[tile.SpriteId]
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
		a.movementRestrictions[location] = MovementNotAllowed{}
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
						MoveSpeed:       characterSpeed,
					},
					Animations: map[input.Direction]*anim.AnimatedSprite{
						input.Down:  anim.RobotDown(),
						input.Up:    anim.RobotUp(),
						input.Right: anim.RobotRight(),
						input.Left:  anim.RobotLeft(),
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
						MoveSpeed:       2,
					},
					Animations: map[input.Direction]*anim.AnimatedSprite{
						input.Down:  anim.RobotDown(),
						input.Up:    anim.RobotUp(),
						input.Right: anim.RobotRight(),
						input.Left:  anim.RobotLeft(),
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
				npc.MoveableEntity.MoveSpeed = 6
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
