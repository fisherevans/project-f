package adventure

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
)

func init() {
	newRegistrarBuilder().
		byTile(tiles.Torch, tiles.TorchRight, tiles.TorchLeft).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			renderer := NewBasicEntityRenderer(entity)
			renderer.WithLights(NewLightWithModifier(colors.FromString("#db9a3d"), 2, "flicker"))
			switch *params.SpriteId {
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
			entity.SetRenderer(renderer)
			return entity, nil
		})
	newRegistrarBuilder().
		byTile(tiles.LightCircle, tiles.LightTable, tiles.LightTall, tiles.LightWide, tiles.LightFork, tiles.LightDoubleL, tiles.LightDoubleR).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			entity := system.RegisterEntity(params.EntityId, params.Location)
			renderer := NewBasicEntityRenderer(entity)
			renderer.WithLights(NewLight(colors.FromString("#fff"), 1.333))
			renderer.WithAnimations(anim.NewStaticAnimationFromId(atlas, *params.SpriteId))
			if *params.SpriteId == tiles.LightFork {
				renderer.WithZPriority(10)
			}
			entity.SetRenderer(renderer)
			return entity, nil
		})
	newRegistrarBuilder().
		byTile(tiles.GlowRed, tiles.GlowOrange, tiles.GlowAqua, tiles.GlowPurple, tiles.GlowPink, tiles.GlowTBD).
		registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
			colorMask := colors.HexString("#fff")
			params.Properties.GetFloat("size", 2)
			switch *params.SpriteId {
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

			entity := system.RegisterEntity(params.EntityId, params.Location)
			renderer := NewBasicEntityRenderer(entity)
			renderer.WithLights(NewLight(colorMask, 2))
			if *params.SpriteId == tiles.LightFork {
				renderer.WithZPriority(10)
			}
			entity.SetRenderer(renderer)
			return entity, nil
		})
}
