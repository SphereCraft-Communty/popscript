package main

import (
	"fmt"
	"os"

	"popscript/interpreter"
	"popscript/lexer"
	"popscript/parser"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pop run <file.pscript>")
			os.Exit(1)
		}
		runFile(os.Args[2])

	case "init":
		fmt.Println("Initialized new PopScript project.")

	case "build":
		fmt.Println("Note: 'pop build' is not yet implemented. Use 'pop run' to interpret.")

	case "get":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pop get <package>")
			os.Exit(1)
		}
		fmt.Printf("Fetching package %q... (stub)\n", os.Args[2])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func runFile(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(src))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Lexer error: %v\n", err)
		os.Exit(1)
	}

	// Parse
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Interpret
	interp := interpreter.New()
	if err := interp.Run(prog); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`PopScript interpreter v1.0.1

Usage:
  pop run <file.pscript>    Run a PopScript file
  pop init                  Initialize a new project
  pop build <file.pscript>  Build (stub)
  pop get <package>         Install a package (stub)`)
}
