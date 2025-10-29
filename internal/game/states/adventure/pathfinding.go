package adventure

import (
	"fmt"
	"math"
	"slices"
	"time"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/astar"
	"github.com/rs/zerolog/log"
)

type Path struct {
	PathFound     bool
	Tiles         []MapLocation
	TotalDistance int
}

type pathfindingContext struct {
	entityId string
	system   *EntitySystem
}

func (l MapLocation) PathNeighbors(ctx pathfindingContext) []MapLocation {
	var neighbors []MapLocation
	for _, d := range input.Directions {
		neighbor := l.Moved(d)
		if !ctx.system.isValidTransition(ctx.entityId, l, d, false) {
			continue
		}
		if !ctx.system.isValidTransition(ctx.entityId, neighbor, d, true) {
			continue
		}
		neighbors = append(neighbors, neighbor)
	}
	return neighbors
}

func (l MapLocation) PathNeighborCost(ctx pathfindingContext, to MapLocation) float64 {
	return 1 // todo - add suggested paths for npcs (i.e. middle of a hall)
}

func (l MapLocation) PathEstimatedCost(ctx pathfindingContext, to MapLocation) float64 {
	dx := l.X - to.X
	dy := l.Y - to.Y
	return math.Abs(float64(dx)) + math.Abs(float64(dy))
}

func (es *EntitySystem) FindPath(from, to MapLocation, entityId string) Path {
	start := time.Now()

	ctx := pathfindingContext{
		entityId: entityId,
		system:   es,
	}

	pathers, distance, found := astar.Path(from, to, ctx)
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
	slices.Reverse(path.Tiles)  // astar library returns reverse order slice (to > from)
	path.Tiles = path.Tiles[1:] // it also includes the initial location

	dur := time.Since(start)
	if dur > time.Millisecond {
		log.Warn().Str("elapsed", fmt.Sprintf("%dus", dur.Microseconds())).
			Any("from", from).
			Any("to", to).
			Int("direct_distance", int(from.PathEstimatedCost(ctx, to))).
			Int("path_distance", path.TotalDistance).
			Bool("found", path.PathFound).
			Msgf("pathfinding took a long time")
	}
	return path
}
