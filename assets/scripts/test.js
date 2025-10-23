// @ts-check
/** @type {import('./global').EntityHandler} */
const handler = {
    Init(self, world, state) {
        return {
            state: state || {
                counter: 0
            },
        };
    },

    OnInteract(self, world, state, event) {
        if (event.targetId !== self.Id()) {
            return null;
        }

        if (self.Mode() === "mined") {
            return null;
        }

        state.counter++;
        log.Info("Node mined " + state.counter + " times!");

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
                        mode: "mined",
                    },
                    yieldElythium: {
                        amount: 3,
                    }
                }
            ]
        };
    },

    OnTimerComplete(self, world, state, event) {
        if (event.createdBy !== self.Id() || event.timerId !== "reset") {
            return null;
        }
        log.Info("Node reset");
        return {
            state: state,
            effects: [
                {
                    mutateEntity: {
                        entityId: self.Id(),
                        mode: "",
                    }
                }
            ]
        };
    }
};