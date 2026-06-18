package algo

import (
	"lem-in/internal/graph"
	"math"
)

type edge struct {
	to       int
	capacity int
	flow     int
	rev      int
}

type flowGraph struct {
	n      int
	edges  [][]edge
	s, t   int
	nodeID map[string]int
	idNode map[int]string
}

func newFlowGraph(roomCount int) *flowGraph {
	return &flowGraph{
		n:      roomCount * 2,
		edges:  make([][]edge, roomCount*2),
		nodeID: make(map[string]int),
		idNode: make(map[int]string),
	}
}

func (fg *flowGraph) addEdge(from, to, cap int) {
	fg.edges[from] = append(fg.edges[from], edge{to, cap, 0, len(fg.edges[to])})
	fg.edges[to] = append(fg.edges[to], edge{from, 0, 0, len(fg.edges[from]) - 1})
}

func (fg *flowGraph) bfs(parent []int) bool {
	for i := range parent {
		parent[i] = -1
	}
	parent[fg.s] = -2
	q := []int{fg.s}
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		for _, e := range fg.edges[v] {
			if parent[e.to] == -1 && e.capacity-e.flow > 0 {
				parent[e.to] = v
				if e.to == fg.t {
					return true
				}
				q = append(q, e.to)
			}
		}
	}
	return false
}

func (fg *flowGraph) pathExists() bool {
	visited := make([]bool, fg.n)
	var dfs func(v int) bool
	dfs = func(v int) bool {
		if v == fg.t {
			return true
		}
		visited[v] = true
		for _, e := range fg.edges[v] {
			if e.flow > 0 && !visited[e.to] {
				if dfs(e.to) {
					return true
				}
			}
		}
		return false
	}
	return dfs(fg.s)
}

func FindPaths(g *graph.Graph) [][]string {
	if g.Start == g.End || g.Rooms[g.Start] == nil || g.Rooms[g.End] == nil {
		return nil
	}

	roomCount := len(g.Rooms)
	fg := newFlowGraph(roomCount)

	idx := 0
	for name := range g.Rooms {
		fg.nodeID[name] = idx
		fg.idNode[idx] = name
		idx++
	}

	// Node splitting: in -> out
	for name := range g.Rooms {
		inNode := fg.nodeID[name]
		outNode := inNode + roomCount
		if name == g.Start || name == g.End {
			fg.addEdge(inNode, outNode, math.MaxInt32)
		} else {
			fg.addEdge(inNode, outNode, 1)
		}
	}

	// Tunnel edges: out -> neighbor_in
	for a := range g.Rooms {
		for _, b := range g.Links[a] {
			aOut := fg.nodeID[a] + roomCount
			bIn := fg.nodeID[b]
			fg.addEdge(aOut, bIn, 1)
		}
	}

	fg.s = fg.nodeID[g.Start]
	fg.t = fg.nodeID[g.End] + roomCount

	// Edmonds-Karp: find augmenting paths
	parent := make([]int, fg.n)
	for fg.bfs(parent) {
		v := fg.t
		for v != fg.s {
			u := parent[v]
			for i := range fg.edges[u] {
				if fg.edges[u][i].to == v && fg.edges[u][i].capacity-fg.edges[u][i].flow > 0 {
					fg.edges[u][i].flow++
					fg.edges[v][fg.edges[u][i].rev].flow--
					break
				}
			}
			v = u
		}
	}

	var paths [][]string
	for {
		path := extractPath(fg, g.Start, g.End, roomCount)
		if path == nil {
			break
		}
		paths = append(paths, path)
	}

	return paths
}

func extractPath(fg *flowGraph, start, end string, roomCount int) []string {
	startID := fg.nodeID[start]
	endID := fg.nodeID[end]
	target := endID + roomCount

	visited := make([]bool, fg.n)
	parent := make([]int, fg.n)
	for i := range parent {
		parent[i] = -1
	}

	var dfs func(v int) bool
	dfs = func(v int) bool {
		if v == target {
			return true
		}
		visited[v] = true
		for _, e := range fg.edges[v] {
			if e.flow > 0 && !visited[e.to] {
				parent[e.to] = v
				if dfs(e.to) {
					return true
				}
			}
		}
		return false
	}

	if !dfs(startID) {
		return nil
	}

	var revRooms []string
	v := target
	for v != startID {
		u := parent[v]
		for i := range fg.edges[u] {
			if fg.edges[u][i].to == v && fg.edges[u][i].flow > 0 {
				fg.edges[u][i].flow--
				fg.edges[v][fg.edges[u][i].rev].flow++
				break
			}
		}
		if v >= roomCount {
			name := fg.idNode[v-roomCount]
			if name != "" {
				revRooms = append(revRooms, name)
			}
		}
		v = u
	}
	revRooms = append(revRooms, start)

	for i, j := 0, len(revRooms)-1; i < j; i, j = i+1, j-1 {
		revRooms[i], revRooms[j] = revRooms[j], revRooms[i]
	}

	var filtered []string
	for _, room := range revRooms {
		if len(filtered) == 0 || filtered[len(filtered)-1] != room {
			filtered = append(filtered, room)
		}
	}

	return filtered
}
