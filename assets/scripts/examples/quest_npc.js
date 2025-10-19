// Example: Quest NPC with Multi-Step Plan
// Demonstrates using plans for complex sequences

const SUBSCRIPTION = {
  Trigger: ["QuestComplete"],
  Priority: 10
};

function OnInit(self, world, state) {
  return {
    state: {
      questGiven: false,
      questComplete: false,
      rewardGiven: false
    }
  };
}

function OnInteract(self, world, state, event) {
  if (!state.questGiven) {
    // Give quest with a multi-step plan
    return {
      effects: [{
        type: "ShowDialog",
        data: {
          text: "I need your help! Please retrieve the ancient artifact from the temple.",
          speaker: "Elder"
        }
      }],
      plan: {
        type: "Seq",
        children: [
          {
            type: "Effect",
            data: {
              type: "SetVar",
              data: {
                key: "global.quests.ancient_artifact.active",
                value: true
              }
            }
          },
          {
            type: "Wait",
            data: { duration: 2.0 }
          },
          {
            type: "Effect",
            data: {
              type: "ShowDialog",
              data: {
                text: "Be careful out there!",
                speaker: "Elder"
              }
            }
          }
        ]
      },
      state: {
        questGiven: true,
        questComplete: false,
        rewardGiven: false
      }
    };
  } else if (state.questGiven && !state.questComplete) {
    // Quest in progress
    const hasArtifact = world.getVar("global.inventory.ancient_artifact");
    
    if (hasArtifact) {
      return {
        effects: [{
          type: "Trigger",
          data: {
            triggerName: "QuestComplete",
            targetEntity: self.id
          }
        }]
      };
    } else {
      return {
        effects: [{
          type: "ShowDialog",
          data: {
            text: "Have you found the artifact yet?",
            speaker: "Elder"
          }
        }]
      };
    }
  } else if (state.questComplete && !state.rewardGiven) {
    // Give reward
    return {
      plan: {
        type: "Seq",
        children: [
          {
            type: "Effect",
            data: {
              type: "ShowDialog",
              data: {
                text: "You found it! Thank you, brave adventurer!",
                speaker: "Elder"
              }
            }
          },
          {
            type: "Wait",
            data: { duration: 1.5 }
          },
          {
            type: "Effect",
            data: {
              type: "ShowDialog",
              data: {
                text: "Please accept this reward.",
                speaker: "Elder"
              }
            }
          },
          {
            type: "Effect",
            data: {
              type: "SetVar",
              data: {
                key: "global.inventory.gold",
                value: 100
              }
            }
          }
        ]
      },
      state: {
        questGiven: true,
        questComplete: true,
        rewardGiven: true
      }
    };
  } else {
    // Quest complete
    return {
      effects: [{
        type: "ShowDialog",
        data: {
          text: "Thank you again for your help!",
          speaker: "Elder"
        }
      }]
    };
  }
}

function OnTrigger(self, world, state, event) {
  if (event.triggerName === "QuestComplete") {
    return {
      state: {
        questGiven: true,
        questComplete: true,
        rewardGiven: false
      }
    };
  }
}
