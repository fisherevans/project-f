# Types

Skills will always have a **Skill Type**, identifying what kind of damage they deal. eg: Kinetic, Thermal, etc.

Specimen will always have one **Body Type** which align with their physical characteristics. Some specimen might have two. eg: Organic, Metal, etc.

Specimen might have an **Affinity**, which will grant them attack bonuses when using skills of that type, and also grant them defense bonuses when receiving damage of that type. Some items or abilities might grant an Affinity, meaning that a Primortal could have multiple affinities at once.

When an Animech imprints a Primortal, it assumes the body type and affinities of that Specimen.

Body Types and Skills are different (unlike games like Pokemon where the same types are used for both use cases). The one exception is "Abyssal" which is both a Skill Type and Body type.

Primortals will have a correlation of what types of skills they might be able to unlock - but it won't be a rules.

## Body Types

- **Organic**: Squishy, like mammals and reptiles. Life you'd find on Earth.
- **Metal**: Could be a metal exoskeleton, or metal bones.
- **Rock**: Think a ball of rocks, or sentient gravel.
- **Crystal**: Cosmic, strong but brittle.
- **Synthetic**: Robots and plastics.
- **Liquid**: Slug-type things, jelly fish
- **Gas**: Ghost-ish, vapor beings.
- **Abyssal**: Of another dimension, from another plane of reality. Skills of this type are more mechanic and less damage focused.

## Skill Types

- **Kinetic**: Punches, body slams, slices, the like
- **Voltaic**: Electric shocks, plasma beams, and so on
- **Thermal**: Fire, magma, steam
- **Sonic**: Vibrations, sound, disorienting
- **Magnetic**: Control, telekinesis
- **Acidic**: Corrosive, lingering
- **Gamma**: Debuffs over time, not immediate, grows over time
- **Abyssal**: The unknown, magical-esque, metaphysical powers

## Strengths & Weaknesses

When a skill is dealt damage to a body type, the damage is either:


- Strong `x2`: Damage is multiplied by two
- Neutral `x1`: Damage is not effected
- Weak `x0.5`: Damage is halved
- Immune `x0`: Damage is not dealt

These are those relationships (Skill types on the left, Body types on top):

|              | **Organic** | **Metal** | **Rock** | **Crystal** | **Synthetic** | **Liquid** | **Gas** | **Abyssal** |
| :----------: | :---------: | :-------: | :------: | :---------: | :-----------: | :--------: | :-----: | :---------: |
| **Kinetic**  |             |   x0.5    |    x2    |     x2      |      x2       |            |   x0    |    x0.5     |
| **Voltaic**  |     x2      |   x0.5    |    x0    |             |      x2       |     x2     |         |             |
| **Thermal**  |     x2      |    x2     |          |             |               |     x0     |         |    x0.5     |
|  **Sonic**   |    x0.5     |           |    x2    |    x0.5     |               |            |   x2    |             |
| **Magnetic** |    x0.5     |    x2     |          |     x0      |      x2       |            |  x0.5   |     x2      |
|  **Acidic**  |     x2      |    x2     |          |    x0.5     |               |    x0.5    |         |             |
|  **Gamma**   |     x2      |    x0     |   x0.5   |             |     x0.5      |            |   x2    |     x2      |
| **Abyssal**  |             |           |          |             |               |            |         |             |

> Abyssal skills aren't strong or weak against anything since their power comes from "elsewhere" - those skills are more mechanic, like damage based on level, or other non-traditional combat calculations.

### Affinities

If a Primortal has an affinity:

- Attacks dealt of that skill type receive a `x1.5` multiplier
- Damage received of that type get a `x0.667` multiplier.

### Examples

If a Specimen has an affinity, skills of that type get an additional `x1.5` multiplier.

> 100 Voltaic damage to a Metal body type by a specimen with an affinity to Voltaic:
>
> - 100 voltaic damage
> - x1.5 (skill affinity)
> - x0.5 (weak against metal)
> - = 75 damage

If a Specimen has two body type, the multipliers are multiplied together.

>100 Thermal damage is dealt to an Organic Metal body type:
>
>- 100 thermal damage
>- x2 (strong against organic)
>- x2 (strong against metal)
>- = 400 damage

These all stack together.

> 100 Acidic damage is dealt to a Liquid Metal body type. Both Specimen have an affinity to Acidic:
>
> - 100 acidic damage
> - x1.5 (attacker affinity)
> - x0.5 (weak against organic)
> - x2 (strong against metal)
> - x0.667 (defender affinity)
> - = 100 damage
>

Immunities will always cancel out damage.

> 100 Magnetic damage is dealt to an Abyssal Crystal body type by a Specimen with an affinity to Magnetic:
>
> - 100 magnetic damage
> - x1.5 (skill affinity)
> - x0 (crystal is immune)
> - x2 (strong against abyssal)
> - = 0 damage
>

