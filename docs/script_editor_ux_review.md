# Script Editor UX Review

From the perspective of a game designer building cutscenes, NPC encounters,
and entity interactions.

---

## Workflow analysis

### 1. Creating a new NPC with escalating dialogue

Goal: Add a shopkeeper NPC that greets the player differently based on visit
count, offers a focused dialogue sequence, and reacts to a global quest state.

**Steps taken:**

1. Open the script file, click + to add a handler named `hq.shopkeeper`.
2. Add `on_interact_self` hook via the hook picker - good, the hook picker
   shows descriptions so I know which one handles player-initiated interaction.
3. Add a rule, add a condition (`global_gt` on a visit counter), add steps.
4. For the dialogue itself, I want a `focused_sequence` with effects. I click
   "Add step," search "focused" - the picker finds it, description tells me
   what it does. This works.
5. I need to put dialogue steps *inside* the focused_sequence. It auto-expands
   and shows an "effects" sub-step list. I add `dialogue` steps there.
6. For the escalating part, I add a second rule with a different condition and
   different dialogue. Rules currently always run first-match-wins, but the UI
   doesn't indicate this. I can't even reorder the rules.

**Friction points:**

- Rule evaluation order is implicit. No visual indicator of first-match-wins
  behavior, and no way to reorder rules.
- To create "first visit" vs "repeat visit" logic, I need two rules with
  mirrored conditions. No validation that conditions are exhaustive or
  non-overlapping. A `return` step would let me write this as a flat
  sequence instead of mirrored conditions across rules.

### 2. Setting up a key-card door

Goal: Create a door entity that blocks passage until the player uses a key
card at a slot, then opens with sound and animation.

**Steps taken:**

1. Look at `shared/common.yaml` for the existing `door.run_state_based`
   handler. It references custom actions like `door_sync_blocking` and
   `door_configure_shield`.
2. Click on the handler. I see the `on_init` hook with a `switch` step and
   several `custom_action` calls.
3. I want to understand what `door_sync_blocking` does. The step shows
   `custom_action` badge and the name, but **I cannot click through to see
   the action's steps**. I have to manually find it in the left panel,
   losing my place in the handler.
4. For the key slot handler, I look at `intro.equipment_door.key_slot` in
   hallway.yaml. It calls `equipment_key_slot_interact` with a `with` param.
   The `with` map editor shows raw key-value pairs with **no hint what each
   param means or which are required vs optional**.

**Friction points:**

- No inline introspection of referenced custom actions.
- The `with` map is raw key-value pairs - it should be a structured form
  driven by the action's declared params, showing descriptions, types,
  required/optional status, and defaults.
- Cross-file references are invisible. No way to see where an action is
  defined or jump to it.

### 3. Building an escalating guard encounter

Goal: Replicate the `intro.guarded_entry` pattern - a guard that pushes the
player back with increasingly exasperated dialogue.

**Steps taken:**

1. Open `hallway.yaml`, click on `intro.guarded_entry`. Props section shows
   `zone_id` and `walk_back_direction` - useful, documents the interface.
   Tiled Entities section shows which map objects use it - genuinely helpful.
2. The `on_zone_activity` rule calls `guarded_entry_deny`. To trace the
   escalating dialogue: click custom action in left panel, see it calls
   `focused_sequence` with `pick_dialogue` referencing the const list
   `guard_deny_messages`. Click the const to see the 20 messages.
3. Three separate context switches to trace one gameplay behavior.

**Friction points:**

- Tracing logic flow requires multiple panel switches (handler -> custom
  action -> const list). Each switch loses context.
- The collapsed `custom_action` step doesn't show the action's description.
- Const references in steps like `pick_dialogue` are just strings - no
  link, no preview of the list contents.

### 4. Debugging deeply nested if/switch/while

Goal: Understand `equipment_key_slot_interact` which has 3 levels of nested
`if` statements.

**Steps taken:**

1. Click on the custom action. See a top-level `if` step.
2. Auto-expands. `else` branch contains another `if`, which contains another
   `if` in its `else`. Three levels deep.
3. Indentation compresses available width at each level. Expression inputs
   at the innermost level are very narrow.
4. No collapse-all / expand-all, no summary view of branching logic.
5. This entire structure could be flat if we had a `return` step - check
   condition, do thing, return. Check next condition, do thing, return.
   The current nesting exists because there's no way to break out early.

**Friction points:**

