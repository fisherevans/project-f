package adventure

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerEventHandler("intro.training.enter.listener", func(*util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if event.ZoneId != "intro.training.enter" || globals.Get(thisEntity.GetId()).AsString("") == game.InstanceId {
					return nil
				}
				return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
					for _, key := range s.globals.baseGlobals.KeysWithPrefix("intro.training.") {
						s.globals.baseGlobals.Delete(key)
					}
					game.CurrentSave().Animech.SkillSet = &rpg.SkillSet{
						Skill1: rpg.Skill_Jab.Id,
					}
					game.CurrentSave().ControlledUnlockedSkills = map[rpg.SkillId]struct{}{
						rpg.Skill_Jab.Id: {},
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
					s.globals.baseGlobals.Set(thisEntity.GetId(), game.InstanceId)
					s.globals.baseGlobals.Delete(globalVariableNameHasXenologAccess)
					s.globals.baseGlobals.Delete(globalVariableNameHasXenologPrimortals)
					s.globals.baseGlobals.Set(rpg.GlobalKeyElythium, 0)
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
	registerEventHandler("intro.training.1.chair", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != thisEntity.GetId() {
					return nil
				}
				if globals.Get(globalVariableNameHasXenologPrimortals).AsBool(false) {
					return NewOutput().WithEffects(NewDialogueEffect("Uhhh. I don't think I need to do that again..."))
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("Hesitantly, you climb into the chair..."),
					NewPlaySoundEffect("adventure/beeps/success"),
					NewDialogueEffect("Without any warning, you feel it. Your mind slips backwards. Not quite falling, but... floating?"),
					NewTeleportPlayerEffect().
						WithToEntityId("intro.training.1.destination").
						WithInterstitialEffects([]Effect{
							NewChangePlayerRendererEffect("animech"),
							NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Right),
						}),
					NewMutateEntityBehaviorEffect(event.SourceId).WithDisableBy(thisEntity.GetId()),
					NewTimerEffect(1),
					NewDialogueEffect("There's someone looking at you. But not. Their eyes are open, but they're looking through you."),
					NewDialogueEffect("They're your eyes. Your looking at yourself."),
					NewMutateEntityBehaviorEffect(event.SourceId).WithEnableBy(thisEntity.GetId()),
				)
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
						NewDialogueEffect("Well, at least your can move around in that Animech."),
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
		return BasicHandlerBuilder[None]{
			CombatComplete: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventCombatComplete) *HandlerOutput {
				if event.CombatId != "intro.training.4" {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("INTERCOM: Nice job recruit. Most folks curl up in a ball when hey face their first foe."),
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
				if globals.Get(globalVariableNameHasXenologAccess).AsBool(false) {
					return NewOutput().WithSerialPlan(
						NewDialogueEffect("Don't chicken out now."),
					)
				}
				return NewFocusedSequenceBuilder(thisEntity.GetId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Now that your first battle is under your belt, you've gained some experience. With it, we can enhance the synchronization between your soul and the Animech."),
						NewDialogueEffect("On the table next to me is Xenolog. A device that allows you to track your explorations and see the status of your Animech."),
						NewDialogueEffect("Try opening it up, and spend those experience points to enhance your Animech's abilities."),
						NewSetWorldStateEffect(globalVariableNameHasXenologAccess, true),
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
						NewSetWorldStateEffect(globalVariableNameHasXenologAccess, true),
					).
					BuildEffects()
				effects = append(effects, NewFocusedSequenceBuilder("intro.training.5.door", playerId).
					WithMiddleEffects(
						NewSetWorldStateEffect(keyDoor, true),
					).
					BuildEffects()...)
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
				return NewOutput().WithEffects(NewFunctionEffect(func(s *State) {
					s.openXenolog()
				}))
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.6.listener", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			CombatComplete: func(thisEntity EntityReader, globals StateGlobalsReader, state None, event *EventCombatComplete) *HandlerOutput {
				if event.CombatId != "intro.training.6" {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("INTERCOM: Not bad, kid.. not bad."),
					NewSetRunStateEffect(doorKey(6), true),
				)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.training.7.npc", func(_ *util.Properties) EventHandler {
		keyDoor := doorKey(7)
		keyElythium := "intro.training.7.elythium"
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
				doneKey := "intro.training.7.elythium.done"
				if event.Id != "intro.training.7.elythium" && globals.Get(doneKey).AsBool(false) {
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
				need := 4
				if have < need {
					return NewOutput().WithEffects(NewDialogueEffect(fmt.Sprintf("The screen readout says I need to hae %d elythium, but I only have %d.", need, have)))
				}
				return NewOutput().WithSerialPlan(
					NewDialogueEffect("You deposit the elythium and press the button..."),
					NewDialogueEffect("At first nothing, but you notice you can't move. Not an inch..."),
					NewPlaySoundEffect("adventure/beeps/success"),
					NewDialogueEffect("Suddenly, your vision swims. You feel... liquid? Somewhere between melting and floating?"),
					NewDialogueEffect("................."),
					NewTeleportPlayerEffect().
						WithToEntityId("intro.training.1.return").
						WithInterstitialEffects([]Effect{
							NewChangePlayerRendererEffect("human"),
							NewEntityFaceDirectionEffect(event.SourceId).WithDirection(input.Up),
						}),
					NewDialogueEffect("You stare at yourself, across the gap."),
					NewDialogueEffect("But it isn't you."),
				)
			},
		}.CreateHandler()
	})
}
