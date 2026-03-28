package export

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseMEML parses a MEML persona file and returns a PersonaConfig.
// It extracts tone, verbosity, voice, and metadata fields from the MEML format.
func ParseMEML(content string) (*PersonaConfig, error) {
	cfg := &PersonaConfig{}

	// Regex patterns for parsing
	sectionRe := regexp.MustCompile(`\b(meta|tone|verbosity|voice)\b`)
	quotedStringRe := regexp.MustCompile(`^(\w+)\s*=\s*"([^"]*)"`)
	floatRe := regexp.MustCompile(`^(\w+)\s*=\s*(-?[0-9]+(?:\.[0-9]+)?)`)
	quotedArrayRe := regexp.MustCompile(`^(\w+)\s*=\s*\[([^\]]*)\]`)
	arrayItemRe := regexp.MustCompile(`"([^"]*)"`)

	currentSection := ""
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Strip inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Check for section header
		if sectionMatch := sectionRe.FindStringSubmatch(line); len(sectionMatch) > 0 {
			currentSection = sectionMatch[1]
			continue
		}

		// Parse key-value pairs based on current section
		switch currentSection {
		case "meta":
			if m := quotedStringRe.FindStringSubmatch(line); len(m) == 3 {
				key, val := m[1], m[2]
				if key == "name" {
					cfg.Name = val
				} else if key == "description" {
					cfg.Description = val
				}
			}

		case "tone":
			if m := floatRe.FindStringSubmatch(line); len(m) == 3 {
				key, val := m[1], m[2]
				fval, _ := strconv.ParseFloat(val, 64)
				switch key {
				case "formality":
					cfg.Tone.Formality = fval
				case "directness":
					cfg.Tone.Directness = fval
				case "humor":
					cfg.Tone.Humor = fval
				case "hedging":
					cfg.Tone.Hedging = fval
				case "warmth":
					cfg.Tone.Warmth = fval
				}
			}

		case "verbosity":
			if m := quotedStringRe.FindStringSubmatch(line); len(m) == 3 {
				key, val := m[1], m[2]
				switch key {
				case "response_length":
					cfg.Verbosity.ResponseLength = val
				case "reading_level":
					cfg.Verbosity.ReadingLevel = val
				case "format_preference":
					cfg.Verbosity.FormatPreference = val
				}
			}

		case "voice":
			if m := quotedArrayRe.FindStringSubmatch(line); len(m) == 3 {
				key, content := m[1], m[2]
				// Extract quoted items from array content
				items := arrayItemRe.FindAllStringSubmatch(content, -1)
				var values []string
				for _, item := range items {
					if len(item) == 2 {
						values = append(values, item[1])
					}
				}
				switch key {
				case "avoid_words":
					cfg.Voice.AvoidWords = values
				case "prefer_words":
					cfg.Voice.PreferWords = values
				case "custom_phrases":
					cfg.Voice.CustomPhrases = values
				}
			}
		}
	}

	// Validate required fields
	if cfg.Name == "" {
		return nil, fmt.Errorf("ParseMEML: missing required field: name")
	}

	return cfg, nil
}