- Deep nesting is a consequence of missing `return`. With `return`, this
  would be a flat list of guarded blocks.
- The `if` step's collapsed summary shows nothing - the `when` condition
  (the most important piece) is missing.
- No way to see the decision tree in aggregate.

---

## What works well

**Hook picker with descriptions.** Grouped picker with descriptions for each
hook type. A designer knows what `on_zone_activity` vs `on_broadcast` means
without checking docs.

**Expression editor.** CodeMirror with syntax highlighting, autocomplete for
`var.*` and `const.*` keys, and live server-side validation. The expression
help modal is thorough.

**Step kind picker.** Categorized, searchable, with descriptions and param
hints. A designer can discover available steps without memorizing the schema.

**Color-coded step categories.** Left-border accent colors on step cards
(blue for text, violet for flow, amber for state, teal for entity) provide
visual scanning of what a step sequence does at a glance.

**Props and Tiled cross-reference.** The Properties section documenting
expected Tiled properties, and the Tiled Entities section showing which map
objects reference this handler with their actual property values - exactly the
information a designer needs to understand a handler's contract and usage.

**Inline helper text.** Small gray descriptions scattered throughout teach
usage without separate documentation.

**Schema-driven smart inputs.** Entity ref inputs, zone ID pickers, sound
pickers, and global key inputs with autocomplete for their respective domains.

**Diff viewer.** Proper unified diff of unsaved changes builds confidence
when editing complex scripts.

---

## Problems and gaps

### P1: No inline introspection of custom actions or sequences

When a handler step calls `custom_action: escalating_denial`, the only
information visible is the action name. The action's description, parameters,
and steps are completely opaque without navigating away. This is the biggest
friction point because custom actions are the primary abstraction mechanism.

Cross-references need to be introspectable without losing your place. A
designer often has unsaved changes in a handler, so navigating away risks
losing context. The right approach is a read-only modal or popover showing
the action's description, params, and steps - with a button to "open in
another tab" or select it in the left panel if it's in the same file.

For const references (in `pick_dialogue`, `pick_chatter`, expressions), a
similar preview would help - show the const's value inline or on hover. In
collapsed step summaries, something like `list = guard_deny_messages (20
items)` gives enough context.

### P2: Collapsed step summaries are near-useless

The `getStepSummary` function returns the params value for simple strings,
or nothing meaningful for complex steps:

- `if:` shows nothing - the `when` condition is the key info
- `switch:` shows nothing - the `on` expression and case count are missing
- `set_run_state:` / `set_var:` shows nothing - key and value are missing
- `focused_sequence:` shows nothing - target info is missing
- `custom_action:` doesn't show the action description
- `pick_dialogue:` / `pick_chatter:` doesn't show the list name

A designer should be able to scan 10 collapsed steps and understand the flow
without expanding anything.

### P3: Rule evaluation model is rigid and invisible

Rules currently always use first-match-wins semantics, which the UI doesn't
communicate. Two changes would help:

1. **Add a toggle per hook**: "run first match only" (current behavior) vs
   "run all matching rules". This gives designers flexibility per use case.
2. **Add rule reordering**: up/down move buttons matching the step reorder UX.

When first-match-only is selected, the UI should visually indicate evaluation
order. When run-all is selected, order still matters for execution sequence
but the semantics are different.

This pairs with the `return` step proposal (see Improvements section) - with
`return`, a single rule with flat steps can replace multiple rules with
mirrored conditions.

### P4: Custom action params are raw key-value pairs

The `with` map in `custom_action` steps is a raw key-value editor with no
connection to the action's declared params. Problems:

- No descriptions, types, or required/optional indicators from the action def
- No validation that keys match declared params
- No suggested fields for params not yet specified
- Extra undeclared params can't be added (for pass-through scenarios)

This should be a structured form driven by the action's param definitions,
with expression inputs for each value, and an option to add extra custom
params beyond the declared ones.

Also: `with` is a weird name. `params` would be clearer and more consistent
with the action definition itself which uses `params:`.

### P5: No way to trace cross-references without losing context

Custom actions and consts can be referenced across files. There's no way to:

- Click from a `custom_action: X` step to the action definition
- Click from a `const.list_name` reference to the const
- See which handlers call a given custom action (reverse reference)
- See which handlers use a given const

The solution must preserve unsaved state. A read-only modal/popover showing
the referenced item (with a "jump to" button that navigates) is safer than
inline navigation that loses the current handler.

