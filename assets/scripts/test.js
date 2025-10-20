// entity.js
// @ts-check
function Init(self, world, state) {
    state = state || {
        counter: 0
    }
    log("state in init: " + JSON.stringify(state))
    return {
        state: state
    }
}

function OnInteract(self, world, state, event) {
    if (event.targetId !== self.id || state.complete)  {
        return
    }
    log("state in interact: " + JSON.stringify(state))
    log("state.counter before increment: " + state.counter)
    state.counter++
    log("counter at " + state.counter)
    const effects = [];
    if(state.counter > 5) {
        state.complete = true
        log("target hit")
        effects.push({
            type: "dialogue",
            data: {
                text: "You have interacted with me 5 times!",
            }
        });
        effects.push({
            type: "set_world_var",
            data: {
                key: "counter_hit",
                value: true,
            }
        });
    }
    return {
        state: state,
        effects: effects
    }
}