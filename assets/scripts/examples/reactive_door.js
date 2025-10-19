// Example: Reactive Door that responds to world variable changes
// Demonstrates VarPrefix subscription and FlagChanged event

const SUBSCRIPTION = {
    VarPrefix: ["global.power.*"],
    Priority: 5
};

function OnInit(self, world, state) {
    // Check initial power state
    const powerOn = world.getVar("global.power.grid_online");
    
    return {
        effects: powerOn ? [] : [{
            type: "CloseDoor",
            data: { doorId: self.id }
        }],
        state: {
            isPowered: powerOn || false,
            locked: !powerOn
        }
    };
}

function OnFlagChanged(self, world, state, event) {
    // React to power grid changes
    if (event.varName === "global.power.grid_online") {
        const powerOn = event.newValue;
        
        if (powerOn && state.locked) {
            // Power restored - unlock door
            return {
                effects: [
                    {
                        type: "ShowDialog",
                        data: {
                            text: "The electronic lock disengages with a soft click."
                        }
                    },
                    {
                        type: "OpenDoor",
                        data: { doorId: self.id }
                    },
                    {
                        type: "PlaySound",
                        data: { soundId: "door_unlock" }
                    }
                ],
                state: {
                    isPowered: true,
                    locked: false
                }
            };
        } else if (!powerOn && !state.locked) {
            // Power lost - lock door
            return {
                effects: [
                    {
                        type: "ShowDialog",
                        data: {
                            text: "The door's electronic lock engages as power fails."
                        }
                    },
                    {
                        type: "CloseDoor",
                        data: { doorId: self.id }
                    },
                    {
                        type: "PlaySound",
                        data: { soundId: "door_lock" }
                    }
                ],
                state: {
                    isPowered: false,
                    locked: true
                }
            };
        }
    }
    
    return { state: state };
}

function OnInteract(self, world, state, event) {
    if (state.locked) {
        return {
            effects: [{
                type: "ShowDialog",
                data: {
                    text: "The door is locked. It requires power to operate."
                }
            }]
        };
    } else {
        return {
            effects: [{
                type: "ShowDialog",
                data: {
                    text: "The door slides open smoothly."
                }
            }]
        };
    }
}
