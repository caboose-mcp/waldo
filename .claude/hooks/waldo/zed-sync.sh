#!/bin/bash
# waldo zed-sync.sh — Sync active waldo persona to Zed's Rules system
#
# Generates a Markdown rules file from your active persona JSON and writes it
# to ~/.config/zed/rules/waldo.md for use in Zed's Rules Library, or to a
# project's .rules file for per-project injection.
#
# Usage:
#   zed-sync.sh                    # Write to ~/.config/zed/rules/waldo.md
#   zed-sync.sh project [dir]      # Write .rules to project dir (default: cwd)
#   zed-sync.sh print              # Print rules markdown to stdout
#
# After running (global mode), open Zed's Rules Library (ctrl-alt-l),
# find waldo.md, and pin it as a default rule.

set -euo pipefail

WALDO_CONFIG="${WALDO_CONFIG:-$HOME/.config/waldo}"
PERSONAS_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/personas"
ZED_RULES_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/zed/rules"
ACTIVE_FILE="$WALDO_CONFIG/.active"
MOOD_FILE="$WALDO_CONFIG/.mood"

# ---------------------------------------------------------------------------
# Helpers: translate numeric tone values to prose labels (pure awk, no bc)
# ---------------------------------------------------------------------------

tone_label() {
  local val="$1" type="$2"
  awk -v v="$val" -v t="$type" 'BEGIN {
    if (t == "formality") {
      if      (v < 0.3) print "very casual"
      else if (v < 0.5) print "casual"
      else if (v < 0.7) print "professional"
      else              print "formal"
    } else if (t == "directness") {
      if      (v < 0.3) print "roundabout, lots of context"
      else if (v < 0.6) print "balanced"
      else if (v < 0.85) print "direct"
      else               print "blunt, no fluff"
    } else if (t == "humor") {
      if      (v < 0.2) print "none"
      else if (v < 0.5) print "dry, occasional"
      else if (v < 0.75) print "moderate wit"
      else               print "frequent, casual"
    } else if (t == "hedging") {
      if      (v < 0.3) print "confident, no qualifiers"
      else if (v < 0.6) print "light hedging"
      else               print "heavily qualified"
    } else if (t == "warmth") {
      if      (v < 0.3) print "cold, transactional"
      else if (v < 0.6) print "neutral"
      else               print "warm, enthusiastic"
    }
  }'
}

# ---------------------------------------------------------------------------
# Preflight checks
# ---------------------------------------------------------------------------

if ! command -v jq &>/dev/null; then
  echo "waldo zed-sync: jq is required (pacman -S jq)" >&2
  exit 1
fi

if [[ ! -f "$ACTIVE_FILE" ]]; then
  echo "waldo zed-sync: no active persona found at $ACTIVE_FILE" >&2
  echo "  Run: /waldo use agent/<name>" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Load persona
# ---------------------------------------------------------------------------

ACTIVE=$(cat "$ACTIVE_FILE")
PERSONA_FILE="$PERSONAS_DIR/${ACTIVE}.json"

if [[ ! -f "$PERSONA_FILE" ]]; then
  echo "waldo zed-sync: persona file not found: $PERSONA_FILE" >&2
  exit 1
fi

jq_get() { jq -r "${1} // ${2}" "$PERSONA_FILE"; }

NAME=$(jq_get '.meta.name' '"unknown"')
DESC=$(jq_get '.meta.description' '""')
VERSION=$(jq_get '.meta.version' '"1.0.0"')

FORMALITY=$(jq_get '.tone.formality' '0.5')
DIRECTNESS=$(jq_get '.tone.directness' '0.5')
HUMOR=$(jq_get '.tone.humor' '0.5')
HEDGING=$(jq_get '.tone.hedging' '0.5')
WARMTH=$(jq_get '.tone.warmth' '0.5')

RESPONSE_LENGTH=$(jq_get '.verbosity.response_length' '"adaptive"')
READING_LEVEL=$(jq_get '.verbosity.reading_level' '"professional"')
FORMAT_PREF=$(jq_get '.verbosity.format_preference' '"adaptive"')
BULLET_THRESHOLD=$(jq_get '.verbosity.bullet_threshold' '"multi_step_only"')

AVOID_WORDS=$(jq -r '.voice.avoid_words // [] | if length > 0 then map("`\(.)` ") | join("") else "" end' "$PERSONA_FILE")
PREFER_WORDS=$(jq -r '.voice.prefer_words // [] | if length > 0 then map("`\(.)` ") | join("") else "" end' "$PERSONA_FILE")
CUSTOM_PHRASES=$(jq -r '.voice.custom_phrases // [] | if length > 0 then map("\"\(.)\" ") | join("") else "" end' "$PERSONA_FILE")
SIGN_OFF=$(jq_get '.voice.sign_off' 'null')

