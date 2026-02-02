package warming

import (
	"time"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

func Warmup(targetX pixel.Target) {
	log.Info().Msgf("Warming sprites")
	start := time.Now()
	atlas := resources.DefaultAtlas()
	batch := atlas.NewBatch()
	for _, d := range atlas.GetAllSprites() {
		d.Draw(batch, pixel.IM)
	}
	for _, d := range atlas.GetAllFrameSprites() {
		d.Draw(batch, pixel.IM)
	}
	for _, d := range atlas.GetAllTilesheetSprites() {
		d.Draw(batch, pixel.IM)
	}
	for _, f := range atlas.GetAllFonts() {
		tb := textbox.NewInstance(f, tbcfg.NewConfig(0, 0))
		c := tb.NewComplexContent("hello")
		c.Render(batch, pixel.IM)
	}
	batch.Draw(targetX)
	log.Info().Dur("duration", time.Since(start)).Msgf("Warming sprites complete")
}
