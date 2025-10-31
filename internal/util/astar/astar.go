package astar

import (
	"container/heap"

	"github.com/rs/zerolog/log"
)

type Pather[T any, C any] interface {
	PathNeighbors(ctx C) []WeightedNeighbor[T]
	PathHeuristic(ctx C, to T) float64
}

type WeightedNeighbor[T any] struct {
	Neighbor T
	Cost     float64
}

type node[T any] struct {
	pather T
	cost   float64
	rank   float64
	parent *node[T]
	open   bool
	closed bool
	index  int
}

type nodeMap[T comparable] map[T]*node[T]

func (nm nodeMap[T]) get(p T) *node[T] {
	n, ok := nm[p]
	if !ok {
		n = &node[T]{
			pather: p,
			cost:   1e9, // Initialize with a large cost (infinity)
		}
		nm[p] = n
	}
	return n
}

type priorityQueue[T any] []*node[T]

func (pq priorityQueue[T]) Len() int {
	return len(pq)
}

func (pq priorityQueue[T]) Less(i, j int) bool {
	return pq[i].rank < pq[j].rank
}

func (pq priorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue[T]) Push(x interface{}) {
	n := x.(*node[T])
	n.index = len(*pq)
	*pq = append(*pq, n)
}

func (pq *priorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func Path[T comparable, C any](from, to T, ctx C, maxNodes int) (path []T, distance float64, found bool) {
	if _, ok := any(from).(Pather[T, C]); !ok {
		return nil, 0, false
	}

	nm := nodeMap[T]{}
	nq := &priorityQueue[T]{}
	heap.Init(nq)
	fromNode := nm.get(from)
	fromNode.cost = 0
	fromNode.open = true
	heap.Push(nq, fromNode)

	nodesChecked := 0
	for {
		if nq.Len() == 0 {
			return nil, 0, false
		}
		current := heap.Pop(nq).(*node[T])
		current.open = false
		current.closed = true

		nodesChecked++
		if maxNodes > 0 && nodesChecked >= maxNodes {
			// Exceeded node limit, path not found
			log.Warn().Int("max_nodes", maxNodes).Msg("exceeded node limit, path not found")
			return nil, 0, false
		}

		if current.pather == to {
			p := []T{}
			curr := current
			for curr != nil {
				p = append(p, curr.pather)
				curr = curr.parent
			}
			return p, current.cost, true
		}

		currentPather := any(current.pather).(Pather[T, C])
		for _, weightedNeighbor := range currentPather.PathNeighbors(ctx) {
			neighbor := weightedNeighbor.Neighbor
			cost := current.cost + weightedNeighbor.Cost
			neighborNode := nm.get(neighbor)

			// Skip if already processed with a better cost
			if neighborNode.closed && cost >= neighborNode.cost {
				continue
			}

			// If we found a better path, update it
			if cost < neighborNode.cost {
				if neighborNode.open {
					heap.Remove(nq, neighborNode.index)
				}
				neighborPather := any(neighbor).(Pather[T, C])
				neighborNode.cost = cost
				neighborNode.rank = cost + neighborPather.PathHeuristic(ctx, to)
				neighborNode.parent = current
				neighborNode.open = true
				neighborNode.closed = false
				heap.Push(nq, neighborNode)
			}
		}
	}
}
