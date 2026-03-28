package terminal

import (
	"os"
	"strings"
)

type Capability struct {
	Emoji   bool // Terminal supports emoji
	Color   bool // Terminal supports ANSI colors
	RawMode bool // Terminal supports raw input mode
	UTF8    bool // Terminal encoding is UTF-8
}

// Detect inspects the environment and determines terminal capabilities.
func Detect() *Capability {
	term := os.Getenv("TERM")
	lang := os.Getenv("LANG")

	cap := &Capability{
		Emoji:   true,
		Color:   true,
		RawMode: true,
		UTF8:    strings.Contains(lang, "UTF-8") || strings.Contains(lang, "utf8"),
	}

	// Downgrade for known dumb terminals
	switch term {
	case "dumb", "vt100", "vt220", "vt52":
		cap.Emoji = false
		cap.Color = false
		cap.RawMode = false
		cap.UTF8 = false
	case "screen", "screen-256color":
		// tmux compatibility: raw mode can cause issues
		if _, inTmux := os.LookupEnv("TMUX"); inTmux {
			cap.RawMode = false
		}
	case "linux":
		// Linux console: emoji support depends on kernel + fonts
		cap.Emoji = false
	}

	// CI/CD environments: disable emoji (often stripped)
	if _, inGitHub := os.LookupEnv("GITHUB_ACTIONS"); inGitHub {
		cap.Emoji = false
	}
	if _, inGitLab := os.LookupEnv("CI"); inGitLab {
		cap.Emoji = false
	}
	if _, inDocker := os.LookupEnv("DOCKER_CONTAINER"); inDocker {
		cap.Emoji = false
	}

	// Command-line overrides
	if _, noEmoji := os.LookupEnv("WALDO_NO_EMOJI"); noEmoji {
		cap.Emoji = false
	}
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		cap.Color = false
	}

	return cap
}

// String returns a human-readable summary of capabilities.
func (c *Capability) String() string {
	var parts []string
	if c.Emoji {
		parts = append(parts, "emoji")
	}
	if c.Color {
		parts = append(parts, "color")
	}
	if c.RawMode {
		parts = append(parts, "raw-mode")
	}
	if c.UTF8 {
		parts = append(parts, "utf8")
	}
	return "[" + strings.Join(parts, " ") + "]"
}
