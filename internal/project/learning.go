package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Learning tracks user preferences across projects.
type Learning struct {
	ProjectsCreated   int
	PreferredFormats  map[string]int // meml, yaml, json, toml
	TemplateUsage     map[string]int // cli, library, service, tool
	LanguagesUsed     map[string]int // go, node, python, rust
	ThemesUsed        map[string]int // dark, light, auto
	AvgProjectSize    string         // small, medium, large
	LastUpdated       string
}

// LoadLearning reads learning history from persona deltas.
// TODO: Parse from ~/.config/waldo/personas/agent/.deltas
func LoadLearning() (*Learning, error) {
	return &Learning{
		PreferredFormats: make(map[string]int),
		TemplateUsage:    make(map[string]int),
		LanguagesUsed:    make(map[string]int),
		ThemesUsed:       make(map[string]int),
	}, nil
}

// SuggestFormat recommends config format based on history.
func (l *Learning) SuggestFormat() string {
	maxCount := 0
	var best string

	for fmt, count := range l.PreferredFormats {
		if count > maxCount {
			maxCount = count
			best = fmt
		}
	}

	if best == "" {
		return "meml" // Default if no history
	}

	return best
}

// SuggestTemplate recommends project template based on history.
func (l *Learning) SuggestTemplate() string {
	maxCount := 0
	var best string

	for tmpl, count := range l.TemplateUsage {
		if count > maxCount {
			maxCount = count
			best = tmpl
		}
	}

	if best == "" {
		return "cli" // Default if no history
	}

	return best
}

// SuggestLanguage recommends language based on history.
func (l *Learning) SuggestLanguage() string {
	maxCount := 0
	var best string

	for lang, count := range l.LanguagesUsed {
		if count > maxCount {
			maxCount = count
			best = lang
		}
	}

	if best == "" {
		return "go" // Default if no history
	}

	return best
}

// SuggestTheme recommends editor theme based on history.
func (l *Learning) SuggestTheme() string {
	if len(l.ThemesUsed) == 0 {
		return "dark" // Default
	}

	maxCount := 0
	var best string

	for theme, count := range l.ThemesUsed {
		if count > maxCount {
			maxCount = count
			best = theme
		}
	}

	return best
}

// RecordProject adds a new project to learning history.
// TODO: Save to ~/.config/waldo/personas/agent/.deltas
func (l *Learning) RecordProject(format, template, language string) error {
	l.PreferredFormats[format]++
	l.TemplateUsage[template]++
	l.LanguagesUsed[language]++
	l.ProjectsCreated++

	// TODO: Marshal to JSON and append to .deltas file
	return nil
}

// Summary returns a human-readable summary of preferences.
func (l *Learning) Summary() string {
	summary := fmt.Sprintf(`
Preference Learning Summary (%.0f projects created):

Favorite config format: %s (used %d times)
Favorite template:      %s (used %d times)
Favorite language:      %s (used %d times)
Preferred theme:        %s (used %d times)
`,
		float64(l.ProjectsCreated),
		l.SuggestFormat(), l.PreferredFormats[l.SuggestFormat()],
		l.SuggestTemplate(), l.TemplateUsage[l.SuggestTemplate()],
		l.SuggestLanguage(), l.LanguagesUsed[l.SuggestLanguage()],
		l.SuggestTheme(), l.ThemesUsed[l.SuggestTheme()],
	)

	return summary
}

// SaveToDeltaFile appends learning record to persona deltas.
// TODO: Implement full JSONL delta appending
func (l *Learning) SaveToDeltaFile(personaName string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	deltasFile := filepath.Join(home, ".config", "waldo", "personas", "agent", ".deltas")

	// Create file if it doesn't exist
	if _, err := os.Stat(deltasFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(deltasFile), 0755); err != nil {
			return err
		}
	}

	// TODO: Append JSON line to deltas file
	// For now, just log
	fmt.Printf("TODO: Save learning record to %s\n", deltasFile)
	return nil
}

// ToJSON serializes learning to JSON.
func (l *Learning) ToJSON() (string, error) {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
