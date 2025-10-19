package scripting

import (
	"encoding/json"
	"fmt"
	"time"
)

// PlanType represents the type of plan node
type PlanType string

const (
	PlanTypeSeq    PlanType = "Seq"    // Sequential execution
	PlanTypePar    PlanType = "Par"    // Parallel execution
	PlanTypeWait   PlanType = "Wait"   // Wait for duration
	PlanTypeIf     PlanType = "If"     // Conditional execution
	PlanTypeChoice PlanType = "Choice" // Player choice
	PlanTypeEffect PlanType = "Effect" // Single effect
)

// Plan represents a multi-step sequence of actions
type Plan struct {
	Type     PlanType               `json:"type"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Children []*Plan                `json:"children,omitempty"`
}

// PlanCursor tracks execution progress through a plan
type PlanCursor struct {
	PlanID      string                 `json:"planId"`
	EntityID    string                 `json:"entityId"`
	CurrentPath []int                  `json:"currentPath"` // Path to current node
	State       map[string]interface{} `json:"state"`       // Execution state
	StartTime   time.Time              `json:"startTime"`
	WaitUntil   *time.Time             `json:"waitUntil,omitempty"`
}

// PlanRunner executes plans over time
type PlanRunner struct {
	activePlans map[string]*PlanCursor
	applier     *EffectApplier
}

// NewPlanRunner creates a new plan runner
func NewPlanRunner(applier *EffectApplier) *PlanRunner {
	return &PlanRunner{
		activePlans: make(map[string]*PlanCursor),
		applier:     applier,
	}
}

// StartPlan begins executing a plan for an entity
func (r *PlanRunner) StartPlan(entityID string, plan *Plan) string {
	planID := fmt.Sprintf("%s_%d", entityID, time.Now().UnixNano())
	cursor := &PlanCursor{
		PlanID:      planID,
		EntityID:    entityID,
		CurrentPath: []int{0},
		State:       make(map[string]interface{}),
		StartTime:   time.Now(),
	}

	r.activePlans[planID] = cursor
	return planID
}

// StopPlan stops a running plan
func (r *PlanRunner) StopPlan(planID string) {
	delete(r.activePlans, planID)
}

// Update advances all active plans
func (r *PlanRunner) Update(plan *Plan, deltaTime float64) ([]Effect, error) {
	allEffects := make([]Effect, 0)

	for planID, cursor := range r.activePlans {
		// Check if waiting
		if cursor.WaitUntil != nil {
			if time.Now().Before(*cursor.WaitUntil) {
				continue
			}
			cursor.WaitUntil = nil
		}

		// Execute current node
		effects, done, err := r.executeNode(plan, cursor)
		if err != nil {
			return nil, fmt.Errorf("plan %s failed: %w", planID, err)
		}

		allEffects = append(allEffects, effects...)

		if done {
			delete(r.activePlans, planID)
		}
	}

	return allEffects, nil
}

func (r *PlanRunner) executeNode(plan *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	node := r.getNodeAtPath(plan, cursor.CurrentPath)
	if node == nil {
		return nil, true, nil // Plan complete
	}

	switch node.Type {
	case PlanTypeEffect:
		return r.executeEffect(node, cursor)

	case PlanTypeWait:
		return r.executeWait(node, cursor)

	case PlanTypeSeq:
		return r.executeSeq(plan, node, cursor)

	case PlanTypePar:
		return r.executePar(plan, node, cursor)

	case PlanTypeIf:
		return r.executeIf(plan, node, cursor)

	case PlanTypeChoice:
		return r.executeChoice(plan, node, cursor)

	default:
		return nil, false, fmt.Errorf("unknown plan type: %s", node.Type)
	}
}

func (r *PlanRunner) executeEffect(node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	// Convert node.Data to Effect
	jsonBytes, err := json.Marshal(node.Data)
	if err != nil {
		return nil, false, err
	}

	var effect Effect
	if err := json.Unmarshal(jsonBytes, &effect); err != nil {
		return nil, false, err
	}

	// Advance to next node
	r.advancePath(cursor)

	return []Effect{effect}, false, nil
}

func (r *PlanRunner) executeWait(node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	duration, ok := node.Data["duration"].(float64)
	if !ok {
		return nil, false, fmt.Errorf("wait node missing duration")
	}

	waitUntil := time.Now().Add(time.Duration(duration * float64(time.Second)))
	cursor.WaitUntil = &waitUntil

	// Advance to next node
	r.advancePath(cursor)

	return nil, false, nil
}

func (r *PlanRunner) executeSeq(plan *Plan, node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	// Get current child index
	childIdx := 0
	if idx, ok := cursor.State["seqIndex"].(float64); ok {
		childIdx = int(idx)
	}

	if childIdx >= len(node.Children) {
		// Sequence complete, advance to next sibling
		r.advancePath(cursor)
		delete(cursor.State, "seqIndex")
		return nil, false, nil
	}

	// Execute current child
	cursor.CurrentPath = append(cursor.CurrentPath, childIdx)
	effects, done, err := r.executeNode(plan, cursor)
	if err != nil {
		return nil, false, err
	}

	if done {
		// Child complete, move to next
		cursor.CurrentPath = cursor.CurrentPath[:len(cursor.CurrentPath)-1]
		cursor.State["seqIndex"] = float64(childIdx + 1)
	}

	return effects, false, nil
}

func (r *PlanRunner) executePar(plan *Plan, node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	// Execute all children in parallel
	allEffects := make([]Effect, 0)
	allDone := true

	for i := range node.Children {
		cursor.CurrentPath = append(cursor.CurrentPath, i)
		effects, done, err := r.executeNode(plan, cursor)
		if err != nil {
			return nil, false, err
		}

		allEffects = append(allEffects, effects...)
		if !done {
			allDone = false
		}

		cursor.CurrentPath = cursor.CurrentPath[:len(cursor.CurrentPath)-1]
	}

	if allDone {
		r.advancePath(cursor)
	}

	return allEffects, false, nil
}

func (r *PlanRunner) executeIf(plan *Plan, node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	// Evaluate condition
	condition, ok := node.Data["condition"].(bool)
	if !ok {
		return nil, false, fmt.Errorf("if node missing condition")
	}

	var childIdx int
	if condition {
		childIdx = 0 // then branch
	} else {
		if len(node.Children) > 1 {
			childIdx = 1 // else branch
		} else {
			// No else branch, skip
			r.advancePath(cursor)
			return nil, false, nil
		}
	}

	cursor.CurrentPath = append(cursor.CurrentPath, childIdx)
	effects, done, err := r.executeNode(plan, cursor)
	if err != nil {
		return nil, false, err
	}

	if done {
		cursor.CurrentPath = cursor.CurrentPath[:len(cursor.CurrentPath)-1]
		r.advancePath(cursor)
	}

	return effects, false, nil
}

func (r *PlanRunner) executeChoice(plan *Plan, node *Plan, cursor *PlanCursor) ([]Effect, bool, error) {
	// Wait for player choice (stored in cursor.State)
	choice, ok := cursor.State["choice"]
	if !ok {
		// Still waiting for choice
		return nil, false, nil
	}

	choiceIdx, ok := choice.(float64)
	if !ok || int(choiceIdx) >= len(node.Children) {
		return nil, false, fmt.Errorf("invalid choice index")
	}

	cursor.CurrentPath = append(cursor.CurrentPath, int(choiceIdx))
	effects, done, err := r.executeNode(plan, cursor)
	if err != nil {
		return nil, false, err
	}

	if done {
		cursor.CurrentPath = cursor.CurrentPath[:len(cursor.CurrentPath)-1]
		r.advancePath(cursor)
		delete(cursor.State, "choice")
	}

	return effects, false, nil
}

func (r *PlanRunner) getNodeAtPath(plan *Plan, path []int) *Plan {
	if len(path) == 0 {
		return plan
	}

	current := plan
	for _, idx := range path {
		if idx >= len(current.Children) {
			return nil
		}
		current = current.Children[idx]
	}
	return current
}

func (r *PlanRunner) advancePath(cursor *PlanCursor) {
	if len(cursor.CurrentPath) == 0 {
		return
	}

	// Increment last index
	cursor.CurrentPath[len(cursor.CurrentPath)-1]++
}

// GetActivePlans returns all active plan cursors
func (r *PlanRunner) GetActivePlans() map[string]*PlanCursor {
	result := make(map[string]*PlanCursor)
	for k, v := range r.activePlans {
		result[k] = v
	}
	return result
}

// SetPlayerChoice sets a player's choice for a Choice node
func (r *PlanRunner) SetPlayerChoice(planID string, choiceIdx int) error {
	cursor, ok := r.activePlans[planID]
	if !ok {
		return fmt.Errorf("plan %s not found", planID)
	}

	cursor.State["choice"] = float64(choiceIdx)
	return nil
}

// Serialize converts active plans to JSON
func (r *PlanRunner) Serialize() ([]byte, error) {
	return json.Marshal(r.activePlans)
}

// Deserialize loads active plans from JSON
func (r *PlanRunner) Deserialize(data []byte) error {
	return json.Unmarshal(data, &r.activePlans)
}
