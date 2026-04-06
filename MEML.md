# MEML: Persona Configuration Format

waldo uses **MEML** (Emoji Markup Language) to store personas. MEML is a TOML-inspired config format with first-class emoji support.

## Why MEML?

- **Human-readable** — comments inline, emoji section decorators
- **Self-documenting** — emoji hints clarify intent (`🎭 tone`, `🗣️ voice`)
- **Lintable** — `meml validate` catches syntax errors with line numbers
- **Convertible** — `meml dump` converts `.meml` → JSON at read time
- **Expressive** — `✅`/`❌` booleans, emoji atoms, type hints with emoji

## Anatomy of a Persona MEML File

```meml
# Header comment
[🪪 meta]
name        = "chris-marasco"
description = "Casual, direct, self-aware persona"
version     = "0.1.0"
created_at  = "2026-03-26T00:00:00Z"

[🎭 tone]
formality   = 0.28  # 0 = very casual, 1 = very formal
directness = 0.82  # 0 = roundabout, 1 = blunt
humor       = 0.68  # 0 = none, 1 = frequent wit
hedging     = 0.12  # 0 = confident, 1 = heavily qualified
warmth      = 0.58  # 0 = cold, 1 = enthusiastic

[📢 verbosity]
response_length   = "concise"       # concise | adaptive | verbose
reading_level     = "casual"        # casual | professional | technical
format_preference = "prose"         # prose | bullets | adaptive
bullet_threshold  = "only_when_truly_a_list"

[😊 emoji]
enabled   = ✅
frequency = "sparse"  # sparse | moderate | frequent
contexts  = ["emphasis", "mood"]

[🗣️ voice]
avoid_words    = ["certainly", "leverage", "utilize"]
prefer_words   = ["use", "build", "fix", "get"]
custom_phrases = ["yea", "no dice", "here's the deal"]
sign_off       = ~  # null value

[♿ vtt]
enabled            = ❌
line_length        = 42
words_per_minute   = 160
pacing_indicators  = ❌
caption_format     = "srt"

[⚙️ workflow]
[pre_pr]
required = ✅
steps    = ["owasp_top10_review", "lint", "test", "build"]
note     = ~  # optional context shown alongside the checklist

[git]
branch_from       = "main"         # always cut branches from here
no_direct_push    = ["main"]       # branches that must never be pushed to directly
commit_convention = "conventional" # conventional | none
ai_attribution    = ❌             # never add Co-Authored-By AI footer
```

## MEML Syntax Crash Course

### Sections
```meml
[name]           # Plain section
[🎭 tone]        # Emoji-decorated section (emoji is metadata, optional)
[🔑]             # Emoji IS the section name
["my section"]   # Quoted section (allows spaces)
```

### Values
```meml
string = "double quotes" or 'single quotes'
multi  = """
triple-quoted string
with newlines
"""
number = 42 or 3.14 or -1
bool   = true or false or ✅ or ❌
null   = null or ~
emoji  = 🎭 or 🟢 (semantic emoji atoms)
array  = ["a", "b", "c"] or [1, 2, 3] or ["mixed", 123, true]
table  = { key = "value", port = 8080 }
```

### Comments
```meml
# Hash comments
💬 Emoji comments work too
```

### Emoji Annotations (on keys)
```meml
🔑 secret_token = "xoxb-..."  # 🔑 marks sensitive data
📁 path = "/home/user"        # 📁 marks file paths
🌍 url = "https://..."        # 🌍 marks URLs
🔢 count = 42                 # 🔢 hints integer
📋 items = [1, 2, 3]          # 📋 hints list
⚠️ old_field = "deprecated"   # ⚠️ marks deprecated
```

## Creating a New Persona in MEML

1. **Copy the template:**
   ```bash
   cp example.meml ~/.config/waldo/personas/agent/my-voice.meml
   ```

2. **Edit the `[🪪 meta]` section:**
   ```meml
   name        = "my-voice"
   description = "Your persona description"
   version     = "0.1.0"
   ```

3. **Adjust tone values (0.0 to 1.0):**
   ```meml
   [🎭 tone]
   formality   = 0.3    # your target formality
   directness = 0.8    # how direct you are
   humor       = 0.6    # how funny
   hedging     = 0.1    # how much you qualify statements
   warmth      = 0.5    # how friendly
   ```

4. **Set verbosity:**
   ```meml
   [📢 verbosity]
   response_length   = "concise"  # how long responses are
   reading_level     = "casual"   # complexity
   format_preference = "prose"    # bullets or flowing text
   ```

5. **Define your voice:**
   ```meml
   [🗣️ voice]
   avoid_words  = ["certainly", "leverage"]  # Don't use these
   prefer_words = ["use", "build", "fix"]    # Use these instead
   custom_phrases = ["yea", "no dice"]       # Your catchphrases
   ```

6. **Validate syntax:**
   ```bash
   meml validate ~/.config/waldo/personas/agent/my-voice.meml
   ```

7. **Activate:**
   ```bash
   /waldo use agent/my-voice
   ```

## How Personas Get Injected

The `inject-persona.sh` hook runs on `UserPromptSubmit`:

```bash
# Read active persona name
PERSONA=$(cat ~/.config/waldo/.active)  # e.g., "agent/chris-marasco"

# Convert .meml → JSON at read time
PERSONA_JSON=$(meml dump ~/.config/waldo/personas/$PERSONA.meml)

# Inject into Claude context (via hook)
echo "{...persona context...}" >> prompt
```

