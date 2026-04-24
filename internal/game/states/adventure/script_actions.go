package adventure

import (
	"fmt"
	"math/rand"
	"strings"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
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
		entityId, _ := params["entity_id"].(string)
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

	RegisterScriptAction("guarded_entry_deny", func(s *State, source EntityReader, params map[string]any) {
		messages := []string{
			"Sorry, restricted zone. You'll need clearance to pass.",
			"Hold up - your ID's not on the list.",
			"Access denied. You know the rules.",
			"You've been flagged for repeated entry attempts.",
			"Central already notified us about you.",
			"You're persistent, I'll give you that.",
			"Security's keeping an eye on your route.",
			"Still not authorized. Try the proper channels.",
			"You're on record, just so you know.",
			"Another attempt? You're starting to stand out.",
			"Command says to remind you this area is sealed.",
			"We've got your ID tagged facility-wide now.",
			"You must really like closed doors.",
			`Everyone's been briefed about your "tours."`,
			"You're not going to find a guard who says yes.",
			"You're pushing your luck here.",
			"They told us to expect you.",
			"Next stop's a security report if you keep at it.",
			"You've officially become a training example.",
			"You're done here. Move along.",
		}
		attemptsKey := "intro.do_not_enter.attempts"
		attempts := s.globals.Get(attemptsKey).AsInt(0)
		message := messages[min(attempts, len(messages)-1)]
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		playerId := s.globals.Get(globalVariableNamePlayerId).AsString("")
		walkBackDir, _ := params["walk_back_direction"].(string)
		builder := NewFocusedSequenceBuilder(entityId, playerId).
			WithMoveCamera(true).
			WithFacePlayer(true).
			WithMiddleEffects(NewDialogueEffect(message)).
			WithPostEffects(
				NewTriggerMovementEffect(playerId).WithDirection(input.DirectionFromString(walkBackDir)),
				NewSetRunStateEffect(attemptsKey, attempts+1),
			)
		s.ExecuteSystemEffectsInOrder(builder.BuildEffects()...)
	})

	RegisterScriptAction("guarded_entry_chat", func(s *State, source EntityReader, params map[string]any) {
		dialogues := []string{
			"Everything in order, recruit? Stay clear of restricted zones.",
			"You again? Don't wander too close to the barriers.",
			"I've seen your name on a few access logs. Keep it clean.",
			"You're getting a reputation for curiosity. That's not a compliment.",
			"You've tested every guard in this facility. Maybe take a hint.",
		}
		attemptsKey := "intro.do_not_enter.attempts"
		attempts := s.globals.Get(attemptsKey).AsInt(0)
		message := dialogues[min(attempts/len(dialogues), len(dialogues)-1)]
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		playerId := s.globals.Get(globalVariableNamePlayerId).AsString("")
		builder := NewFocusedSequenceBuilder(entityId, playerId).
			WithMiddleEffects(NewDialogueEffect(message))
		s.ExecuteSystemEffectsInOrder(builder.BuildEffects()...)
	})

	RegisterScriptAction("hall_npc_random_chatter", func(s *State, source EntityReader, params map[string]any) {
		messages := util.StringList{
			"I respect curiosity. I just don't have time for it right now.",
			"If it's not urgent, take it to your supervisor, not me.",
			"You're not cleared for curiosity yet, kid. Stick to your assignments.",
			"Whatever you're thinking, don't. Trust me, it's not worth the paperwork.",
			"You're asking the wrong person, and probably at the wrong time.",
			"Don't worry, we were all clueless once. Some of us still are.",
			"Curiosity's fine. Just don't make it my problem.",
			"You're eager. That'll fade after the first audit.",
			"Ask your mentor before you start poking at things.",
			"Whatever you heard, forget it. You'll sleep better.",
			"That question sounds expensive. I'd drop it.",
			"Keep that energy. The system loves optimism - at first.",
			"You're not in trouble yet. Emphasis on yet.",
			"The less you know, the longer you last here.",
			"If you value sanity, stay out of the details.",
			"First day? You'll stop asking by the third.",
			"Everyone starts curious. Then the reports pile up.",
			"Save the questions for someone still enthusiastic.",
			"Don't look at me like I have answers.",
			"I admire initiative. I also ignore it.",
			"You'll learn fast. Or not at all.",
			"Be careful what you notice. It notices back.",
			"Advice? Stop asking for advice.",
			"The rules change faster than explanations.",
			"Curiosity's contagious. Try not to spread it.",
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		s.ExecuteSystemEffectsInOrder(
			NewPushEntityBehaviorEffect(entityId).WithFacingEntityId(s.globals.Get(globalVariableNamePlayerId).AsString("")),
			NewChatterEffect(entityId, 3, messages.Random()),
			NewPopEntityBehaviorEffect(entityId),
		)
	})

	RegisterScriptAction("hall_npc_start_random_motion", func(s *State, source EntityReader, params map[string]any) {
		targets := []string{
			"intro.hall_way_end.west",
			"intro.hall_way_end.north",
			"intro.hall_way_end.east",
		}
		entityId := ""
		if source != nil {
			entityId = source.GetId()
		}
		target := targets[rand.Intn(len(targets))]
		wasCanceled, _ := params["was_canceled"].(bool)
		if wasCanceled {
			s.ExecuteSystemEffectsInOrder(
				NewTimerEffect(rand.Float64()*3),
				NewStartScriptedMotionEffect(entityId).WithToEntityId(target),
			)
		} else {
			s.ExecuteSystemEffects(NewStartScriptedMotionEffect(entityId).WithToEntityId(target))
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

	RegisterScriptAction("equipment_key_slot_interact", func(s *State, source EntityReader, params map[string]any) {
		variable, _ := params["run_state_key"].(string)
		hasEquipmentKey := s.globals.Get("intro.has_equipment_key").AsBool(false)
		if variable == "" || variable == "not_it" {
			if hasEquipmentKey {
				s.ExecuteSystemEffectsInOrder(
					NewPlaySoundEffect("adventure/beeps/error"),
					NewSelfDialogueEffect("Yeah, that key card doesn't work in this slot..."),
				)
			} else {
				s.ExecuteSystemEffects(NewSelfDialogueEffect("This looks like a slot for a key card..."))
			}
			return
		}
		equipmentDoorState := s.globals.Get("intro.equipment_door_state").AsString(doorClosed)
		if equipmentDoorState == doorOpen {
			s.ExecuteSystemEffects(NewSelfDialogueEffect("The door is already open."))
			return
		}
		if !hasEquipmentKey {
			s.ExecuteSystemEffects(NewSelfDialogueEffect("This looks like a slot for a key card..."))
			return
		}
		s.ExecuteSystemEffectsInOrder(
			NewPlaySoundEffect("adventure/beeps/success"),
			NewSelfDialogueEffect("That worked!"),
			NewSetRunStateEffect(variable, doorOpen),
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
