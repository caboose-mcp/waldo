package cli

import (
	"fmt"
	"os"

	"github.com/caboose-mcp/waldo/internal/config"
	"github.com/caboose-mcp/waldo/internal/github"
	"github.com/caboose-mcp/waldo/internal/s3"
	"github.com/caboose-mcp/waldo/internal/tui"
)

func RunBucketPicker() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Config load warning: %v\n", err)
		// Continue — not fatal
	}

	fmt.Print("📦 Fetching S3 buckets...")
	buckets, err := s3.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		os.Exit(1)
	}

	if len(buckets) == 0 {
		fmt.Fprintf(os.Stderr, "\n❌ No S3 buckets found. Create one with:\n")
		fmt.Fprintf(os.Stderr, "   aws s3api create-bucket --bucket my-personas\n")
		os.Exit(1)
	}

	fmt.Printf(" found %d bucket(s)\n\n", len(buckets))

	names := s3.Names(buckets)
	menu := tui.NewMenu(names, "Select S3 bucket:", 10)
	selected, cancelled, err := menu.Run(os.Stdin)

	if cancelled {
		fmt.Println("Cancelled.")
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	if cfg == nil {
		cfg, err = config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	if err := cfg.SaveS3Bucket(selected); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to save bucket: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ S3 bucket set: %s\n", selected)
	fmt.Printf("  Updated: ~/.claude/settings.json\n")
}

func RunScriptFetch(rawURL string, args []string) {
	fmt.Printf("📥 Fetching: %s\n", rawURL)

	script, err := github.Fetch(rawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %d lines\n\n", len(script.Lines()))

	github.Preview(script, os.Stdout)
	fmt.Println()

	if !github.Confirm("Execute this script?") {
		fmt.Println("Skipped.")
		return
	}

	fmt.Println("Running...")
	if err := script.Execute(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Done")
}

func RunStatus() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Config load warning: %v\n", err)
		fmt.Println()
	}

	fmt.Println("waldo status")
	fmt.Println("─────────────────────────────────────────")

	if cfg != nil {
		fmt.Printf("Active persona   %s\n", cfg.ActivePersona)
		fmt.Printf("Personas dir     ~/.claude/personas/\n")
		if cfg.S3Bucket != "" {
			fmt.Printf("S3 bucket        %s\n", cfg.S3Bucket)
		}
		if cfg.AWSProfile != "" {
			fmt.Printf("AWS profile      %s\n", cfg.AWSProfile)
		}
		fmt.Printf("Config root      ~/.config/waldo/\n")
	} else {
		fmt.Println("(no config loaded)")
	}
}
