# nothanksona — Cross-Machine Persona Sync

A comprehensive persona management system for Claude Code that persists voice/tone preferences, learns over time, and syncs across machines via S3.

## Quick Start

### 1. Setup Personas Directory

```bash
mkdir -p ~/.claude/personas/{agent,code}
touch ~/.claude/personas/.active
echo "agent/default" > ~/.claude/personas/.active
```

### 2. Create Default Persona

```bash
cat > ~/.claude/personas/agent/default.json << 'EOF'
{
  "meta": {
    "name": "default",
    "description": "Neutral baseline persona",
    "version": "1.0.0",
    "created_at": "2026-03-26T00:00:00Z"
  },
  "tone": {
    "formality": 0.5,
    "directness": 0.5,
    "humor": 0.5,
    "hedging": 0.5,
    "warmth": 0.5
  },
  "verbosity": {
    "response_length": "adaptive",
    "reading_level": "professional",
    "format_preference": "adaptive",
    "bullet_threshold": "multi_step_only"
  },
  "voice": {
    "custom_phrases": [],
    "avoid_words": [],
    "prefer_words": [],
    "sign_off": null
  }
}
EOF
```

### 3. Copy Hooks to ~/.claude/hooks/nothanksona/

From `nothanksona-SKILL-v5.md` or check the `/home/caboose/.claude/hooks/nothanksona/` directory:

- `inject-persona.sh` — UserPromptSubmit hook for injecting persona context
- `session-counter.sh` — Tracks messages, triggers learning nudge at 50
- `s3-sync.sh` — Push/pull personas to S3 (cross-machine sync)
- `accumulate-deltas.sh` — Merge deltas with time-decay weighting
- `scan-code-style.sh` — Auto-detect code conventions
- `fingerprint-cache.sh` — SHA256 fingerprints for token optimization

### 4. Update ~/.claude/settings.json

Add SessionStart hook for S3 pull and PostToolUse hook for S3 push:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/home/caboose/.claude/hooks/nothanksona/s3-sync.sh pull",
            "timeout": 15,
            "async": true
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.command // empty' | grep -q 'accumulate-deltas' && /home/caboose/.claude/hooks/nothanksona/s3-sync.sh push 2>/dev/null || true",
            "timeout": 15,
            "async": true
          }
        ]
      }
    ]
  }
}
```

### 5. (Optional) Setup S3 for Cross-Machine Sync

```bash
# Create S3 bucket (or reuse existing)
aws s3api create-bucket --bucket my-personas --region us-east-1

# Update ~/.claude/settings.json env:
{
  "env": {
    "AWS_PROFILE": "default",
    "AWS_REGION": "us-east-1"
  }
}
```

## Features

### ✨ Natural Language Mood Tweaks

Apply temporary tone adjustments:

```bash
/nothanksona mood make me sound happier
/nothanksona mood pissed
/nothanksona mood more professional
```

Session-only by default. Bake into persona permanently with `/nothanksona mood save`.

### 📚 Incremental Learning

Analyze conversation patterns and suggest persona updates:

```bash
/nothanksona learn
```

Proposed updates show confidence scores. Apply with `y` → merges deltas into persona JSON with 30-day time-decay weighting.

### 🔄 Cross-Machine Sync

- **SessionStart**: Auto-pulls latest personas from S3
- **After `/nothanksona learn --accumulate`**: Auto-pushes updated personas to S3
- Graceful fallback if AWS not configured

### ⚡ Token Optimization

Persona context reduces from ~200 tokens (first session) to ~10 tokens (subsequent sessions) via fingerprint caching + model shorthand learning.

### 🎨 Code Style Personas

Scan repositories for coding conventions:

```bash
/nothanksona code-scan /path/to/repo
/nothanksona code-style
```

Stored at `~/.claude/personas/code/coding-style.json` (separate from agent domain).

## File Structure

```
~/.claude/personas/
├── .active                    # Current persona (e.g., "agent/nothanksona")
├── .mood                      # Session-only mood overlay (optional)
├── .session-count             # Messages this session (auto-reset)
├── agent/
│   ├── default.json           # Baseline neutral persona
│   ├── nothanksona.json       # Example persona (casual, direct)
│   └── .deltas                # JSONL incremental learning deltas
└── code/
    ├── coding-style.json      # Code conventions template
    └── .deltas                # Code style evolution deltas
```

## Workflow Example

**Machine A (Mac):**
```bash
# 1. Use a custom persona
/nothanksona use agent/nothanksona

# 2. Have a conversation...

