package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/shapes"
	"fisherevans.com/project/f/internal/util/textbox"
	"fmt"
	"github.com/gopxl/pixel/v2"
)

var combatStatFrame = resources.Frames["combat_combatant_stats_frame"]
var statBarFrame = resources.Frames["combat_stat_bar_frame"]

var combatantNameText = textbox.NewInstance(textbox.FontBoldTitle, textbox.NewConfig(200).RenderFrom(textbox.TopLeft).ExpandMode(textbox.ExpandFit))
var nameContent = combatantNameText.NewComplexContent("{+o:#cfcfcf,+c:black}Fumalug")
var namePadding = 3

var combatantStatText = textbox.NewInstance(textbox.FontMicro, textbox.NewConfig(200).RenderFrom(textbox.BottomLeft).ExpandMode(textbox.ExpandFit))

var noneSelectedSprite = resources.Sprites["combat_tick_none_selected"]

type StatBoxOriginLocation int

const (
	StatBoxOriginTopLeft StatBoxOriginLocation = iota
	StatBoxOriginTopRight
)

type StatBox struct {
	bars           []*StatBar
	originLocation StatBoxOriginLocation
}

var statBoxContentPadding = 1

func (sb *StatBox) Draw(ctx *game.Context, target pixel.Target, matrix pixel.Matrix) {
	width, height := sb.Dimensions()

	switch sb.originLocation {
	case StatBoxOriginTopLeft:
		matrix = matrix.Moved(pixel.V(0, -float64(height)))
	case StatBoxOriginTopRight:
		matrix = matrix.Moved(pixel.V(-float64(width), -float64(height)))
	}

	frames.Draw(target, combatStatFrame, pixel.R(0, 0, float64(width), float64(height)), matrix)

	contentX := float64(statBoxContentPadding + combatStatFrame.Padding[resources.FrameLeft])
	contentY := float64(statBoxContentPadding + combatStatFrame.Padding[resources.FrameBottom])
	matrix = matrix.Moved(pixel.V(contentX, contentY))

	lastBarId := len(sb.bars) - 1
	for barId := lastBarId; barId >= 0; barId-- {
		bar := sb.bars[barId]
		if barId != lastBarId {
			matrix = matrix.Moved(pixel.V(0, 1))
		}
		bar.Draw(ctx, target, matrix)
		matrix = matrix.Moved(pixel.V(0, float64(bar.Height())))
	}
}

func (sb *StatBox) Dimensions() (int, int) {
	contentWidth, contentHeight := 0, 0
	for barId, bar := range sb.bars {
		if barId > 0 {
			contentHeight += 1
		}
		contentWidth = max(contentWidth, bar.width)
		contentHeight += bar.Height()
	}
	width := contentWidth + statBoxContentPadding*2 + combatStatFrame.HorizontalPadding()
	height := contentHeight + statBoxContentPadding*2 + combatStatFrame.VerticalPadding()
	return width, height
}

type StatBarLine int

const (
	StatBarLabel StatBarLine = iota
	StatBarVisual
)

func (l StatBarLine) Height() int {
	switch l {
	case StatBarLabel:
		return 5
	case StatBarVisual:
		return 5
	default:
		return 0
	}
}

type StatBar struct {
	lines                         []StatBarLine
	labelSprite                   *resources.SpriteReference
	colorDark, color, colorBright pixel.RGBA
	current, max                  int
	width                         int
	nameFirst                     bool
}

func (sb *StatBar) Height() int {
	height := 0
	for id, line := range sb.lines {
		if id > 0 {
			height++
		}
		height += line.Height()
	}
	return height
}

func (sb *StatBar) Draw(ctx *game.Context, target pixel.Target, matrix pixel.Matrix) int {
	lastLineId := len(sb.lines) - 1
	for lineId := lastLineId; lineId >= 0; lineId-- {
		line := sb.lines[lineId]
		if lineId != lastLineId {
			matrix = matrix.Moved(pixel.V(0, 1))
		}
		switch line {
		case StatBarLabel:
			sb.labelSprite.Sprite.DrawColorMask(target, matrix.Moved(pixel.V(float64(2), float64(5)-sb.labelSprite.Sprite.Frame().H())).Moved(sb.labelSprite.MoveVecBottomLeft()), sb.color)
			content := combatantStatText.NewComplexContent(fmt.Sprintf("{+c:%s}%d{+c:%s}/%d", colors.ToHex(sb.colorBright), sb.current, colors.ToHex(sb.color), sb.max)) // TODO don't compute hex
			combatantStatText.Render(ctx, target, matrix.Moved(pixel.V(float64((sb.width-content.Width())-2), 0)), content)
		case StatBarVisual:
			maxRectWidth := sb.width - 2
			currentRectWidth := int(float64(maxRectWidth) * float64(sb.current) / float64(sb.max))
			shapes.DrawRect2(target, matrix.Moved(pixel.V(float64(1), float64(1))), shapes.BottomLeft, currentRectWidth, 3, sb.colorDark)
			if currentRectWidth < maxRectWidth {
				shapes.DrawRect2(target, matrix.Moved(pixel.V(float64(1+currentRectWidth-1), float64(1))), shapes.BottomLeft, 1, 3, colors.ScaleColor(sb.colorBright, 1.5))
				shapes.DrawRect2(target, matrix.Moved(pixel.V(float64(1+currentRectWidth), float64(1))), shapes.BottomLeft, maxRectWidth-currentRectWidth, 3, sb.colorBright)
			}
			frames.Draw(target, statBarFrame, pixel.R(0, 0, float64(sb.width), float64(5)), matrix, frames.WithColor(sb.color))
		}
		matrix = matrix.Moved(pixel.V(0, float64(line.Height())))
	}
	return sb.Height()
}
