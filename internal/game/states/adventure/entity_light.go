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
	registerDynamicEntity().
		byTile(tiles.Torch, tiles.TorchRight, tiles.TorchLeft).
		register(func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
			e := NewDynamicEntity(entityId, location)
			e.SetDefaultIsPassable(true)
			e.WithLights("", NewDynamicLight(colors.FromString("#db9a3d"), 2, "flicker"))
			switch *mapEntity.SpriteId {
			case tiles.Torch:
				e.WithAnimations("", anim.Torch(atlas))
			case tiles.TorchRight:
				e.WithAnimations("", anim.TorchRight(atlas))
				e.WithLightOriginOffset("", pixel.V(float64(resources.MapTileSize/2), 0))
				e.renderZPriority = 10
			case tiles.TorchLeft:
				e.WithAnimations("", anim.TorchLeft(atlas))
				e.WithLightOriginOffset("", pixel.V(-float64(resources.MapTileSize/2), 0))
				e.renderZPriority = 10
			}
			return e, nil
		})
	registerDynamicEntity().
		byTile(tiles.LightCircle, tiles.LightTable, tiles.LightTall, tiles.LightWide, tiles.LightFork, tiles.LightDoubleL, tiles.LightDoubleR).
		register(func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
			e := NewDynamicEntity(entityId, location)
			e.SetDefaultIsPassable(true)
			e.WithLights("", NewDynamicLight(colors.FromString("#fff"), 1.333, ""))
			e.WithAnimations("", anim.NewStaticAnimationFromId(atlas, *mapEntity.SpriteId))
			if *mapEntity.SpriteId == tiles.LightFork {
				e.renderZPriority = 10
			}
			return e, nil
		})
	registerDynamicEntity().
		byTile(tiles.GlowRed, tiles.GlowOrange, tiles.GlowAqua, tiles.GlowPurple, tiles.GlowPink, tiles.GlowTBD).
		register(func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
			colorMask := colors.HexString("#fff")
			switch *mapEntity.SpriteId {
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
			e := NewDynamicEntity(entityId, location)
			e.SetDefaultIsPassable(true)
			e.WithLights("", NewDynamicLight(colorMask, 2, ""))
			return e, nil
		})
}

type LightEntity struct {
	InnateEntity
	Light
	Animations []*anim.AnimatedSprite
	Passable
}

func (i *LightEntity) Update(adv *State, timeDelta float64) {
	i.Light.Update(timeDelta)
	for _, a := range i.Animations {
		a.Update(timeDelta)
	}
}

func (i *LightEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	for _, a := range i.Animations {
		a.Sprite().Draw(target, matrix)
	}
}

func (i *LightEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	i.Light.Render(target, matrix)
}
