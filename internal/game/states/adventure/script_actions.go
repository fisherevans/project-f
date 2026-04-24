package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/highlighter"
)

func init() {
	RegisterScriptAction("open_xenolog", func(s *State, source EntityReader, params map[string]any) {
		s.openXenolog()
	})

	RegisterScriptAction("save_game", func(s *State, source EntityReader, params map[string]any) {
		if err := game.CurrentSave().Save(); err != nil {
			s.ExecuteSystemEffectsInOrder(
				NewSelfDialogueEffect("oh no."),
				NewSelfDialogueEffect(err.Error()),
			)
			return
		}
		s.ExecuteSystemEffectsInOrder(
			NewSelfDialogueEffect("SAVING... DON'T TURN OFF THE POWER."),
			NewSelfDialogueEffect("You saved the game."),
		)
	})

	RegisterScriptAction("animech_level_description", func(s *State, source EntityReader, params map[string]any) {
		level := game.CurrentSave().Animech.Upgrades.GetLevel()
		msg := fmt.Sprintf("It's a shiny, level %d Animech.", level)
		s.ExecuteSystemEffects(NewSelfDialogueEffect(msg))
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
			effects = append(effects, NewSetWorldStateEffect(key, nil))
		}
		for _, key := range s.globals.baseGlobals.KeysWithPrefix("combat.training.") {
			effects = append(effects, NewSetWorldStateEffect(key, nil))
		}
		game.CurrentSave().Animech.SkillSet = &rpg.SkillSet{
			Skill1: rpg.Skill_Tackle.Id,
			Skill2: rpg.Skill_Guard.Id,
		}
		game.CurrentSave().ControlledUnlockedSkills = map[rpg.SkillId]struct{}{
			rpg.Skill_Tackle.Id: {},
			rpg.Skill_Guard.Id:  {},
		}
		game.CurrentSave().Animech.Upgrades.ShieldLevel = 0
		game.CurrentSave().Animech.Upgrades.SyncLevel = 0
		game.CurrentSave().Animech.Experience = 100
		for _, p := range game.CurrentSave().Primortals {
			p.ResearchPoints = 0
		}
		game.CurrentSave().Primortals[rpg.Primortal_Toxmidge.Type] = &rpg.PrimortalProgress{
			Visibility:     rpg.PrimortalVisibilityDefeated,
			ResearchPoints: 5,
		}
		effects = append(effects, NewSetWorldStateEffect(globalVariableNameHasXenologAccess, nil))
		effects = append(effects, NewSetWorldStateEffect(globalVariableNameHasXenologPrimortals, nil))
		effects = append(effects, NewSetWorldStateEffect(rpg.GlobalKeyElythium, 0))
		effects = append(effects, NewSetWorldStateEffect(entityId, game.InstanceId))
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
		fmt.Println("resetting!")
		var effects []Effect
		for _, key := range s.globals.baseGlobals.KeysWithPrefix("intro.") {
			effects = append(effects, NewSetWorldStateEffect(key, nil))
		}
		effects = append(effects, NewSetWorldStateEffect(rpg.GlobalKeyElythium, 0))
		effects = append(effects, NewSetWorldStateEffect(entityId, game.InstanceId))
		s.ExecuteSystemEffects(effects...)
	})

	RegisterScriptAction("clear_globals_prefix", func(s *State, source EntityReader, params map[string]any) {
		prefix, _ := params["prefix"].(string)
		if prefix == "" {
			return
		}
		var effects []Effect
		for _, key := range s.globals.baseGlobals.KeysWithPrefix(prefix) {
			effects = append(effects, NewSetWorldStateEffect(key, nil))
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
			Skill1: rpg.Skill_Tackle.Id,
			Skill2: rpg.Skill_Guard.Id,
		}
		triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
			WithCombatId("intro.training.4.combat_over").
			WithOpponent(game.CombatOpponent{
				Type:      rpg.Primortal_Dummy.Type,
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
			Skill1: rpg.Skill_Tackle.Id,
			Skill2: rpg.Skill_Guard.Id,
		}
		triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
			WithCombatId("intro.training.6.combat_over").
			WithOpponent(game.CombatOpponent{
				Type:      rpg.Primortal_Toxmidge.Type,
				Archetype: "training.2",
			}).
			WithReward(game.CombatReward{
				ExperiencePoints: 25,
				ResearchType:     rpg.Primortal_Toxmidge.Type,
				ResearchPoints:   4,
			}).
			WithPlayer(combatPlayer).
			WithTrainingSequence("training.2")
		s.ExecuteSystemEffects(triggerCombat)
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

	RegisterScriptAction("indexed_self_dialogue", func(s *State, source EntityReader, params map[string]any) {
		counterKey, _ := params["counter_key"].(string)
		messagesRaw, _ := params["messages"].([]any)
		var messages []string
		for _, m := range messagesRaw {
			if str, ok := m.(string); ok {
				messages = append(messages, str)
			}
		}
		if len(messages) == 0 || counterKey == "" {
			return
		}
		attempts := s.globals.Get(counterKey).AsInt(0)
		message := messages[attempts%len(messages)]
		s.ExecuteSystemEffectsInOrder(
			NewSelfDialogueEffect(message),
			NewSetRunStateEffect(counterKey, attempts+1),
		)
	})

	RegisterScriptAction("pick_up_paper", func(s *State, source EntityReader, params map[string]any) {
		attemptsKey, _ := params["attempts_key"].(string)
		attempts := s.globals.Get(attemptsKey).AsInt(0)
		var message string
		if attempts == 0 {
			message = "Okay, I've got my papers. I need to give them to the instructor."
		} else if attempts < 3 {
			message = "Oh good, this one's actually mine. That shouldn't have been so hard."
		} else {
			message = "Ah, the elusive papers that are actually mine."
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		s.ExecuteSystemEffectsInOrder(
			NewSelfDialogueEffect(message),
			NewDeleteEntityEffect(entityId),
			NewSetRunStateEffect("intro.has_papers", true),
		)
	})

	RegisterScriptAction("turn_in_papers", func(s *State, source EntityReader, params map[string]any) {
		attemptsKey, _ := params["attempts_key"].(string)
		attempts := s.globals.Get(attemptsKey).AsInt(0)
		message := "Ah, thank you. Now go find the equipment specialist for your Animech training."
		if attempts >= 2 {
			message = "What were you doing snooping around everyone else's paperwork? I almost wrote you up. Anyways, go find the equipment specialist for your Animech training."
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		playerId := s.globals.Get(globalVariableNamePlayerId).AsString("")
		builder := NewFocusedSequenceBuilder(entityId, playerId).
			WithMiddleEffects(
				NewDialogueEffect(message),
				NewMutateFollowCameraEffect().WithFollowEntityId("intro.papers_door"),
				NewTimerEffect(1.5),
				NewSetRunStateEffect("intro.papers_turned_in", true),
				NewTimerEffect(1.5),
				NewMutateFollowCameraEffect().WithFollowEntityId(playerId),
			)
		s.ExecuteSystemEffectsInOrder(builder.BuildEffects()...)
	})

	RegisterScriptAction("door_sync_blocking", func(s *State, source EntityReader, params map[string]any) {
		variable, _ := params["variable"].(string)
		if variable == "" {
			return
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		stateValue := doorRunStateValue(s.globals, variable)
		s.ExecuteSystemEffects(
			NewMutateBlockingPresenceEffect(entityId).WithIsBlockingIngress(stateValue == doorClosed),
			NewMutateModeBasedEntityEffect(entityId).WithMode(stateValue),
		)
	})

	RegisterScriptAction("show_elythium_highlight", func(s *State, source EntityReader, params map[string]any) {
		elythiumHighlight := func(msg string) highlighter.Target {
			return highlighter.NewTarget(util.R(200, 135, 40, 24)).
				WithMessage(highlighter.NewMessage(msg, highlighter.MessageOnLeft).Wrapped(140)).
				WithBadge(highlighter.NewBadge(highlighter.BadgeOnBottomMiddle).WithLabel("Okay"))
		}
		s.ExecuteSystemEffects(
			NewSetHighlightSequenceEffect([]highlighter.Target{
				elythiumHighlight("This gauge tells you how much Elythium you've collected."),
				elythiumHighlight("Once it's full, you're ready to transfer back to your body."),
			}),
		)
	})
}

func doorRunStateValue(globals StateGlobalsReader, variable string) string {
	v := globals.Get(variable)
	if !v.Exists() {
		return doorClosed
	}
	if b, ok := v.Value().(bool); ok {
		if b {
			return doorOpen
		}
		return doorClosed
	}
	return v.AsString(doorClosed)
}
