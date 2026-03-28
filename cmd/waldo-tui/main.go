package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/caboose-mcp/waldo/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bucket":
		cli.RunBucketPicker()
	case "fetch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "fetch: URL required\nUsage: waldo-tui fetch <url> [args...]\n")
			os.Exit(1)
		}
		cli.RunScriptFetch(os.Args[2], os.Args[3:])
	case "status":
		cli.RunStatus()
	case "--help", "-h", "help":
		cli.PrintUsage()
	case "--version", "-v":
		fmt.Println("waldo-tui v1.0.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		cli.PrintUsage()
		os.Exit(1)
	}
}

func init() {
	flag.Parse()
}
