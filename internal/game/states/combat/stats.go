package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
	"fmt"
	"github.com/gopxl/pixel/v2"
)

var (
	combatStatFrame            = frames.New("combat/combatant_stats/box", atlas)
	statBarFrame               = frames.New("combat/combatant_stats/bar", atlas)
	combatantNameText          = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard), tbcfg.NewConfig(200, tbcfg.WithExpandMode(tbcfg.ExpandFit)))
	statBorderPadding          = 3
	statBorderPaddingNameExtra = 6
	combatantStatText          = textbox.NewInstance(atlas.GetFont(resources.FontNameFF), tbcfg.NewConfig(200, tbcfg.WithExpandMode(tbcfg.ExpandFit)))
	noneSelectedSprite         = atlas.GetSprite("combat/tick_bar/skill_none_selected")
	statNameBoxSprite          = atlas.GetTilesheetSprite("combat/combatant_stats/background", 1, 1)
	statRightSprite            = atlas.GetTilesheetSprite("combat/combatant_stats/background", 2, 1)
	statBottomSprite           = atlas.GetTilesheetSprite("combat/combatant_stats/background", 3, 1)
	statBoxWidth               = 80
)

func (s *State) drawPlayerStats(ctx *game.Context) {
	syncBar := &StatBar{
		lines:       []StatBarLine{StatBarVisual, StatBarLabel},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 6, 1),
		label:       "sync",
		colorDark:   colors.HexColor("772712"),
		color:       colors.HexColor("d58d7a"),
		colorBright: colors.HexColor("ebc4bb"),
		current:     s.Player.GetCombatant().GetCurrentSync().Current,
		max:         s.Player.GetCombatant().GetCurrentSync().Max,
	}

	shieldBar := &StatBar{
		lines:       []StatBarLine{StatBarLabel, StatBarVisual},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 5, 1),
		label:       "shield",
		colorDark:   colors.HexColor("126177"),
		color:       colors.HexColor("73bed3"),
		colorBright: colors.HexColor("bbe0eb"),
		current:     s.Player.GetCombatant().GetCurrentShield().Current,
		max:         s.Player.GetCombatant().GetCurrentShield().Max,
	}

	statBox := &StatBox{
		bars:           []*StatBar{shieldBar, syncBar},
		originLocation: StatBoxOriginTopLeft,
	}

	s.drawCombatantStatBox(ctx, "Fumalug", statBox, gfx.IVec(0, game.GameHeight), gfx.TopLeft)
}

func (s *State) drawOpponentStats(ctx *game.Context) {
	healthBar := &StatBar{
		lines:       []StatBarLine{StatBarLabel, StatBarVisual},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 7, 1),
		label:       "health",
		colorDark:   colors.HexColor("127839"),
		color:       colors.HexColor("73d398"),
		colorBright: colors.HexColor("bcebce"),
		current:     s.Opponent.GetHealth().Current,
		max:         s.Opponent.GetHealth().Max,
	}

	statBox := &StatBox{
		bars:           []*StatBar{healthBar},
		originLocation: StatBoxOriginTopRight,
	}

	s.drawCombatantStatBox(ctx, "Plent", statBox, gfx.IVec(game.GameWidth, game.GameHeight), gfx.TopRight)
}

func (s *State) drawCombatantStatBox(ctx *game.Context, name string, statBox *StatBox, origin pixel.Vec, originLocation gfx.OriginLocation) {
	renderScale := pixel.V(1, 1)
	var nameContentOpts []textbox.ContentOpt
	if originLocation == gfx.TopRight {
		nameContentOpts = append(nameContentOpts, textbox.WithAlignment(tbcfg.AlignRight))
		renderScale = pixel.V(-1, 1)
	}
	nameContent := combatantNameText.NewComplexContent("{+o:#cfcfcf,+c:black}"+name, nameContentOpts...)
	paddedNameHeight := statBorderPadding + combatantNameText.Metadata.GetFullLineHeight() + 2 // 1 for outline, 1 for spacing

	matrix := pixel.IM.Moved(origin)

	nameBoxHeight := paddedNameHeight + statBox.FrameHeight() - int(statBottomSprite.Bounds().H())
	nameBoxWidth := statBorderPadding + statBorderPaddingNameExtra + nameContent.Width() + statBorderPaddingNameExtra
	nameBoxSpriteScaleX := float64(nameBoxWidth) / statNameBoxSprite.Bounds().W()
	nameBoxSpriteScaleY := float64(nameBoxHeight) / statNameBoxSprite.Bounds().H()
	statNameBoxSprite.Draw(s.batch, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(nameBoxSpriteScaleX, nameBoxSpriteScaleY).ScaledXY(renderScale)).
		Chained(matrix).
		Moved(originLocation.AlignInt(nameBoxWidth, nameBoxHeight)))

	statRightSprite.Draw(s.batch, pixel.IM.ScaledXY(pixel.ZV, renderScale).Chained(matrix).
		Moved(originLocation.Align(statRightSprite)).
		Moved(gfx.IVec(nameBoxWidth-1, 0).ScaledXY(renderScale)))

	statBottomSprite.Draw(s.batch, pixel.IM.ScaledXY(pixel.ZV, renderScale).Chained(matrix).
		Moved(originLocation.Align(statBottomSprite)).
		Moved(gfx.IVec(0, -nameBoxHeight).ScaledXY(renderScale)))

	combatantNameText.Render(ctx, s.batch,
		matrix.Moved(gfx.IVec(statBorderPadding+statBorderPaddingNameExtra, -statBorderPadding).ScaledXY(renderScale)),
		nameContent,
		tbcfg.RenderFrom(originLocation))

	statBox.Draw(ctx, s.batch, matrix.Moved(gfx.IVec(statBorderPadding, -paddedNameHeight).ScaledXY(renderScale)), statBoxWidth)
}

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

