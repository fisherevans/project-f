package adventure

// All intro hallway handlers migrated to assets/scripts/intro/hallway.yaml
// Named actions (reset_intro_state, indexed_self_dialogue, pick_up_paper,
// turn_in_papers, guarded_entry_deny, guarded_entry_chat,
// hall_npc_random_chatter, hall_npc_start_random_motion,
// equipment_key_slot_interact) remain in script_actions.go

func init() {
	registerPropertyTemplate("intro.hallway_npc", map[string]any{
		"script_ref": "intro.hall_way_npc",
		"movement":   "static",
	})
}
