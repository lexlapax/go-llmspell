// ABOUTME: Entry point for the llmspell CLI command
// ABOUTME: Provides command-line interface for running LLM spells

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		scriptPath = flag.String("script", "", "Path to spell script file")
		engine     = flag.String("engine", "lua", "Script engine to use (lua, js, tengo)")
		verbose    = flag.Bool("v", false, "Enable verbose output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *scriptPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("🧙‍♂️ Casting spell from %s using %s engine...\n", *scriptPath, *engine)
	
	if *verbose {
		log.Printf("Verbose mode enabled")
	}

	fmt.Println("✨ Spell cast successfully!")
}