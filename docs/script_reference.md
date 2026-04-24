# Script System Reference

Auto-generated from `internal/schema/script_schema.json`. Do not edit manually.

## Step Kinds

Each step in a handler's rule is a single-key YAML map: `{kind: params}`.

### audio

#### `play_sound`

Play a sound effect

- **Param style:** string
**Parameters:**

- `sound` (string (required)): Sound resource path (e.g. 'adventure/beeps/success')

### camera

#### `mutate_camera`

Modify the current camera's follow target or position without pushing/popping

- **Param style:** map
**Parameters:**

- `follow_entity` (string): New entity ID for the camera to follow
- `reset_position` (bool): If true, snap camera to the follow target's current position

#### `override_camera`

Push a new camera override that follows a specific entity

- **Param style:** map
**Parameters:**

- `follow` (map (required)): Camera follow configuration with 'entity' (string) and optional 'reset_position' (bool)

#### `pop_camera`

Remove the current camera override and restore the previous camera

- **Param style:** string_or_map
**Parameters:**

- `maintain_location` (bool, default: false): If true, the restored camera starts from the overridden camera's current position. Can be provided as a bare bool value.

### combat

#### `trigger_combat`

Initiate a combat encounter

- **Param style:** map
**Parameters:**

- `combat_id` (string): Identifier for the combat (used in on_combat_complete filter)
- `background` (string (required)): Combat background scene identifier
- `opponent_type` (string (required)): Primortal type of the opponent
- `opponent_archetype` (string): Combat archetype for the opponent
- `training_sequence` (string): Training sequence ID (for tutorial combats)

### entity

#### `change_player_renderer`

Change the player entity's visual style/renderer

- **Param style:** string
**Parameters:**

- `style` (string (required)): Renderer style name

#### `delete_entity`

Remove an entity from the map

- **Param style:** string
**Parameters:**

- `entity_id` (string (required)): ID of the entity to delete

#### `disable_behavior`

Disable an entity's current behavior with a named lock

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `by` (string (required)): Lock name for later re-enabling

#### `enable_behavior`

Re-enable an entity's behavior by releasing a named lock

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `by` (string (required)): Lock name to release (must match the disable_behavior lock)

#### `face_direction`

Make an entity face a cardinal direction

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID to reorient
- `direction` (string, values: [up, down, left, right]): Direction to face

#### `face_entity`

Make an entity turn to face another entity

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID to reorient
- `target` (string (required)): Entity ID to face toward

#### `pop_behavior`

Pop the top behavior from an entity's behavior stack

- **Param style:** string
**Parameters:**

- `entity_id` (string (required)): Entity ID to pop behavior from

#### `push_behavior`

Push a new behavior onto an entity's behavior stack

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `scripted_motion` (bool): Enable scripted motion behavior
- `active_player_zone` (string): Zone that activates scripted motion (requires scripted_motion: true)
- `facing_entity` (string): Entity ID to face continuously

#### `reset_animation`

Reset a ModeBasedEntity's animation to the beginning of its current mode

- **Param style:** string
**Parameters:**

- `entity_id` (string (required)): Entity ID to reset animation for

#### `reset_movement`

Reset an entity's movement state to idle

- **Param style:** string
**Parameters:**

- `entity_id` (string (required)): Entity ID to reset

#### `scripted_motion`

Start pathfinding-based scripted motion toward a target entity

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID to move
- `to_entity` (string): Target entity ID to move toward

#### `set_blocking`

Set whether an entity blocks player ingress

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `is_blocking` (bool (required)): Whether the entity should block movement

#### `set_mode`

Set the visual mode of a ModeBasedEntity (changes which animation/sprite it displays)

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `mode` (string (required)): Mode name to set

#### `trigger_movement`

Trigger one tick of movement on an entity in a direction

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID
- `direction` (string, values: [up, down, left, right]): Direction to move

### flow

#### `action`

Invoke a registered named Go action. Can be a bare string (action name) or a map with 'name' and additional params.

- **Param style:** string_or_map
**Parameters:**

- `name` (string (required)): Registered action name. Can be provided as a bare string value.
- `...` (any): Additional params are passed through to the action function

#### `broadcast`

Send a broadcast event that other handlers can listen for

- **Param style:** map
**Parameters:**

- `id` (string (required)): Broadcast identifier
- `data` (any): Optional data payload

#### `focused_sequence`

A scripted interaction sequence that disables player controls, optionally moves the camera, faces entities, runs effects, then restores everything. Replicates FocusedSequenceBuilder.

- **Param style:** map
- **Accepts sub-steps:** yes
**Parameters:**

