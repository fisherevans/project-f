package adventure

import (
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"time"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/astar"
	"github.com/rs/zerolog/log"
)

const (
	MaxPathfindingNodes = 100000
	PathJitterAmount    = 0.1
)

type Path struct {
	PathFound     bool
	Tiles         []MapLocation
	TotalDistance int
}

type pathfindingContext struct {
	entity Entity
}

func (l MapLocation) PathNeighbors(ctx pathfindingContext) []astar.WeightedNeighbor[MapLocation] {
	var neighbors []astar.WeightedNeighbor[MapLocation]
	for _, d := range input.Directions {
		neighbor := l.Moved(d)
		validEgress, egressImpedance := ctx.entity.GetSystem().isValidTransition(ctx.entity, l, d, false)
		validIngress, ingressImpedance := ctx.entity.GetSystem().isValidTransition(ctx.entity, l, d, true)
		cost := egressImpedance + ingressImpedance
		if (!validIngress || !validEgress) && cost >= ImpedanceImpassable {
			continue
		}
		cost += getPathJitter(ctx.entity.GetId(), neighbor)
		neighbors = append(neighbors, astar.WeightedNeighbor[MapLocation]{
			Neighbor: neighbor,
			Cost:     cost,
		})
	}
	return neighbors
}

// getPathJitter returns a deterministic jitter value in range [-PathJitterAmount, +PathJitterAmount]
// based on entity ID and location. This ensures each NPC has consistent but unique path preferences.
func getPathJitter(entityId string, loc MapLocation) float64 {
	h := fnv.New64a()
	h.Write([]byte(entityId))
	h.Write([]byte(fmt.Sprintf("%d,%d", loc.X, loc.Y)))
	hash := h.Sum64()
	// Map hash to [-1, 1] range
	normalized := float64(hash%10000) / 10000.0      // [0, 1)
	return (normalized*2.0 - 1.0) * PathJitterAmount // [-PathJitterAmount, +PathJitterAmount]
}

func (l MapLocation) PathHeuristic(ctx pathfindingContext, to MapLocation) float64 {
	dx := l.X - to.X
	dy := l.Y - to.Y
	return (math.Abs(float64(dx)) + math.Abs(float64(dy))) * ImpedanceBase
}

func (es *EntitySystem) FindPath(from, to MapLocation, entity Entity) Path {
	start := time.Now()

	ctx := pathfindingContext{
		entity: entity,
	}

	pathers, distance, found := astar.Path(from, to, ctx, MaxPathfindingNodes)
	if !found {
		return Path{
			PathFound: false,
		}
	}
	path := Path{
		PathFound:     true,
		Tiles:         nil,
		TotalDistance: int(distance),
	}
	path.Tiles = pathers
	slices.Reverse(path.Tiles) // astar library returns reverse order slice (to > from)

	dur := time.Since(start)
	if dur > time.Millisecond {
		log.Warn().Str("elapsed", fmt.Sprintf("%dus", dur.Microseconds())).
			Any("from", from).
			Any("to", to).
			Int("direct_distance", int(from.PathHeuristic(ctx, to))).
			Int("path_distance", path.TotalDistance).
			Bool("found", path.PathFound).
			Msgf("pathfinding took a long time")
	}
	return path
}
