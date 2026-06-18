package graph

type Room struct {
	Name    string
	X, Y    int
	IsStart bool
	IsEnd   bool
}

type Graph struct {
	Rooms    map[string]*Room
	Links    map[string][]string
	linkSet  map[string]map[string]bool
	AntCount int
	Start    string
	End      string
}

func NewGraph() *Graph {
	return &Graph{
		Rooms:   make(map[string]*Room),
		Links:   make(map[string][]string),
		linkSet: make(map[string]map[string]bool),
	}
}

func (g *Graph) AddRoom(name string, x, y int) *Room {
	r := &Room{Name: name, X: x, Y: y}
	g.Rooms[name] = r
	return r
}

func (g *Graph) AddLink(a, b string) {
	if g.linkSet[a] == nil {
		g.linkSet[a] = make(map[string]bool)
	}
	if g.linkSet[b] == nil {
		g.linkSet[b] = make(map[string]bool)
	}
	if !g.linkSet[a][b] {
		g.Links[a] = append(g.Links[a], b)
		g.linkSet[a][b] = true
	}
	if !g.linkSet[b][a] {
		g.Links[b] = append(g.Links[b], a)
		g.linkSet[b][a] = true
	}
}

func (g *Graph) HasLink(a, b string) bool {
	return g.linkSet[a] != nil && g.linkSet[a][b]
}
