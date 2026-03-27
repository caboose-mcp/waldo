package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/caboose-mcp/waldo/internal/config"
)

func main() {
	typeFlag := flag.String("type", "cli", "Project type: cli, library, service, tool")
	langFlag := flag.String("lang", "go", "Language: go, node, python, rust")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}

	if args[0] != "new" {
		usage()
		os.Exit(1)
	}

	projectName := args[1]

	// Load active persona
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Could not load waldo config (proceeding with defaults): %v\n", err)
		cfg = &config.Config{
			ActivePersona: "unknown",
		}
	}

	// Display project creation info
	fmt.Printf(`
✓ Creating project: %s
  Type: %s
  Language: %s
  Persona: %s

This is a stub. Full implementation coming in v0.3.

TODO:
  - Generate Go project structure (cmd/, internal/, pkg/)
  - Create Makefile, go.mod, .gitignore
  - Generate README.md template
  - Create example.meml config template
  - Customize template based on persona preferences

For now, use standard tools:
  mkdir -p %s/{cmd,internal,pkg}
  cd %s
  go mod init github.com/yourname/%s

`, projectName, *typeFlag, *langFlag, cfg.ActivePersona, projectName, projectName, projectName)
}

func usage() {
	fmt.Fprintf(os.Stderr, `waldo project new <name> [options]

Create a new project using waldo persona templates.

Options:
  -type go|node|python|rust  Project type (default: cli)
  -lang go|node|python|rust  Language (default: go)

Examples:
  waldo project new my-cli --type cli
  waldo project new my-lib --type library --lang go
  waldo project new my-app --type service --lang node

Templates are customized based on your active waldo persona:
  - Config format (MEML, YAML, JSON)
  - Directory structure
  - Testing style
  - Documentation style
  - Theme recommendations
`)
}
