package anim

import (
	"fmt"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type AnimatedSprite struct {
	frames          []frame
	framesPerSecond float64
	randomized      bool
	progression     float64
	currentFrame    int
}

type frame struct {
	drawable pixelutil.BoundedDrawable
	weight   float64
}

func NewStaticAnimation(f pixelutil.BoundedDrawable) *AnimatedSprite {
	return &AnimatedSprite{
		frames: []frame{{
			drawable: f,
			weight:   1,
		}},
		framesPerSecond: 1,
	}
}

func NewStaticAnimationFromId(atlas *resources.Atlas, tileId resources.TilesheetSpriteId) *AnimatedSprite {
	return NewStaticAnimation(atlas.GetTilesheetSpriteById(tileId))
}

func (a *AnimatedSprite) ApplyPingPong() *AnimatedSprite {
	for i := len(a.frames) - 2; i > 0; i-- {
		a.frames = append(a.frames, a.frames[i])
	}
	return a
}

func FromTilesheetRow(atlas *resources.Atlas, tilesheet string, row int, framesPerSecond float64) *AnimatedSprite {
	ts := resources.GetTilesheet(tilesheet)
	return FromTilesheetRowPartial(atlas, tilesheet, row, 1, ts.Columns, framesPerSecond)
}

func FromTilesheetRowPartial(atlas *resources.Atlas, tilesheet string, row, colFrom, columns int, framesPerSecond float64) *AnimatedSprite {
	var tiles []*resources.SpriteTilesheetAnimationTile
	for i := colFrom; i < colFrom+columns; i++ {
		tiles = append(tiles, &resources.SpriteTilesheetAnimationTile{
			Column: i,
			Row:    row,
			Weight: 1,
		})
	}
	return FromTilesheetTiles(atlas, tilesheet, framesPerSecond, false, tiles)
}

func FromTilesheetTiles(atlas *resources.Atlas, tilesheet string, framesPerSecond float64, randomized bool, tiles []*resources.SpriteTilesheetAnimationTile) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
		randomized:      randomized,
	}
	for _, tile := range tiles {
		animated.frames = append(animated.frames, frame{
			drawable: atlas.GetTilesheetSprite(tilesheet, tile.Column, tile.Row),
			weight:   tile.Weight,
		})
	}
	fmt.Printf("tilesheet animation: %s:\n%#v\n", tilesheet, animated.frames)
	return animated
}

func Load(atlas *resources.Atlas, tilesheetName string, animationName string) *AnimatedSprite {
	metadata := resources.GetTilesheetAnimation(tilesheetName, animationName)
	msgf := func(format string, args ...interface{}) string {
		return fmt.Sprintf(
			"invalid tilesheet animation (%s/%s): %s",
			tilesheetName,
			animationName,
			fmt.Sprintf(format, args...),
		)
	}
	if metadata == nil {
		panic(msgf("not found"))
	}
	tilesheet := resources.GetTilesheet(tilesheetName)
	if tilesheet == nil {
		panic(msgf("invalid tilesheet"))
	}
	if metadata.Sequence != nil && len(metadata.Tiles) > 0 {
		panic(msgf("cannot specify both sequence and tiles"))
	}
	var tiles []*resources.SpriteTilesheetAnimationTile
	if metadata.Sequence != nil {
		if metadata.Sequence.Row <= 0 || metadata.Sequence.Row > tilesheet.Rows {
			panic(msgf("invalid row %d", metadata.Sequence.Row))
		}
		from := metadata.Sequence.ColumnFrom
		to := metadata.Sequence.ColumnTo
		if from == 0 && to == 0 {
			to = tilesheet.Columns
		}
		if from <= 0 || from >= to || to > tilesheet.Columns {
			panic(msgf("invalid column range %d-%d", from, to))
		}
		if len(metadata.Sequence.FrameWeights) == 0 {
			for i := from; i <= to; i++ {
				metadata.Sequence.FrameWeights = append(metadata.Sequence.FrameWeights, 1)
			}
		}
		if len(metadata.Sequence.FrameWeights) != to-from+1 {
			panic(msgf("invalid frame weights length %d", len(metadata.Sequence.FrameWeights)))
		}
		for i := from; i <= to; i++ {
			tiles = append(tiles, &resources.SpriteTilesheetAnimationTile{
				Column: i,
				Row:    metadata.Sequence.Row,
				Weight: metadata.Sequence.FrameWeights[i-from],
			})
		}
	}
	for tileId, tile := range metadata.Tiles {
		if tile.Row <= 0 || tile.Row > tilesheet.Rows {
			panic(msgf("invalid row %d", tile.Row))
		}
		if tile.Column < 0 || tile.Column > tilesheet.Columns {
			panic(msgf("invalid column %d", tile.Column))
		}
		if tile.Weight < 0 {
			panic(msgf("invalid weight tileId %d: %f", tileId, tile.Weight))
		}
		if tile.Weight == 0 {
			tile.Weight = 1
		}
		tiles = append(tiles, &resources.SpriteTilesheetAnimationTile{
			Column: tile.Column,
			Row:    tile.Row,
			Weight: tile.Weight,
		})
	}
	if metadata.FramesPerSecond < 0 {
		panic(msgf("invalid frames per second %f", metadata.FramesPerSecond))
	}
	if metadata.FramesPerSecond == 0 {
		metadata.FramesPerSecond = 1
	}
	//j, _ := json.MarshalIndent(tiles, "", "  ")
	//fmt.Printf("tilesheet animation: %s/%s:\n%s\n", tilesheetName, animationName, j)
	return FromTilesheetTiles(atlas, tilesheetName, metadata.FramesPerSecond, metadata.Randomize, tiles)
}

func (a *AnimatedSprite) Sprite() pixelutil.BoundedDrawable {
	return a.frames[a.currentFrame].drawable
}

func (a *AnimatedSprite) Update(timeDelta float64) {
	a.progression += timeDelta

	for {
		w := a.frames[a.currentFrame].weight
		if w <= 0 {
			w = 1
		}
		dur := w / a.framesPerSecond // secondsPerFrame * weight
		if a.progression < dur {
			break
		}
		a.progression -= dur
		a.currentFrame = (a.currentFrame + 1) % len(a.frames)
	}
}

func (a *AnimatedSprite) Reset() {
	a.currentFrame = 0
	a.progression = 0
}