# 3. Learn and update
/nothanksona learn
→ Suggested: humor 0.65 → 0.75, directness 0.85 → 0.9
→ Apply? (y/n)
→ yes
→ Automatically pushed to S3
```

**Machine B (Linux, 1 hour later):**
```bash
# 1. Start new session
# SessionStart hook auto-pulls from S3

# 2. Your persona from Machine A is now active
/nothanksona list
→ agent/default
→ agent/nothanksona  [active]  ← Updated with your learning

# 3. Continue with latest persona
```

## Commands Reference

| Command | Purpose |
|---------|---------|
| `/nothanksona use <name>` | Switch active persona |
| `/nothanksona list` | Show all personas (mark active) |
| `/nothanksona new <name>` | Create persona interactively |
| `/nothanksona edit <name>` | Modify existing persona |
| `/nothanksona export <name>` | Export as JSON (shareable) |
| `/nothanksona import` | Import from pasted JSON |
| `/nothanksona slack-import` | Generate from Slack messages |
| `/nothanksona mood <desc>` | Temporary tone adjustment |
| `/nothanksona mood reset` | Clear mood overlay |
| `/nothanksona mood save` | Bake mood into persona |
| `/nothanksona learn` | Analyze session, suggest updates |
| `/nothanksona learn --accumulate` | Merge deltas into persona (triggers S3 push) |
| `/nothanksona code-scan <path>` | Detect code style conventions |
| `/nothanksona code-style` | View current code style |

## Mood Mapping (Natural Language)

| Input | Tone Changes |
|-------|--------------|
| happier / upbeat | humor +0.2, warmth +0.2 |
| pissed / annoyed | directness +0.3, warmth -0.2 |
| passive aggressive | hedging +0.1, warmth -0.1 |
| chill / relaxed | formality -0.1, humor +0.1 |
| more professional | formality +0.3, humor -0.2 |
| more concise | response_length = concise |
| tired / low energy | humor -0.1, response_length = concise |
| enthusiastic | warmth +0.3, humor +0.1 |

## Persona JSON Schema

```json
{
  "meta": {
    "name": "slug-name",
    "description": "Human-readable description",
    "version": "semver",
    "created_at": "ISO 8601"
  },
  "tone": {
    "formality": "0.0 (casual) to 1.0 (formal)",
    "directness": "0.0 (roundabout) to 1.0 (blunt)",
    "humor": "0.0 (dry) to 1.0 (witty)",
    "hedging": "0.0 (confident) to 1.0 (qualified)",
    "warmth": "0.0 (cold) to 1.0 (enthusiastic)"
  },
  "verbosity": {
    "response_length": "concise | adaptive | verbose",
    "reading_level": "casual | professional | technical",
    "format_preference": "prose | bullets | adaptive",
    "bullet_threshold": "only_when_truly_a_list | multi_step_only | always | never"
  },
  "voice": {
    "custom_phrases": ["use occasionally"],
    "avoid_words": ["never use"],
    "prefer_words": ["over synonyms"],
    "sign_off": "string or null"
  }
}
```

## Troubleshooting

### Persona not injecting?
```bash
echo '{}' | bash ~/.claude/hooks/nothanksona/inject-persona.sh
# Should output JSON with `additionalContext` field
```

### S3 sync errors?
Check AWS credentials:
```bash
aws --profile default sts get-caller-identity
# If fails, reconfigure: aws configure --profile default
```

### Session counter not nudging?
Set counter manually and test:
```bash
printf '49' > ~/.claude/personas/.session-count
echo '{}' | bash ~/.claude/hooks/nothanksona/session-counter.sh
# Should see systemMessage nudge at count=50
```

### Deltas not merging?
Run accumulate manually:
```bash
bash ~/.claude/hooks/nothanksona/accumulate-deltas.sh agent nothanksona 30
```

## Token Optimization Details

**Session 1** (~200 tokens):
```
[PERSONA: nothanksona]
Tone: formality 0.2, directness 0.85, humor 0.65, hedging 0.1, warmth 0.55
Verbosity: response_length concise, reading_level casual
Voice: custom_phrases: ["yea breh"], avoid_words: []
... (full JSON)
```

**Session 2+** (~10 tokens):
```
[PERSONA FINGERPRINT: nk:nothanksona:1.0:f3d6b2a1e9c4]
(Model recognizes shorthand from Session 1)
```

~95% reduction in persona context tokens.

## See Also

- `nothanksona-SKILL-v5.md` — Full skill documentation
- `/home/caboose/.claude/hooks/nothanksona/` — Hook scripts
- `~/.claude/personas/` — Persona storage
- `~/.claude/settings.json` — Hook configuration

