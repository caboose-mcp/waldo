package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caboose-mcp/waldo/internal/export"
)

func main() {
	outputFlag := flag.String("output", ".rules", "Output file path (default: .rules)")
	stdoutFlag := flag.Bool("stdout", false, "Print to stdout instead of writing file")
	flag.Parse()

	// Get current working directory (project root)
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Load waldo config
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	personasDir := filepath.Join(home, ".config", "waldo", "personas")
	activeFile := filepath.Join(personasDir, ".active")

	// Read active persona
	activeData, err := os.ReadFile(activeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ No active persona found. Run: waldo use agent/default\n")
		os.Exit(1)
	}

	activeName := strings.TrimSpace(string(activeData))
	personaPath := filepath.Join(personasDir, activeName+".meml")

	// Read persona file
	personaData, err := os.ReadFile(personaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Persona file not found: %s\n", personaPath)
		os.Exit(1)
	}

	// Parse MEML to PersonaConfig
	persona, err := export.ParseMEML(string(personaData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to parse persona: %v\n", err)
		os.Exit(1)
	}

	// Generate Zed rules
	zedExporter := &export.ZedExporter{Persona: persona}
	rulesContent := zedExporter.RulesFile()

	// Output to stdout or file
	if *stdoutFlag {
		fmt.Print(rulesContent)
		return
	}

	// Write .rules file in project root (or specified output path)
	outputPath := filepath.Join(projectRoot, *outputFlag)
	if err := os.WriteFile(outputPath, []byte(rulesContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("✓ Synced persona '%s' to %s\n", activeName, outputPath)
}
