package xenolog

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/dag"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type skillNodeState int

const (
	skillNodeStateHidden skillNodeState = iota
	skillNodeStateUnlocked
	skillNodeStateUnlockable
)

var (
	skillNodeSpacingX                    = 48
	skillNodeSpacingY                    = 24
	skillNodeSpriteHidden                = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 4, 1)
	skillNodeSpriteUnlockedHighlighted   = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 3, 1)
	skillNodeSpriteUnlockableHighlighted = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 2, 1)
	skillNodeSpriteHighlightCursor       = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 5, 1)
	skillNodeSpriteUnlocked              = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 6, 1)
	skillNodeSpriteUnlockable            = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 7, 1)
)

type skillTree struct {
	menu          *primortalDetailMenu
	nodes         map[rpg.SkillId]*skillNode
	selectedSkill rpg.SkillId
	frame         pixel.Rect
}

func newSkillTree(primortal rpg.PrimortalType, m *primortalDetailMenu) *skillTree {
	t := &skillTree{
		menu: m,
	}
	t.nodes = buildSkillTreeNodes(t, primortal.Primortal())
	t.selectInitialSkill()
	return t
}

func (t *skillTree) handleInput(dir input.Direction) bool {
	if len(t.nodes) == 0 || t.selectedSkill == "" {
		return false
	}
	cur, ok := t.nodes[t.selectedSkill]
	if !ok {
		log.Warn().Msgf("selected skill %s not found in tree", t.selectedSkill)
		return false
	}

	type candidate struct {
		id                 rpg.SkillId
		primary, secondary int
	}
	best := candidate{"", int(^uint(0) >> 1), int(^uint(0) >> 1)}

	cx, cy := cur.x, cur.y

	inDir := func(dx, dy int) bool { return true }
	primary := func(dx, dy int) int { return 0 }
	secondary := func(dx, dy int) int { return 0 }

	switch dir {
	case input.Up:
		inDir = func(dx, dy int) bool { return dy > 0 }
		primary = func(dx, dy int) int { return dy }
		secondary = func(dx, dy int) int { return abs(dx) }
	case input.Down:
		inDir = func(dx, dy int) bool { return dy < 0 }
		primary = func(dx, dy int) int { return abs(dy) }
		secondary = func(dx, dy int) int { return abs(dx) }
	case input.Left:
		inDir = func(dx, dy int) bool { return dx < 0 }
		primary = func(dx, dy int) int { return abs(dx) }
		secondary = func(dx, dy int) int { return abs(dy) }
	case input.Right:
		inDir = func(dx, dy int) bool { return dx > 0 }
		primary = func(dx, dy int) int { return abs(dx) }
		secondary = func(dx, dy int) int { return abs(dy) }
	default:
		return false
	}

	for id, n := range t.nodes {
		if id == t.selectedSkill {
			continue
		}
		dx, dy := n.x-cx, n.y-cy
		if !inDir(dx, dy) {
			continue
		}
		p := primary(dx, dy)
		s := secondary(dx, dy)
		if p < best.primary || (p == best.primary && s < best.secondary) || (p == best.primary && s == best.secondary && fmt.Sprint(id) < fmt.Sprint(best.id)) {
			best = candidate{id: id, primary: p, secondary: s}
		}
	}

	if best.id == "" {
		return false
	}
	t.selectedSkill = best.id
	return true
}

func (t *skillTree) renderTree(target pixel.Target, treeOrigin pixel.Matrix, timeDelta float64) {
	for _, node := range t.nodes {
		node.RenderEdges(target, treeOrigin)
	}
	for _, node := range t.nodes {
		node.RenderNode(target, treeOrigin, timeDelta)
	}
}