- `entity` (string): The entity performing the sequence (defaults to {{self}})
- `target` (string): The target entity (defaults to {{player}})
- `move_camera` (bool, default: false): Whether to pan the camera to the focused entity
- `face_player` (bool, default: false): Whether the entity should face the player during the sequence
- `pre_effects` (steps): Steps to execute before the main effects
- `effects` (steps (required)): Main steps of the focused sequence
- `post_effects` (steps): Steps to execute after the main effects, before restoring state

#### `parallel`

Execute a list of sub-steps concurrently instead of sequentially

- **Param style:** list
- **Accepts sub-steps:** yes
**Parameters:**

- `steps` (steps (required)): List of steps to run in parallel

#### `ref`

Include a named sequence inline with optional parameter substitution

- **Param style:** map
**Parameters:**

- `name` (string (required)): Name of the sequence to reference
- `with` (map): Parameter values to pass to the referenced sequence

#### `timer`

Wait for a duration before continuing the sequence

- **Param style:** string_or_map
**Parameters:**

- `duration` (number (required)): Seconds to wait. Can be provided as a bare number value.
- `id` (string): Optional timer ID for later reference via on_timer_complete

#### `wait_for`

Pause the sequence until a named condition becomes true (checked each frame)

- **Param style:** map
**Parameters:**

- `condition` (string (required)): Name of a registered script condition to wait for
- `...` (any): Additional params are passed through to the condition factory

#### `wait_for_animation`

Pause the sequence until an entity's current animation completes

- **Param style:** string
**Parameters:**

- `entity_id` (string (required)): Entity ID whose animation to wait for

### rpg

#### `yield_elythium`

Award Elythium currency to the player

- **Param style:** string_or_map
**Parameters:**

- `amount` (number (required)): Amount of Elythium to award. Can be provided as a bare number value.

### state

#### `set_run_state`

Set a run state variable (persists only during current game session)

- **Param style:** map
**Parameters:**

- `key` (string (required)): Run state variable key
- `value` (any (required)): Value to set

#### `set_world_state`

Set a persistent world state global variable (persists across save/load)

- **Param style:** map
**Parameters:**

- `key` (string (required)): Global variable key
- `value` (any (required)): Value to set (null to delete)

### text

#### `chatter`

Show floating chatter text above an entity for a duration

- **Param style:** map
**Parameters:**

- `entity` (string (required)): Entity ID to show chatter above
- `message` (string (required)): The chatter text
- `duration` (number, default: 3): Duration in seconds to show chatter

#### `dialogue`

Display dialogue text from the handler's entity speaker

- **Param style:** string
**Parameters:**

- `text` (string (required)): The dialogue text to display

#### `highlight_sequence`

Display a series of UI highlights with messages, used for tutorials

- **Param style:** map
**Parameters:**

- `targets` (any (required)): List of highlight target definitions (region, message, badge)

#### `self_dialogue`

Display dialogue text as the player character's inner monologue

- **Param style:** string
**Parameters:**

- `text` (string (required)): The self-dialogue text to display

#### `tooltip`

Push a tooltip message to the screen

- **Param style:** string
**Parameters:**

- `message` (string (required)): The tooltip text

### transition

#### `deactivate_fade`

Deactivate a running fade effect by its ID

- **Param style:** string
**Parameters:**

- `fade_id` (string (required)): ID of the fade to deactivate

#### `fade`

Apply a screen fade effect transitioning between colors

- **Param style:** map
**Parameters:**

- `duration` (number, default: 0.33): Fade duration in seconds
- `transitions` (number, default: 1): Number of color transitions
- `from_color` (string): Starting color (hex string)
- `to_color` (string): Ending color (hex string)
- `auto_deactivate` (bool): Whether the fade automatically deactivates when complete

#### `load_map`

Load a different map, optionally spawning at a waypoint

- **Param style:** map
**Parameters:**

- `map` (string (required)): Map name to load
- `waypoint` (string): Waypoint name to spawn at in the new map

#### `teleport_player`

Teleport the player to a reference point or entity location

- **Param style:** map
- **Accepts sub-steps:** yes
**Parameters:**

- `to_reference` (string): Map reference point name to teleport to
- `to_entity` (string): Entity ID to teleport to (alternative to to_reference)
- `transition_style` (string): Visual transition style
- `interstitial` (steps): Steps to execute during the teleport transition (between fade out and fade in)

## Named Actions

Invoked via `action: name` or `action: {name: ..., param: value}` in YAML steps.

### `animech_level_description`

Show a self-dialogue describing the player's current Animech level


### `clear_globals_prefix`

Delete all global variables matching a key prefix

**Parameters:**

- `prefix` (string (required)): Prefix to match (e.g. 'intro.' deletes all intro globals)

### `door_sync_blocking`

