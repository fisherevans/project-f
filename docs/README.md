# Project F

"Project F" is a singke player, 2D pixelart, sci-fi, catch-em-all roguelike built in Go using a custom engine inspired by retro RPG games, pulling in modern themes and mechanics. The game is designed to be played in the form factor of a Game Boy Advance.


You play as an astrobiologist remotely piloting a humanoid Animech to explore alien worlds, capturing life forms and harnessing their powers. While playing, you uncover the truth behind the corporation that employs you in this dystopian world. Combat is a tactical mix or turn based skill selection, and well timed execution.

> *Click the image below to watch a recent demo.*

[![Watch a demo](https://img.youtube.com/vi/4j0PABJWh3c/hqdefault.jpg)](https://www.youtube.com/watch?v=4j0PABJWh3c)

## Quick Links

- [Devlog](https://www.tumblr.com/fishwingdev) (with screenshots & demo footage)
- [Story](story.md)
- [Gameplay loop](core-loop.md)

# Game Summary

> ***Bold** words can be found in the [Glossary](glossary.md)...*

You've been living a life sequestered on a small outer-system planet, researching local ecosystems and fauna. It suffers a terrible tragedy, killing everyone you've grown close to. You survived, but only because of your own selfish actions. Your small group friends and family don't share the same fate. Your roots are completely upended and you are forced to start a new life. You take a job as an **Astrobiologist** working for **Thanadox Industries** - the corporation that effectively rules the space frontier. You're tasked with exploring newly discovered planets in order to research fauna found in the outer fringes of explored space.

The nature of your job means you live a lonely existence on a one-man research vessel far away from any real civilization. Using a Thanadox **Animech** (a humanoid mech you can control "remotely"), you explore planets without having to land your research vessel. By imbuing your soul into the humanoid robot and leaving your human body in a form of stasis, you can fully control the robot as if it were yourself. The creatures found on these planets have unique strengths and abilities. Your job is to find these **Specimen** and capture the essence of their lifeforms - cataloging both their physical and metaphysical properties: their **Primortal** forms. You do this with a tool called a **XenoCrypt**. These are sent back to Thanadox headquarters for further study.

Your Animech is able to learn from these captured specimen, allowing you to use the skills they themselves used, enhancing your combat abilities. As you gain experience within your Animech you will be able increase its strength and the skills it has access to. The more Primortal forms you capture, the more you are able to learn from their innate abilities and unlock new and stronger skills and traits.

As you explore a planetary **Biome** you will find **Elythium**, a powerful energy resource used to power space travel. You must capture enough of it in order to transfer your Animech's soul back to your ship. If your Animech is destroyed while exploring you lose any physical items acquired (such as captured Primortals) and your own soul slowly finds its way back to your ship, reanimating your human body. You may find checkpoints throughout a Biome that can be used to repair your Animech if it is damaged, as well as offer an opportunity to upload any digitial acquisitions, such as Animech usage data which is considered experience points.

Throughout your explorations you'll make discoveries about Thanadox Industries, the civilization at the outer edge of space exploration, and yourself. Is what you do moral? Is what others do your responsibility? More on the story [here](story.md)...

# Game Overview

- The [Core Loop](core-loop.md) explains the general shape of the game: deploy, extract, upgrade, repeat.
- You [explore planets](exploration.md), researching specimen and collecting resources.
- You'll face alien life forms and other NPCs in [Battle](combat-system.md).
- With the experience and research collected while deployed, you'll be able to [upgrade your Animech's abilities](progression.md) between runs.

# Constraints

The primary constraint I've given myself while implementing this game is that it would need to be playable in the form-factor of a Gameboy Advanced:

- 240x160px resolution
- Limited inputs: A, B, D-Pad, Start, Select

The aim is to keep the game complexity low and reduce scope, while also ensuring it's easy to learn and play on various devices.

Other goals I have for this game are that:

- Players can achieve meaningful progress in a 15m session
- Combat is interesting and difficult to master.
- The story is meaningful and engaging for a more mature audience.

# Influences

- Pokemon
  - 1v1 turn based combat
  - Capturing creatures to be used later for combat
  - Affinities and types
  - Tile based adventure with random encounters
- Returnal, Risk of Rain 2, Witchfire
  - Rouge like runs in similar biomes, new biomes unlocked with story progression
  - Story progressed via small tidbits found during runs
- Golden Sun
  - Djinns influence abilities available
  - Turn based combat
- Guild Wars
  - Players build load-outs based on skills they acquire throughout the game.
- Neon Genesis
  - Humanoid mechs, powered/controlled by the souls of another
- Battlestar Galactica
  - The soul is separate from the physical body - upon death can be transmitted back to an origin
- Final Fantasy
  - Various moral quandaries
  - Mix of real time + turn based combat

# Tools Used

- [Aseprite](https://www.aseprite.org/) to edit sprites
- [Tiled](https://www.mapeditor.org/) to create maps
- Libraries:
  - [Pixel](https://github.com/gopxl/pixel), a lightweight Golang OpenGL wrapper
- Sprites (temporary and altered):
  - [SnowHex](https://snowhex.itch.io/)
  - [CyberMonkeyAssets](https://cybermonkeyassets.itch.io/sci-fi-spacial-pack)
- Fonts:
  - [AddStandard](https://www.dafont.com/addstandardbitmap.font)
  - [3-by-5 Pixel Font](https://fontstruct.com/fontstructions/show/716744/3_by_5_pixel_font)
  - [m3x7](https://fontstruct.com/fontstructions/show/2372824/3x7-font)
  - [m3x5](https://fontstruct.com/fontstructions/show/716744/3_by_5_pixel_font)
- AI usage disclaimer:
  - I sometimes use AI to generate images:
    - For smaller sprites, as references/templates
    - For larger backdrops (processed them in post)
  - Some code snipets/blocks are AI generated and altered to fit my needs
  - AI tools are used in my IDE for autocomplete/predictive typing
