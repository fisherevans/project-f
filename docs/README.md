# Project F (Primortal?)

> The title is a work in progress. The letter F might indicate how many times I've started a game and not finished it. I'm leaning towards "Primortal" which is [fairly unique](https://www.google.com/search?q=primortal&source=hp&oq=primortal).

A single player, party-based, rogue-like, collect-em-all, space RPG, designed to be played in the form factor of a Game Boy Advance.

- [Game Summary (Below)](#game-summary)
- [Glossary](glossary.md)
- [Story](story.md)
- [Types](types.md)
- [Combat and Progression](combat-and-progression.md)
- [Requirements](requirements.md)

# Game Summary

> ***Bold** words can be found in the [Glossary](glossary.md)...*

You've lived a life sequestered on a small outer-rim planet that has recently suffered a terrible tragedy, killing most civilians. You survived, but only because of your own selfish actions. Your small group friends and family don't share the same fate. Your roots are completely upended and you are forced to start a new life. You take a job as an **Astrobiologist** working for **Thanadox Industries** - the corporation that effectively rules the space frontier. You're tasked with exploring newly discovered planets in order to research fauna found in the outer fringes of explored space.

The nature of your job means you live a lonely existence on one-man research vessel far away from any real civilization. Using a Thanadox **Animech** (a humanoid mech you can control remotely), you explore planets without having to land your research vessel. By imbuing your soul into the humanoid robot and leaving your human body in a form of stasis, you can fully control the robot as if it were yourself. The creatures found on these planets have both special **Physical** and **Aethereal** strengths and abilities. Your job is to find these **Specimen** and capture the essence of their lifeforms - cataloging both their physical and metaphysical properties: their **Primortal** forms. You do this with a tool called a **XenoCrypt**. These are sent back to Thanadox headquarters for further study.

In addition to imbuing your soul into an Animech, it can also harness previously captured Primortal forms within capsules called **Animetric Cores** that can be loaded into the mech. This is called **Imprinting**. These Primortal forms grant access to special powers and abilities useful in combat and exploration. As you gain experience within your Animech you will be able increase its strength and the number of capsules it carry at the same time. The more Primortal forms you capture, the more you are able to learn from their innate abilities and unlock new and stronger skills and traits.

As you explore a planetary **Biome** you will find **Elythium**, a powerful energy resource used to power space travel. You must capture enough of it in order to launch your Animech back to orbit and return to your ship. If your Animech is destroyed while exploring you lose any physical items acquired (such as captured Primortals) and your soul slowly finds its way back to your ship, reannimating your human body. You may find checkpoints throughout a Biome that can be used to repair your Animech if it is damaged, as well as offer an opportunity to upload any digitial acquisitions, such as Animech usage data which is considered experience points.

Throughout your explorations you'll make discoveries about Thanadox Industries, the civilization at the outer edge of space exploration, and yourself. Is what you do moral? Is what others do your responsibility? More on the story [here](story.md)...

# Game Loop

## Starting a Run

You're headquarters is always your research vessel.

- You choose a destination biome and send your Animech down to the Planet.
- While exploring, you'll encounter, battle, and capture Specimen.
- You might also find other Astrobiologists, mining colonies, raiders, pirates, highwaymen, small settlements and other special areas and events that offer opportunities to battle NPCs, buy & trade items, and progress the story.
- Once you gather enough Elythium by killing creatures or finding deposits throughout the biome to return to your ship, you can do so at any checkpoint.
- Most biomes will have at least one special encounter that will feel like a boss, granting hefty rewards and enough Elythium to return to the ship.

## During a Run

While exploring as your Animech:

- You will need to find enough Elythium to return back to your ship. This resource is collected from dying life forms (i.e. defeating wild Specimen), or captured in other forms (i.e. extracted from other defeated Animechs or caches found throughout the biome).
- You will encounter wild Specimen:
  - This will happen in areas with random encounters (i.e. "tall grass") or as predicted encounters - by colliding with a Specimen "NPC" found in the world, or entering an aggressive Specimen's line of sight.
  - When encountering wild Specimen, you can either:
    - Defeat the Specimen (granting experience points, research points for the Primortals users in combat, and Elythium)
    - Capture the Specimen with your XenoCrypt. This is done with a mini game that can change based on Animech upgrades. The mini game is easier with a higher resonance. Resonance is increased by certain skills you can use in combat.
    - Or, attempt escape which may not be successful based on various RPG factors.
- You will encounter non-specimen enemies:
  - These could be other human characters, other Animechs, military robots, and more.
  - You cannot capture Specimen in these encounters, nor try to escape.
  - Defeating them grant experience points, research points, and other items & resources.
- When battling:
  - Your Animech can load a Animetric Core of a Primortal to gain it's abilities, strengths, weaknesses, and physical traits.
  - When fighting, combatants use skills to deal damage to their opponent of a certain skill/damage type. The recipient's body type determine how effective that damage is.
  
- You will likely find one or more boss encounters through a Biome. These are stronger enemies, but offer greater rewards.
  - Combat during these encounters may have special rules you need to follow or adapt to.
  - Some bosses are Specimen that can be captured.
- You will come across NPCs, settlements, traders, items, and other types interactive encounters that will progress the story or offer unique opportunities to obtain various resources.
  - These encounters will progress the story and unlock coordinates of other planets with new biomes to explore.
- You will encounter checkpoints. These are locations that:
  - Have the equipment necessary to heal your Animech
  - Re-sync your Primortals recovering any lost SYNC in battle
  - Upload digital data to your space ship
  - Or, if you have enough Elythium, return to your ship

> Learn more about combat [here](combat-and-progression.md)...

## Completing a Run

Once you return to your ship, you will receive rewards for your exploration:
- **Expeirence Points**: Used to level up your Animech. This means:
  
  - Increasing it's base Shield and Integrity stats, or
  - Increasing the number of Animetric Cores it can hold at the same time
  > *Experience points can be uploaded digitally at Checkpoints.*

- **Primortals**: Unlock new Animetric Cores that will grant new abilities to your Animech when loaded
  
  - Capturing more of the same species will instead grant Research Points for the existing Primortal.
  > *Primortals must be physically returned to your ship.*
  
- **Research Points**: Are granted from captured specimen of combat experience while using a Primortal form. They can be used strengthen the abilities of a Primortal. Such as:
  
  - Upgrading Sync (health), Aether (special attack), or Kinetic (special attack) stats to increase the amount of damage your Animech can take or deal while using this Primortal Form
  - Unlocking new skills from a skill tree that grant access to new combat or passive abilities. Skill trees are capped by the level of your Animech.
  > *Research Points obtained during combat can be uploaded digitally at Checkpoints.*
  
- **Special Items**: Most cannot be fully processed by the Animech while on-planet and require equipment on your research vessel to decode or reconstruct. If successfully brought back to your ship, these items can be used to progress the story, unlock special Primortals, grant access to skills not available in skill trees, and other special benefits.
  
  > *Most Special Items must be physically returned to your ship.*

- **Memories**: Certain pieces of information found while in your Animech (such as a code on a note, or parts of a conversation) can be retained in your Soul's memory. Most memories will be used to progress the story, but some may also be used to unlock certain areas.

  > *Memories are retained the moment they happen. They do not need to be uploaded or physically returned to the ship.*


## Between Runs

In between runs, you can use facilities on your research vessel.

- **Manage Loadouts**: Animech loadouts can be created, updated, saved, and loaded into your Animech. This allows you to alter what capabilities, combat strategies, and Animetric Cores you have available for your next run.
- **Receive Communications**: You may be sent messages from Thanadox Industries or other people you come across in your travels.

## Special Events

Beyond the core game loop, special events can happen that grant special rewards or progress the story:

- **Altered Course**: While en route to a target biome, your Animech may detect a previously unknown moon colony, pirate ship, meteor, abandoned space stations, etc. that you may choose to explore instead of your intended destination. These micro-biomes could offer special challenges and rewards, unique items and other benefits.
- **Settlements**: As mentioned in the main Game Loop, while exploring you may stumble accross small settlements or other pockets of civilization. Some of these are ephemeral and aren't easy to return to (and if you do, may not be the same). But some of them are large enough that you are able to keep track of their location. You will be able to choose these locations later on as destinations allowing you to explore them further. There might be future story beats that unlock NPCs, or abilities needed to solve certain puzzles granting access to previously inaccessible locations.
