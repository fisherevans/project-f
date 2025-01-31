package adventure

import (
	resources2 "fisherevans.com/project/f/internal/resources"
	"fmt"
)

var characterSpriteId = resources2.SpriteId{
	Tilesheet: "ui",
	Column:    1,
	Row:       2,
}

var npcRandomSpriteId = resources2.SpriteId{
	Tilesheet: "ui",
	Column:    2,
	Row:       2,
}

var npcStaticSpriteId = resources2.SpriteId{
	Tilesheet: "ui",
	Column:    3,
	Row:       2,
}

var npcHorizSpriteId = resources2.SpriteId{
	Tilesheet: "ui",
	Column:    4,
	Row:       2,
}

func initializeMap(a *State, m resources2.Map) {
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
	for _, layerName := range []resources2.MapLayerName{resources2.LayerBase, resources2.LayerDecor, resources2.LayerOverlay} {
		thisRenderLayer := renderLayer{
			tiles: make([][]*resources2.SpriteReference, a.mapWidth),
		}
		for x := 0; x < a.mapWidth; x++ {
			thisRenderLayer.tiles[x] = make([]*resources2.SpriteReference, a.mapHeight)
		}
		for _, tile := range m.Layers[layerName].Tiles {
			ref := resources2.Sprites[tile.SpriteId]
			thisRenderLayer.tiles[tile.X+dx][tile.Y+dy] = ref
		}
		if layerName == resources2.LayerOverlay {
			a.overlayRenderLayers = append(a.overlayRenderLayers, thisRenderLayer)
		} else {
			a.baseRenderLayers = append(a.baseRenderLayers, thisRenderLayer)
		}
	}
	a.occupiedLocations = map[MapLocation]EntityReference{}
	for _, collisionTile := range m.Layers[resources2.LayerCollision].Tiles {
		location := adjustedLocation(collisionTile.X, collisionTile.Y)
		a.movementRestrictions[location] = MovementNotAllowed{}
	}
	for id, entityTile := range m.Layers[resources2.LayerEntities].Tiles {
		location := adjustedLocation(entityTile.X, entityTile.Y)
		npcRef := EntityReference{
			EntityId: fmt.Sprintf("npc-%d", id),
		}
		switch entityTile.SpriteId {
		case characterSpriteId:
			a.player = &Player{
				MoveableEntity: MoveableEntity{
					EntityReference: EntityReference{
						EntityId: "player",
					},
					CurrentLocation: location,
					MoveSpeed:       characterSpeed,
				},
			}
			a.currentCameraLocation = a.player.RenderMapLocation()
			a.targetCameraLocation = a.currentCameraLocation
			a.AddEntity(a.player)
		case npcRandomSpriteId:
			a.AddEntity(&NPC{
				MoveableEntity: MoveableEntity{
					EntityReference: npcRef,
					CurrentLocation: location,
					MoveSpeed:       2,
				},
				DoesMove: true,
			})
		case npcStaticSpriteId:
			a.AddEntity(&NPC{
				MoveableEntity: MoveableEntity{
					EntityReference: npcRef,
					CurrentLocation: location,
					MoveSpeed:       2,
				},
			})
		case npcHorizSpriteId:
			a.AddEntity(&NPC{
				MoveableEntity: MoveableEntity{
					EntityReference: npcRef,
					CurrentLocation: location,
					MoveSpeed:       2,
				},
				DoesMove:  true,
				HorizOnly: true,
			})
		}
	}
}
