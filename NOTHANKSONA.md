# nothanksona — Claude Code Persona System

A persistent persona/configuration system for Claude Code that shapes responses by tone, verbosity, and voice.

## What It Does

**nothanksona** lets you:
- Create and manage response personas (pre-configured tone + voice profiles)
- Switch personas on the fly with `/nothanksona use <name>`
- Auto-generate personas from your Slack message samples with `/nothanksona slack-import`
- Export/import personas as JSON to share with others
- Configure tone (formality, directness, humor, hedging, warmth)
- Set verbosity preferences (concise/verbose, casual/professional, prose/bullets)
- Define voice traits (words to avoid, words to prefer, characteristic phrases)
- Enable accessibility features like VTT/caption styling

**Active personas influence every response**, not just when you explicitly invoke the skill. The hook runs on every prompt.

---

## Quick Start

### List available personas

```bash
/nothanksona list
```

Output:
```
default: Neutral, professional Claude default behavior. No special persona adjustments. [active]
nothanksona: Chill, slightly snarky, self-aware. Doesn't take itself too seriously. Gets the job done without the corporate gloss.
```

### Switch personas

```bash
/nothanksona use default
```

Personas instantly affect all future responses in this session.

### See the active persona's traits

Check `~/.claude/personas/.active` to see which persona is active:

```bash
cat ~/.claude/personas/.active
# Output: nothanksona
```

Then inspect the persona JSON:

```bash
cat ~/.claude/personas/nothanksona.json | jq .
```

---

## Architecture

- **`~/.claude/personas/*.json`** — Persona config files (default.json, nothanksona.json, custom ones you create)
- **`~/.claude/personas/.active`** — Plain text file pointing to the active persona name
- **`~/.claude/hooks/nothanksona/inject-persona.sh`** — `UserPromptSubmit` hook that reads `.active` and injects persona context into every prompt
- **`~/.claude/skills/nothanksona/SKILL.md`** — Slash command skill with subcommands (use, list, new, edit, export, import, slack-import)

---

## Persona JSON Schema

```json
{
  "meta": {
    "name": "slug-name",
    "description": "Human-readable description",
    "version": "1.0.0",
    "created_at": "2026-03-26T00:00:00Z"
  },
  "tone": {
    "formality": 0.0 - 1.0,     // 0 = casual, 1 = formal
    "directness": 0.0 - 1.0,    // 0 = roundabout, 1 = blunt
    "humor": 0.0 - 1.0,         // 0 = dry, 1 = frequent wit
    "hedging": 0.0 - 1.0,       // 0 = confident, 1 = heavily qualified
    "warmth": 0.0 - 1.0         // 0 = cold, 1 = enthusiastic
  },
  "verbosity": {
    "response_length": "concise|adaptive|verbose",
    "reading_level": "casual|professional|technical",
    "format_preference": "prose|bullets|adaptive",
    "bullet_threshold": "only_when_truly_a_list|multi_step_only|always|never"
  },
  "vtt": {
    "enabled": false,
    "line_length": 42,
    "words_per_minute": 160,
    "pacing_indicators": false,
    "caption_format": "srt|webvtt"
  },
  "keyboard_shortcuts": {
    "note": "Documentation-only (no API yet)",
    "shortcuts": [
      { "key": "ctrl+shift+p", "action": "Switch persona", "note": "/nothanksona use <name>" }
    ]
  },
  "voice": {
    "custom_phrases": ["phrase1", "phrase2"],
    "avoid_words": ["word1", "word2"],
    "prefer_words": ["word3", "word4"],
    "sign_off": null
  }
}
```

---

## Skill Subcommands

### `/nothanksona list`
List all personas, mark the active one.

### `/nothanksona use <name>`
Switch to a different persona. The hook injects its context on your next message.

### `/nothanksona new <name>`
Create a new persona interactively. Walk through tone, verbosity, and voice settings.

### `/nothanksona edit <name>`
Edit an existing persona. Modify individual fields.

### `/nothanksona export <name>`
Export a persona as JSON for sharing.

### `/nothanksona import`
Paste a persona JSON to import it locally.

### `/nothanksona slack-import`
Paste 10–15 of your own Slack messages. Claude analyzes your tone and auto-generates a persona JSON. Review and save.

---

## Current Personas

### `default`
- Neutral, professional, no special adjustments
- All tone values at middle ground
- Adaptive verbosity
- Standard Claude behavior

### `nothanksona`
- Chill and slightly snarky
- Low formality (0.2), high directness (0.85), moderate humor (0.65)
- Concise responses, casual reading level, prose over bullets
- Avoids: "certainly", "leverage", "utilize", corporate jargon
- Prefers: "use", "build", "fix", "yeah", "nope", "honestly"
- Characteristic phrases: "yeah, so", "here's the deal", "quick heads up"

---

## How the Hook Works

Every time you send a message:

1. The `UserPromptSubmit` hook runs `inject-persona.sh`
2. The hook reads `~/.claude/personas/.active` (e.g., `nothanksona`)
3. It loads `~/.claude/personas/nothanksona.json`
4. It converts the numeric tone values (0.0–1.0) to natural language descriptions
5. It constructs a system-level `additionalContext` block
6. It injects this context into the Claude Code prompt **before** the model sees it

Result: Claude's response follows the persona traits without needing explicit reminders.

---

## Testing

### Verify hook output

```bash
echo '{"session_id":"test","prompt":"hi"}' | bash ~/.claude/hooks/nothanksona/inject-persona.sh | jq .
```

Expected: Valid JSON with `continue: true` and `additionalContext` containing persona instructions.

### Verify files exist

```bash
ls ~/.claude/personas/
cat ~/.claude/personas/.active
```

### Test persona switching

In Claude Code:

```
/nothanksona use default
[Now in default tone]

/nothanksona use nothanksona
[Now in chill/snarky tone]
```

Notice the difference in how Claude responds — shorter answers, more casual language, less corporate phrasing.

---

## Future Extensions

- VTT/caption formatting for accessibility
- Real keyboard shortcut hooks (once Claude Code exposes a keybinding API)
- Persona versioning and rollback
- Persona sharing marketplace
- A/B testing personas
- Context-aware persona switching (e.g., switch to "technical" for code reviews)

---

## Files

```
~/.claude/personas/
  default.json
  nothanksona.json
  .active

~/.claude/hooks/nothanksona/
  inject-persona.sh
  inject-persona.log

~/.claude/skills/nothanksona/
  SKILL.md

~/.claude/settings.json
  (UserPromptSubmit hook entry added)
```
