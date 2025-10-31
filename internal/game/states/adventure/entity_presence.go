package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

type EntityPresence interface {
	PathfindingImpedance
	IsInteractable() bool
	AllowsEgress(side input.Direction, id string) bool
	AllowsIngress(side input.Direction, id string) bool
}

func AttachPresenceFromConfig(entity Entity, properties *util.Properties) EntityPresence {
	cfg := &types.PresenceConfig{}
	impedanceWeight := ImpedanceExtreme
	if !properties.LoadStructFromKey("presence_config", cfg) {
		return AttachBlockIngressPresence(entity, true, NewStaticImpedance(impedanceWeight))
	}
	isInteractable, blockIngress := false, false
	if cfg.IsInteractable != nil {
		isInteractable = *cfg.IsInteractable
	}
	if cfg.BlockIngress != nil {
		blockIngress = *cfg.BlockIngress
	}
	if cfg.Impedance != nil {
		impedanceWeight = *cfg.Impedance
	}
	p := AttachBlockIngressPresence(entity, isInteractable, NewStaticImpedance(impedanceWeight))
	p.isBlockingIngress = blockIngress
	entity.SetPresence(p)
	return p
}

type BlockIngressPresence struct {
	PathfindingImpedance
	isInteractable    bool
	isBlockingIngress bool
}

func AttachBlockIngressPresence(entity Entity, isInteractable bool, impedance PathfindingImpedance) *BlockIngressPresence {
	presence := &BlockIngressPresence{
		PathfindingImpedance: impedance,
		isInteractable:       isInteractable,
		isBlockingIngress:    true,
	}
	entity.SetPresence(presence)
	return presence
}

func (p *BlockIngressPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *BlockIngressPresence) AllowsEgress(side input.Direction, id string) bool {
	return true
}

func (p *BlockIngressPresence) AllowsIngress(side input.Direction, id string) bool {
	return !p.isBlockingIngress
}

type DirectionalPresence struct {
	PathfindingImpedance
	isInteractable bool
	blockedSides   map[input.Direction]bool
}

func AttachDirectionalPresence(entity Entity, isInteractable bool, impedance PathfindingImpedance, blockedSides map[input.Direction]bool) *DirectionalPresence {
	presence := &DirectionalPresence{
		PathfindingImpedance: impedance,
		isInteractable:       isInteractable,
		blockedSides:         blockedSides,
	}
	entity.SetPresence(presence)
	return presence
}

func (p *DirectionalPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *DirectionalPresence) AllowsEgress(side input.Direction, id string) bool {
	if p.blockedSides == nil {
		return true
	}
	return p.blockedSides[side]
}

func (p *DirectionalPresence) AllowsIngress(side input.Direction, id string) bool {
	if p.blockedSides == nil {
		return false
	}
	return p.blockedSides[side]
}

type BidirectionalPresence struct {
	PathfindingImpedance
	isInteractable      bool
	blockedEgressSides  map[input.Direction]bool
	blockedIngressSides map[input.Direction]bool
}

func AttachBidirectionalPresence(entity Entity, isInteractable bool, impedance PathfindingImpedance, blockedEgressSides map[input.Direction]bool, blockedIngressSides map[input.Direction]bool) *BidirectionalPresence {
	presence := &BidirectionalPresence{
		PathfindingImpedance: impedance,
		isInteractable:       isInteractable,
		blockedEgressSides:   blockedEgressSides,
		blockedIngressSides:  blockedIngressSides,
	}
	entity.SetPresence(presence)
	return presence
}

func (p *BidirectionalPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *BidirectionalPresence) AllowsEgress(side input.Direction, id string) bool {
	if p.blockedEgressSides == nil {
		return false
	}
	return p.blockedEgressSides[side]
}

func (p *BidirectionalPresence) AllowsIngress(side input.Direction, id string) bool {
	if p.blockedIngressSides == nil {
		return false
	}
	return p.blockedIngressSides[side]
}

type ConditionalBlockIngressPresence struct {
	PathfindingImpedance
	isInteractable bool
	doBlock        func(id string) bool
}

func AttachConditionalBlockIngressPresence(entity Entity, isInteractable bool, impedance PathfindingImpedance, doBlock func(id string) bool) *ConditionalBlockIngressPresence {
	presence := &ConditionalBlockIngressPresence{
		PathfindingImpedance: impedance,
		isInteractable:       isInteractable,
		doBlock:              doBlock,
	}
	entity.SetPresence(presence)
	return presence
}

func (p *ConditionalBlockIngressPresence) IsInteractable() bool {
	return p.isInteractable
}

func (p *ConditionalBlockIngressPresence) AllowsEgress(side input.Direction, id string) bool {
	return true
}

func (p *ConditionalBlockIngressPresence) AllowsIngress(side input.Direction, id string) bool {
	return !p.doBlock(id)
}
