package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"github.com/gopxl/pixel/v2"
)

type Tempo struct {
	current int
}

func (t *Tempo) GetCurrent() int {
	return t.current
}

func (t *Tempo) Increment() {
	t.IncrementBy(1)
}

func (t *Tempo) IncrementBy(n int) {
	t.current += n
}

func (t *Tempo) Reset() {
	t.current = 0
}

var (
	tempoTextName = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(0))
)

func (t *Tempo) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix) {

}
