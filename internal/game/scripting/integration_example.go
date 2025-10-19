package scripting

import (
	"fmt"
	"os"
)

// IntegrationExample demonstrates how to integrate the scripting system with the game
func IntegrationExample() {
	fmt.Println("=== Scripting System Integration Example ===")

	// 1. Initialize the system
	fmt.Println("1. Initializing scripting system...")
	registry := NewScriptRegistry()
	worldVars := NewWorldVarStore()
	dispatcher := NewEventDispatcher(registry, worldVars)
	saveSystem := NewSaveSystem(dispatcher, worldVars)

	// 2. Load and register scripts
	fmt.Println("2. Registering NPC scripts...")
	
	gatekeeperScript := `
		const SUBSCRIPTION = { Priority: 10 };
		
		function OnInit(self, world, state) {
			return { state: { greeted: false, visits: 0 } };
		}
		
		function OnInteract(self, world, state, event) {
			const visits = (state.visits || 0) + 1;
			
			if (visits >= 3) {
				return {
					effects: [
						{
							type: "SetVar",
							data: { key: "map.meadow.north_gate.open", value: true }
						},
						{
							type: "OpenDoor",
							data: { doorId: "north_gate" }
						}
					],
					state: { greeted: true, visits: visits }
				};
			}
			
			return {
				effects: [{
					type: "ShowDialog",
					data: { text: "Visit " + visits + " of 3", speaker: "Gatekeeper" }
				}],
				state: { greeted: true, visits: visits }
			};
		}
	`

	module, err := registry.Register("npc_gatekeeper", gatekeeperScript)
	if err != nil {
		fmt.Printf("Error registering script: %v\n", err)
		return
	}
	fmt.Printf("   ✓ Registered 'npc_gatekeeper' (handlers: %v)\n", getHandlerNames(module.Handlers))

	// 3. Attach scripts to entities
	fmt.Println("\n3. Attaching scripts to entities...")
	err = dispatcher.AttachScript("guard_001", "npc_gatekeeper", nil)
	if err != nil {
		fmt.Printf("Error attaching script: %v\n", err)
		return
	}
	fmt.Println("   ✓ Attached script to entity 'guard_001'")

	// 4. Dispatch Init event
	fmt.Println("\n4. Dispatching Init event...")
	initEvent := NewInitEvent("guard_001")
	err = dispatcher.Dispatch(initEvent)
	if err != nil {
		fmt.Printf("Error dispatching init: %v\n", err)
		return
	}
	
	state, _ := dispatcher.GetEntityState("guard_001")
	fmt.Printf("   ✓ Entity initialized: %+v\n", state.Data)

	// 5. Simulate player interactions
	fmt.Println("\n5. Simulating player interactions...")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n   Interaction #%d:\n", i)
		interactEvent := NewInteractEvent("guard_001", "player")
		err = dispatcher.Dispatch(interactEvent)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		// Show trace
		trace := dispatcher.GetTrace()
		summary := trace.GetSummary()
		fmt.Printf("   - Handlers executed: %d\n", summary["handlerCount"])
		fmt.Printf("   - Conflicts detected: %d\n", summary["conflictCount"])

		// Show state
		state, _ := dispatcher.GetEntityState("guard_001")
		fmt.Printf("   - Entity state: visits=%v\n", state.Data["visits"])

		// Show world vars
		gateOpen := worldVars.GetBool("map.meadow.north_gate.open")
		fmt.Printf("   - Gate open: %v\n", gateOpen)
	}

	// 6. Demonstrate world variable system
	fmt.Println("\n6. Testing world variable system...")
	worldVars.Set("global.power.grid_online", true)
	worldVars.Set("map.meadow.chest.looted", false)
	worldVars.Set("topic.security.alarm_level", 2)

	fmt.Println("   ✓ Set world variables:")
	for key, value := range worldVars.GetAll() {
		fmt.Printf("     - %s = %v\n", key, value)
	}

	// 7. Test variable listeners
	fmt.Println("\n7. Testing variable change listeners...")
	worldVars.AddListener("topic.security.*", func(key string, oldValue, newValue interface{}) {
		fmt.Printf("   ! Variable changed: %s (%v → %v)\n", key, oldValue, newValue)
	})

	worldVars.Set("topic.security.alarm_level", 3)

	// 8. Create and serialize a save file
	fmt.Println("\n8. Creating save file...")
	save, err := saveSystem.Save()
	if err != nil {
		fmt.Printf("Error creating save: %v\n", err)
		return
	}

	fmt.Printf("   ✓ Save created:\n")
	fmt.Printf("     - Entities: %d\n", len(save.Entities))
	fmt.Printf("     - World vars: %d\n", len(save.World.Vars))
	fmt.Printf("     - Active plans: %d\n", len(save.Plans))

	jsonData, err := saveSystem.SerializeToJSON(save)
	if err != nil {
		fmt.Printf("Error serializing save: %v\n", err)
		return
	}

	fmt.Printf("\n   Save JSON preview:\n%s\n", string(jsonData[:min(len(jsonData), 500)]))

	// 9. Demonstrate plan execution
	fmt.Println("\n9. Testing plan execution...")
	
	planScript := `
		function OnInteract(self, world, state, event) {
			return {
				plan: {
					type: "Seq",
					children: [
						{
							type: "Effect",
							data: {
								type: "ShowDialog",
								data: { text: "Step 1: Starting sequence..." }
							}
						},
						{
							type: "Wait",
							data: { duration: 1.0 }
						},
						{
							type: "Effect",
							data: {
								type: "ShowDialog",
								data: { text: "Step 2: After delay..." }
							}
						}
					]
				}
			};
		}
	`

	registry.Register("plan_test", planScript)
	dispatcher.AttachScript("npc_002", "plan_test", nil)

	planEvent := NewInteractEvent("npc_002", "player")
	err = dispatcher.Dispatch(planEvent)
	if err != nil {
		fmt.Printf("Error dispatching plan event: %v\n", err)
		return
	}

	activePlans := dispatcher.GetPlanRunner().GetActivePlans()
	fmt.Printf("   ✓ Active plans: %d\n", len(activePlans))
	for planID, cursor := range activePlans {
		fmt.Printf("     - Plan %s: entity=%s, path=%v\n", planID, cursor.EntityID, cursor.CurrentPath)
	}

	// 10. Summary
	fmt.Println("\n=== Integration Example Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("  ✓ Script registration and compilation")
	fmt.Println("  ✓ Event dispatching (Init, Interact)")
	fmt.Println("  ✓ Effect validation and application")
	fmt.Println("  ✓ World variable management")
	fmt.Println("  ✓ State persistence")
	fmt.Println("  ✓ Save/load system")
	fmt.Println("  ✓ Plan execution")
	fmt.Println("  ✓ Conflict detection")
	fmt.Println("  ✓ Event tracing")
}

// RunIntegrationExample is a convenience function to run the example
func RunIntegrationExample() {
	IntegrationExample()
}

// SaveIntegrationExampleToFile saves the example output to a file
func SaveIntegrationExampleToFile(filename string) error {
	// Redirect stdout to file
	oldStdout := os.Stdout
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	os.Stdout = file
	IntegrationExample()
	os.Stdout = oldStdout

	return nil
}

func getHandlerNames(handlers map[EventType]bool) []string {
	names := make([]string, 0, len(handlers))
	for eventType, exists := range handlers {
		if exists {
			names = append(names, string(eventType))
		}
	}
	return names
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
