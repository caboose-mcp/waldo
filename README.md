[![CI](https://github.com/caboose-mcp/waldo/actions/workflows/ci.yml/badge.svg)](https://github.com/caboose-mcp/waldo/actions/workflows/ci.yml)
[![Lint](https://img.shields.io/badge/Lint-MEML%20%2B%20ShellCheck-blue)](https://github.com/caboose-mcp/waldo/actions)
[![Test](https://img.shields.io/badge/Test-E2E%20%2B%20Smoke-green)](https://github.com/caboose-mcp/waldo/actions)
[![Coverage](https://img.shields.io/badge/Coverage-Docs%20100%25-brightgreen)](./MEML.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Persistent persona system for AI tools.** Define your voice, tone, and style once — then apply it everywhere.

Works with **Claude Code**, **Cursor**, **ChatGPT**, **Gemini**, and **Codeium**.

## What It Does

- **Create personas** from your Slack messages, or manually define them
- **Switch voices** with `/waldo use agent/your-voice`
- **Learn over time** — Claude observes your patterns and suggests tone updates
- **Sync across machines** — S3 or local sync
- **Agent agnostic** — same system works in Claude Code, Cursor, and other AI tools

## Quick Start

**Try it first:** Visit the [interactive playground](https://caboose-mcp.github.io/waldo/ui/playground.html) — adjust tone sliders, see ChatGPT/Gemini prompts update in real-time, no installation needed.

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/setup-waldo.sh | bash

# Create a persona from your Slack
/waldo slack-import
# → Paste 10+ Slack messages you wrote
# → Claude analyzes tone → saves as agent/my-voice

# Use it
/waldo use agent/my-voice
# → Every Claude response now reflects your voice

# Learn from conversations
/waldo learn
# → Suggests tone updates based on this session
# → Apply them or skip
```

## Persona Format: MEML

Personas are stored as `.meml` files (MEML = TOML-inspired with emoji).

```meml
[🪪 meta]
name        = "my-voice"
description = "Casual, direct, no corporate speak"
version     = "0.1.0"

[🎭 tone]
formality   = 0.2   # 0 = very casual, 1 = very formal
directness = 0.8   # 0 = roundabout, 1 = blunt
humor       = 0.6   # 0 = dry, 1 = frequent wit
hedging     = 0.1   # 0 = confident, 1 = heavily qualified
warmth      = 0.5   # 0 = cold, 1 = enthusiastic

[📢 verbosity]
response_length   = "concise"       # concise | adaptive | verbose
reading_level     = "casual"        # casual | professional | technical
format_preference = "prose"         # prose | bullets | adaptive

[🗣️ voice]
avoid_words    = ["certainly", "leverage", "utilize"]
prefer_words   = ["use", "build", "fix"]
custom_phrases = ["yea", "here's the deal"]
```

See [example.meml](./example.meml) for a full example.

## Commands

| Command | What It Does |
|---------|-------------|
| `/waldo list` | Show all personas, mark active |
| `/waldo use <name>` | Switch persona |
| `/waldo new <name>` | Create manually |
| `/waldo slack-import` | Generate from Slack messages |
| `/waldo mood <desc>` | Temp tone overlay (happier, pissed, professional, etc.) |
| `/waldo mood save` | Keep mood forever |
| `/waldo mood reset` | Clear mood |
| `/waldo learn` | Analyze conversation, suggest updates |
| `/waldo code-scan <path>` | Detect code style conventions |
| `/waldo code-style` | View code conventions |
| `/waldo export <name>` | Share persona as JSON |
| `/waldo import` | Paste JSON from someone else |

## Agent Support

| Tool | Support | Notes |
|------|---------|-------|
| **Claude Code** | ✅ Full | Native hook integration |
| **Cursor** | ✅ Full | Workspace rules sync |
| **ChatGPT** | ✅ Manual | Copy hooks, use `/waldo` commands |
| **Gemini** | ✅ Manual | Same as ChatGPT |
| **Codeium** | ✅ Manual | Same as ChatGPT |

All tools use the same `~/.config/waldo/` config directory.

## Architecture

**Core:** `~/.config/waldo/personas/` (XDG standard, agent-agnostic)
- `agent/` — response personas
- `code/` — code style profiles
- `.active` — current persona
- `.deltas` — learning history

**Adapters:**
- **Claude Code** — `UserPromptSubmit` hook in `~/.claude/settings.json`
- **Cursor** — workspace rules update on session start
- **CLI** — `/waldo inject` prints context to stdout

**Format:** MEML config with emoji annotations for readability and tooling hints.

## Setup

### Requirements

- AWS CLI (optional, for S3 sync)
- `meml` CLI (`go install github.com/caboose-mcp/meml/cmd/meml@latest`)
- `jq` (JSON query tool)

### Installation

**Standard:**
```bash
# One command (handles everything)
bash setup-waldo.sh

# Or manually
mkdir -p ~/.config/waldo/personas/{agent,code}
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/example.meml > ~/.config/waldo/personas/agent/default.meml
echo "agent/default" > ~/.config/waldo/.active
```

**Hardened (Sandbox Runtime — macOS/Linux):**

Install with network, filesystem, and capability restrictions using [Anthropic Sandbox Runtime](https://github.com/anthropics/sandbox-runtime):

```bash
# Install srt
npm install -g @anthropic-ai/sandbox-runtime

# Download setup script and sandbox config
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/setup-waldo.sh -o /tmp/setup-waldo.sh
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/waldo.srt-settings.json -o /tmp/waldo.srt-settings.json

# Run inside sandbox
srt --settings /tmp/waldo.srt-settings.json bash /tmp/setup-waldo.sh

# Or install from local copy
srt --settings ./waldo.srt-settings.json bash ./setup-waldo.sh
```

The sandbox config (`waldo.srt-settings.json`) enforces:
- **Network**: GitHub, AWS S3 APIs only (no localhost)
- **Filesystem**: Read-only on secrets (~/.ssh, ~/.gpg); write allowed to ~/.claude and ~/.config/waldo
- **Execution**: Only bash, aws, jq, git allowed; no sudo/rm
- **Audit**: All operations logged to ~/.waldo/audit.log

### S3 Sync (Optional)

Enable cross-machine sync during setup:

```bash
bash setup-waldo.sh
# → "Setup S3 sync? (y/n)" → y
# → Pick existing bucket or create new
```

Now personas auto-sync between machines on `SessionStart` hook.

## Workflow Example

**Machine A (your laptop):**

```bash
# Create persona from your Slack
/waldo slack-import
# → Save as: chris-marasco

# Use it
/waldo use agent/chris-marasco

# Have a conversation
(chat with Claude...)

# Learn from it
/waldo learn
# → humor: 0.65 → 0.75 (0.8 confidence)
# → directness: 0.85 → 0.9 (0.9 confidence)
# → Apply? (y/n) → yes
# → Auto-pushes to S3
```

**Machine B (your Linux box, 1 hour later):**

```bash
# SessionStart hook auto-pulls from S3

# Your persona is active with new tone updates
/waldo list
# → agent/chris-marasco [active]

# Continue chatting with your exact voice
(chat with Claude...)
```

## Troubleshooting

**Persona not injecting?**

```bash
echo '{}' | bash ~/.claude/hooks/waldo/inject-persona.sh
# Should output JSON with your persona context
```

**MEML validation error?**

```bash
meml validate ~/.config/waldo/personas/agent/my-voice.meml
# Shows line number and error
```

**S3 not syncing?**

```bash
aws s3 ls s3://my-personas/
# If fails, check: aws configure
```

**Lost a persona?**

```bash
ls ~/.config/waldo/personas/agent/*.backup.*
# Restore: cp persona.json.backup.TIMESTAMP persona.json
```

## Development

### Building

```bash
# Test setup script
bash setup-waldo.sh

# Validate all MEML files
meml validate example.meml
find ~/.config/waldo/personas -name "*.meml" -exec meml validate {} \;

# Run linter
shellcheck setup-waldo.sh .claude/hooks/waldo/*.sh
```

### Contributing

This is an experimental system. Feedback and PRs welcome.

- Add new tone dimensions to the schema
- Build new adapters (Codeium, Vim, etc.)
- Design new mood overlays
- Improve MEML emoji conventions

## License

MIT

---

**Full docs:**
- [Setup Guide](./WALDO-SETUP.md)
- [Skill Reference](./waldo-SKILL-v5.md)
- [MEML Format Guide](./MEML.md) — How to write personas
- [Quick Start](./QUICK-START.md)
- [DEMO](./DEMO.md)

**Related:**
- [meml](https://github.com/caboose-mcp/meml) — config language
- [caboose-mcp](https://github.com/caboose-mcp) — AI tools org

---

> not because of that one (but it is funny though) — he sort of disappears, iykyk 👀 still love and respect him tho
