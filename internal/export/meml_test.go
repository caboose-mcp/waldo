package export

import (
	"strings"
	"testing"
)

const testMEMLBasic = `
[🪪 meta]
name        = "ask-a-dev-bot"
description = "Direct, curious, dev-centric."

[🎭 tone]
formality   = 0.6
directness  = 0.8
humor       = 0.2
hedging     = 0.7
warmth      = 0.7

[📢 verbosity]
response_length   = "adaptive"
reading_level     = "professional"
format_preference = "bullets"

[🗣️ voice]
avoid_words      = ["certainly", "absolutely", "as an AI"]
prefer_words     = ["devs report", "common pattern"]
custom_phrases   = ["here's the deal"]
`

func TestParseMEML_MetaFields(t *testing.T) {
	cfg, err := ParseMEML(testMEMLBasic)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Name != "ask-a-dev-bot" {
		t.Errorf("expected name 'ask-a-dev-bot', got %q", cfg.Name)
	}
	if cfg.Description != "Direct, curious, dev-centric." {
		t.Errorf("expected description, got %q", cfg.Description)
	}
}

func TestParseMEML_ToneFloats(t *testing.T) {
	cfg, err := ParseMEML(testMEMLBasic)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Tone.Formality != 0.6 {
		t.Errorf("expected formality 0.6, got %v", cfg.Tone.Formality)
	}
	if cfg.Tone.Directness != 0.8 {
		t.Errorf("expected directness 0.8, got %v", cfg.Tone.Directness)
	}
	if cfg.Tone.Humor != 0.2 {
		t.Errorf("expected humor 0.2, got %v", cfg.Tone.Humor)
	}
	if cfg.Tone.Hedging != 0.7 {
		t.Errorf("expected hedging 0.7, got %v", cfg.Tone.Hedging)
	}
	if cfg.Tone.Warmth != 0.7 {
		t.Errorf("expected warmth 0.7, got %v", cfg.Tone.Warmth)
	}
}

func TestParseMEML_ToneWithInlineComment(t *testing.T) {
	meml := `
[🪪 meta]
name = "test"

[🎭 tone]
formality = 0.28  # Low formality
`
	cfg, err := ParseMEML(meml)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Tone.Formality != 0.28 {
		t.Errorf("expected formality 0.28 with inline comment, got %v", cfg.Tone.Formality)
	}
}

func TestParseMEML_VerbosityStrings(t *testing.T) {
	cfg, err := ParseMEML(testMEMLBasic)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Verbosity.ResponseLength != "adaptive" {
		t.Errorf("expected 'adaptive', got %q", cfg.Verbosity.ResponseLength)
	}
	if cfg.Verbosity.ReadingLevel != "professional" {
		t.Errorf("expected 'professional', got %q", cfg.Verbosity.ReadingLevel)
	}
	if cfg.Verbosity.FormatPreference != "bullets" {
		t.Errorf("expected 'bullets', got %q", cfg.Verbosity.FormatPreference)
	}
}

func TestParseMEML_VoiceArrays(t *testing.T) {
	cfg, err := ParseMEML(testMEMLBasic)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}

	if len(cfg.Voice.AvoidWords) != 3 || cfg.Voice.AvoidWords[0] != "certainly" {
		t.Errorf("avoid_words mismatch: %v", cfg.Voice.AvoidWords)
	}
	if len(cfg.Voice.PreferWords) != 2 || cfg.Voice.PreferWords[0] != "devs report" {
		t.Errorf("prefer_words mismatch: %v", cfg.Voice.PreferWords)
	}
	if len(cfg.Voice.CustomPhrases) != 1 || cfg.Voice.CustomPhrases[0] != "here's the deal" {
		t.Errorf("custom_phrases mismatch: %v", cfg.Voice.CustomPhrases)
	}
}

