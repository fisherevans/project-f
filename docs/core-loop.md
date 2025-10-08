# Core Game Loop

The core gameplay of *Project F* is structured around repeatable **expeditions**: preparing at HQ, deploying your Animech to a planet, exploring its zones, battling creatures, extracting with resources, and returning to HQ to upgrade and prepare for the next run.

## Story Context

You play as an astrobiologist contracted to explore the galaxy using an **Animech** - a machine vessel piloted by your soul. The corporation that funds your expeditions seeks data and samples from distant planets.  
- This document focuses on the mechanics of the core game loop.  
- For details on the overarching main story, see [Story](./story.md).

## The Loop Overview

1. **HQ (Preparation & Progression)**  
2. **Deployment to Planet**  
3. **Exploration & Encounters**  
4. **Combat**  
5. **Extraction**  
6. **Return to HQ**  

This cycle repeats, with each run pushing deeper into planets and unlocking new abilities, moves, and narrative progression.

### HQ (Preparation & Progression)

Your headquarters is always your research vessel, where you set up for the next expedition:  

- **Loadouts**  
  - Choose 4 moves from your permanently unlocked pool.  
  - You can save and load named "loadouts" for easy swapping.  

- **Upgrades**
  - Spend **Experience (XP)** to improve Animech stats (Shield, Sync, Regen, Xenocrypt efficiency).  
  - Unlock new utility abilities (e.g., Dash) once discovered in the world.  

- **Research**
  - Spend **Research Points (RP)** to unlock moves in creature‑specific skill trees.  
  - After a tree is complete, future RP from that creature type converts into sellable intel (economy TBD).  

- **Planet Selection**  
  - Choose which planet to visit.  
  - Select a **landing zone** (initially limited, more unlock as you discover outposts/settlements).  

### Deployment to Planet

- Your human body never travels to the planets, instead your soul is extracted and used to control a robotic "Animech"
- Launching to a planet sends your Animech through deep space on a one way trip. The soul of you Animech and yourself are able to return back to you and your ship outside of the material world.
- **Random travel events** may occur during transit (e.g., anomalies, derelicts, asteroids). These events are not fully defined yet but will add variety to the journey.  
  - **Altered Course**: While en route to a target biome, your Animech may detect a previously unknown moon colony, pirate ship, meteor, abandoned space stations, etc. that you may choose to explore instead of your intended destination. These micro-biomes could offer special challenges and rewards, unique items and other benefits.
  - **Settlements**: As mentioned in the main Game Loop, while exploring you may stumble accross small settlements or other pockets of civilization. Some of these are ephemeral and aren't easy to return to (and if you do, may not be the same). But some of them are large enough that you are able to keep track of their location. You will be able to choose these locations later on as destinations allowing you to explore them further. There might be future story beats that unlock NPCs, or abilities needed to solve certain puzzles granting access to previously inaccessible locations.
- Once deployed, the Animech descends to the selected landing zone and the run begins.  

### Exploration & Encounters

Exploration is described in detail in the [**Exploration & Adventure**](exploration.md) doc, but at a high level:  

- Planets are built from **forward‑only, branching zone webs**.  
- Encounters include:  
  - **Shadow mobs** (wild creatures that chase or flee).  
  - **Trainer‑like characters** (NPCs with scripted battles).  
  - **Bosses or chained battles** (sequential fights against multiple foes).  
- **Caches** can be found containing **run‑specific mods** (temporary combat buffs).  
- **Settlements and outposts** act as safe zones and unlock future landing options.  

### Combat

Combat is described in detail in the [**Combat System**](combat-system.md) doc. A quick summary:  

- Battles are a mix **real-time** and **turn‑based**, allowing users to take their time, but rewarding them when they maintain tempo.
- Skills are composed of ticks, which may deal damage, apply statuses, or put the Animech into stances.  
- Encounters always deal some damage; survival relies on Shield regen and Sync integrity.  
- The **Xenocrypt** can capture creatures for more RP (but less XP).  

### Extraction

- To keep your progress, you must reach an **extraction node** and spend enough **Elythium** (collected during the run).  
- Extraction ends the run and beams your soul and gathered data back to HQ.  

**Outcomes:**  

- **Successful Extraction** → All XP and RP banked at HQ.  
- **Death (Sync = 0)** → You lose all unbanked RP, but some **XP is preserved**.  
- **No Elythium** → You cannot extract until you find more.  

### Return and Rest

After extraction (or death), you return to HQ: 

- 
  This is the same as the first phase of the loop - you can upgrade your character and prepare for the next deployment.
- You will be forced to rest between runs. This time is used to further the narrative, triggering memories and flash backs that progress the story and expand on general lore.

## Core Loop Summary

The rhythm of play is:  

- **Prepare** - Build loadouts, upgrade Animech, plan landing zone.  
- **Deploy** - Travel to a planet, possibly encounter random events.  
- **Explore** - Traverse one‑way zones, encounter enemies, discover caches.  
- **Fight** - Defeat or capture creatures in tactical tick‑based combat.  
- **Extract** - Secure Elythium, reach an endpoint, and beam data home.  
- **Rest** - Return to HQ to rest, potentially triggering memories that progress the story.

This loop blends **long‑term progression** (Animech upgrades, move unlocks) with **short‑term tension** (run‑specific mods & load-outs, risk of failing to extract), encouraging careful planning and bold exploration.  
