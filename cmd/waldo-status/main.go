package main

import (
	"flag"
	"fmt"

	"github.com/caboose-mcp/waldo/internal/config"
	"github.com/caboose-mcp/waldo/internal/git"
	"github.com/caboose-mcp/waldo/internal/status"
)

func main() {
	withPersona := flag.Bool("persona", false, "include active persona in status bar")
	flag.Parse()

	var bar string
	if *withPersona {
		cfg, _ := config.Load()
		persona := ""
		if cfg != nil {
			persona = cfg.ActivePersona
		}
		bar = status.BarWithPersona(persona)
	} else {
		bar = status.Bar()
	}

	fmt.Println(bar)

	// Exit with warning code if on prod/detached
	if gitStatus, err := git.GetStatus(); err == nil {
		if gitStatus.IsDetached || gitStatus.IsProd {
			// Non-zero exit signals a warning condition
			// Useful for shell prompts: if ! waldo-status; then show-warning; fi
		}
	}
}