func (sb *StatBox) Draw(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, frameWidth int) {
	frameHeight := sb.FrameHeight()

	ctx.DebugBR("frame height: %d", frameHeight)

	switch sb.originLocation {
	case StatBoxOriginTopLeft:
		matrix = matrix.Moved(pixel.V(0, -float64(frameHeight)))
	case StatBoxOriginTopRight:
		matrix = matrix.Moved(pixel.V(-float64(frameWidth), -float64(frameHeight)))
	}

	combatStatFrame.Draw(target, pixel.R(0, 0, float64(frameWidth), float64(frameHeight)), matrix)

	// draw bars inside padding of frame
	matrix = matrix.Moved(gfx.IVec(
		statBoxContentPadding+combatStatFrame.Padding[resources.FrameLeft],
		statBoxContentPadding+combatStatFrame.Padding[resources.FrameTop]))

	maxBarWidth := float64(frameWidth - 2*statBarFrame.HorizontalPadding())
	maxMax := 0
	for _, bar := range sb.bars {
		if bar.max > maxMax {
			maxMax = bar.max
		}
	}

	lastBarId := len(sb.bars) - 1
	for barId := lastBarId; barId >= 0; barId-- {
		bar := sb.bars[barId]
		width := int(float64(bar.max) / float64(maxMax) * maxBarWidth)
		if barId != lastBarId {
			matrix = matrix.Moved(pixel.V(0, 1))
		}
		bar.Draw(ctx, target, matrix, width)
		matrix = matrix.Moved(pixel.V(0, float64(bar.Height())))
	}
}

func (sb *StatBox) FrameHeight() int {
	contentHeight := 0
	for barId, bar := range sb.bars {
		if barId > 0 {
			contentHeight += 1
		}
		contentHeight += bar.Height()
	}
	return contentHeight + statBoxContentPadding*2 + combatStatFrame.VerticalPadding()
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
	labelSprite                   pixelutil.BoundedDrawable
	label                         string
	colorDark, color, colorBright pixel.RGBA
	current, max                  int
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

func (sb *StatBar) Draw(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, width int) int {
	lastLineId := len(sb.lines) - 1
	for lineId := lastLineId; lineId >= 0; lineId-- {
		line := sb.lines[lineId]
		if lineId != lastLineId {
			matrix = matrix.Moved(pixel.V(0, 1))
		}
		switch line {
		case StatBarLabel:
			//moveVec := sb.labelSprite.Bounds().Center()
			labelContent := combatantStatText.NewComplexContent(fmt.Sprintf("{+c:%s}%s", colors.ToHex(sb.color), sb.label)) // TODO don't compute hex
			combatantStatText.Render(ctx, target, matrix.Moved(pixel.V(float64(2), 0)), labelContent)
			//sb.labelSprite.DrawColorMask(target, matrix.Moved(pixel.V(float64(2), float64(5)-sb.labelSprite.Bounds().H())).Moved(moveVec), sb.color)

			valueContent := combatantStatText.NewComplexContent(fmt.Sprintf("{+c:%s}%d{+c:%s}/%d", colors.ToHex(sb.colorBright), sb.current, colors.ToHex(sb.color), sb.max)) // TODO don't compute hex
			combatantStatText.Render(ctx, target, matrix.Moved(pixel.V(float64((width-valueContent.Width())-2), 0)), valueContent)
		case StatBarVisual:
			maxRectWidth := width - 2
			currentRectWidth := int(float64(maxRectWidth) * float64(sb.current) / float64(sb.max))
			gfx.DrawRect(atlas, target, matrix.Moved(pixel.V(float64(1), float64(1))), gfx.BottomLeft, currentRectWidth, 3, sb.colorDark)
			if currentRectWidth < maxRectWidth {
				gfx.DrawRect(atlas, target, matrix.Moved(pixel.V(float64(1+currentRectWidth-1), float64(1))), gfx.BottomLeft, 1, 3, colors.ScaleColor(sb.colorBright, 1.5))
				gfx.DrawRect(atlas, target, matrix.Moved(pixel.V(float64(1+currentRectWidth), float64(1))), gfx.BottomLeft, maxRectWidth-currentRectWidth, 3, sb.colorBright)
			}
			statBarFrame.Draw(target, pixel.R(0, 0, float64(width), float64(5)), matrix, frames.WithColor(sb.color))
		}
		matrix = matrix.Moved(pixel.V(0, float64(line.Height())))
	}
	return sb.Height()
}
