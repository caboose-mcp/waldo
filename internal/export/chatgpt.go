package export

import (
	"fmt"
	"strings"
)

type ChatGPTExporter struct {
	Persona *PersonaConfig
}

type PersonaConfig struct {
	Name        string
	Description string
	Tone        struct {
		Formality   float64
		Directness  float64
		Humor       float64
		Hedging     float64
		Warmth      float64
	}
	Verbosity struct {
		ResponseLength string
		ReadingLevel   string
		FormatPreference string
	}
	Voice struct {
		AvoidWords   []string
		PreferWords  []string
		CustomPhrases []string
	}
}

// SystemPrompt generates a ChatGPT-compatible system prompt.
// Can be pasted directly into ChatGPT Custom Instructions.
func (ce *ChatGPTExporter) SystemPrompt() string {
	p := ce.Persona

	tone := fmt.Sprintf(`## Tone Settings
- **Formality:** %.1f (0=very casual, 1=very formal)
- **Directness:** %.1f (0=roundabout, 1=blunt)
- **Humor:** %.1f (0=dry, 1=frequent wit)
- **Hedging:** %.1f (0=confident, 1=heavily qualified)
- **Warmth:** %.1f (0=cold, 1=enthusiastic)`,
		p.Tone.Formality, p.Tone.Directness, p.Tone.Humor, p.Tone.Hedging, p.Tone.Warmth)

	verbosity := fmt.Sprintf(`## Verbosity
- **Response length:** %s
- **Reading level:** %s
- **Format preference:** %s`,
		p.Verbosity.ResponseLength, p.Verbosity.ReadingLevel, p.Verbosity.FormatPreference)

	voice := fmt.Sprintf(`## Voice
- **Avoid:** %s
- **Prefer:** %s
- **Characteristic phrases:** %s`,
		strings.Join(p.Voice.AvoidWords, ", "),
		strings.Join(p.Voice.PreferWords, ", "),
		strings.Join(p.Voice.CustomPhrases, ", "))

	prompt := fmt.Sprintf(`You are designed to respond with this personality profile:

%s

%s

%s

Apply these traits consistently to all your responses.`,
		tone, verbosity, voice)

	return prompt
}

// Markdown returns a formatted markdown version suitable for documentation.
func (ce *ChatGPTExporter) Markdown() string {
	return fmt.Sprintf(`# %s

%s

%s`, ce.Persona.Name, ce.Persona.Description, ce.SystemPrompt())
}

// JSON returns the persona as JSON (compatible with Claude API, etc.)
func (ce *ChatGPTExporter) JSON() string {
	// TODO: Use json.Marshal with proper formatting
	return "{}"
}

// Instructions returns ChatGPT-specific instruction format.
func (ce *ChatGPTExporter) Instructions() string {
	return fmt.Sprintf(`Copy this into ChatGPT's Custom Instructions:

%s

---

**How to use:**
1. Go to ChatGPT Settings → Personalization → Custom Instructions
2. Paste the above text into the "What would you like ChatGPT to know about you?" field
3. Start a new conversation for the instructions to take effect
`, ce.SystemPrompt())
}
