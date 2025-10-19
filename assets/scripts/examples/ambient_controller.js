// Example: Ambient Controller that manages music and atmosphere
// Demonstrates Step event for periodic updates and complex state management

const SUBSCRIPTION = {
    VarPrefix: ["map.*", "topic.combat.*"],
    Priority: 15
};

const MUSIC_TRACKS = {
    peaceful: "ambient_meadow",
    tense: "ambient_tension",
    combat: "battle_theme",
    victory: "victory_fanfare"
};

function OnInit(self, world, state) {
    return {
        effects: [{
            type: "PlayMusic",
            data: {
                musicId: MUSIC_TRACKS.peaceful,
                fadeIn: 2.0
            }
        }],
        state: {
            currentTrack: MUSIC_TRACKS.peaceful,
            combatActive: false,
            tensionLevel: 0,
            lastUpdate: getTime()
        }
    };
}

function OnStep(self, world, state, event) {
    const effects = [];
    const now = getTime();
    const timeSinceUpdate = (now - (state.lastUpdate || 0)) / 1000;
    
    // Only update every 5 seconds
    if (timeSinceUpdate < 5.0) {
        return { state: state };
    }
    
    // Check combat state
    const inCombat = world.getVar("topic.combat.active") || false;
    
    if (inCombat && !state.combatActive) {
        // Entered combat
        effects.push({
            type: "PlayMusic",
            data: {
                musicId: MUSIC_TRACKS.combat,
                fadeIn: 0.5
            }
        });
        
        return {
            effects: effects,
            state: {
                currentTrack: MUSIC_TRACKS.combat,
                combatActive: true,
                tensionLevel: state.tensionLevel,
                lastUpdate: now
            }
        };
    } else if (!inCombat && state.combatActive) {
        // Exited combat
        const won = world.getVar("topic.combat.victory") || false;
        
        if (won) {
            // Victory sequence
            return {
                plan: {
                    type: "Seq",
                    children: [
                        {
                            type: "Effect",
                            data: {
                                type: "PlayMusic",
                                data: {
                                    musicId: MUSIC_TRACKS.victory,
                                    fadeIn: 0.5
                                }
                            }
                        },
                        {
                            type: "Wait",
                            data: { duration: 5.0 }
                        },
                        {
                            type: "Effect",
                            data: {
                                type: "PlayMusic",
                                data: {
                                    musicId: MUSIC_TRACKS.peaceful,
                                    fadeIn: 2.0
                                }
                            }
                        }
                    ]
                },
                state: {
                    currentTrack: MUSIC_TRACKS.peaceful,
                    combatActive: false,
                    tensionLevel: 0,
                    lastUpdate: now
                }
            };
        } else {
            // Return to peaceful
            effects.push({
                type: "PlayMusic",
                data: {
                    musicId: MUSIC_TRACKS.peaceful,
                    fadeIn: 2.0
                }
            });
            
            return {
                effects: effects,
                state: {
                    currentTrack: MUSIC_TRACKS.peaceful,
                    combatActive: false,
                    tensionLevel: 0,
                    lastUpdate: now
                }
            };
        }
    }
    
    // Check tension level (based on alarm)
    if (!inCombat) {
        const alarmLevel = world.getVar("topic.security.alarm_level") || 0;
        
        if (alarmLevel >= 3 && state.currentTrack !== MUSIC_TRACKS.tense) {
            effects.push({
                type: "PlayMusic",
                data: {
                    musicId: MUSIC_TRACKS.tense,
                    fadeIn: 1.0
                }
            });
            
            return {
                effects: effects,
                state: {
                    currentTrack: MUSIC_TRACKS.tense,
                    combatActive: false,
                    tensionLevel: alarmLevel,
                    lastUpdate: now
                }
            };
        } else if (alarmLevel < 3 && state.currentTrack === MUSIC_TRACKS.tense) {
            effects.push({
                type: "PlayMusic",
                data: {
                    musicId: MUSIC_TRACKS.peaceful,
                    fadeIn: 2.0
                }
            });
            
            return {
                effects: effects,
                state: {
                    currentTrack: MUSIC_TRACKS.peaceful,
                    combatActive: false,
                    tensionLevel: alarmLevel,
                    lastUpdate: now
                }
            };
        }
    }
    
    return {
        state: {
            currentTrack: state.currentTrack,
            combatActive: state.combatActive,
            tensionLevel: world.getVar("topic.security.alarm_level") || 0,
            lastUpdate: now
        }
    };
}

function OnFlagChanged(self, world, state, event) {
    // Immediate response to critical changes
    if (event.varName === "topic.combat.active" && event.newValue === true) {
        return {
            effects: [{
                type: "PlayMusic",
                data: {
                    musicId: MUSIC_TRACKS.combat,
                    fadeIn: 0.3
                }
            }],
            state: {
                currentTrack: MUSIC_TRACKS.combat,
                combatActive: true,
                tensionLevel: state.tensionLevel,
                lastUpdate: getTime()
            }
        };
    }
    
    return { state: state };
}
