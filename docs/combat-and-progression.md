# Combat & Progression

### Combat Modes

When fighting within an Animech, there are two combat modes:

- **Void Mode** - when no Animetric Core is loaded

  - This mode is primarily focused on capturing Specimen
  - The Animech has no direct benefits from a Primortal.
  - This mode is focused on capturing specimen using the XenoCrypt.
  - Damage is dealt to your Animech's Shield (SHLD) then Integrity (INTG).

- **Imprinted Mode** - when an Animetric Core is loaded

  - This mode is primarily focused on combat and defeating enemies
  - A single Primortal is Imprinted onto the Animech, based on which Animetric Core is loaded.
  - Damage is dealt to your Animech's Shield (SHLD), then the Primortal Synchronization (SYNC), then your Animech's Integrity (INTG).

  The "primary focus" of these modes is mostly characterized by the types of skills available when building your load-out. The mechanisms while battling in each mode feel very similar. The only real difference is the lack of extra health when in Void Mode.

### Combat Actions

When in battle, these options available to the player:

- Fight
- Swap
- When in a wild Specimen encounter:
  - Capture
  - Escape
- (There might be the ability to use items, but at this time, I'm not sure what those would be)

#### Fight

When in the Fight Menu, the player has 4 skills available to them, based on the loadout they configured for the Void Mode or loaded Animetric Core. Combat isn't truly real-time - but skills can be pre-selected while prior actions are still resolving. If you have a skill pre-selected when the last action completes, you will build up a "Tempo Combo". This combo increases your base stats based on set thresholds and affects certain skills.

Skills from both sides are executed on a standard tick-based time system. Each tack lasts some short period of time. Each skill has a predefined timing. The total duration of a skill is always at least 1 tick. Some straw-man examples:

- A simple damage skill might:
  - trigger damage immediately
  - sleep 1 tick
- A wind-up skill might:
  - sleep for 2 ticks
  - deal damage at the end of those 2 ticks
- A barage-type skill might:
  - trigger damage
  - sleep 1 tick
  - trigger damage
  - sleep 1 tick
  - trigger damage again

The ticks remaining for your current skill and the opponents current skill is visible to the player allowing you to react to their moves.

The "shape" of the opponents next skill may also be visible, indicating what you might need to expect. Your next skill, if selected, will also be shown in the tick display.

If two actions, one from each combat participant, would trigger an effect at the same time - the longer running skill's effect is triggered first - existing ties are resolved by the level of the player and opponent.

#### Swap

xxx

#### Capture

xxx

#### Escape

xxx

## Effects

Effects can be applied to a creature:

- i.e. burn, poison, sleep - tbd



