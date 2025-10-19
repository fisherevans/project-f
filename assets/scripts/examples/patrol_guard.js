// Example: Patrol Guard with Timer-based Movement
// Demonstrates timer events and state-based behavior

const SUBSCRIPTION = {
  Priority: 5
};

function OnInit(self, world, state) {
  return {
    effects: [{
      type: "StartTimer",
      data: {
        timerName: "patrol",
        duration: 3.0
      }
    }],
    state: {
      patrolIndex: 0,
      patrolPoints: [
        { x: 10, y: 5 },
        { x: 15, y: 5 },
        { x: 15, y: 10 },
        { x: 10, y: 10 }
      ],
      suspicious: false
    }
  };
}

function OnTimer(self, world, state, event) {
  if (event.timerName === "patrol") {
    const patrolIndex = state.patrolIndex || 0;
    const patrolPoints = state.patrolPoints;
    const nextIndex = (patrolIndex + 1) % patrolPoints.length;
    const nextPoint = patrolPoints[nextIndex];
    
    return {
      effects: [
        {
          type: "MoveTo",
          data: {
            x: nextPoint.x,
            y: nextPoint.y
          }
        },
        {
          type: "StartTimer",
          data: {
            timerName: "patrol",
            duration: 3.0
          }
        }
      ],
      state: {
        patrolIndex: nextIndex,
        patrolPoints: patrolPoints,
        suspicious: state.suspicious
      }
    };
  }
}

function OnInteract(self, world, state, event) {
  const alarmLevel = world.getVar("topic.security.alarm_level") || 0;
  
  if (alarmLevel > 2) {
    return {
      effects: [
        {
          type: "ShowDialog",
          data: {
            text: "Intruder! Sound the alarm!",
            speaker: "Guard"
          }
        },
        {
          type: "StartBattle",
          data: {
            battleId: "guard_encounter"
          }
        }
      ]
    };
  } else if (state.suspicious) {
    return {
      effects: [{
        type: "ShowDialog",
        data: {
          text: "I'm watching you...",
          speaker: "Guard"
        }
      }],
      state: {
        patrolIndex: state.patrolIndex,
        patrolPoints: state.patrolPoints,
        suspicious: true
      }
    };
  } else {
    return {
      effects: [{
        type: "ShowDialog",
        data: {
          text: "Move along, citizen.",
          speaker: "Guard"
        }
      }],
      state: {
        patrolIndex: state.patrolIndex,
        patrolPoints: state.patrolPoints,
        suspicious: true
      }
    };
  }
}
