function Init(self, world, state) {
    return {
        state: state || {
            ready: true
        },
    }
}

function OnTimerComplete(self, world, state, event) {
    if (event.createdBy !== self.Id() || event.timerId !== "reset") {
        return;
    }
    state.ready = true
    return {
        state: state,
    };
}

function OnEntityZoneActivity(self, world, state, event) {
    if (event.zoneId === "exit" &&
        event.isEntering === true &&
        event.entityId === world.Get("player_id") &&
        state.ready) {
        state.ready = false;
        return {
            state: state,
            effects: [
                {
                    timer: {
                        id: "reset",
                        durationSeconds: 10,
                    },
                }
            ]
        }
    }
}