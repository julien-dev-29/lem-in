package solver

import (
	"lem-in/internal/graph"
	"testing"
)

func makeAnts(pathCounts ...int) []AntStatus {
	var ants []AntStatus
	for pidx, count := range pathCounts {
		for i := 0; i < count; i++ {
			ants = append(ants, AntStatus{
				PathIndex: pidx,
				Step:      0,
				Finished:  false,
			})
		}
	}
	return ants
}

func TestDistributeOnePath(t *testing.T) {
	paths := [][]string{{"0", "1"}}
	ants := distributeAnts(5, paths)
	if len(ants) != 5 {
		t.Fatalf("expected 5 ants, got %d", len(ants))
	}
	for _, a := range ants {
		if a.PathIndex != 0 {
			t.Fatalf("expected all ants on path 0")
		}
	}
}

func TestDistributeTwoEqualPaths(t *testing.T) {
	paths := [][]string{{"0", "2", "1"}, {"0", "3", "1"}}
	ants := distributeAnts(4, paths)
	if len(ants) != 4 {
		t.Fatalf("expected 4 ants, got %d", len(ants))
	}
}

func TestSimulateDirectPath(t *testing.T) {
	paths := [][]string{{"0", "1"}}
	ants := makeAnts(3)
	turns := simulate(ants, paths)
	if len(turns) != 3 {
		t.Fatalf("expected 3 turns for 3 ants on direct path, got %d", len(turns))
	}
}

func TestSimulateSingleAnt(t *testing.T) {
	paths := [][]string{{"0", "2", "3", "1"}}
	ants := makeAnts(1)
	turns := simulate(ants, paths)
	if len(turns) != 3 {
		t.Fatalf("expected 3 turns for single ant on path of length 3, got %d", len(turns))
	}
}

func TestSimulateTwoPathsNoConflict(t *testing.T) {
	paths := [][]string{{"0", "2", "1"}, {"0", "3", "1"}}
	ants := makeAnts(1, 1)
	turns := simulate(ants, paths)
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns for 2 ants on disjoint paths of length 2, got %d", len(turns))
	}
}

func TestSelectBestPathSet(t *testing.T) {
	paths := [][]string{
		{"0", "2", "1"},
		{"0", "3", "4", "1"},
	}
	selected := selectBestPathSet(2, paths)
	if len(selected) == 0 {
		t.Fatal("expected at least 1 path")
	}
}

func TestComputeTotalTurns(t *testing.T) {
	paths := [][]string{{"0", "2", "1"}}
	turns := computeTotalTurns(3, paths)
	if turns != 4 {
		t.Fatalf("expected 4 turns for 3 ants on path of length 2, got %d", turns)
	}
}

func TestSolveBasic(t *testing.T) {
	g := newTestGraph()
	paths := [][]string{{"0", "1"}}
	turns := Solve(g, paths)
	if len(turns) == 0 {
		t.Fatal("expected some turns")
	}
}

func newTestGraph() *graph.Graph {
	g := graph.NewGraph()
	g.AntCount = 1
	g.AddRoom("0", 0, 0)
	g.AddRoom("1", 10, 0)
	g.AddLink("0", "1")
	g.Start = "0"
	g.End = "1"
	return g
}
