package btest

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
	"image/color"
	"strings"
)

type state struct {
	batch *pixel.Batch
}

var (
	atlas = resources.CreateAtlas(resources.AtlasFilter{
		FontNames: []string{resources.FontNameFF},
	})
)

func (s state) ClearColor() color.Color {
	return color.Black
}

func (s state) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()
	m := pixel.IM.Moved(gfx.IVec(game.GameWidth*0.75, game.GameHeight-10))
	for id, st := range rpg.AllSkillTypes {
		badges.Using(atlas).OfSkillType(st, id%2 == 0).Render(ctx, s.batch, m.Moved(gfx.IVec(1, 1)), gfx.BottomLeft)
		badges.Using(atlas).OfSkillType(st, id%2 != 0).Render(ctx, s.batch, m.Moved(gfx.IVec(-1, 1)), gfx.BottomRight)
		badges.Using(atlas).OfSkillType(st, id%2 != 0).Render(ctx, s.batch, m.Moved(gfx.IVec(1, -1)), gfx.TopLeft)
		badges.Using(atlas).OfSkillType(st, id%2 == 0).Render(ctx, s.batch, m.Moved(gfx.IVec(-1, -1)), gfx.TopRight)
		m = m.Moved(pixel.V(0, -20))
	}
	gfx.DrawRect(
		atlas,
		s.batch,
		pixel.IM,
		gfx.BottomLeft,
		100, game.GameHeight,
		colors.SkillTypeAbyssal.RGBA,
	)
	m = pixel.IM.Moved(gfx.IVec(10, 10))
	actions := [][2]string{
		{"a", "choose"},
		{"A", "select"},
		{"select", "info"},
		{"start", "menu"},
		{"b", "cancel"},
		{"B", "back"},
	}
	for _, a := range actions {
		pressed := false
		switch strings.ToLower(a[0]) {
		case "a":
			pressed = ctx.Controls.ButtonA().IsPressed()
		case "b":
			pressed = ctx.Controls.ButtonB().IsPressed()
		case "select":
			pressed = ctx.Controls.ButtonSelect().IsPressed()
		case "start":
			pressed = ctx.Controls.ButtonStart().IsPressed()
		}
		badges.Using(atlas).ButtonAction(a[0], a[1]).Highlighted(pressed).Render(ctx, s.batch, m, gfx.BottomLeft)
		m = m.Moved(pixel.V(0, 15))
	}
	s.batch.Draw(target)
}

func New() game.State {
	return &state{
		batch: atlas.NewBatch(),
	}
}
