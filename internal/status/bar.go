package status

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/caboose-mcp/waldo/internal/build"
	"github.com/caboose-mcp/waldo/internal/git"
)

// Bar generates a status bar string for terminal display.
// Shows: git branch status + build status + persona info
// Uses emoji and color coding for quick visual feedback.
func Bar() string {
	var parts []string

	// Git status
	if gitStatus, err := git.GetStatus(); err == nil {
		statusStr := gitStatus.StatuslineFormat()
		color := gitStatus.WarningLevel()
		parts = append(parts, colorize(statusStr, color))
	}

	// Build status
	parts = append(parts, build.GetStatus())

	return strings.Join(parts, " | ")
}

// BarWithPersona includes the current active persona.
func BarWithPersona(persona string) string {
	var parts []string

	// Git status
	if gitStatus, err := git.GetStatus(); err == nil {
		statusStr := gitStatus.StatuslineFormat()
		color := gitStatus.WarningLevel()
		parts = append(parts, colorize(statusStr, color))
	}

	// Build status
	parts = append(parts, build.GetStatus())

	// Persona
	if persona != "" {
		parts = append(parts, fmt.Sprintf("🎭 %s", persona))
	}

	return strings.Join(parts, " | ")
}

func colorize(s, color string) string {
	style := lipgloss.NewStyle()

	switch color {
	case "red":
		style = style.Foreground(lipgloss.Color("1")).Bold(true)
	case "yellow":
		style = style.Foreground(lipgloss.Color("3"))
	case "green":
		style = style.Foreground(lipgloss.Color("2"))
	}

	return style.Render(s)
}
