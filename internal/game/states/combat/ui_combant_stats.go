package combat

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	combatStatFrame            *frames.Instance
	statBarFrame               *frames.Instance
	combatantNameText          *textbox.Instance
	statBorderPadding          = 1
	statBorderPaddingNameExtra = 6
	combatantStatText          *textbox.Instance
	noneSelectedSprite         pixelutil.BoundedDrawable
	statNameBoxSprite          pixelutil.BoundedDrawable
	statRightSprite            pixelutil.BoundedDrawable
	statBottomSprite           pixelutil.BoundedDrawable
	statBoxWidth               = 80
)

func (s *State) drawPlayerStats(timeDelta float64) {
	syncBar := &StatBar{
		lines:       []StatBarLine{StatBarVisual, StatBarLabel},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 6, 1),
		label:       "sync",
		colorDark:   colors.HexString("772712"),
		color:       colors.HexString("d58d7a"),
		colorBright: colors.HexString("ebc4bb"),
		current:     s.Player.GetCurrentSync().GetCurrentInt(),
		max:         s.Player.GetCurrentSync().Max,
	}

	shieldBar := &StatBar{
		lines:       []StatBarLine{StatBarLabel, StatBarVisual},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 5, 1),
		label:       "shield",
		colorDark:   colors.HexString("126177"),
		color:       colors.HexString("73bed3"),
		colorBright: colors.HexString("bbe0eb"),
		current:     s.Player.GetCurrentShield().GetCurrentInt(),
		max:         s.Player.GetCurrentShield().Max,
	}

	statBox := &StatBox{
		bars:           []*StatBar{shieldBar, syncBar},
		originLocation: StatBoxOriginTopLeft,
	}

	s.drawCombatantStatBox(s.Player.Name(), statBox, s.Player.GetStatuses(), s.Battle.GetPlayerCurrentTickProgress(), gfx.IVec(0, game.GameHeight), gfx.TopLeft, s.visibilityPlayerStats, timeDelta)
}

func (s *State) drawOpponentStats(timeDelta float64) {
	healthBar := &StatBar{
		lines:       []StatBarLine{StatBarLabel, StatBarVisual},
		labelSprite: atlas.GetTilesheetSprite("combat/combatant_stats/background", 7, 1),
		label:       "health",
		colorDark:   colors.HexString("127839"),
		color:       colors.HexString("73d398"),
		colorBright: colors.HexString("bcebce"),
		// TODO render target
		current: s.Opponent.GetHealth().GetCurrentInt(),
		max:     s.Opponent.GetHealth().Max,
	}

	statBox := &StatBox{
		bars:           []*StatBar{healthBar},
		originLocation: StatBoxOriginTopRight,
	}

	s.drawCombatantStatBox(s.Opponent.Name(), statBox, s.Opponent.GetStatuses(), s.Battle.GetOpponentCurrentTickProgress(), gfx.IVec(game.GameWidth, game.GameHeight), gfx.TopRight, s.visibilityOpponentStats, timeDelta)
}

func (s *State) drawCombatantStatBox(name string, statBox *StatBox, statuses *AppliedStatuses, currentTickProgress float64, origin pixel.Vec, originLocation gfx.OriginLocation, visibility *util.Visibility, timeDelta float64) {
	renderScale := pixel.V(1, 1)
	var nameContentOpts []textbox.ContentOpt
	if originLocation == gfx.TopRight {
		nameContentOpts = append(nameContentOpts, textbox.WithAlignment(tbcfg.HAlignRight))
		renderScale = pixel.V(-1, 1)
	}
	nameContent := combatantNameText.NewComplexContent("{+o:#cfcfcf,+c:black}"+name, nameContentOpts...)
	paddedNameHeight := statBorderPadding + combatantNameText.Metadata.GetFullLineHeight() + 3 // 1 for outline, 1 for spacing, 1 for letter tails

	matrix := pixel.IM.Moved(origin)
	dy := visibility.GetInvisibleAmount() * 40
	dx := visibility.GetInvisibleAmount() * -80
	matrix = matrix.Moved(pixel.V(dx, dy).ScaledXY(renderScale))

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

	nameContent.Render(s.batch,
		matrix.Moved(gfx.IVec(statBorderPadding+statBorderPaddingNameExtra, -statBorderPadding).ScaledXY(renderScale)),
		tbcfg.RenderFrom(originLocation))

	statBox.Draw(s.batch, matrix.Moved(gfx.IVec(statBorderPadding, -paddedNameHeight).ScaledXY(renderScale)), statBoxWidth)

	sidePadding := 5
	topCorner := matrix.
		Moved(pixel.V(0, -float64(statBox.FrameHeight()+paddedNameHeight))).
		Moved(originLocation.AlignFrom(gfx.Centered, float64(sidePadding*2), 0))
	topCorner = topCorner.Moved(gfx.IVec(0, -1))
	statuses.Render(topCorner, s.batch, timeDelta, originLocation, currentTickProgress)
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

func (sb *StatBox) Draw(target pixel.Target, matrix pixel.Matrix, frameWidth int) {
	frameHeight := sb.FrameHeight()

	game.DebugBRf("frame height: %d", frameHeight)

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
		bar.Draw(target, matrix, width)
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

func (sb *StatBar) Draw(target pixel.Target, matrix pixel.Matrix, width int) int {
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
			labelContent.Render(target, matrix.Moved(pixel.V(float64(2), 0)))
			//sb.labelSprite.DrawColorMask(target, matrix.MovedDelta(pixel.V(float64(2), float64(5)-sb.labelSprite.Bounds().H())).MovedDelta(moveVec), sb.color)

			valueContent := combatantStatText.NewComplexContent(fmt.Sprintf("{+c:%s}%d{+c:%s}/%d", colors.ToHex(sb.colorBright), sb.current, colors.ToHex(sb.color), sb.max)) // TODO don't compute hex
			valueDx := float64((width - valueContent.Width()) - 2)
			valueDx = math.Max(valueDx, float64(labelContent.Width()+4))
			valueContent.Render(target, matrix.Moved(pixel.V(valueDx, 0)))
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
