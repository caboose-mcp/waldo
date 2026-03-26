---
name: nothanksona
description: Manage Claude response personas — agent (tone, voice) and code (style, conventions). Subcommands: use, list, edit, export, import, slack-import, mood, learn (agent); code-scan, code-style (code). Use when switching personas, tweaking voice/tone, analyzing Slack, applying mood overlays, learning from session (agent), or scanning/viewing code style conventions (code).
user_invocable: true
---

# nothanksona

Manage response personas that shape Claude's tone, verbosity, and voice style.

## Persona file location

All personas live at `~/.claude/personas/<name>.json`. The active persona name is stored in `~/.claude/personas/.active` as plain text. A `UserPromptSubmit` hook reads `.active` on every prompt and injects the persona as context.

## Subcommands

### `/nothanksona use <name>`

Switch the active persona.

1. Read `~/.claude/personas/.active` to see the current persona.
2. Verify `~/.claude/personas/<name>.json` exists. If not, list available personas and ask the user to choose one or create a new one.
3. Write the name (and nothing else) to `~/.claude/personas/.active`:
   ```bash
   printf '%s' "<name>" > ~/.claude/personas/.active
   ```
4. Confirm: "Persona switched to <name>. It will apply starting from your next message."

### `/nothanksona list`

List all available personas.

1. Run:
   ```bash
   ls ~/.claude/personas/*.json 2>/dev/null
   ```
2. Read `~/.claude/personas/.active` to identify the currently active one.
3. For each persona file, extract `.meta.name` and `.meta.description` using jq:
   ```bash
   jq -r '"\(.meta.name): \(.meta.description)"' ~/.claude/personas/<name>.json
   ```
4. Present as a formatted list. Mark the active persona with `[active]`.

### `/nothanksona new <name>`

Create a new persona interactively.

1. Ask the user for each section if they don't provide it upfront:
   - Description (free text)
   - Tone: formality, directness, humor, hedging, warmth — each as 0.0–1.0. Offer named presets: "casual" (0.2), "professional" (0.6), "high" (0.8).
   - Verbosity: response_length (concise/adaptive/verbose), reading_level (casual/professional/technical), format_preference (prose/bullets/adaptive)
   - Voice: words to avoid, words to prefer, custom phrases
   - VTT: whether to enable caption mode
2. Build the JSON structure following the schema.
3. Write to `~/.claude/personas/<name>.json`.
4. Ask if the user wants to activate it now.

### `/nothanksona edit <name>`

Edit an existing persona.

1. Read the current persona file:
   ```bash
   cat ~/.claude/personas/<name>.json
   ```
2. Ask the user which fields they want to change.
3. Apply changes with the Edit tool (surgical field updates, not full rewrites).
4. Confirm what changed.

### `/nothanksona export <name>`

Export a persona to a shareable JSON snippet.

1. Read `~/.claude/personas/<name>.json`.
2. Print the full JSON contents in a code block so the user can copy it.
3. Note: "You can share this with others. They can import it with `/nothanksona import`."

### `/nothanksona import`

Import a persona from JSON the user pastes.

1. Ask the user to paste the persona JSON.
2. Validate required fields: `meta.name`, `meta.version`, `tone`, `verbosity`, `voice`.
3. Sanitize the name: only letters, numbers, underscores, hyphens. Reject path traversal attempts.
4. If a persona with that name already exists, ask whether to overwrite or rename.
5. Write to `~/.claude/personas/<name>.json`.
6. Confirm and ask if they want to activate it.

### `/nothanksona slack-import`

Generate a persona from Slack message samples.

See the "Slack Import Flow" section below for full instructions.

### `/nothanksona mood <description>`

Apply a session-only mood overlay with natural language.

Examples: "make me sound happier", "pissed", "passive aggressive", "more concise", "more professional"

