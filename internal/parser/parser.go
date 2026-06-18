package parser

import (
	"fmt"
	"lem-in/internal/graph"
	"os"
	"strconv"
	"strings"
)

type ParseResult struct {
	Graph *graph.Graph
	Lines []string
}

func ParseFile(path string) (*ParseResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	content := string(data)
	rawLines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(rawLines) == 0 || (len(rawLines) == 1 && rawLines[0] == "") {
		return nil, fmt.Errorf("empty file")
	}

	g := graph.NewGraph()
	seenRooms := make(map[string]bool)
	seenLinks := make(map[string]bool)

	var originalLines []string
	nextIsStart := false
	nextIsEnd := false
	parsedAnts := false
	type pendingLine struct {
		text        string
		isStartMark bool
		isEndMark   bool
	}
	var pendingTunnels []string

	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if trimmed[0] == '#' {
			switch trimmed {
			case "##start":
				nextIsStart = true
			case "##end":
				nextIsEnd = true
			}
			continue
		}

		if !parsedAnts {
			antCount, err := strconv.Atoi(trimmed)
			if err != nil || antCount < 1 {
				return nil, fmt.Errorf("invalid number of ants")
			}
			g.AntCount = antCount
			parsedAnts = true
			originalLines = append(originalLines, trimmed)
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) >= 3 {
			if nextIsStart {
				originalLines = append(originalLines, "##start")
			}
			if nextIsEnd {
				originalLines = append(originalLines, "##end")
			}

			name := fields[0]
			if name[0] == 'L' || name[0] == '#' {
				return nil, fmt.Errorf("invalid room name: %s", name)
			}
			x, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid x coordinate for room %s", name)
			}
			y, err := strconv.Atoi(fields[2])
			if err != nil {
				return nil, fmt.Errorf("invalid y coordinate for room %s", name)
			}
			if seenRooms[name] {
				return nil, fmt.Errorf("duplicate room: %s", name)
			}
			seenRooms[name] = true

			r := g.AddRoom(name, x, y)
			if nextIsStart {
				if g.Start != "" {
					return nil, fmt.Errorf("multiple start rooms")
				}
				r.IsStart = true
				g.Start = name
				nextIsStart = false
			}
			if nextIsEnd {
				if g.End != "" {
					return nil, fmt.Errorf("multiple end rooms")
				}
				r.IsEnd = true
				g.End = name
				nextIsEnd = false
			}
			originalLines = append(originalLines, trimmed)
		} else if strings.Contains(trimmed, "-") {
			pendingTunnels = append(pendingTunnels, trimmed)
		} else {
			return nil, fmt.Errorf("invalid line: %s", trimmed)
		}
	}

	if g.AntCount == 0 {
		return nil, fmt.Errorf("no ants specified")
	}
	if g.Start == "" {
		return nil, fmt.Errorf("no start room found")
	}
	if g.End == "" {
		return nil, fmt.Errorf("no end room found")
	}

	for _, trimmed := range pendingTunnels {
		parts := splitLink(trimmed, g.Rooms)
		if parts == nil {
			return nil, fmt.Errorf("invalid link format: %s", trimmed)
		}
		if parts[0] == parts[1] {
			return nil, fmt.Errorf("self-link not allowed: %s", trimmed)
		}
		linkKey := linkKey(parts[0], parts[1])
		if seenLinks[linkKey] {
			return nil, fmt.Errorf("duplicate link: %s", trimmed)
		}
		seenLinks[linkKey] = true
		g.AddLink(parts[0], parts[1])
		originalLines = append(originalLines, trimmed)
	}

	return &ParseResult{Graph: g, Lines: originalLines}, nil
}

func splitLink(link string, rooms map[string]*graph.Room) []string {
	for i := 1; i < len(link); i++ {
		if link[i] == '-' {
			a, b := link[:i], link[i+1:]
			if rooms[a] != nil && rooms[b] != nil {
				return []string{a, b}
			}
		}
	}
	return nil
}

func linkKey(a, b string) string {
	if a < b {
		return a + "-" + b
	}
	return b + "-" + a
}
