package export

import (
	"fmt"
	"strings"
)

type GeminiExporter struct {
	Persona *PersonaConfig
}

// SystemPrompt generates a Gemini-compatible system prompt.
// Can be pasted directly into Gemini's System Instructions.
func (ge *GeminiExporter) SystemPrompt() string {
	p := ge.Persona

	tone := fmt.Sprintf(`Tone Settings:
- Formality: %.1f (0=casual, 1=formal)
- Directness: %.1f (0=roundabout, 1=blunt)
- Humor: %.1f (0=dry, 1=frequent)
- Hedging: %.1f (0=confident, 1=qualified)
- Warmth: %.1f (0=cold, 1=enthusiastic)`,
		p.Tone.Formality, p.Tone.Directness, p.Tone.Humor, p.Tone.Hedging, p.Tone.Warmth)

	verbosity := fmt.Sprintf(`Response Style:
- Length: %s
- Reading level: %s
- Format: %s`,
		p.Verbosity.ResponseLength, p.Verbosity.ReadingLevel, p.Verbosity.FormatPreference)

	voice := fmt.Sprintf(`Voice:
- Avoid: %s
- Prefer: %s
- Phrases: %s`,
		strings.Join(p.Voice.AvoidWords, ", "),
		strings.Join(p.Voice.PreferWords, ", "),
		strings.Join(p.Voice.CustomPhrases, ", "))

	prompt := fmt.Sprintf(`You respond with this personality:

%s

%s

%s

Use these traits in all responses.`, tone, verbosity, voice)

	return prompt
}

// Markdown returns a markdown-formatted version.
func (ge *GeminiExporter) Markdown() string {
	return fmt.Sprintf(`# %s

%s

## System Prompt

%s`, ge.Persona.Name, ge.Persona.Description, ge.SystemPrompt())
}

// Instructions returns Gemini-specific setup instructions.
func (ge *GeminiExporter) Instructions() string {
	return fmt.Sprintf(`Copy this into Gemini's System Instructions:

%s

---

**How to use:**
1. Go to gemini.google.com
2. Click "Customize" (gear icon)
3. Paste the above text into System Instructions
4. Start a new conversation
`, ge.SystemPrompt())
}
