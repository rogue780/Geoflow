// Package repl implements the Read-Eval-Print Loop for GeoFlow.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/rogue780/geoflow/internal/eval"
	"github.com/rogue780/geoflow/internal/lexer"
	"github.com/rogue780/geoflow/internal/object"
	"github.com/rogue780/geoflow/internal/parser"
	"github.com/rogue780/geoflow/internal/stdlib"
)

const PROMPT = "geoflow> "

// Start begins the REPL loop.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	// Register builtins
	for name, builtin := range object.GetBuiltins() {
		env.Set(name, builtin, false)
	}
	for name, builtin := range stdlib.GetMathBuiltins() {
		env.Set(name, builtin, false)
	}
	for name, builtin := range stdlib.GetGeoBuiltins() {
		env.Set(name, builtin, false)
	}

	fmt.Fprintf(out, "GeoFlow v0.1.0 - Interactive REPL\n")
	fmt.Fprintf(out, "Type 'exit' or 'quit' to exit.\n\n")

	for {
		fmt.Fprint(out, PROMPT)
		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Fprintln(out, "Goodbye!")
			return
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		result := eval.Eval(program, env)
		if result != nil && result != object.NIL {
			fmt.Fprintln(out, result.Inspect())
		}
	}
}

func printParseErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintf(out, "  Parse error: %s\n", msg)
	}
}

// Execute runs a GeoFlow program from source code.
func Execute(source string, env *object.Environment) object.Object {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		var msgs []string
		for _, e := range p.Errors() {
			msgs = append(msgs, e)
		}
		return &object.Error{Message: fmt.Sprintf("parse errors:\n%s", strings.Join(msgs, "\n"))}
	}

	return eval.Eval(program, env)
}

// NewDefaultEnvironment creates an environment with all builtins pre-loaded.
func NewDefaultEnvironment() *object.Environment {
	env := object.NewEnvironment()
	for name, builtin := range object.GetBuiltins() {
		env.Set(name, builtin, false)
	}
	for name, builtin := range stdlib.GetMathBuiltins() {
		env.Set(name, builtin, false)
	}
	for name, builtin := range stdlib.GetGeoBuiltins() {
		env.Set(name, builtin, false)
	}
	return env
}
