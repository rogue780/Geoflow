// GeoFlow - Composition-oriented functional language for geospatial data transformation.
package main

import (
	"fmt"
	"os"

	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/repl"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	switch args[0] {
	case "--version", "-v":
		fmt.Printf("GeoFlow %s\n", version)
	case "--help", "-h":
		printUsage()
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: geoflow run <file.gf>")
			os.Exit(1)
		}
		runFile(args[1])
	default:
		runFile(args[0])
	}
}

func runFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	env := repl.NewDefaultEnvironment()
	result := repl.Execute(string(data), env)

	if result != nil {
		if errObj, ok := result.(*object.Error); ok {
			fmt.Fprintf(os.Stderr, "%s\n", errObj.Inspect())
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("GeoFlow - Composition-oriented functional language for geospatial data transformation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  geoflow                    Start interactive REPL")
	fmt.Println("  geoflow <file.gf>          Run a GeoFlow script")
	fmt.Println("  geoflow run <file.gf>      Run a GeoFlow script")
	fmt.Println("  geoflow --version, -v      Print version")
	fmt.Println("  geoflow --help, -h         Print this help")
}
