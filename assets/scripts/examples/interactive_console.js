// Example: Interactive Console with Player Choices
// Demonstrates the Choice plan type for branching dialogue

const SUBSCRIPTION = {
    Priority: 10
};

function OnInit(self, world, state) {
    return {
        state: {
            powerRestored: false,
            securityOverridden: false,
            conversationComplete: false
        }
    };
}

function OnInteract(self, world, state, event) {
    if (state.conversationComplete) {
        return {
            effects: [{
                type: "ShowDialog",
                data: {
                    text: "System ready. All functions operational."
                }
            }]
        };
    }

    // Multi-step conversation with choices
    return {
        plan: {
            type: "Seq",
            children: [
                {
                    type: "Effect",
                    data: {
                        type: "ShowDialog",
                        data: {
                            text: "SYSTEM CONSOLE v2.4.1\nSelect operation:",
                            speaker: "Console"
                        }
                    }
                },
                {
                    type: "Choice",
                    data: {
                        prompt: "What would you like to do?",
                        options: [
                            "Restore power grid",
                            "Override security",
                            "Check system status",
                            "Exit"
                        ]
                    },
                    children: [
                        // Option 0: Restore power
                        {
                            type: "Seq",
                            children: [
                                {
                                    type: "Effect",
                                    data: {
                                        type: "ShowDialog",
                                        data: { text: "Initializing power grid..." }
                                    }
                                },
                                {
                                    type: "Wait",
                                    data: { duration: 2.0 }
                                },
                                {
                                    type: "Effect",
                                    data: {
                                        type: "SetVar",
                                        data: {
                                            key: "global.power.grid_online",
                                            value: true
                                        }
                                    }
                                },
                                {
                                    type: "Effect",
                                    data: {
                                        type: "ShowDialog",
                                        data: { text: "Power grid restored successfully." }
                                    }
                                },
                                {
                                    type: "Effect",
                                    data: {
                                        type: "Trigger",
                                        data: {
                                            triggerName: "PowerRestored"
                                        }
                                    }
                                }
                            ]
                        },
                        // Option 1: Override security
                        {
                            type: "Seq",
                            children: [
                                {
                                    type: "Effect",
                                    data: {
                                        type: "ShowDialog",
                                        data: { text: "WARNING: Security override requires authorization." }
                                    }
                                },
                                {
                                    type: "Wait",
                                    data: { duration: 1.0 }
                                },
                                {
                                    type: "If",
                                    data: {
                                        condition: world.hasVar("global.inventory.security_keycard")
                                    },
                                    children: [
                                        // Has keycard
                                        {
                                            type: "Seq",
                                            children: [
                                                {
                                                    type: "Effect",
                                                    data: {
                                                        type: "ShowDialog",
                                                        data: { text: "Keycard accepted. Security systems disabled." }
                                                    }
                                                },
                                                {
                                                    type: "Effect",
                                                    data: {
                                                        type: "SetVar",
                                                        data: {
                                                            key: "topic.security.alarm_level",
                                                            value: 0
                                                        }
                                                    }
                                                }
                                            ]
                                        },
                                        // No keycard
                                        {
                                            type: "Effect",
                                            data: {
                                                type: "ShowDialog",
                                                data: { text: "ERROR: Authorization failed. Access denied." }
                                            }
                                        }
                                    ]
                                }
                            ]
                        },
                        // Option 2: Check status
                        {
                            type: "Effect",
                            data: {
                                type: "ShowDialog",
                                data: {
                                    text: "SYSTEM STATUS:\n" +
                                          "Power: " + (world.getVar("global.power.grid_online") ? "ONLINE" : "OFFLINE") + "\n" +
                                          "Security: Level " + (world.getVar("topic.security.alarm_level") || 0)
                                }
                            }
                        },
                        // Option 3: Exit
                        {
                            type: "Effect",
                            data: {
                                type: "ShowDialog",
                                data: { text: "Exiting console..." }
                            }
                        }
                    ]
                }
            ]
        },
        state: {
            powerRestored: world.getVar("global.power.grid_online") || false,
            securityOverridden: (world.getVar("topic.security.alarm_level") || 0) === 0,
            conversationComplete: true
        }
    };
}
