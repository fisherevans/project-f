package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"math"
)

var (
	skillFrame                  = frames.New("combat/menu/skill_frame", atlas)
	skillPendingFrame           = frames.New("combat/menu/skill_pending_frame", atlas)
	skillFrameWidth             = 84
	skillFrameHeight            = 13
	skillFrameHorizontalSpacing = 26

	skillText = textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6), textbox.
			NewConfig(skillFrameWidth-skillFrame.HorizontalPadding()).
			Foreground(colors.Black.RGBA).
			Aligned(textbox.AlignCenter))
)

func (s *State) renderSkills(ctx *game.Context, target pixel.Target, bottomLeft pixel.Vec, timeDelta float64) {
	s.skillFlashTimeElapsed += timeDelta
	sin := (math.Sin(s.skillFlashTimeElapsed*10) + 1.0) / 2.0 // [0-1]
	s.skillFlashAlpha = 0.5 + sin*0.5                         // [0.5-1]
	s.skillFlashAlphaInverse = 0.5 + (1.0-sin)*0.5
	for optionId := 0; optionId < 4; optionId++ {
		text := ""
		option := s.Player.GetCombatant().GetFightOption(optionId)
		if option != nil {
			text = option.Get().Name
		}
		content := s.simpleSkillContent(text)
		matrix := pixel.IM.Moved(bottomLeft)
		rightDx := skillFrameWidth + skillFrameHorizontalSpacing
		switch typeOptionKey[optionId] {
		case input.Up:
			matrix = matrix.Moved(pixel.V(float64(rightDx/2), float64(skillFrameHeight-1)*2))
		case input.Down:
			matrix = matrix.Moved(pixel.V(float64(rightDx/2), 0))
		case input.Right:
			matrix = matrix.Moved(pixel.V(float64(rightDx), float64(skillFrameHeight-1)))
		case input.Left:
			matrix = matrix.Moved(pixel.V(0, float64(skillFrameHeight-1)))
		}
		frame := skillFrame
		if s.Player.NextSkill != nil && option != nil && option.Get().Id == *s.Player.NextSkill {
			frame = skillPendingFrame
		}
		frameRect := pixel.R(0, 0, float64(skillFrameWidth), float64(skillFrameHeight))
		frame.Draw(target, frameRect, matrix)
		textDy := (skillFrameHeight - skillText.Metadata.GetFullLineHeight()) / 2
		matrix = matrix.Moved(pixel.V(0, float64(textDy)))
		skillText.Render(ctx, target, matrix, content)
	}
	centerMatrix := pixel.IM.Moved(bottomLeft).Moved(pixel.V(
		float64(skillFrameWidth+(skillFrameHorizontalSpacing/2)),
		math.Ceil(float64(skillFrameHeight-1)*1.5)))
	atlas.GetTilesheetSprite("combat/menu/skill_arrows", 1, 1).Draw(target, centerMatrix)
	s.combatArrowAlpha -= timeDelta * 0.75
	switch ctx.Controls.DPad().PressedDirection() {
	case input.Up:
		s.combatArrowAlpha, s.combatArrowColumn = 1, 2
	case input.Right:
		s.combatArrowAlpha, s.combatArrowColumn = 1, 3
	case input.Down:
		s.combatArrowAlpha, s.combatArrowColumn = 1, 4
	case input.Left:
		s.combatArrowAlpha, s.combatArrowColumn = 1, 5
	}
	ctx.DebugTR("arrow: %.2f, %d", s.combatArrowAlpha, s.combatArrowColumn)
	if s.combatArrowAlpha > 0 {
		atlas.GetTilesheetSprite("combat/menu/skill_arrows", s.combatArrowColumn, 1).DrawColorMask(target, centerMatrix, colors.Alpha(s.combatArrowAlpha))
	}
}

func (s *State) simpleSkillContent(text string) *textbox.Content {
	c, exist := s.cachedContents[text]
	if !exist {
		c = skillText.NewComplexContent(text)
		s.cachedContents[text] = c
	}
	return c
}

type TickHandler func(ctx *game.Context, s *State, tick int, source Combatant, target Combatant) []rpg.DamageResult

type SkillInstance struct {
	Skill       *rpg.Skill
	NextTick    int
	Duration    int // 1+
	OnTick      TickHandler
	DisplayType func(int) rpg.TickDisplayType
}

func (i *SkillInstance) String() string {
	if i == nil {
		return "<none>"
	}
	return fmt.Sprintf("SkillInstance{NextTick: %d, Duration: %d}", i.NextTick, i.Duration)
}

func newInstance(skillId rpg.SkillId) *SkillInstance {
	skill := skillId.Get()
	var handlers []TickHandler
	for _, tick := range skill.Ticks {
		handlers = append(handlers, func(ctx *game.Context, s *State, tickId int, source Combatant, target Combatant) []rpg.DamageResult {
			sourceStats := source.GetStats()
			targetStats := target.GetStats()
			var allDamage []rpg.DamageResult
			for _, effect := range tick.Effects {
				if effect.StaticDamage != nil {
					result := rpg.ComputeDamage(rpg.DamageSource{
						BaseDamage:     effect.StaticDamage.Amount,
						Affinities:     sourceStats.Affinities,
						DamageMedium:   effect.StaticDamage.Medium,
						SkillType:      skill.Type,
						PhysicalAttack: sourceStats.PhysicalAttack,
						AetherAttack:   sourceStats.AetherAttack,
					}, rpg.DamageTarget{
						TargetType:      targetStats.BodyType,
						Affinities:      targetStats.Affinities,
						PhysicalDefence: targetStats.PhysicalDefense,
						AetherDefence:   targetStats.AetherDefense,
					})
					allDamage = append(allDamage, result)
					outcome := target.ApplyDamage(result)
					if outcome.Perished {
						ctx.Notify("%s DIED!", target.Name())
					}
				}
			}
			return allDamage
		})
	}
	return &SkillInstance{
		Skill:    &skill,
		Duration: len(handlers) - 1,
		OnTick: func(ctx *game.Context, s *State, tickId int, source Combatant, target Combatant) []rpg.DamageResult {
			return handlers[tickId](ctx, s, tickId, source, target)
		},
		DisplayType: func(tickId int) rpg.TickDisplayType {
			return skill.Ticks[tickId].DisplayType
		},
	}
}