func TestParseMEML_ArrayWithMultipleItems(t *testing.T) {
	meml := `
[🪪 meta]
name = "test"

[🗣️ voice]
avoid_words = ["certainly", "absolutely", "I'd be happy to", "leverage"]
`
	cfg, err := ParseMEML(meml)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if len(cfg.Voice.AvoidWords) != 4 {
		t.Errorf("expected 4 avoid_words, got %d: %v", len(cfg.Voice.AvoidWords), cfg.Voice.AvoidWords)
	}
	if cfg.Voice.AvoidWords[2] != "I'd be happy to" {
		t.Errorf("expected apostrophe in item 2, got %q", cfg.Voice.AvoidWords[2])
	}
}

func TestParseMEML_MissingNameReturnsError(t *testing.T) {
	meml := `
[🎭 tone]
formality = 0.5
`
	cfg, err := ParseMEML(meml)
	if err == nil {
		t.Fatalf("expected error for missing name, got nil")
	}
	if cfg != nil {
		t.Fatalf("expected cfg to be nil on error, got %v", cfg)
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error message should mention 'name', got: %v", err)
	}
}

func TestParseMEML_UnknownSectionsIgnored(t *testing.T) {
	meml := `
[🪪 meta]
name = "test"

[😊 emoji]
enabled = true

[🎭 tone]
formality = 0.5

[♿ vtt]
enabled = false
`
	cfg, err := ParseMEML(meml)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Name != "test" {
		t.Errorf("name parsing failed: got %q", cfg.Name)
	}
	if cfg.Tone.Formality != 0.5 {
		t.Errorf("tone parsing failed with unknown sections: got %v", cfg.Tone.Formality)
	}
}

func TestParseMEML_FullPersona(t *testing.T) {
	meml := `
[🪪 meta]
name        = "chris-marasco"
description = "Casual, direct, no corporate speak"
version     = "0.1.0"

[🎭 tone]
formality   = 0.28
directness  = 0.9
humor       = 0.6
hedging     = 0.2
warmth      = 0.7

[📢 verbosity]
response_length   = "concise"
reading_level     = "casual"
format_preference = "prose"

[🗣️ voice]
avoid_words    = ["certainly", "leverage", "utilize"]
prefer_words   = ["use", "build", "fix"]
custom_phrases = ["yea", "here's the deal"]
`
	cfg, err := ParseMEML(meml)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}

	if cfg.Name != "chris-marasco" {
		t.Errorf("Name: expected 'chris-marasco', got %q", cfg.Name)
	}
	if cfg.Description != "Casual, direct, no corporate speak" {
		t.Errorf("Description mismatch")
	}
	if cfg.Tone.Formality != 0.28 {
		t.Errorf("Formality: expected 0.28, got %v", cfg.Tone.Formality)
	}
	if cfg.Tone.Directness != 0.9 {
		t.Errorf("Directness: expected 0.9, got %v", cfg.Tone.Directness)
	}
	if cfg.Verbosity.ResponseLength != "concise" {
		t.Errorf("ResponseLength: expected 'concise', got %q", cfg.Verbosity.ResponseLength)
	}
	if len(cfg.Voice.AvoidWords) != 3 {
		t.Errorf("AvoidWords: expected 3 items, got %d", len(cfg.Voice.AvoidWords))
	}
	if len(cfg.Voice.PreferWords) != 3 {
		t.Errorf("PreferWords: expected 3 items, got %d", len(cfg.Voice.PreferWords))
	}
}

func TestParseMEML_EmptyValues(t *testing.T) {
	meml := `
[🪪 meta]
name = "empty-test"
description = ""

[🗣️ voice]
avoid_words = []
`
	cfg, err := ParseMEML(meml)
	if err != nil {
		t.Fatalf("ParseMEML failed: %v", err)
	}
	if cfg.Name != "empty-test" {
		t.Errorf("Name parsing failed")
	}
	if cfg.Description != "" {
		t.Errorf("Empty description should parse as empty string, got %q", cfg.Description)
	}
	if len(cfg.Voice.AvoidWords) != 0 {
		t.Errorf("Empty array should parse as empty slice, got %v", cfg.Voice.AvoidWords)
	}
}
