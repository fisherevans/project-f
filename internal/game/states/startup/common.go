package startup

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/gfx"
)

var atlas = resources.CreateAtlas(resources.AtlasFilter{
	DoIncludeSprite: resources.RequireSpritePrefix("startup/", "1x1", "2x2"),
	FontNames: []string{
		resources.FontNameFF,
	},
})

var centerMatrix = gfx.Moved(game.GameWidth/2, game.GameHeight/2)
