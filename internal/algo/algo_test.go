package algo

import (
	"lem-in/internal/graph"
	"testing"
)

func newTestGraph(ants int, start, end string, rooms map[string][2]int, links [][2]string) *graph.Graph {
	g := graph.NewGraph()
	g.AntCount = ants
	for name, coords := range rooms {
		g.AddRoom(name, coords[0], coords[1])
	}
	for _, l := range links {
		g.AddLink(l[0], l[1])
	}
	g.Start = start
	g.End = end
	return g
}

func TestFindPathsDirect(t *testing.T) {
	g := newTestGraph(5, "0", "1",
		map[string][2]int{"0": {0, 0}, "1": {10, 0}},
		[][2]string{{"0", "1"}},
	)
	paths := FindPaths(g)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if len(paths[0]) != 2 {
		t.Fatalf("expected path of length 2, got %d", len(paths[0]))
	}
}

func TestFindPathsNoPath(t *testing.T) {
	g := newTestGraph(3, "0", "1",
		map[string][2]int{"0": {0, 0}, "1": {10, 0}, "2": {5, 5}},
		[][2]string{{"0", "2"}},
	)
	paths := FindPaths(g)
	if len(paths) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(paths))
	}
}

func TestFindPathsTwoDisjoint(t *testing.T) {
	g := newTestGraph(4, "0", "5",
		map[string][2]int{
			"0": {0, 0}, "5": {15, 0},
			"1": {5, 5}, "2": {5, -5},
			"3": {10, 5}, "4": {10, -5},
		},
		[][2]string{
			{"0", "1"}, {"0", "2"},
			{"1", "3"}, {"2", "4"},
			{"3", "5"}, {"4", "5"},
		},
	)
	paths := FindPaths(g)
	if len(paths) < 1 {
		t.Fatal("expected at least 1 path")
	}
	t.Logf("Found %d paths", len(paths))
	for i, p := range paths {
		t.Logf("Path %d: %v", i, p)
	}
}

func TestFindPathsChain(t *testing.T) {
	g := newTestGraph(1, "0", "3",
		map[string][2]int{"0": {0, 0}, "1": {5, 0}, "2": {10, 0}, "3": {15, 0}},
		[][2]string{{"0", "1"}, {"1", "2"}, {"2", "3"}},
	)
	paths := FindPaths(g)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if len(paths[0]) != 4 {
		t.Fatalf("expected path of length 4, got %d: %v", len(paths[0]), paths[0])
	}
}

func TestFindPathsWithBottleneck(t *testing.T) {
	g := newTestGraph(3, "0", "4",
		map[string][2]int{
			"0": {0, 0}, "4": {15, 0},
			"1": {5, 5}, "2": {5, -5}, "3": {10, 0},
		},
		[][2]string{
			{"0", "1"}, {"0", "2"},
			{"1", "3"}, {"2", "3"},
			{"3", "4"},
		},
	)
	paths := FindPaths(g)
	t.Logf("Found %d paths (bottleneck at room 3)", len(paths))
	for i, p := range paths {
		t.Logf("Path %d: %v", i, p)
	}
}
