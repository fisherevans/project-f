// Example: Zone Trigger Script
// Triggers events when player enters specific zones

const SUBSCRIPTION = {
  EnterZone: ["power_room", "security_office"],
  LeaveZone: ["power_room"],
  Priority: 5
};

function OnEnterZone(self, world, state, event) {
  const effects = [];
  
  if (event.zoneName === "power_room") {
    // Check if power is already on
    const powerOn = world.getVar("global.power.grid_online");
    
    if (!powerOn) {
      effects.push({
        type: "ShowDialog",
        data: {
          text: "The power grid is offline. Find the main switch."
        }
      });
      
      effects.push({
        type: "SetVar",
        data: {
          key: "map.facility.power_room.visited",
          value: true
        }
      });
    } else {
      effects.push({
        type: "ShowDialog",
        data: {
          text: "The power grid hums with energy."
        }
      });
    }
  } else if (event.zoneName === "security_office") {
    // Increase security alert level
    const currentLevel = world.getVar("topic.security.alarm_level") || 0;
    
    effects.push({
      type: "IncVar",
      data: {
        key: "topic.security.alarm_level",
        value: 1
      }
    });
    
    if (currentLevel >= 2) {
      effects.push({
        type: "Trigger",
        data: {
          triggerName: "SecurityAlert",
          data: { level: currentLevel + 1 }
        }
      });
    }
  }
  
  return { effects: effects };
}

function OnLeaveZone(self, world, state, event) {
  if (event.zoneName === "power_room") {
    return {
      effects: [{
        type: "ShowDialog",
        data: {
          text: "You leave the power room."
        }
      }]
    };
  }
}
