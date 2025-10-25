package adventure

import "fisherevans.com/project/f/internal/game/input"

type EntityPresence interface {
	IsInteractable() bool
	AllowsEgress(side input.Direction, id string) bool
	AllowsIngress(side input.Direction, id string) bool
}

type blockIngressPresence struct {
	isInteractable    bool
	isBlockingIngress bool
}

func newBlockIngressPresence(isInteractable bool) *blockIngressPresence {
	return &blockIngressPresence{
		isInteractable:    isInteractable,
		isBlockingIngress: true,
	}
}

func (p *blockIngressPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *blockIngressPresence) AllowsEgress(side input.Direction, id string) bool {
	return true
}

func (p *blockIngressPresence) AllowsIngress(side input.Direction, id string) bool {
	return !p.isBlockingIngress
}

type directionalPresence struct {
	isInteractable bool
	blockedSides   map[input.Direction]bool
}

func (p *directionalPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *directionalPresence) AllowsEgress(side input.Direction, id string) bool {
	if p.blockedSides == nil {
		return true
	}
	return p.blockedSides[side]
}

func (p *directionalPresence) AllowsIngress(side input.Direction, id string) bool {
	if p.blockedSides == nil {
		return false
	}
	return p.blockedSides[side]
}

type bidirectionalPresence struct {
	isInteractable      bool
	blockedEgressSides  map[input.Direction]bool
	blockedIngressSides map[input.Direction]bool
}

func (p *bidirectionalPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *bidirectionalPresence) AllowsEgress(side input.Direction, id string) bool {
	if p.blockedEgressSides == nil {
		return false
	}
	return p.blockedEgressSides[side]
}

func (p *bidirectionalPresence) AllowsIngress(side input.Direction, id string) bool {
	if p.blockedIngressSides == nil {
		return false
	}
	return p.blockedIngressSides[side]
}
