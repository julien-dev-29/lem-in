package solver

import (
	"lem-in/internal/graph"
	"math"
)

type Turn struct {
	Moves []Move
}

type Move struct {
	Ant  int
	Room string
}

type AntStatus struct {
	PathIndex int
	Step      int
	Finished  bool
}

func Solve(g *graph.Graph, paths [][]string) []Turn {
	if len(paths) == 0 {
		return nil
	}

	bestPaths := selectBestPathSet(g.AntCount, paths)
	if len(bestPaths) == 0 {
		bestPaths = paths
	}

	ants := distributeAnts(g.AntCount, bestPaths)

	return simulate(ants, bestPaths)
}

func selectBestPathSet(antCount int, paths [][]string) [][]string {
	if len(paths) == 1 {
		return paths
	}

	bestTurns := math.MaxInt32
	bestIdx := 0

	for k := 1; k <= len(paths); k++ {
		subset := paths[:k]
		turns := computeTotalTurns(antCount, subset)
		if turns < bestTurns {
			bestTurns = turns
			bestIdx = k
		}
	}

	return paths[:bestIdx]
}

func computeTotalTurns(antCount int, paths [][]string) int {
	distribution := make([]int, len(paths))
	base := make([]int, len(paths))
	for i, p := range paths {
		base[i] = len(p) - 1
	}

	for ant := 0; ant < antCount; ant++ {
		minCost := math.MaxInt32
		chosen := 0
		for i := range paths {
			cost := base[i] + distribution[i]
			if cost <= minCost {
				minCost = cost
				chosen = i
			}
		}
		distribution[chosen]++
	}

	maxTurns := 0
	for i := range paths {
		turns := base[i] + distribution[i] - 1
		if turns > maxTurns {
			maxTurns = turns
		}
	}
	return maxTurns
}

func distributeAnts(antCount int, paths [][]string) []AntStatus {
	dist := make([]int, len(paths))
	pathLen := make([]int, len(paths))
	for i, p := range paths {
		pathLen[i] = len(p) - 1
	}

	antPaths := make([]int, antCount)
	for ant := 0; ant < antCount; ant++ {
		minCost := math.MaxInt32
		chosen := 0
		for i := range paths {
			cost := pathLen[i] + dist[i]
			if cost <= minCost {
				minCost = cost
				chosen = i
			}
		}
		dist[chosen]++
		antPaths[ant] = chosen
	}

	ants := make([]AntStatus, antCount)
	for i, pidx := range antPaths {
		ants[i] = AntStatus{
			PathIndex: pidx,
			Step:      0,
			Finished:  false,
		}
	}
	return ants
}

func tunnelKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func simulate(ants []AntStatus, paths [][]string) []Turn {

	var turns []Turn
	finished := false
	for !finished {
		finished = true
		var moves []Move
		occupied := make(map[string]bool)
		usedTunnels := make(map[string]bool)

		for i := range ants {
			if ants[i].Finished {
				continue
			}

			path := paths[ants[i].PathIndex]
			nextStep := ants[i].Step + 1

			if nextStep >= len(path) {
				ants[i].Finished = true
				continue
			}

			currentRoom := path[ants[i].Step]
			nextRoom := path[nextStep]

			roomFree := !occupied[nextRoom] || nextRoom == path[len(path)-1]
			tunnelFree := !usedTunnels[tunnelKey(currentRoom, nextRoom)]

			if roomFree && tunnelFree {
				moves = append(moves, Move{Ant: i + 1, Room: nextRoom})
				occupied[nextRoom] = true
				usedTunnels[tunnelKey(currentRoom, nextRoom)] = true
				ants[i].Step = nextStep
				if nextRoom == path[len(path)-1] {
					ants[i].Finished = true
				}
			}
		}

		if len(moves) > 0 {
			turns = append(turns, Turn{Moves: moves})
		}

		finished = true
		for i := range ants {
			if !ants[i].Finished {
				finished = false
				break
			}
		}
	}

	return turns
}