1. **SFW guardrails**: Block any mood request containing slurs, explicit content, identity changes, hate speech, or offensive language. Reject requests that would add unsafe words to avoid_words/prefer_words/custom_phrases.
2. **Map natural language to tone delta**: Use this mapping—
   - "happier" / "upbeat" → humor +0.2, warmth +0.2
   - "pissed" / "annoyed" → directness +0.3, warmth -0.2
   - "passive aggressive" → hedging +0.1, warmth -0.1
   - "chill" / "relaxed" → formality -0.1, humor +0.1
   - "more professional" → formality +0.3, humor -0.2
   - "more concise" → response_length = "concise"
   - "tired" / "low energy" → humor -0.1, response_length = "concise"
   - "enthusiastic" → warmth +0.3, humor +0.1
   - Custom mappings also allowed, within reason
3. **Build and write `.mood` file** at `~/.claude/personas/.mood`:
   ```json
   {
     "source": "<description>",
     "expires": "session",
     "overrides": {
       "tone": { /* delta fields */ },
       "verbosity": { /* delta fields */ }
     }
   }
   ```
4. Confirm: "Mood overlay applied: <description>. It's session-only — run `/nothanksona mood save` to make it permanent, or `/nothanksona mood reset` to clear it."

### `/nothanksona mood reset`

Clear the active mood overlay.

1. Delete `~/.claude/personas/.mood` if it exists.
2. Confirm: "Mood overlay cleared. Back to base persona."

### `/nothanksona mood save`

Bake the active mood overlay permanently into the persona JSON.

1. Read `~/.claude/personas/.mood` to get the overrides.
2. Load the active persona from `~/.claude/personas/.active`.
3. Merge mood overrides into the persona's tone and verbosity fields (update, don't replace).
4. Write updated persona back to its JSON file.
5. Delete `~/.claude/personas/.mood`.
6. Confirm: "Mood saved to <persona-name>. It's now permanent (until you edit it again)."

### `/nothanksona learn`

Analyze this session's conversation patterns and suggest persona updates.

1. **Review session context**: Look back at the user's prompts and your responses in this session. Identify:
   - Average message length (terse vs. long-form)
   - Use of slang, technical terms, humor, formality
   - Directness level (leading with conclusions vs. hedging)
   - Emoji or punctuation patterns
   - Topics and depth
2. **Compare against active persona**: Load the active persona JSON and analyze deltas.
3. **Produce suggested updates** with reasoning. Example:
   ```
   Suggested updates to nothanksona:
   - humor: 0.65 → 0.75  (you've been more playful/sarcastic this session)
   - directness: 0.85 → 0.9  (consistently leading with conclusions)
   - response_length: concise → adaptive  (you asked several detailed questions)
   ```
4. Ask: "Apply these? (y/n) — or 'save as new <name>' to fork a new persona with these changes?"
5. On yes, write updates to active persona JSON using Edit tool. On "save as new", create a new persona with these settings.

---

## Slack Import Flow

This command analyzes a user's Slack writing style and generates a persona config that mirrors it.

### Step 1 — Collect samples

Ask the user to paste at least 10–15 Slack messages they wrote. Instruct them:

> "Paste a representative sample of your own Slack messages — things you wrote, not replies to you. The more variety the better. You can paste them as a raw block."

### Step 2 — Analyze the samples

Analyze the pasted messages for these signals:

**Tone signals:**
- Formality: count contractions, slang, emoji, lowercase-only sentences vs. proper punctuation
- Directness: average sentence length, ratio of qualifiers ("maybe", "I think", "sort of") to assertions
- Humor: presence of jokes, self-deprecation, memes, absurdist phrasing
- Hedging: frequency of "might", "could", "perhaps", "I think", "probably", "not sure but"
- Warmth: greeting patterns, emoji, affirmations ("nice!", "love this", "sounds good")

**Verbosity signals:**
- Average message length in words
- Bullet vs. prose ratio
- Use of headers or structure in longer messages

**Voice signals:**
- Characteristic phrases that appear 2+ times
- Words used significantly more often than baseline English
- Corporate buzzwords (to flag for avoid_words if the user uses many)
- Sign-off patterns

### Step 3 — Produce scores

Convert the analysis into 0.0–1.0 scores with brief reasoning. Show your work:

