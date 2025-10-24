package adventure

// Helper functions for wiring up entities with movement and behavior

// RegisterNPC registers an NPC entity with movement controller and behavior
func (s *State) RegisterNPC(npc *NPC, location MapLocation, speeds map[MoveState]float64, doesMove bool, horizOnly bool) {
	// Register movement
	movementState := s.movementController.Register(npc.GetEntityId(), location, speeds)
	npc.SetMovementState(movementState)
	
	// Register behavior
	behavior := NewNPCBehavior(npc.GetEntityId(), doesMove, horizOnly)
	s.behaviors[npc.GetEntityId()] = behavior
	
	// Add to entity list
	s.AddEntity(npc)
}

// RegisterPlayer registers the player entity with movement controller and behavior
func (s *State) RegisterPlayer(player *Player, location MapLocation, speeds map[MoveState]float64) {
	// Register movement
	movementState := s.movementController.Register(player.GetEntityId(), location, speeds)
	player.SetMovementState(movementState)
	
	// Register behavior
	behavior := NewPlayerBehavior(player.GetEntityId())
	s.behaviors[player.GetEntityId()] = behavior
	
	// Add to entity list
	s.AddEntity(player)
}
