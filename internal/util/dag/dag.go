package dag

import (
	"fmt"
	"sort"
)

type Pos struct{ X, Y int }

// LayoutIndices lays out a DAG from a children→parents map.
// - X = generation (column), integer
// - Y = row (integer), bottom-left origin but TOP-aligned across columns
// - Depth = number of columns
// - Height = max nodes in any column
func LayoutIndices[T comparable](childrenToParents map[T][]T) (pos map[T]Pos, Depth, Height int) {
	// ----- build nodes -----
	type node struct {
		id       T
		key      string
		Parents  []*node
		Children []*node
		depth    int // column index
		order    int // row within column (0 = top)
	}
	nodes := map[T]*node{}
	get := func(id T) *node {
		if n, ok := nodes[id]; ok {
			return n
		}
		n := &node{id: id, key: fmt.Sprint(id)}
		nodes[id] = n
		return n
	}
	// create all nodes and wire parents
	for c, ps := range childrenToParents {
		nc := get(c)
		for _, p := range ps {
			nc.Parents = append(nc.Parents, get(p))
		}
	}
	// build children directly from parents
	for _, n := range nodes {
		n.Children = nil
	}
	for _, n := range nodes {
		for _, p := range n.Parents {
			p.Children = append(p.Children, n)
		}
	}

	// ----- topo layering: longest path depth -----
	in := map[*node]int{}
	q := []*node{}
	for _, n := range nodes {
		for _, c := range n.Children {
			in[c]++
		}
	}
	for _, n := range nodes {
		if in[n] == 0 {
			q = append(q, n)
		}
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		for _, c := range n.Children {
			if c.depth < n.depth+1 {
				c.depth = n.depth + 1
			}
			in[c]--
			if in[c] == 0 {
				q = append(q, c)
			}
		}
	}

	// ----- bucket by depth -----
	maxd := 0
	for _, n := range nodes {
		if n.depth > maxd {
			maxd = n.depth
		}
	}
	layers := make([][]*node, maxd+1)
	for _, n := range nodes {
		layers[n.depth] = append(layers[n.depth], n)
	}
	Depth = len(layers)
	for _, L := range layers {
		if len(L) > Height {
			Height = len(L)
		}
	}

	// ----- deterministic order, then light crossing reduction -----
	for _, L := range layers {
		sort.SliceStable(L, func(i, j int) bool { return L[i].key < L[j].key })
		for i, n := range L {
			n.order = i
		}
	}
	median := func(a []int) float64 {
		if len(a) == 0 {
			return 0
		}
		sort.Ints(a)
		m := len(a) / 2
		if len(a)%2 == 1 {
			return float64(a[m])
		}
		return float64(a[m-1]+a[m]) / 2
	}
	down := func() {
		for d := 0; d < len(layers)-1; d++ {
			L := layers[d]
			type pair struct {
				n *node
				w float64
			}
			ps := make([]pair, len(L))
			for i, n := range L {
				if len(n.Children) == 0 {
					ps[i] = pair{n, float64(n.order)}
				} else {
					neis := make([]int, len(n.Children))
					for k, c := range n.Children {
						neis[k] = c.order
					}
					ps[i] = pair{n, median(neis)}
				}
			}
			sort.SliceStable(ps, func(i, j int) bool { return ps[i].w < ps[j].w })
			for i, p := range ps {
				layers[d][i] = p.n
				p.n.order = i
			}
		}
	}
	up := func() {
		for d := len(layers) - 1; d >= 1; d-- {
			L := layers[d]
			type pair struct {
				n *node
				w float64
			}
			ps := make([]pair, len(L))
			for i, n := range L {
				if len(n.Parents) == 0 {
					ps[i] = pair{n, float64(n.order)}
				} else {
					neis := make([]int, len(n.Parents))
					for k, p := range n.Parents {
						neis[k] = p.order
					}
					ps[i] = pair{n, median(neis)}
				}
			}
			sort.SliceStable(ps, func(i, j int) bool { return ps[i].w < ps[j].w })
			for i, p := range ps {
				layers[d][i] = p.n
				p.n.order = i
			}
		}
	}
	// a couple of sweeps are enough for clean ordering
	down()
	up()

	// ----- positions: uniform grid, TOP-aligned across columns -----
	pos = make(map[T]Pos, len(nodes))
	for x, L := range layers {
		for _, n := range L {
			// n.order is top-origin (0=top). Put top at global Y = Height-1.
			y := Height - 1 - n.order // bottom-left origin
			pos[n.id] = Pos{X: x, Y: y}
		}
	}
	return
}
