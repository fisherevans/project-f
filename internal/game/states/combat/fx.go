package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/colors/typecolors"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/text"
)

type FX interface {
	Update(ctx *game.Context, s *State, timeDelta float64) bool
	Render(ctx *game.Context, target pixel.Target)
}

type DamageFX struct {
	Damage rpg.DamageResult

	Position pixel.Vec
	Velocity pixel.Vec
	Age      float64
}

var damageFxMaxAge = 4.0
var damageFxText = text.New(pixel.ZV, resources.Fonts.M6.Atlas)
var damageFxGravity = -100.0

func (fx *DamageFX) Update(ctx *game.Context, s *State, timeDelta float64) bool {
	fx.Age += timeDelta
	fx.Position = fx.Position.Add(fx.Velocity.Scaled(timeDelta))
	fx.Velocity = pixel.V(fx.Velocity.X, fx.Velocity.Y+damageFxGravity*timeDelta)
	return fx.Age > damageFxMaxAge
}

func (fx *DamageFX) Render(ctx *game.Context, target pixel.Target) {
	color := typecolors.SkillTypeColor(fx.Damage.DamageType)
	color = colors.WithAlpha(color, 1.0-(fx.Age/damageFxMaxAge))

	str := fmt.Sprintf("%d", fx.Damage.TotalDamage)

	damageFxText.Clear()
	damageFxText.Dot = pixel.ZV
	damageFxText.Color = colors.ScaleColor(color, 0.1)
	damageFxText.WriteString(str)
	damageFxText.Draw(target, pixel.IM.Moved(fx.Position))

	damageFxText.Dot = pixel.ZV.Add(pixel.V(-1, 1))
	damageFxText.Color = color
	damageFxText.WriteString(str)
	damageFxText.Draw(target, pixel.IM.Moved(fx.Position))

	ctx.DebugTR("damage: %d, pod: %.0f, %.0f", fx.Damage.TotalDamage, fx.Position.X, fx.Position.Y)
}
