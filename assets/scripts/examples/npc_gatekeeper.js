// Example: NPC Gatekeeper Script
// This NPC guards a gate and only opens it after being greeted multiple times

const SUBSCRIPTION = {
  Priority: 10
};

const STATE_VERSION = 1;

function OnInit(self, world, state) {
  return {
    state: {
      greeted: false,
      visits: 0,
      gateOpen: false
    }
  };
}

function OnInteract(self, world, state, event) {
  const visits = (state.visits || 0) + 1;
  const greeted = state.greeted || false;
  
  const effects = [];
  
  if (!greeted) {
    // First interaction
    effects.push({
      type: "ShowDialog",
      data: {
        text: "Halt! State your business, traveler.",
        speaker: "Gatekeeper"
      }
    });
    
    return {
      effects: effects,
      state: {
        greeted: true,
        visits: visits,
        gateOpen: false
      }
    };
  } else if (visits < 3) {
    // Subsequent visits before opening gate
    effects.push({
      type: "ShowDialog",
      data: {
        text: "You again? I'm watching you... (" + visits + "/3)",
        speaker: "Gatekeeper"
      }
    });
    
    return {
      effects: effects,
      state: {
        greeted: true,
        visits: visits,
        gateOpen: false
      }
    };
  } else if (!state.gateOpen) {
    // Third visit - open the gate
    effects.push({
      type: "ShowDialog",
      data: {
        text: "Alright, alright. You're persistent. The gate is open.",
        speaker: "Gatekeeper"
      }
    });
    
    effects.push({
      type: "OpenDoor",
      data: {
        doorId: "north_gate"
      }
    });
    
    effects.push({
      type: "SetVar",
      data: {
        key: "map.meadow.north_gate.open",
        value: true
      }
    });
    
    return {
      effects: effects,
      state: {
        greeted: true,
        visits: visits,
        gateOpen: true
      }
    };
  } else {
    // Gate already open
    effects.push({
      type: "ShowDialog",
      data: {
        text: "The gate is open. Move along.",
        speaker: "Gatekeeper"
      }
    });
    
    return {
      effects: effects,
      state: state
    };
  }
}

function Migrate(state, fromVersion) {
  if (fromVersion === 1) {
    // Example migration from v1 to v2
    return {
      greeted: state.greeted || false,
      visits: state.visits || 0,
      gateOpen: state.gateOpen || false,
      // Add new field in v2
      lastInteractionTime: 0
    };
  }
  return state;
}
