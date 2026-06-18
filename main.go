package main

import (
	"fmt"
	"os"

	"lem-in/internal/algo"
	"lem-in/internal/output"
	"lem-in/internal/parser"
	"lem-in/internal/solver"
)

func main() {
	// Si il n'y a pas d'arguments
	if len(os.Args) != 2 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	// Je parse le fichier passé en argument
	result, err := parser.ParseFile(os.Args[1])
	if err != nil {
		fmt.Println("ERROR: invalid data format")
		return
	}

	paths := algo.FindPaths(result.Graph)
	if len(paths) == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	turns := solver.Solve(result.Graph, paths)
	output.PrintResult(result.Lines, turns)
}