**Why this approach?**
- MEML files stay readable with comments and emoji
- No need to convert ahead of time (`.meml` is source of truth)
- `meml dump` handles any MEML syntax
- Falls back to `.json` if `.meml` doesn't exist

## Tone Values Explained

Each tone dimension is a **0.0 to 1.0 scale**:

### Formality
- 0.0 = very casual (lowercase, contractions, slang)
- 0.5 = moderate (professional but approachable)
- 1.0 = very formal (corporate, no contractions)

### Directness
- 0.0 = roundabout (lots of context, hedges)
- 0.5 = balanced (some hedges, some directness)
- 1.0 = blunt (get to the point, no preamble)

### Humor
- 0.0 = dry/none (serious, no jokes)
- 0.5 = moderate (light wit, occasional jokes)
- 1.0 = frequent wit (jokes in most responses)

### Hedging
- 0.0 = confident (assertions without qualifiers)
- 0.5 = balanced (some "maybe", some "definitely")
- 1.0 = heavily qualified (lots of "might", "perhaps")

### Warmth
- 0.0 = cold/clinical (no enthusiasm, task-focused)
- 0.5 = moderate (friendly and approachable)
- 1.0 = enthusiastic (lots of emoji, exclamation marks)

## Emoji Configuration

```meml
[😊 emoji]
enabled   = ✅            # true/false (or ✅/❌)
frequency = "sparse"      # sparse | moderate | frequent
contexts  = ["emphasis", "mood", "transitions"]
```

- **sparse** — 0–2 emoji per response
- **moderate** — 2–5 emoji per response
- **frequent** — 5+ emoji, liberal use

Claude respects these as hints but decides final usage based on context.

## Verbosity Options

### response_length
- `"concise"` — short, punchy responses
- `"adaptive"` — match input length (default)
- `"verbose"` — long-form, detailed explanations

### reading_level
- `"casual"` — conversational, accessible
- `"professional"` — business English, clear
- `"technical"` — specialized terms, assumes domain knowledge

### format_preference
- `"prose"` — flowing paragraphs
- `"bullets"` — lists and bullet points
- `"adaptive"` — use what fits (default)

### bullet_threshold
- `"only_when_truly_a_list"` — bullets only for actual lists
- `"multi_step_only"` — bullets for instructions or sequences
- `"always"` — format everything as bullets
- `"never"` — always prose

## Voice Configuration

### avoid_words
List of words/phrases to never use:
```meml
avoid_words = [
  "certainly",
  "absolutely",
  "I'd be happy to",
  "leverage",
  "synergy"
]
```

### prefer_words
List of alternatives to use instead:
```meml
prefer_words = [
  "use",
  "build",
  "fix",
  "check",
  "get"
]
```

When you'd normally say "utilize", Claude sees "prefer: use" and uses that instead.

### custom_phrases
List of phrases that are characteristic of your voice:
```meml
custom_phrases = [
  "yea",
  "no dice",
  "here's the deal",
  "not gonna lie"
]
```

Claude uses these occasionally (not every message) to capture your authentic voice.

### attribution
Whether to include AI agent attribution in responses (e.g., "Generated by Claude Code"):
```meml
attribution = ❌  # Don't mention AI agent by default
```

- `❌` (false) — No attribution
- `✅` (true) — Include attribution line

This is a persona-level choice, not enforced. Claude respects the hint.

## Migrating from JSON to MEML

If you have existing JSON personas:

```bash
# Convert JSON to MEML manually
# (Or keep both — waldo supports .json fallback)

cp ~/.claude/personas/agent/my-voice.json.backup my-voice.meml
# Edit my-voice.meml manually to add emoji decorators and comments
meml validate my-voice.meml
```

The hook will use `.meml` if it exists, otherwise fall back to `.json`.

## Validating Your Persona

```bash
meml validate ~/.config/waldo/personas/agent/my-voice.meml
```

Output on success:
```
✓ Valid MEML syntax
```

Output on error:
```
Line 12: Expected '=' after key 'formality'
```

Fix the line number and re-validate.

## Sharing Personas

Export as JSON (still readable):
```bash
/waldo export chris-marasco
```

Or manually convert:
```bash
meml dump ~/.config/waldo/personas/agent/chris-marasco.meml | jq .
```

Share the JSON with others. They can import it:
```bash
/waldo import
# Paste the JSON
```

## Advanced: Custom Emoji Hints

Use emoji annotations on keys to hint at tooling (documentation only; Claude doesn't enforce):

```meml
[🗣️ voice]
🔑 secret_phrase    = "..."   # Sensitive custom phrase
⚠️ deprecated_avoid = ["old"]  # No longer used
```

These are just metadata for humans reading the file. The hook ignores them.

## Performance: Token Overhead

When a persona is injected, waldo computes a fingerprint:

```
nk:waldo:0.1:abc123def456
```

- **First session:** Full persona context (~200 tokens)
- **Subsequent sessions:** Fingerprint + short summary (~10 tokens)

Claude learns the shorthand, reducing per-message overhead by 95%.

---

**See also:**
- [example.meml](./example.meml) — Working example
- [meml GitHub](https://github.com/caboose-mcp/meml) — MEML spec and CLI
- [WALDO.md](./WALDO.md) — Core system docs
