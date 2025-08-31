package adventure

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/tiles"
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
		entityType := entity.GetStringMetadata("type", "")
		switch entity.SpriteId {
		case tiles.Torch, tiles.TorchRight, tiles.TorchLeft:
			entityType = "torch"
		case tiles.RedCoin:
			entityType = "red_coin"
		}
		switch entityType {
		case "player":
			normalPlayerLight := &Light{
				RenderDetails: LightRenderDetails{
					SizeScale: 1.5,
					ColorMask: colors.HexColor("#888"),
				},
			}
			dashPlayerLight := &Light{
				RenderDetails: LightRenderDetails{
					SizeScale: 1.5,
					ColorMask: colors.HexColor("#88f"),
				},
			}
			a.player = &Player{
				AnimatedMoveableEntity: AnimatedMoveableEntity{
					MoveableEntity: MoveableEntity{
						BaseEntity: BaseEntity{
							Id: entityId,
						},
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
					Lights: map[MoveState]*Light{
						MoveStateIdle:    normalPlayerLight,
						MoveStateWalking: normalPlayerLight,
						MoveStateRunning: normalPlayerLight,
						MoveStateDashing: dashPlayerLight,
					},
				},
			}
			a.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedPlayerDefault)
			a.AddEntity(a.player)
		case "blob":
			npc := &NPC{
				AnimatedMoveableEntity: AnimatedMoveableEntity{
					MoveableEntity: MoveableEntity{
						BaseEntity: BaseEntity{
							Id: entityId,
						},
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
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				hasItem: true,
				item:    entity.GetStringMetadata("item", "a sock"),
			})
		case "interest":
			a.AddEntity(&EntityInterest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				topic: entity.GetStringMetadata("topic", ""),
			})
		case "combat":
			a.AddEntity(&EntityCombatTest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
			})
		case "menu_test":
			a.AddEntity(&EntityMenuTest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
			})
		case "stairs":
			ref := TeleportReference(entity.GetStringMetadata("ref", ""))
			if _, exists := a.teleports[ref]; exists {
				log.Fatal().Msgf("stairs reference %s already exists", ref)
				break
			}
			dest := TeleportReference(entity.GetStringMetadata("destination", ""))
			a.teleports[ref] = Teleport{
				Destination:   dest,
				Location:      location,
				ExitDirection: input.DirectionFromString(entity.GetStringMetadata("exit_direction", "")),
			}
			if _, exists := a.movementRestrictions[location]; exists {
				log.Fatal().Msgf("stairs location at %s is already occupied", location)
				break
			}
			a.movementRestrictions[location] = TeleportTile{
				Reference: ref,
			}
			log.Info().Msgf("stairs added: %s (%s) -> %s", ref, location, dest)
		case "torch":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:              entityId,
						RenderZPriority: 10,
						Passable:        true,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 2,
						ColorMask: colors.HexColor("#db9a3d"),
					},
					Modifiers: []LightModifier{
						&LightModifierFlicker{
							Jitter: &LightModifierJitterUpdate{
								FrequencySeconds:   0.15,
								FrequencyVariation: 0.05,
							},
							SizeVariation:       0.1,
							BrightnessVariation: 0.1,
						},
					},
				},
			}
			switch entity.SpriteId {
			case tiles.Torch:
				t.Animation = anim.Torch(atlas)
			case tiles.TorchRight:
				t.Animation = anim.TorchRight(atlas)
				t.Light.RenderDetails.PositionDelta = pixel.V(float64(resources.MapTileSize/2), 0)
			case tiles.TorchLeft:
				t.Animation = anim.TorchLeft(atlas)
				t.Light.RenderDetails.PositionDelta = pixel.V(-float64(resources.MapTileSize/2), 0)
			}
			a.AddEntity(t)
			log.Info().Msgf("torch added: %s", location)
		case "red_coin":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:       entityId,
						Passable: true,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 0.5,
						ColorMask: colors.HexColor("#f00"),
					},
					Modifiers: []LightModifier{
						&LightModifierPulse{
							PeriodSeconds:       2,
							SizeIntensity:       0.1,
							BrightnessIntensity: 0.4,
						},
					},
				},
				Animation: anim.RedCoin(atlas),
			}
			a.AddEntity(t)
			log.Info().Msgf("red coin added: %s", location)
		default:
			log.Warn().Msgf("Unknown entity type: %s (sprite:%#v)", entityType, entity.SpriteId)
		}
	}
}
