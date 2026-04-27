package adventure

import (
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/overlays"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	RegisterScriptAction("open_xenolog", func(s *State, source EntityReader, params map[string]any) {
		s.openXenolog()
	})

	RegisterScriptAction("save_game", func(s *State, source EntityReader, params map[string]any) {
		if err := game.CurrentSave().Save(); err != nil {
			s.ExecuteSystemEffects(NewSelfDialogueEffect(err.Error()))
		}
	})

	RegisterScriptAction("reset_training_state", func(s *State, source EntityReader, params map[string]any) {
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		if s.globals.Get(entityId).AsString("") == game.InstanceId {
			return
		}
		var effects []Effect
		for _, key := range s.globals.baseGlobals.KeysWithPrefix("intro.training.") {
			effects = append(effects, NewSetGlobalEffect(key, nil))
		}
		for _, key := range s.globals.baseGlobals.KeysWithPrefix("combat.training.") {
			effects = append(effects, NewSetGlobalEffect(key, nil))
		}
		game.CurrentSave().Animech.SkillSet = &rpg.SkillSet{
			Skill1: "tackle",
			Skill2: "guard",
		}
		game.CurrentSave().ControlledUnlockedSkills = map[rpg.SkillId]struct{}{
			"tackle": {},
			"guard":  {},
		}
		game.CurrentSave().Animech.Upgrades.ShieldLevel = 0
		game.CurrentSave().Animech.Upgrades.SyncLevel = 0
		game.CurrentSave().Animech.Experience = 100
		for _, p := range game.CurrentSave().Primortals {
			p.ResearchPoints = 0
		}
		game.CurrentSave().Primortals["toxmidge"] = &rpg.PrimortalProgress{
			Visibility:     rpg.PrimortalVisibilityDefeated,
			ResearchPoints: 5,
		}
		effects = append(effects, NewSetGlobalEffect(globalVariableNameHasXenologAccess, nil))
		effects = append(effects, NewSetGlobalEffect(globalVariableNameHasXenologPrimortals, nil))
		effects = append(effects, NewSetGlobalEffect(rpg.GlobalKeyElythium, 0))
		effects = append(effects, NewSetGlobalEffect(entityId, game.InstanceId))
		s.ExecuteSystemEffects(effects...)
	})

	RegisterScriptAction("reset_intro_state", func(s *State, source EntityReader, params map[string]any) {
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		if s.globals.Get(entityId).AsString("") == game.InstanceId {
			return
		}
		var effects []Effect
		for _, key := range s.globals.baseGlobals.KeysWithPrefix("intro.") {
			effects = append(effects, NewSetGlobalEffect(key, nil))
		}
		effects = append(effects, NewSetGlobalEffect(rpg.GlobalKeyElythium, 0))
		effects = append(effects, NewSetGlobalEffect(entityId, game.InstanceId))
		s.ExecuteSystemEffects(effects...)
	})

	RegisterScriptAction("clear_globals_prefix", func(s *State, source EntityReader, params map[string]any) {
		prefix, _ := params["prefix"].(string)
		if prefix == "" {
			return
		}
		var effects []Effect
		for _, key := range s.globals.baseGlobals.KeysWithPrefix(prefix) {
			effects = append(effects, NewSetGlobalEffect(key, nil))
		}
		if len(effects) > 0 {
			s.ExecuteSystemEffects(effects...)
		}
	})

	RegisterScriptAction("open_computer", func(s *State, source EntityReader, params map[string]any) {
		game.SetActiveStateIntent(game.ComputerIntent{
			Background: s,
		})
	})

	RegisterScriptAction("trigger_training_combat_1", func(s *State, source EntityReader, params map[string]any) {
		combatPlayer := game.NewCombatPlayer(game.CurrentSave().Animech)
		combatPlayer.SkillSet = &rpg.SkillSet{
			Skill1: "tackle",
			Skill2: "guard",
		}
		triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
			WithCombatId("intro.training.4.combat_over").
			WithOpponent(game.CombatOpponent{
				Type:      "dummy",
				Archetype: "training.1",
			}).
			WithPlayer(combatPlayer).
			WithReward(game.CombatReward{
				ExperiencePoints: 25,
			}).
			WithTrainingSequence("training.1")
		s.ExecuteSystemEffects(triggerCombat)
	})

	RegisterScriptAction("trigger_training_combat_2", func(s *State, source EntityReader, params map[string]any) {
		combatPlayer := game.NewCombatPlayer(game.CurrentSave().Animech)
		combatPlayer.SkillSet = &rpg.SkillSet{
			Skill1: "tackle",
			Skill2: "guard",
		}
		triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
			WithCombatId("intro.training.6.combat_over").
			WithOpponent(game.CombatOpponent{
				Type:      "toxmidge",
				Archetype: "training.2",
			}).
			WithReward(game.CombatReward{
				ExperiencePoints: 25,
				ResearchType:     "toxmidge",
				ResearchPoints:   4,
			}).
			WithPlayer(combatPlayer).
			WithTrainingSequence("training.2")
		s.ExecuteSystemEffects(triggerCombat)
	})

	RegisterScriptAction("random_chatter_interact", func(s *State, source EntityReader, params map[string]any) {
		chattersRaw, _ := params["chatters"].(string)
		if chattersRaw == "" {
			return
		}
		chatters := util.StringList(strings.Split(chattersRaw, "\n"))
		duration := 4.0
		if d, ok := params["duration"].(float64); ok {
			duration = d
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		if !IsBehaviorType[*NPCBehavior](source) {
			return
		}
		playerId := s.globals.Get(globalVariableNamePlayerId).AsString("")
		s.ExecuteSystemEffectsInOrder(
			NewPushEntityBehaviorEffect(entityId).WithFacingEntityId(playerId),
			NewChatterEffect(entityId, duration, chatters.Random()),
			NewPopEntityBehaviorEffect(entityId),
		)
	})

	RegisterScriptAction("random_dialogue_interact", func(s *State, source EntityReader, params map[string]any) {
		dialoguesRaw, _ := params["dialogues"].(string)
		if dialoguesRaw == "" {
			return
		}
		dialogues := util.StringList(strings.Split(dialoguesRaw, "\n"))
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		playerId := s.globals.Get(globalVariableNamePlayerId).AsString("")
		s.ExecuteSystemEffectsInOrder(
			NewPushEntityBehaviorEffect(entityId).WithFacingEntityId(playerId),
			NewDialogueEffect(dialogues.Random()),
			NewPopEntityBehaviorEffect(entityId),
		)
	})

	RegisterScriptAction("show_elythium_highlight", func(s *State, source EntityReader, params map[string]any) {
		flow, ok := overlays.GetFlow("elythium_intro")
		if !ok {
			return
		}
		targets, err := overlays.ResolveTargets(flow)
		if err != nil {
			return
		}
		s.ExecuteSystemEffects(NewSetHighlightSequenceEffect(targets))
	})

	RegisterScriptCondition("skill_equipped", func(params map[string]any) ConditionCheck {
		skillId, _ := params["skill"].(string)
		return func(s *State, td float64) bool {
			ss := game.CurrentSave().Animech.SkillSet
			return ss.Skill1 == rpg.SkillId(skillId) ||
				ss.Skill2 == rpg.SkillId(skillId) ||
				ss.Skill3 == rpg.SkillId(skillId) ||
				ss.Skill4 == rpg.SkillId(skillId)
		}
	})

	upgradeLevelGtFactory := func(params map[string]any) ConditionCheck {
		level := 0
		if l, ok := params["level"].(int); ok {
			level = l
		} else if l, ok := params["level"].(float64); ok {
			level = int(l)
		}
		return func(s *State, td float64) bool {
			return game.CurrentSave().Animech.Upgrades.GetLevel() > level
		}
	}
	RegisterScriptCondition("animech_upgrade_level_gt", upgradeLevelGtFactory)
	RegisterScriptCondition("upgrade_level_gt", upgradeLevelGtFactory)

	RegisterScriptCondition("entity_not_moving", func(params map[string]any) ConditionCheck {
		entityId, _ := params["entity"].(string)
		return func(s *State, td float64) bool {
			e, ok := s.entities.GetEntity(entityId)
			if !ok {
				return true
			}
			return !e.IsMoving()
		}
	})

	RegisterScriptCondition("player_in_zone", func(params map[string]any) ConditionCheck {
		zone, _ := params["zone"].(string)
		return func(s *State, td float64) bool {
			return s.globals.GetZonesAt(s.globals.Player().GetLocation()).Contains(zone)
		}
	})

	RegisterScriptCondition("not_run_this_instance", func(params map[string]any) ConditionCheck {
		key, _ := params["key"].(string)
		return func(s *State, td float64) bool {
			return s.globals.Get(key).AsString("") != game.InstanceId
		}
	})

	RegisterScriptCondition("has_face_behavior", func(params map[string]any) ConditionCheck {
		entityId, _ := params["entity"].(string)
		return func(s *State, td float64) bool {
			e, ok := s.entities.GetEntity(entityId)
			if !ok {
				return false
			}
			return IsBehaviorType[*FaceEntityBehavior](e)
		}
	})
}
