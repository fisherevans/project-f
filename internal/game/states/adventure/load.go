package adventure

import (
	"fmt"
	"math/rand"
	"slices"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/tiles"
)

func initializeMap(a *State, m *resources.Map) {
	a.sceneClear = m.SceneClearColor
	a.lightClear = m.LightingClearColor
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
			a.AddEntity(&EntityCollision{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: EntityId(fmt.Sprintf("collision-%d-%d", location.X, location.Y)),
					},
					MapLocation: location,
				},
				Passable: newPassablePreventIngress(true),
			})
			// TODO dashing!
			//case resources.TileCollisionJumpHorizontal:
			//	a.movementRestrictions[location] = MovementJumpTile{}
			//case resources.TileCollisionJumpVertical:
			//	a.movementRestrictions[location] = MovementJumpTile{}
			//case resources.TileCollisionJumpAll:
			//	a.movementRestrictions[location] = MovementJumpTile{}
		}
	}
	for stringEntityId, entity := range m.Entities {
		if entity == nil {
			log.Fatal().Str("entityId", stringEntityId).Msg("entity cannot be nil")
		}
		entityId := EntityId(stringEntityId)
		location := adjustedLocation(entity.X, entity.Y)
		if a.registerDynamicEntity(entityId, location, entity) {
			log.Info().Msgf("added dynamic entity '%s'", stringEntityId)
			continue
		}
		entityType := entity.GetStringMetadata("type", "")
		if entity.SpriteId != nil {
			switch *entity.SpriteId {
			case tiles.Torch, tiles.TorchRight, tiles.TorchLeft:
				entityType = "torch"
			case tiles.LightCircle, tiles.LightTable, tiles.LightTall, tiles.LightWide, tiles.LightFork, tiles.LightDoubleL, tiles.LightDoubleR:
				entityType = "light"
			case tiles.GlowRed,
				tiles.GlowOrange,
				tiles.GlowAqua,
				tiles.GlowPurple,
				tiles.GlowPink,
				tiles.GlowTBD:
				entityType = "glow"
			case tiles.RedCoin:
				entityType = "red_coin"
			case tiles.Knight:
				entityType = "interest"
			case tiles.PDAEnabled:
				entityType = "pda"
			case tiles.LevelUp:
				entityType = "level_up"
			case tiles.Rocket:
				entityType = "rocket"
			case tiles.ShadowMob:
				entityType = "shadow"
			case tiles.DummyFightRobot:
				entityType = "dummy_robot"
			}
		}
		switch entityType {
		case "player":
			normalPlayerLight := &Light{
				RenderDetails: LightRenderDetails{
					SizeScale: 1.5,
					ColorMask: colors.HexString("#888"),
				},
			}
			dashPlayerLight := &Light{
				RenderDetails: LightRenderDetails{
					SizeScale: 1.5,
					ColorMask: colors.HexString("#88f"),
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
						Passable: newPassablePreventIngress(true),
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
			a.worldState.Set("player_id", string(entityId))
		case "blob":
			npc := &NPC{
				AnimatedMoveableEntity: AnimatedMoveableEntity{
					MoveableEntity: MoveableEntity{
						BaseEntity: BaseEntity{
							Id:           entityId,
							Interactable: true,
						},
						CurrentLocation: location,
						MoveSpeeds: map[MoveState]float64{
							MoveStateWalking: 2,
						},
						Passable: newPassablePreventIngress(true),
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
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				hasItem:  true,
				item:     entity.GetStringMetadata("item", "a sock"),
				Passable: newPassablePreventIngress(true),
			})
		case "interest":
			a.AddEntity(&EntityInterest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				topic:    entity.GetStringMetadata("topic", ""),
				Passable: newPassablePreventIngress(true),
			})
		case "dummy_robot":
			a.AddEntity(NewEntityCombatTest(entityId, location))
		case "menu_test":
			a.AddEntity(&EntityMenuTest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				Passable: newPassablePreventIngress(true),
			})
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
		case "torch":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 2,
						ColorMask: colors.HexString("#db9a3d"),
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
				Passable: newPassablePreventIngress(false),
			}
			switch *entity.SpriteId {
			case tiles.Torch:
				t.Animations = []*anim.AnimatedSprite{anim.Torch(atlas)}
			case tiles.TorchRight:
				t.Animations = []*anim.AnimatedSprite{anim.TorchRight(atlas)}
				t.RenderZPriority = 10
				t.Light.RenderDetails.PositionDelta = pixel.V(float64(resources.MapTileSize/2), 0)
			case tiles.TorchLeft:
				t.Animations = []*anim.AnimatedSprite{anim.TorchLeft(atlas)}
				t.RenderZPriority = 10
				t.Light.RenderDetails.PositionDelta = pixel.V(-float64(resources.MapTileSize/2), 0)
			}
			a.AddEntity(t)
		case "red_coin":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 0.5,
						ColorMask: colors.HexString("#f00"),
					},
					Modifiers: []LightModifier{
						&LightModifierPulse{
							PeriodSeconds:       0.5,
							SizeIntensity:       0.1,
							BrightnessIntensity: 0.4,
						},
					},
				},
				Animations: []*anim.AnimatedSprite{anim.RedCoin(atlas)},
				Passable:   newPassablePreventIngress(false),
			}
			a.AddEntity(t)
		case "light":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 1,
						ColorMask: colors.HexString("#fff"),
					},
				},
				Animations: []*anim.AnimatedSprite{anim.NewStaticAnimation(entity.SpriteId.From(atlas))},
				Passable:   newPassablePreventIngress(false),
			}
			if *entity.SpriteId == tiles.LightFork {
				t.RenderZPriority = 10
			}
			a.AddEntity(t)
		case "glow":
			colorMask := colors.HexString("#fff")
			switch *entity.SpriteId {
			case tiles.GlowRed:
				colorMask = colors.HexString("#e05050")

			case tiles.GlowOrange:
				colorMask = colors.HexString("#e09b50")
			case tiles.GlowAqua:
				colorMask = colors.HexString("#50d5e0")
			case tiles.GlowPurple:
				colorMask = colors.HexString("#50d5e0")
			case tiles.GlowPink:
				colorMask = colors.HexString("#e050cb")
			}
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id: entityId,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 2,
						ColorMask: colorMask,
					},
				},
				Passable: newPassablePreventIngress(false),
			}
			if *entity.SpriteId == tiles.LightFork {
				t.RenderZPriority = 10
			}
			a.AddEntity(t)
		case "pda":
			a.AddEntity(&EntityAnimatedInterest{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				Passable:     newPassablePreventIngress(true),
				OnAnimation:  anim.PDA(atlas),
				OffAnimation: anim.PDADisabled(atlas),
				OnMessage:    "You've got mail!",
				OffMessage:   "There's nothing new here...",
				toggled:      true,
			})
		case "shadow":
			a.AddMob(NewShadowMob(entityId, location))
		case "rocket":
			dest := TeleportReference("teleport:" + entity.GetStringMetadata("destination", ""))
			e := &EntityTeleport{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				RequiredElythium: 2,
				Destination:      dest,
				Passable:         newPassablePreventIngress(true),
			}
			a.AddEntity(e)
		case "level_up":
			t := &LightEntity{
				InnateEntity: InnateEntity{
					BaseEntity: BaseEntity{
						Id:           entityId,
						Interactable: true,
					},
					MapLocation: location,
				},
				Light: Light{
					RenderDetails: LightRenderDetails{
						SizeScale: 1,
						ColorMask: colors.HexString("#91f5a8"),
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
				Animations: []*anim.AnimatedSprite{
					anim.Load(atlas, "space_base", "level_up"),
				},
				Passable: newPassablePreventIngress(true),
			}
			a.AddEntity(t)
		default:
			log.Warn().Msgf("Unknown entity type: %s / %s", entityId, entityType)
		}
	}
	a.processEffects(a.eventDispatcher.Init(a.worldState))
}
