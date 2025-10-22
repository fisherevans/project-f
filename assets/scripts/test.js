// entity.js
// @ts-check
function Init(self, world, state) {
    return {
        state: state || {
            counter: 0
        },
    }
}

function OnInteract(self, world, state, event) {
    if (event.targetId !== self.Id()) {
        return;
    }

    if (self.Mode() === "mined") {
        return;
    }

    state.counter++
    log.Info("Node mined " + state.counter + " times!")

    return {
        state: state,
        effects: [
            {
                timer: {
                    timerId: "reset",
                    durationSeconds: 3,
                },
                mutateEntity: {
                    entityId: self.Id(),
                    newMode: "mined",
                },
                yieldElythium: {
                    amount: 3,
                }
            }
        ]
    };
}

function OnTimerComplete(self, world, state, event) {
    if (event.createdBy !== self.Id() || event.timerId !== "reset") {
        return;
    }
    log.Info("Node reset")
    return {
        state: state,
        effects: [
            {
                mutateEntity: {
                    entityId: self.Id(),
                    newMode: "",
                }
            }
        ]
    };
}