# ---------------------------------------------------------------------------
# Mood overlay (session-scoped, optional)
# ---------------------------------------------------------------------------

MOOD_BLOCK=""
if [[ -f "$MOOD_FILE" ]]; then
  MOOD=$(cat "$MOOD_FILE")
  MOOD_BLOCK="
## Active Mood Overlay

A session mood is active: **${MOOD}**

Adjust tone to reflect this mood while keeping the base persona intact. This is temporary and not saved to the persona.
"
fi

# ---------------------------------------------------------------------------
# Voice section (only emit non-empty fields)
# ---------------------------------------------------------------------------

VOICE_BLOCK=""
[[ -n "$AVOID_WORDS" ]]   && VOICE_BLOCK+="- **Never use:** $AVOID_WORDS"$'\n'
[[ -n "$PREFER_WORDS" ]]  && VOICE_BLOCK+="- **Prefer:** $PREFER_WORDS"$'\n'
[[ -n "$CUSTOM_PHRASES" ]] && VOICE_BLOCK+="- **Use occasionally:** $CUSTOM_PHRASES"$'\n'
[[ -n "$SIGN_OFF" && "$SIGN_OFF" != "null" ]] && VOICE_BLOCK+="- **Sign off with:** ${SIGN_OFF}"$'\n'

if [[ -n "$VOICE_BLOCK" ]]; then
  VOICE_BLOCK=$'\n## Voice\n\n'"$VOICE_BLOCK"
fi

# ---------------------------------------------------------------------------
# Bullet threshold → human prose
# ---------------------------------------------------------------------------

bullet_prose() {
  case "$1" in
    always)                 echo "always use bullets" ;;
    never)                  echo "never use bullets, prose only" ;;
    multi_step_only)        echo "only for multi-step sequences" ;;
    only_when_truly_a_list) echo "only when content is genuinely a list" ;;
    *)                      echo "$1" ;;
  esac
}

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# ---------------------------------------------------------------------------
# Build rules markdown
# ---------------------------------------------------------------------------

RULES=$(cat <<MARKDOWN
# Waldo Persona: ${NAME}

> Generated by waldo on ${TIMESTAMP}
> Persona \`${ACTIVE}\` · v${VERSION}
> ${DESC}

You are acting as an AI assistant whose responses reflect the following persona.
Apply these preferences consistently across all replies in this session.

## Tone

| Dimension  | Score | Style |
|------------|-------|-------|
| Formality  | ${FORMALITY} | $(tone_label "$FORMALITY" formality) |
| Directness | ${DIRECTNESS} | $(tone_label "$DIRECTNESS" directness) |
| Humor      | ${HUMOR} | $(tone_label "$HUMOR" humor) |
| Hedging    | ${HEDGING} | $(tone_label "$HEDGING" hedging) |
| Warmth     | ${WARMTH} | $(tone_label "$WARMTH" warmth) |

## Verbosity

- **Response length:** ${RESPONSE_LENGTH}
- **Reading level:** ${READING_LEVEL}
- **Format preference:** ${FORMAT_PREF}
- **Use bullets:** $(bullet_prose "$BULLET_THRESHOLD")
${VOICE_BLOCK}${MOOD_BLOCK}
---
*Managed by [waldo](https://github.com/caboose-mcp/waldo) · run \`bash ~/.claude/hooks/waldo/zed-sync.sh\` to update after switching personas.*
MARKDOWN
)

# ---------------------------------------------------------------------------
# Output routing
# ---------------------------------------------------------------------------

MODE="${1:-global}"

case "$MODE" in
  global)
    mkdir -p "$ZED_RULES_DIR"
    printf '%s\n' "$RULES" > "$ZED_RULES_DIR/waldo.md"
    echo "waldo: ✓ wrote Zed rules → $ZED_RULES_DIR/waldo.md"
    echo ""
    echo "  To activate in Zed:"
    echo "    1. Open Rules Library  ctrl-alt-l"
    echo "    2. Locate waldo.md and click the 📎 pin icon to set as default"
    echo "    3. Rules will inject automatically into every Agent Panel session"
    ;;

  project)
    TARGET_DIR="${2:-.}"
    OUTPUT="$TARGET_DIR/.rules"
    printf '%s\n' "$RULES" > "$OUTPUT"
    echo "waldo: ✓ wrote project rules → $OUTPUT"
    echo "  Zed will auto-load this file in Agent Panel sessions for this project."
    ;;

  print)
    printf '%s\n' "$RULES"
    ;;

  *)
    echo "Usage: zed-sync.sh [global|project [dir]|print]" >&2
    echo "" >&2
    echo "  global           Write to $ZED_RULES_DIR/waldo.md  (default)" >&2
    echo "  project [dir]    Write .rules to project dir (default: cwd)" >&2
    echo "  print            Print markdown to stdout" >&2
    exit 1
    ;;
esac