### P6: Large message lists are tedious to edit

The const list editor for `guard_deny_messages` (20 entries) shows each entry
as a single-line input. Problems:

- No bulk-add (paste multiple lines to create entries)
- No export/copy-all for moving lists between files
- No expand/fullscreen for long text entries
- Const lists (arrays) lack drag-to-reorder that data lists have

Also: data lists should be deprecated in favor of consts. The UI should guide
designers toward consts and eventually remove the data list section. The
`getDataList()` fallback already unifies them at runtime.

### P7: No search

No way to search within a script file (handler names, step kinds, expression
text, string literals) or globally across all script files. With large scripts
having 14+ handlers, finding a specific handler or tracing a global key across
files requires manual scanning.

Both in-file search (filter the left panel, highlight matches in the detail
view) and global search (search across all script files, link to results)
would be valuable.

### P8: Condition types should be consolidated

The current condition system has 8 global_* comparators (global_eq, global_ne,
global_gt, global_gte, global_lt, global_lte, global_exists, global_not_exists)
plus 2 handler_state comparators. All of these are replaceable by the `expr`
condition:

- `global_eq: {intro.has_papers: true}` -> `expr: "global['intro.has_papers'] == true"`
- `global_gt: {key: count, value: 5}` -> `expr: "int(global.count) > 5"`
- `handler_state_eq: {blabbed: true}` -> `expr: "var.blabbed == true"`

Keep `expr` as the primary condition type and game-coupled conditions that
can't be expressed as simple expressions (`player_in_zone`, `skill_equipped`,
`entity_not_moving`, `has_face_behavior`, `not_run_this_instance`). Deprecate
the generic comparators - they add UI complexity without adding capability.

### P9: `set_state` on rules should be a step, not a special attribute

The `set_state` section in rules mutates handler state as a side effect of a
rule matching. It's positioned between the condition and steps with no
explanation of timing. This should just be a `set_var` step at the beginning
of the steps list - it's the same operation, already exists as a step kind,
and is more explicit about when it runs relative to other steps.

Remove `set_state` from the rule schema and migrate existing uses to
`set_var` steps.

---

## Proposed improvements

### Engine changes (prerequisites for UI improvements)

**E1. Add `return` step** [M - engine + UI]
Add a `return` step that exits the current handler's step processing,
similar to an early return in a function. This eliminates the need for
deeply nested if/else chains - instead of nesting, write flat guarded
blocks:

```yaml
steps:
  - if:
      when: "global['door_state'] == 'open'"
      then:
        - self_dialogue: "The door is already open."
        - return: true
  - if:
      when: "global['has_key'] != true"
      then:
        - self_dialogue: "You need a key..."
        - return: true
  - play_sound: "adventure/beeps/success"
  - self_dialogue: "That worked!"
```

Why: The deeply nested if/else pattern in `equipment_key_slot_interact` and
similar handlers exists solely because there's no early exit. `return` makes
logic flat and readable. This also reduces the need for multiple rules with
mirrored conditions.

**E2. Configurable rule evaluation mode** [M - engine + UI]
Add a per-hook toggle: "first match only" (current behavior) vs "run all
matching". Store as a flag on the hook definition in YAML.

Why: Some hooks genuinely want all-matching-rules (e.g., multiple
independent zone_activity triggers). Others want first-match priority (e.g.,
interact handlers with fallback). The current forced first-match means
designers work around it with condition gymnastics.

**E3. Rename `with` to `params` in custom_action steps** [S - engine]
Rename the `with` key to `params` for consistency with custom action
definitions. Keep `with` as a deprecated alias during migration.

Why: The action declares `params:`, so the invocation should use `params:`
too. `with` is an unnecessary indirection.

**E4. Deprecate generic condition comparators** [M - engine + UI]
Deprecate `global_eq`, `global_ne`, `global_gt`, `global_gte`, `global_lt`,
`global_lte`, `global_exists`, `global_not_exists`, `handler_state_eq`,
`handler_state_ne` in favor of `expr`. Keep game-coupled conditions
(`player_in_zone`, `skill_equipped`, `entity_not_moving`,
`has_face_behavior`, `not_run_this_instance`).

Why: The `expr` condition handles all comparison logic more flexibly.
Removing 10 condition types simplifies the condition builder UI and reduces
designer confusion about which to use.