Synchronize a door entity's blocking presence and visual mode based on a run state variable. Sets blocking to true when closed, false when open.

**Parameters:**

- `variable` (string (required)): The run state key that controls the door's open/closed state

### `equipment_key_slot_interact`

Handle interaction with a key card slot. Checks if the player has a key card and whether the door is already open.

**Parameters:**

- `run_state_key` (string (required)): Run state key controlling the equipment door state. Use 'not_it' for decoy slots.

### `guarded_entry_chat`

Chat with a guard NPC when interacted with directly (not trying to pass). Shows messages that escalate with the player's entry denial count.


### `guarded_entry_deny`

Deny the player entry to a restricted zone. Shows an escalating series of messages and pushes the player back.

**Parameters:**

- `walk_back_direction` (string (required), values: [up, down, left, right]): Direction to push the player back

### `hall_npc_random_chatter`

NPC faces the player and shows a random chatter message from a hardcoded list of dismissive/busy responses


### `hall_npc_start_random_motion`

Start random pathfinding motion to one of the predefined hallway endpoints. If the previous motion was canceled, adds a random delay before moving.

**Parameters:**

- `was_canceled` (bool): Whether the previous motion was canceled (passed automatically by motion_complete filter)

### `indexed_self_dialogue`

Cycle through a list of self-dialogue messages based on a counter variable. Each invocation shows the next message and increments the counter.

**Parameters:**

- `counter_key` (string (required)): Run state key used to track the message index
- `messages` (any (required)): YAML list of message strings to cycle through

### `open_computer`

Open the research computer interface


### `open_xenolog`

Open the Xenolog (creature log) interface


### `pick_up_paper`

Pick up a paper entity. Shows context-aware dialogue based on how many wrong papers were picked up, deletes the entity, and sets the 'intro.has_papers' flag.

**Parameters:**

- `attempts_key` (string (required)): Run state key tracking wrong-paper pickup attempts
- `entity_id` (string (required)): Entity ID to delete after pickup

### `random_chatter_interact`

Show a random chatter message when an NPC is interacted with. Requires the entity to have NPCBehavior. Messages are passed as a newline-separated string from entity properties.

**Parameters:**

- `chatters` (string (required)): Newline-separated list of chatter messages
- `duration` (number, default: 4): Duration to show chatter in seconds

### `random_dialogue_interact`

Show a random dialogue when an NPC is interacted with. Messages are passed as a newline-separated string from entity properties.

**Parameters:**

- `dialogues` (string (required)): Newline-separated list of dialogue messages

### `reset_intro_state`

Reset all intro-related global variables and currency. Idempotent per game instance.


### `reset_training_state`

Reset all training-related global variables, player skills, upgrade levels, primortal progress, and currency. Idempotent per game instance.


### `save_game`

Save the current game to disk. Displays success/error feedback to the player.


### `show_elythium_highlight`

Display a tutorial highlight sequence explaining the Elythium gauge UI element


### `trigger_training_combat_1`

Initiate the first training combat against a Dummy opponent. Awards 25 XP. Uses combat_id 'intro.training.4.combat_over'.


### `trigger_training_combat_2`

Initiate the second training combat against a Toxmidge opponent. Awards 25 XP and 4 research points. Uses combat_id 'intro.training.6.combat_over'.


### `turn_in_papers`

Turn in papers to the instructor. Shows context-aware dialogue, pans camera to the papers door, and sets the 'intro.papers_turned_in' flag.

**Parameters:**

- `attempts_key` (string (required)): Run state key tracking wrong-paper pickup attempts (determines dialogue tone)

## Named Conditions

Used in `wait_for` steps or `when` clauses via `check` delegation.

### `animech_upgrade_level_gt`

Check if the player's Animech upgrade level is greater than a threshold

**Parameters:**

- `level` (number (required)): Level threshold to compare against

### `entity_not_moving`

Check if an entity is currently stationary

**Parameters:**

- `entity` (string (required)): Entity ID to check

### `has_face_behavior`

Check if an entity currently has a FaceEntityBehavior on its behavior stack

**Parameters:**

- `entity` (string (required)): Entity ID to check

### `not_run_this_instance`

Check if a global key does not match the current game instance ID. Used for one-time-per-session initialization.

**Parameters:**

- `key` (string (required)): Global key to check against the current instance ID

### `player_in_zone`

Check if the player is currently standing in a named zone

**Parameters:**

- `zone` (string (required)): Zone ID to check

### `skill_equipped`

Check if a skill is equipped in any of the player's skill slots

**Parameters:**

- `skill` (string (required)): Skill ID to check for

### `upgrade_level_gt`

Alias for animech_upgrade_level_gt

**Parameters:**

- `level` (number (required)): Level threshold to compare against

## Built-in Conditions

