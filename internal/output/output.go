package output

import (
	"fmt"
	"lem-in/internal/solver"
	"strings"
)

func PrintResult(originalLines []string, turns []solver.Turn) {
	fmt.Println(strings.Join(originalLines, "\n"))
	fmt.Println()
	if turns == nil {
		return
	}
	for _, turn := range turns {
		var parts []string
		for _, move := range turn.Moves {
			parts = append(parts, fmt.Sprintf("L%d-%s", move.Ant, move.Room))
		}
		fmt.Println(strings.Join(parts, " "))
	}
}