func (t *skillTree) selectInitialSkill() {
	minX, maxY := -1, -1
	for s, node := range t.nodes {
		if node.x < minX || minX == -1 {
			minX = node.x
			maxY = node.y
			t.selectedSkill = s
		} else if node.x == minX && node.y > maxY {
			maxY = node.y
			t.selectedSkill = s
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func buildSkillTreeNodes(t *skillTree, primortal rpg.Primortal) map[rpg.SkillId]*skillNode {
	childrenToParents := make(map[rpg.SkillId][]rpg.SkillId, len(primortal.UnlockableSkills))
	nodes := make(map[rpg.SkillId]*skillNode, len(primortal.UnlockableSkills))

	// Pass 1: create nodes, record parent IDs, and fill the DAG map once.
	for skillId, us := range primortal.UnlockableSkills {
		n := &skillNode{
			tree:    t,
			id:      skillId,
			parents: make(map[rpg.SkillId]*skillNode),
			animations: []*skillNodeAnimation{
				newSkillNodeAnimation("slow", colors.XenoLogDark.RGBA),
				newSkillNodeAnimation("medium", colors.XenoLogHighlight.RGBA),
				newSkillNodeAnimation("fast", colors.XenoLogText.RGBA),
			},
		}
		nodes[skillId] = n
		childrenToParents[skillId] = append([]rpg.SkillId(nil), us.Prerequisites...)
	}

	log.Info().Msgf("Skill tree children to parents: %v", childrenToParents)
	positions, _, _ := dag.LayoutIndices(childrenToParents)
	log.Info().Msgf("Skill tree positions: %v", positions)

	// Pass 2: set positions and link parent pointers.
	t.frame = pixel.R(0, 0, 0, 0)
	for id, n := range nodes {
		p := positions[id]
		n.x = p.X * skillNodeSpacingX
		n.y = p.Y * skillNodeSpacingY
		for _, pid := range childrenToParents[id] {
			if pNode, ok := nodes[pid]; ok {
				n.parents[pid] = pNode
			}
		}
		t.frame.Min.X = math.Min(t.frame.Min.X, float64(n.x))
		t.frame.Min.Y = math.Min(t.frame.Min.Y, float64(n.y))
		t.frame.Max.X = math.Max(t.frame.Max.X, float64(n.x))
		t.frame.Max.Y = math.Max(t.frame.Max.Y, float64(n.y))
	}
	return nodes
}

type skillNode struct {
	tree    *skillTree
	id      rpg.SkillId
	x, y    int
	parents map[rpg.SkillId]*skillNode

	animations []*skillNodeAnimation
}

type skillNodeAnimation struct {
	animation *anim.AnimatedSprite
	mask      pixel.RGBA
}

func newSkillNodeAnimation(speed string, mask pixel.RGBA) *skillNodeAnimation {
	a := &skillNodeAnimation{
		animation: skillNodeAnimUnlockable(speed),
		mask:      mask,
	}
	a.animation.Update(rand.Float64() * 1000)
	return a
}

func (n *skillNode) RenderEdges(target pixel.Target, treeOrigin pixel.Matrix) {
	lineMask := colors.XenoLogDark.RGBA
	for _, parent := range n.parents {
		from := treeOrigin.Project(pixel.V(float64(parent.x), float64(parent.y)))
		to := treeOrigin.Project(pixel.V(float64(n.x), float64(n.y)))
		gfx.DrawLine(target, atlas, from, to, 1, lineMask)
	}
}

func (n *skillNode) RenderNode(target pixel.Target, treeOrigin pixel.Matrix, timeDelta float64) {
	state := n.getState()
	isHighlighted := n.tree.selectedSkill == n.id
	matrix := treeOrigin.Moved(gfx.IVec(n.x, n.y))
	mask := colors.XenoLogText.RGBA

	us := n.tree.menu.primortal.Primortal().UnlockableSkills[n.id]
	progress, hasProgress := game.CurrentSave().Primortals[n.tree.menu.primortal]
	canUnlock := hasProgress && state == skillNodeStateUnlockable && us.Cost <= progress.ResearchPoints

	if canUnlock {
		mask = flashingHighlight()
		for _, a := range n.animations {
			a.animation.Update(timeDelta)
			a.animation.Sprite().DrawColorMask(target, matrix, a.mask)
		}
	}

	if isHighlighted {
		skillNodeSpriteHighlightCursor.DrawColorMask(target, matrix, colors.XenoLogHighlight.RGBA)
	}

	skillNodeSprite(state, isHighlighted).DrawColorMask(target, matrix, mask)

	var label string
	if state == skillNodeStateUnlockable {
		label = fmt.Sprintf("%d rp", us.Cost)
	} else if isHighlighted {
		if state == skillNodeStateUnlocked {
			label = "unlocked"
		} else {
			label = "blocked"
		}
	}
	if label != "" {
		txt := newTextRenderer(target, smallTextbox)
		txt.matrix = treeOrigin
		txt.render("{+o:xenolog_clear}"+label, n.x, n.y+6, mask, tbcfg.RenderFrom(gfx.BottomCenter))
	}
}

func (n *skillNode) getState() skillNodeState {
	if game.CurrentSave().IsSkillUnlocked(n.id) {
		return skillNodeStateUnlocked
	} else if len(n.parents) == 0 {
		return skillNodeStateUnlockable
	} else {
		for parent := range n.parents {
			if !game.CurrentSave().IsSkillUnlocked(parent) {
				return skillNodeStateHidden
			}
		}
	}
	return skillNodeStateUnlockable
}

func skillNodeSprite(state skillNodeState, isHighlighted bool) pixelutil.BoundedDrawable {
	switch state {
	case skillNodeStateHidden:
		return skillNodeSpriteHidden
	case skillNodeStateUnlocked:
		if isHighlighted {
			return skillNodeSpriteUnlockedHighlighted
		}
		return skillNodeSpriteUnlocked
	case skillNodeStateUnlockable:
		if isHighlighted {
			return skillNodeSpriteUnlockableHighlighted
		}
		return skillNodeSpriteUnlockable
	}
	return nil
}

func skillNodeAnimUnlockable(speed string) *anim.AnimatedSprite {
	return anim.Load(atlas, "xenolog/skill_tree/nodes", fmt.Sprintf("spinning_%s", speed))
}