Used in `when` clauses as single-key maps: `{condition_type: params}`.

### `all`

Logical AND - all sub-conditions must be true

*Composite condition - takes sub-conditions as params.*

**Parameters:**

- `conditions` (condition (required)): List of condition nodes that must all evaluate to true

### `any`

Logical OR - at least one sub-condition must be true

*Composite condition - takes sub-conditions as params.*

**Parameters:**

- `conditions` (condition (required)): List of condition nodes where at least one must evaluate to true

### `global_eq`

Check if a global variable equals a value. Supports both {key: k, value: v} and shorthand {some_key: some_value} syntax.

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (any (required)): Expected value (compared as strings)

### `global_exists`

Check if a global variable exists (has been set to any value)

**Parameters:**

- `key` (string (required)): Global variable key to check

### `global_gt`

Check if a global variable is numerically greater than a value

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (number (required)): Numeric threshold

### `global_gte`

Check if a global variable is numerically greater than or equal to a value

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (number (required)): Numeric threshold

### `global_lt`

Check if a global variable is numerically less than a value

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (number (required)): Numeric threshold

### `global_lte`

Check if a global variable is numerically less than or equal to a value

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (number (required)): Numeric threshold

### `global_ne`

Check if a global variable does not equal a value

**Parameters:**

- `key` (string (required)): Global variable key
- `value` (any (required)): Value to compare against

### `global_not_exists`

Check if a global variable does not exist (has never been set or was deleted)

**Parameters:**

- `key` (string (required)): Global variable key to check

### `handler_state_eq`

Check if a handler state variable equals a value. Handler state is per-handler, set via set_state in rules.

**Parameters:**

- `key` (string (required)): Handler state variable key
- `value` (any (required)): Expected value

### `handler_state_ne`

Check if a handler state variable does not equal a value

**Parameters:**

- `key` (string (required)): Handler state variable key
- `value` (any (required)): Value to compare against

### `not`

Logical NOT - negate a single sub-condition

*Composite condition - takes sub-conditions as params.*

**Parameters:**

- `condition` (condition (required)): Single condition node to negate

## Event Hooks

Each handler can define rules under these YAML keys. Rules fire when the corresponding event occurs.

### `on_broadcast`

Fires when a broadcast event is sent

- **Event type:** EventBroadcast
- **Filter fields:**
  - `id` (string): Broadcast ID to filter on

### `on_combat_complete`

Fires when a combat encounter finishes

- **Event type:** EventCombatComplete
- **Filter fields:**
  - `combat_id` (string): Combat ID to filter on

### `on_global_updated`

Fires when a global variable is updated

- **Event type:** EventGlobalVariableUpdated
- **Filter fields:**
  - `key` (string): Global variable key to filter on

### `on_init`

Fires when the handler is first initialized (map load, entity creation)

- **Event type:** init

### `on_interact`

Fires when any entity is interacted with (global listener)

- **Event type:** EventOnInteract

### `on_interact_self`

Fires when the handler's own entity is the target of an interaction

- **Event type:** EventOnInteract

### `on_motion_complete`

Fires when any entity's scripted motion completes (global listener)

- **Event type:** EventScriptedMotionComplete
- **Filter fields:**
  - `entity` (string): Entity ID to filter on
  - `motion_id` (string): Motion ID to filter on
  - `was_canceled` (bool): Filter by whether the motion was canceled

### `on_motion_complete_self`

Fires when the handler's own entity's scripted motion completes

- **Event type:** EventScriptedMotionComplete

### `on_state_enter`

Fires when the adventure state is entered (e.g. returning from combat or menu)

- **Event type:** EventOnStateEnter

### `on_timer_complete`

Fires when a timer effect completes

- **Event type:** EventTimerComplete
- **Filter fields:**
  - `created_by` (string): Entity that created the timer
  - `timer_id` (string): Timer ID to filter on

### `on_zone_activity`

Fires when an entity enters or leaves a named zone

- **Event type:** EventEntityZoneActivity
- **Filter fields:**
  - `zone` (string): Zone ID to filter on
  - `entity` (string): Entity ID to filter on
  - `entering` (bool): Filter by entering (true) or leaving (false)

## Template Variables

Available in all string values via `{{variable}}` syntax.

- `{{self}}` - ID of the entity that owns this handler
- `{{player}}` - ID of the player entity
- `{{source}}` - ID of the entity that triggered the event (e.g. the interacting entity)
- `{{instance_id}}` - Unique ID for the current game session (changes on restart)
- `{{prop.*}}` - Entity property value from the Tiled map (e.g. {{prop.chatters}} reads the 'chatters' custom property)
- `{{<param>}}` - Custom parameter from a sequence ref's 'with' map or a sequence's declared params

