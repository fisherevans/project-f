# Adventure State

- Camera
  - change targets
- Interactions
  - User interaction (open door, read sign)
  - Movement interaction
    - push item
    - ice
    - teleport
    - one way paths
    - conveyer belts
    - tethered movement
    - recorded movement
    - weight triggers
    - delayed effect (like a door opening 5 seconds after pressing a button)
    - holes/collapsable areas
  - damaging effects, like poison gas or heat 
    - shield/integrity display when entering these
  - darkness/limited view
- Dialogue
  - box slide up/down animation
  - text bubbles for players
  - pop ups for notifications/system details
  - glossary words are highlighted, pressing select pulls up the glossary entry
  - text effects
    - jitter for scared
    - text vs. info
    - variable speed
    - glitch texts (void magic, or computer malfucntioning, etc)
  - dsl for rendering
    - infix for styles, speed, colors, glossary words
  - voice sounds
  - dialogue record for key events
  - skip scenes with summary
- "Cut scenes"
  - programmable actions
    - handle multiple entities moving
      - might require path finding, or require multiple linear paths
      - trigger on target hit destination, or time
      - invisible cameras that aren't tied to entities
- NPC logic
  - tasks/jobs/movement that can be programmed
    - "wandering"
  - following another entity
  - identify line of sight to trigger interactions
- running vs walking

# Combat State

- tick based resolution of skills
- current/next skill visualization
- damage calculator
- state manager (for health)
- tempo 
- log of actions

# RPG Engine

- skill progression
- loadouts
- 