```
Formality: 0.25 — uses lowercase, frequent contractions ("it's", "don't"), no formal salutations
Directness: 0.80 — short messages, leads with conclusion, few qualifiers
Humor: 0.55 — occasional dry jokes, some self-deprecation
Hedging: 0.15 — confident assertions, "maybe" appears only twice in 15 messages
Warmth: 0.60 — occasional emoji, uses "nice" and "sounds good"
```

### Step 4 — Build persona JSON

Construct the full persona JSON and show it to the user for review before saving. Let them adjust any scores.

### Step 5 — Save and optionally activate

Ask: "What should I name this persona?" then write the file and optionally activate it.

---

## Schema reference

```json
{
  "meta": {
    "name": "string — slug, no spaces",
    "description": "string",
    "version": "string — semver",
    "created_at": "ISO 8601 timestamp"
  },
  "tone": {
    "formality":  "0.0 (very casual) to 1.0 (very formal)",
    "directness": "0.0 (roundabout) to 1.0 (blunt and fast)",
    "humor":      "0.0 (dry/none) to 1.0 (frequent wit)",
    "hedging":    "0.0 (confident) to 1.0 (heavily qualified)",
    "warmth":     "0.0 (cold/clinical) to 1.0 (enthusiastic)"
  },
  "verbosity": {
    "response_length":   "concise | adaptive | verbose",
    "reading_level":     "casual | professional | technical",
    "format_preference": "prose | bullets | adaptive",
    "bullet_threshold":  "only_when_truly_a_list | multi_step_only | always | never"
  },
  "vtt": {
    "enabled":           "boolean",
    "line_length":       "integer — characters per caption line",
    "words_per_minute":  "integer — target reading pace",
    "pacing_indicators": "boolean — add [pause] and [beat] hints",
    "caption_format":    "srt | webvtt"
  },
  "keyboard_shortcuts": {
    "note": "Documentation only — Claude Code has no keybinding API",
    "shortcuts": [
      { "key": "string", "action": "string", "note": "string" }
    ]
  },
  "voice": {
    "custom_phrases": "array of strings — use occasionally, not every message",
    "avoid_words":    "array of strings — never use these",
    "prefer_words":   "array of strings — use these over synonyms",
    "sign_off":       "string or null"
  }
}
```

---

## Error handling

- **Persona file not found**: List available personas. Offer to create the missing one or switch to default.
- **Invalid JSON during import**: Show the validation error and ask the user to fix the paste.
- **Empty `.active` file**: Treat as "default". If `default.json` is missing, warn the user and explain how to create it.
- **Name with path separators or special chars**: Reject with a clear message. Only `[a-zA-Z0-9_-]` is allowed.

---

## Manual hook test

To verify the hook is injecting context correctly, run:

```bash
echo '{"session_id":"test","prompt":"hello"}' | bash ~/.claude/hooks/nothanksona/inject-persona.sh
```

Expected output shape:
```json
{"continue": true, "hookSpecificOutput": {"hookEventName": "UserPromptSubmit", "additionalContext": "..."}}
```

---

## Code Domain (Beta)

Manage and evolve coding style profiles alongside agent personas.

### `/nothanksona code-scan <repo-path>`

Auto-scan a repository for coding conventions.

1. Run the quick code scanner: `bash ~/.claude/hooks/nothanksona/scan-code-style.sh <repo-path>`
2. The scanner:
   - Finds 10–20 code files (TS, JS, Python, Go, Rust, etc.)
   - Respects `.gitignore` and skips node_modules/dist/build
   - Extracts: indentation (spaces/tabs), naming conventions (camelCase/snake_case), line length, imports style, comment patterns, error handling, type hints
3. Output saved to `~/.claude/personas/code/coding-style.json`
4. Confirm: "Code style scanned and saved. Review with `/nothanksona code-style`"

### `/nothanksona code-style`

View or edit the current code style profile.

1. Read `~/.claude/personas/code/coding-style.json`
2. Pretty-print the profile — indentation, naming, line length, imports, comments, functions, error handling, types
3. Ask: "Want to edit any of these? (run `/nothanksona edit code-style` to modify)"

### `/nothanksona code-learn`

Analyze code in this session and suggest style updates.

(Future: track code written/reviewed this session, suggest tweaks to coding-style.json based on patterns observed)