**E5. Remove `set_state` from rules, use `set_var` steps** [S - engine]
Remove the `set_state` attribute from rule definitions. Migrate existing
uses to `set_var` steps at the start of the steps list.

Why: `set_state` is a confusing special case that does the same thing as
`set_var` but with unclear timing semantics.

### UI improvements - high priority

**U1. Inline custom action preview** [M]
When a `custom_action` step references an action, show the description as
a subtitle. On click/hover, open a read-only modal showing params (with
descriptions, types, required/optional) and steps. Include a "Jump to" button
that selects the action in the left panel (same file) or opens the file
(cross-file). The modal must not discard unsaved changes.

For const references in steps, show resolved value hints in collapsed
summaries (e.g., `list = guard_deny_messages (20 items)`).

**U2. Better collapsed step summaries** [S]
Extend `getStepSummary` for common steps:
- `if:` - show the `when` expression (truncated)
- `switch:` - show `on` expression + case count
- `set_run_state:` / `set_world_state:` - show `key = value`
- `set_var:` - show `key = value`
- `focused_sequence:` - show `target`
- `custom_action:` - show action `name` + description subtitle
- `pick_dialogue:` / `pick_chatter:` - show `list` name + item count
- `teleport_player:` - show `to_entity`

**U3. Rule reordering** [S]
Add up/down move buttons to rules. Show evaluation mode indicator per hook
(pairs with E2).

**U4. Structured custom action params** [M]
Replace the raw `with`/`params` key-value editor with a structured form
driven by the action's param definitions:
- Show each declared param with its description, type, and required/optional
- Use `ExpressionInput` for values (not plain text inputs)
- Pre-populate required params with empty fields
- Show defaults for optional params as placeholders
- Allow adding extra custom params beyond declared ones
- Warn on unrecognized param keys

### UI improvements - medium priority

**U5. Cross-reference modals** [M]
Read-only preview modals for custom actions, consts, and sequences
referenced from steps. Must not lose unsaved state. Include "Open in
editor" button to navigate.

**U6. Search** [M]
Two levels:
- In-file: filter input on the left panel that searches handler names,
  custom action names, const names, sequence names
- Global: search across all script files for handler names, expressions,
  string literals, global keys. Results link directly to the item.

**U7. Bulk text operations for lists** [S]
Add to const list editor (arrays):
- "Bulk add" textarea mode - paste multiple lines, each becomes an entry
- "Copy all" button for exporting list contents
- Drag-to-reorder (matching data list behavior)
- Expand/fullscreen button for long text entries

**U8. Deprecate data lists in UI** [S]
Add a deprecation notice to the Data Lists section: "Use Constants instead.
Data lists and consts are unified at runtime." Eventually hide the section
when a file has no data lists.

### UI improvements - lower priority

**U9. Condition builder cleanup** [S]
After E4, default new conditions to `expr`. Show game-coupled conditions
(`player_in_zone`, etc.) as a separate category. Remove deprecated
comparators from the picker or move to a "legacy" section.

**U10. Expand-all / collapse-all** [S]
Toggle at the top of step lists to expand or collapse all steps.

**U11. Reverse reference panel** [M]
On custom actions: "Used by" section listing which handlers call this
action. On consts: "Referenced by" listing which handlers/actions use it.
Similar to the Tiled Entities cross-reference on handlers.

---

## Summary

The script editor's foundation is solid - expression editing, step kind
picker, schema-driven inputs, and Tiled cross-reference are genuine
strengths. The problems fall into two categories:

**Engine gaps** that force awkward patterns: no `return` step creates deep
nesting, forced first-match-wins rules limit flexibility, redundant condition
types add noise, `set_state` on rules is a confusing special case. These
should be addressed first because they simplify both the YAML authoring and
the UI that renders it.

**Information hiding** in the UI: collapsed steps show nothing, custom action
references are opaque, cross-references require losing context, params are
raw key-value. The top UI fixes (U1-U4) address the most frequent pain
points.

Recommended implementation order:
1. E1 (`return` step) + U2 (better summaries) - immediate readability wins
2. E3 (rename `with`) + U4 (structured params) - fix the action invocation UX
3. U1 (inline preview) + U5 (cross-reference modals) - fix the navigation UX
4. E2 (rule modes) + U3 (rule reordering) - fix the rule authoring UX
5. E4 + E5 + U9 (condition/set_state cleanup) - simplification pass
6. U6 + U7 + U8 (search, bulk ops, deprecation) - quality of life
