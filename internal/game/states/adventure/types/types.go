package types

type MoveState int

const (
	MoveStateIdle MoveState = iota
	MoveStateWalking
	MoveStateRunning
	MoveStateDashing
)

const (
	MetadataKeyMode      = "mode"
	MetadataKeyIsTalking = "isTalking"
)
