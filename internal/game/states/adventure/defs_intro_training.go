package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

func init() {
	keyElythium := "intro.control.elythium"
	registerEventHandler("intro.training.enter.listener", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId != "intro.training.enter" || globals.Get(thisEntity.GetId()).AsString("") == game.InstanceId {
					return nil
				}
				return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
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
					game.CurrentSave().Animech.AnimechExperience = 100
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
					effects = append(effects, NewSetWorldStateEffect(thisEntity.GetId(), game.InstanceId))
					s.ExecuteSystemEffects(effects...)
				}))
			},
		}.CreateHandler()
	})
	doorKey := func(x int) string {
		return fmt.Sprintf("intro.training.%d.door", x)
	}
	registerEventHandler("intro.training.0.npc", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				key := doorKey(0)
				if globals.Get(key).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("We don't appreciate those who linger. Move on to the next training room, recruit..."))
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Welcome to...argh..! ARGHHRRGG!!! ... {+r}PssszzZZZtttt!{-r} Hello! This is Fisher, sorry for the interruption..."),
						NewDialogueEffect("Thank you for playtesting Primortal! This is a very rough draft of the overall game experience, please excuse the many rough edges. This training segment is extremely condensed just to get you into the core gameplay loop."),
						NewDialogueEffect("Thanks again! Back to actual game content ... {+r}PssszzZZZtttt!{-r} ... your.. err, uhhh. I'm sorry, I don't know what just came over me. Let me start over."),
						NewDialogueEffect("Welcome to your crash course training, recruit. I'm sure you're eager to get started."),
						NewDialogueEffect("This series of training rooms will get you ready to venture out into the unknown to explore."),
						NewTimerEffect(.25),
						NewSetRunStateEffect(key, true),
						NewTimerEffect(.25),
						NewDialogueEffect("Please continue into the next room..."),
					).
					Build()
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.1.npc", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				keyDoor := doorKey(10)
				if globals.Get(keyDoor).AsBool(false) {
					return NewOutput().WithEffects(
						NewDialogueEffect("Don't worry, you'll have plenty of time to practice some more... once you're all alone in deep space."),
						NewDialogueEffect("Now, get through that door."),
					)
				}
				if globals.Get(globalVariableNameHasXenologAccess).AsBool(false) {
					return NewOutput().WithSerialPlan(
						NewFocusedSequenceBuilder(thisEntity.GetId(), globals.Player().GetId()).
							WithMoveCamera(true).
							WithMiddleEffects(
								NewDialogueEffect("You're now ready to explore on your own. Proceed through the final door to begin your adventure."),
								NewMutateFollowCameraEffect().WithFollowEntityId(keyDoor),
								NewTimerEffect(1.5),
								NewSetWorldStateEffect(keyDoor, true),
								NewTimerEffect(1.5),
								NewMutateFollowCameraEffect().WithFollowEntityId(globals.Player().GetId()),
							).
							BuildEffects()...)
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Just take a seat over there. All will be explained..."),
					).
					Build()
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.1.chair", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if globals.Get(globalVariableNameHasXenologPrimortals).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("Uhhh. I don't think I need to do that again..."))
				}
				chairId := "intro.training.1.chair"
				chairFrontId := "intro.training.1.chair_front"
				pId := globals.Player().GetId()
				animechEntityId := "intro.training.1.animech"
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId(chairFrontId),
					NewPopEntityBehaviorEffect(pId),
					NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Down),
					NewDialogueEffect("Hesitantly, you climb into the chair..."),
					NewChangePlayerRendererEffect("hidden"),
					NewMutateModeBasedEntityEffect(chairId).WithMode("enter"),
					NewWaitForConditionEffect(func(s *State, td float64) bool {
						e, r, ok := GetModeBasedRenderer(s, chairId)
						if !ok {
							log.Fatal().Msgf("failed to get renderer for %s", chairId)
						}
						return r.getBasicEntityRenderer(ModeMetadataKey.Get(e)).AreAnimationsComplete()
					}),
					NewPlaySoundEffect("adventure/beeps/success"),
					NewDialogueEffect("Without any warning, you feel it. Your mind slips backwards. Not quite falling, but... floating?"),
					NewTeleportPlayerEffect().
						WithToEntityId(animechEntityId).
						WithInterstitialEffects([]Effect{
							NewChangePlayerRendererEffect("animech"),
							NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Down),
							NewMutateModeBasedEntityEffect(animechEntityId).WithMode("hidden"),
						}),
					NewMutateEntityBehaviorEffect(event.SourceId).WithDisableBy(thisEntity.GetId()),
					NewTimerEffect(1),
					NewDialogueEffect("You open your eyes..... and see yourself. Across the gap."),
					NewEntityFaceDirectionEffect(pId).WithDirection(input.Right),
					NewDialogueEffect("Your limbs feel rigid. But strong."),
					NewDialogueEffect("Your balance feels solid, like its on auto pilot."),
					NewDialogueEffect("You {+u}are{-u} an Animech."),
					NewMutateEntityBehaviorEffect(event.SourceId).WithEnableBy(thisEntity.GetId()),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.2.npc", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Your Animech has a special ability to dash across small gaps."),
						NewDialogueEffect("Walk up to the edge and press A."),
					).
					Build()
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.2.button", func(props *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				key := doorKey(2)
				if globals.Get(key).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("The door has already been opened."))
				}
				return NewOutput().WithSerialPlan(
					NewPlaySoundEffect("adventure/beeps/success"),
					NewSetRunStateEffect(key, true),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.3.npc", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				key := doorKey(3)
				if globals.Get(key).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("There's no going back. You must continue to the next room."))
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Well, at least you can move around in that Animech."),
						NewDialogueEffect("Next, you'll need to know how to battle specimen. Go through to the next room and defeat the little vermin."),
						NewTimerEffect(.25),
						NewSetRunStateEffect(key, true),
						NewTimerEffect(.25),
					).
					Build()
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.4.listener", func(_ *util.Properties) EventHandler {
		combatOver := "intro.training.4.combat_over"
		return BasicHandlerBuilder[None]{
			Broadcast: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventBroadcast) *HandlerOutput {
				if event.Id != "intro.training.4.combat" {
					return nil
				}
				combatPlayer := game.NewCombatPlayer(game.CurrentSave().Animech)
				combatPlayer.SkillSet = &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				}
				triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
					WithCombatId(combatOver).
					WithOpponent(game.CombatOpponent{
						Type:      rpg.Primortal_Dummy.Type,
						Archetype: "training.1",
					}).
					WithPlayer(combatPlayer).
					WithTrainingSequence("training.1")
				return NewOutput().WithEffects(triggerCombat)
			},
			CombatComplete: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventCombatComplete) *HandlerOutput {
				if event.CombatId != combatOver {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("INTERCOM: Nice job recruit. Most folks curl up in a ball when hey face their first foe."),
					NewDialogueEffect("INTERCOM: I don't know why though. That combat dummy can't even attack."),
					NewSetRunStateEffect(doorKey(4), true),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.5.npc", func(_ *util.Properties) EventHandler {
		keyDoor := doorKey(5)
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if globals.Get(keyDoor).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("There's no going back. You must continue to the next room."))
				}
				if globals.Get("intro.training.5.xenolog").AsBool(false) {
					return NewOutput().WithSerialPlan(
						NewDialogueEffect("Don't chicken out now."),
					)
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Now that your first battle is under your belt, you've gained some experience. With it, we can enhance the synchronization between your soul and the Animech."),
						NewDialogueEffect("On the table next to me is a Xenolog. A device that allows you to track your explorations and see the status of your Animech."),
						NewDialogueEffect("Try opening it up, and spend those experience points to enhance your Animech's abilities."),
						NewSetWorldStateEffect("intro.training.5.xenolog", true),
					).
					Build()
			},
			OnStateEnter: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnStateEnter) *HandlerOutput {
				if globals.Get(keyDoor).AsBool(false) {
					return nil
				}
				if game.CurrentSave().Animech.Upgrades.GetLevel() <= 1 {
					return nil
				}
				if !globals.GetZonesAt(globals.Player().GetLocation()).Contains("intro.training.5") {
					return nil
				}
				playerId := globals.Get(globalVariableNamePlayerId).AsString("")
				effects := NewFocusedSequenceBuilder(thisEntity.GetId(), playerId).
					WithMoveCamera(true).
					WithMiddleEffects(
						NewDialogueEffect("Nicely done."),
						NewDialogueEffect("Now, lets try battling a real specimen."),
						NewMutateFollowCameraEffect().WithFollowEntityId(keyDoor),
						NewTimerEffect(1.5),
						NewSetWorldStateEffect(keyDoor, true),
						NewTimerEffect(1.5),
						NewMutateFollowCameraEffect().WithFollowEntityId(globals.Player().GetId()),
					).
					BuildEffects()
				return NewOutput().WithSerialPlan(effects...)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.5.xenolog", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if globals.Get("intro.training.5.xenolog").AsBool(false) {
					return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
						s.openXenolog()
					}))
				}
				return NewOutput().WithSerialPlan(
					NewPlaySoundEffect("adventure/beeps/error"),
					NewChatterEffect(thisEntity.GetId(), 3, "The device is locked."),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.6.listener", func(_ *util.Properties) EventHandler {
		combatOver := "intro.training.6.combat_over"
		return BasicHandlerBuilder[None]{
			Broadcast: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventBroadcast) *HandlerOutput {
				if event.Id != "intro.training.6.combat" {
					return nil
				}
				combatPlayer := game.NewCombatPlayer(game.CurrentSave().Animech)
				combatPlayer.SkillSet = &rpg.SkillSet{
					Skill1: rpg.Skill_Tackle.Id,
					Skill2: rpg.Skill_Guard.Id,
				}
				triggerCombat := NewTriggerCombatEffect(rpg.CombatBGSpaceBase).
					WithCombatId(combatOver).
					WithOpponent(game.CombatOpponent{
						Type:      rpg.Primortal_Toxmidge.Type,
						Archetype: "training.2",
					}).
					WithPlayer(combatPlayer).
					WithTrainingSequence("training.2")
				return NewOutput().WithEffects(triggerCombat)
			},
			CombatComplete: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventCombatComplete) *HandlerOutput {
				if event.CombatId != combatOver {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("INTERCOM: Okay, I'll be honest. I expected you to chicken out of that one."),
					NewSetRunStateEffect(doorKey(6), true),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.7.npc", func(_ *util.Properties) EventHandler {
		keyDoor := doorKey(7)
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if globals.Get(keyDoor).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("There's no getting out of it. Yes, there are creatures out there. You're also in a $371 million dollar military grade robot."))
				}
				if globals.Get(keyElythium).AsBool(false) {
					return NewOutput().WithSerialPlan(
						NewDialogueEffect("Just go smash it with your first."),
					)
				}
				if globals.Get(globalVariableNameHasXenologPrimortals).AsBool(false) {
					return NewOutput().WithSerialPlan(
						NewDialogueEffect("Open up that Xenolog, learn a skill, and then equip it."),
					)
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("As you battled that specimen, your Animech acquired research points."),
						NewDialogueEffect("Using the Xenolog, you can spend those research points to unlock new skills."),
						NewDialogueEffect("Do that, and equip your new skill."),
						NewSetWorldStateEffect("intro.training.5.xenolog", true), // reset, just for testing when we skip that room
						NewSetWorldStateEffect(globalVariableNameHasXenologPrimortals, true),
					).
					Build()
			},
			OnStateEnter: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnStateEnter) *HandlerOutput {
				if globals.Get(keyElythium).AsBool(false) {
					return nil
				}
				if !globals.GetZonesAt(globals.Player().GetLocation()).Contains("intro.training.7") {
					return nil
				}
				skills := 0
				if game.CurrentSave().Animech.SkillSet.Skill1 != "" {
					skills++
				}
				if game.CurrentSave().Animech.SkillSet.Skill2 != "" {
					skills++
				}
				if game.CurrentSave().Animech.SkillSet.Skill3 != "" {
					skills++
				}
				if game.CurrentSave().Animech.SkillSet.Skill4 != "" {
					skills++
				}
				if skills < 2 {
					return NewOutput().WithEffects(NewFocusedSequenceBuilder(thisEntity.GetId(), globals.Player().GetId()).
						WithMoveCamera(true).
						WithMiddleEffects(
							NewDialogueEffect("Make sure your unlock a Toxmidge skill, then equip it in your Animech."),
						).
						BuildEffects()...)
				}
				effects := NewFocusedSequenceBuilder(thisEntity.GetId(), globals.Player().GetId()).
					WithMoveCamera(true).
					WithMiddleEffects(
						NewDialogueEffect("Now. In order to transfer your soul back to your body, you'll need to gather enough Elythium."),
						NewDialogueEffect("Go gather that crystal over there."),
						NewSetRunStateEffect(keyElythium, true),
					).
					BuildEffects()
				return NewOutput().WithSerialPlan(effects...)
			},
			Broadcast: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventBroadcast) *HandlerOutput {
				if event.Id != "intro.training.7.elythium" {
					return nil
				}
				doneKey := "intro.training.7.elythium.done"
				if globals.Get(doneKey).AsBool(false) {
					return nil
				}
				return NewOutput().WithEffects(
					NewSetWorldStateEffect(doneKey, true),
					NewSerialPlan(NewFocusedSequenceBuilder(thisEntity.GetId(), globals.Player().GetId()).
						WithMoveCamera(true).
						WithMiddleEffects(
							NewDialogueEffect("You'll need to collect more than that to complete the soul transfer."),
							NewDialogueEffect("Go below and gather enough to trigger the return sequence."),
							NewSetRunStateEffect(keyDoor, true),
						).
						BuildEffects()...),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.9.exit", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				have := globals.Get(rpg.GlobalKeyElythium).AsInt(0)
				need := 12
				if have < need {
					return NewOutput().WithEffects(NewDialogueEffect(fmt.Sprintf("The screen readout says I need to have %d elythium, but I only have %d.", need, have)))
				}
				pId := globals.Player().GetId()
				chairId := "intro.training.1.chair"
				anmechEntityId := "intro.training.9.animech"
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("You deposit the elythium and press the button..."),
					NewPushEntityBehaviorEffect(pId).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(pId).WithToEntityId(anmechEntityId),
					NewPopEntityBehaviorEffect(pId),
					NewMutateEntityBehaviorEffect(pId).WithDisableBy("exit"),
					NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Down),
					NewTimerEffect(0.25),
					NewMutateModeBasedEntityEffect(anmechEntityId).WithMode("visible"),
					NewChangePlayerRendererEffect("hidden"),
					NewDialogueEffect("Your feet latch into the depressions. Suddenly, you can't move your body. Not an inch..."),
					NewPlaySoundEffect("adventure/beeps/success"),
					NewDialogueEffect("Your vision swims. You feel... liquid. Your mind melting inwards."),
					NewTeleportPlayerEffect().
						WithToEntityId("intro.training.1.return").
						WithInterstitialEffects([]Effect{
							NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Left),
						}),
					NewTimerEffect(1.5),
					NewDialogueEffect("You awake."),
					NewDialogueEffect("It takes enormous effort to climb out of the chair."),
					NewMutateModeBasedEntityEffect(chairId).WithMode("exit"),
					NewWaitForConditionEffect(func(s *State, td float64) bool {
						e, r, ok := GetModeBasedRenderer(s, chairId)
						if !ok {
							log.Fatal().Msgf("failed to get renderer for %s", chairId)
						}
						return r.getBasicEntityRenderer(ModeMetadataKey.Get(e)).AreAnimationsComplete()
					}),
					NewChangePlayerRendererEffect("human"),
					NewMutateModeBasedEntityEffect(chairId).WithMode(""),
					NewTimerEffect(0.5),
					NewDialogueEffect("Your limbs are heavy, sluggish. The light hurts your eyes."),
					NewDialogueEffect("You'r back in your own body."),
					NewSetWorldStateEffect(globalVariableNameHasXenologAccess, true),
					NewSetWorldStateEffect(keyElythium, false),
					NewSetWorldStateEffect(rpg.GlobalKeyElythium, 0),
					NewMutateEntityBehaviorEffect(pId).WithEnableBy("exit"),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.wip.listener", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId != "intro.training.wip" || event.EntityId != globals.Player().GetId() {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("Annnnnnd, that's it for now. Hope to see you soon."),
					NewDialogueEffect("Teleporting you to some other testing bed map..........."),
					NewLoadMapEffect("map1"),
				)
			},
		}.CreateHandler()
	})
}
