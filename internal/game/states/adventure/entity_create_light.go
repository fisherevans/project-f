package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
)

func init() {
	targetRegistration().
		byTile(tiles.Torch, tiles.TorchRight, tiles.TorchLeft).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
			renderer := NewBasicEntityRenderer()
			renderer.WithLights(NewLightWithModifier(colors.FromString("#db9a3d"), 2, "flicker"))
			switch *mapEntity.SpriteId {
			case tiles.Torch:
				renderer.WithAnimations(anim.Torch(atlas))
			case tiles.TorchRight:
				renderer.WithAnimations(anim.TorchRight(atlas))
				renderer.WithLightOriginOffset(pixel.V(float64(resources.MapTileSize/2), 0))
				renderer.WithZPriority(10)
			case tiles.TorchLeft:
				renderer.WithAnimations(anim.TorchLeft(atlas))
				renderer.WithLightOriginOffset(pixel.V(-float64(resources.MapTileSize/2), 0))
				renderer.WithZPriority(10)
			}
			system.RegisterEntity(entityId, location, nil, nil, renderer, nil)
			return nil, nil
		})
	targetRegistration().
		byTile(tiles.LightCircle, tiles.LightTable, tiles.LightTall, tiles.LightWide, tiles.LightFork, tiles.LightDoubleL, tiles.LightDoubleR).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
			renderer := NewBasicEntityRenderer()
			renderer.WithLights(NewLight(colors.FromString("#fff"), 1.333))
			renderer.WithAnimations(anim.NewStaticAnimationFromId(atlas, *mapEntity.SpriteId))
			if *mapEntity.SpriteId == tiles.LightFork {
				renderer.WithZPriority(10)
			}
			system.RegisterEntity(entityId, location, nil, nil, renderer, nil)
			return nil, nil
		})
	targetRegistration().
		byTile(tiles.GlowRed, tiles.GlowOrange, tiles.GlowAqua, tiles.GlowPurple, tiles.GlowPink, tiles.GlowTBD).
		registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) (events.EntityContext, events.EventHandler) {
			colorMask := colors.HexString("#fff")
			mapEntity.Properties.GetFloat("size", 2)
			switch *mapEntity.SpriteId {
			case tiles.GlowRed:
				colorMask = colors.HexString("#e05050")
			case tiles.GlowOrange:
				colorMask = colors.HexString("#e09b50")
			case tiles.GlowAqua:
				colorMask = colors.HexString("#50d5e0")
			case tiles.GlowPurple:
				colorMask = colors.HexString("#9350e0")
			case tiles.GlowPink:
				colorMask = colors.HexString("#e050cb")
			}
			renderer := NewBasicEntityRenderer()
			renderer.WithLights(NewLight(colorMask, 2))
			if *mapEntity.SpriteId == tiles.LightFork {
				renderer.WithZPriority(10)
			}
			system.RegisterEntity(entityId, location, nil, nil, renderer, nil)
			return nil, nil
		})
}
