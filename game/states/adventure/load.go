package adventure

import (
	"fisherevans.com/project/f/resources"
	"fmt"
)

var characterSpriteId = resources.SpriteId{
	Tilesheet: "ui",
	Column:    1,
	Row:       2,
}

var npcRandomSpriteId = resources.SpriteId{
	Tilesheet: "ui",
	Column:    2,
	Row:       2,
}

var npcStaticSpriteId = resources.SpriteId{
	Tilesheet: "ui",
	Column:    3,
	Row:       2,
}

var npcHorizSpriteId = resources.SpriteId{
	Tilesheet: "ui",
	Column:    4,
	Row:       2,
}

func initializeMap(a *AdventureState, m resources.Map) {
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
			ref := resources.Sprites[tile.SpriteId]
			thisRenderLayer.tiles[tile.X+dx][tile.Y+dy] = ref
		}
		if layerName == resources.LayerOverlay {
			a.overlayRenderLayers = append(a.overlayRenderLayers, thisRenderLayer)
		} else {
			a.baseRenderLayers = append(a.baseRenderLayers, thisRenderLayer)
		}
	}
	a.occupiedLocations = map[MapLocation]EntityReference{}
	for _, collisionTile := range m.Layers[resources.LayerCollision].Tiles {
		location := adjustedLocation(collisionTile.X, collisionTile.Y)
		a.occupiedLocations[location] = EntityReference{
			EntityId: fmt.Sprintf("tile:%d-%d", location.X, location.Y),
		}
	}
	for id, entityTile := range m.Layers[resources.LayerEntities].Tiles {
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
