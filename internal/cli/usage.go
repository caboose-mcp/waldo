package cli

import (
	"fmt"
)

func PrintUsage() {
	fmt.Print(`waldo-tui — interactive CLI for the waldo persona system

Commands:
  bucket         Fuzzy-search S3 buckets, pick one for persona sync
  fetch <url>    Fetch a GitHub script, inspect, and optionally run it
  status         Show current active persona and S3 config

Flags:
  --help         Show this message
  --version      Show version

Examples:
  waldo-tui bucket
  waldo-tui fetch caboose-mcp/waldo/.claude/hooks/waldo/inject-persona.sh
  waldo-tui status

Environment:
  AWS_PROFILE    AWS profile to use (default: "default")
  AWS_REGION     AWS region (default: "us-east-1")
`)
}
