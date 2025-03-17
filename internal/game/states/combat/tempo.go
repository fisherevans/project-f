package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"fmt"
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

func newComboText(fontName string) *textbox.Instance {
	return textbox.NewInstance(
		atlas.GetFont(fontName),
		tbcfg.NewConfig(0, 0,
			tbcfg.RenderFrom(gfx.TopLeft),
			tbcfg.VAligned(tbcfg.AlignTop),
		))
}

var (
	tempoTextName        = newComboText(resources.FontNameFF)
	tempoTextComboSmall  = newComboText(resources.FontNameFF)
	tempoTextComboMedium = newComboText(resources.FontNameM3x6)
	tempoTextComboBig    = newComboText(resources.FontNameM5x7)
	tempoTextComboHuge   = newComboText(resources.FontNameAddStandard)
)

func (t *Tempo) Render(ctx *game.Context, target pixel.Target, matrix pixel.Matrix) {
	if t.current == 0 {
		return
	}

	var comboText *textbox.Instance
	var color colors.NamedColor
	if t.current >= 20 {
		//comboText = tempoTextComboHuge
		color = colors.Warm9
	} else if t.current >= 10 {
		//comboText = tempoTextComboBig
		color = colors.Warm8
	} else if t.current >= 5 {
		//comboText = tempoTextComboMedium
		color = colors.Warm7
	} else {
		//comboText = tempoTextComboSmall
		color = colors.Warm5
	}
	comboText = tempoTextComboSmall

	nameContent := tempoTextName.NewComplexContent("{+o:black}tempo")
	tempoTextName.Render(ctx, target, matrix, nameContent, tbcfg.Foreground(colors.Warm2.RGBA))
	matrix = matrix.Moved(gfx.IVec(nameContent.Width()+2, 0))

	comboContentText := fmt.Sprintf("{+o:warm_1}x%d", t.current)
	comboContent := comboText.NewComplexContent(comboContentText)
	comboText.Render(ctx, target, matrix, comboContent, tbcfg.Foreground(color.RGBA))
}
