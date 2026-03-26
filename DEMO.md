# waldo Demo — Complete End-to-End

## Installation

**One command (any machine, any AI tool):**

```bash
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/demo/dual-domain/setup-waldo.sh | bash
```

What happens:
```
✓ AWS authenticated as your-user (Account: 123456789012)
✓ Created: ~/.claude/personas/{agent,code}
✓ Default persona created
✓ Active persona: agent/default
✓ Hook scripts: ~/.claude/hooks/waldo
✓ Learning deltas initialized

Setup Complete!
```

**Time:** ~30 seconds

---

## Step 1: Create a Persona from Your Slack

**Command:**
```bash
/waldo slack-import
```

**What you do:**
Paste 10+ Slack messages you wrote (copy from Slack, paste into prompt)

**Claude analyzes:**
```
Formality: 0.25 — uses lowercase, frequent contractions, no formal salutations
Directness: 0.80 — short messages, leads with conclusion, few qualifiers
Humor: 0.55 — occasional dry jokes, some self-deprecation
Hedging: 0.15 — confident assertions, "maybe" appears only twice
Warmth: 0.60 — occasional emoji, uses "nice" and "sounds good"
```

**You name it:**
```
What should I name this persona?
→ my-voice
```

**Result:**
```
✓ Persona saved to ~/.claude/personas/agent/my-voice.json
✓ Activate it now? (y/n) → y
✓ Active persona: my-voice
```

---

## Step 2: Use Your Persona

**Command:**
```bash
/waldo use agent/my-voice
```

**What happens:**
Every response now uses your exact tone, formality, humor, warmth levels.

**You can check:**
```bash
/waldo list
```

```
agent/default: Neutral baseline persona
agent/my-voice: Slack-derived persona [active]
```

---

## Step 3: Apply Mood Overlay (Optional)

Want to sound different just for this conversation?

**Make it happier:**
```bash
/waldo mood make me sound happier
```

Response becomes: +0.2 humor, +0.2 warmth (session-only)

**Make it professional:**
```bash
/waldo mood more professional
```

Response becomes: +0.3 formality, -0.2 humor

**Get pissed:**
```bash
/waldo mood pissed
```

Response becomes: +0.3 directness, -0.2 warmth

**Clear it:**
```bash
/waldo mood reset
```

Back to base persona.

---

## Step 4: Chat Away

Just talk. Claude uses your persona for every response.

Example conversation:
```
You: hey, can you help me debug this code?

Claude (as my-voice):
yeah breh, what's going on

You: it keeps saying undefined is not a function

Claude (as my-voice):
quick heads up — where are you calling it? probably a scope thing.
post the line
```

Notice: casual tone, direct, no hedging, gets to the point. That's your persona.

---

## Step 5: Learn from Your Style

After talking for 20–50 messages, run:

```bash
/waldo learn
```

Claude analyzes *this session* and suggests updates:

```
Recorded deltas for my-voice:
- humor: 0.55 → 0.68  (0.85 confidence: more playful/sarcastic)
- directness: 0.80 → 0.87  (0.92 confidence: leading with conclusions)
- response_length: adaptive → concise  (0.78 confidence: you asked short questions)
```

**Apply updates?** (y/n)
```
→ y
✓ Persona updated. 3 deltas merged with decay weighting.
(If S3 configured) ✓ Auto-pushed to s3://my-personas/personas/
```

Your persona evolved based on *your actual conversation*.

---

## Step 6: Cross-Machine Sync (Optional S3)

**Machine A (MacBook):**
- Used persona for 30 min
- Ran `/waldo learn --accumulate`
- Auto-pushed to S3

**Machine B (Linux, 1 hour later):**
- SessionStart hook ran
- Auto-pulled from S3
- Your persona arrived with all updates

```bash
/waldo list
```

```
agent/my-voice [active] ← with updates from Machine A!
```

Zero manual steps. It just works.

---

## Step 7: Share a Persona

Made a fire persona? Share it.

**Export:**
```bash
/waldo export agent/my-voice
```

Output (copy this):
```json
{
  "meta": {
    "name": "my-voice",
    "description": "Slack-derived persona",
    "version": "0.1.0",
    "created_at": "2026-03-26T12:00:00Z"
  },
  "tone": {
    "formality": 0.25,
    "directness": 0.87,
    "humor": 0.68,
    "hedging": 0.15,
    "warmth": 0.60
  },
  ...
}
```

**Friend imports:**
```bash
/waldo import
```

Pastes JSON → Done. They now have your voice.

---

## Full Command Reference

| Command | Purpose |
|---------|---------|
| `/waldo list` | Show all personas + active |
| `/waldo use <name>` | Activate a persona |
| `/waldo new <name>` | Create manually |
| `/waldo slack-import` | Generate from Slack messages |
| `/waldo mood <desc>` | Temp tone (happier, pissed, professional) |
| `/waldo mood save` | Keep mood forever |
| `/waldo mood reset` | Clear temp mood |
| `/waldo learn` | Analyze session, suggest updates |
| `/waldo code-scan /path` | Detect code style |
| `/waldo export <name>` | Share persona as JSON |
| `/waldo import` | Import persona from JSON |

---

## Real-World Example: "Chris's Setup"

**Day 1 (Mac):**
```bash
# Setup
curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/demo/dual-domain/setup-waldo.sh | bash

# Create persona
/waldo slack-import
(paste Slack messages...)
→ Save as: chris-marasco

# Use it
/waldo use agent/chris-marasco

# Chat for 30 min
(...)

# Learn
/waldo learn
→ Apply? y
```

**Day 2 (Linux):**
```bash
# Session starts
(SessionStart hook auto-pulls from S3)

# Your persona is here with yesterday's updates
/waldo list
→ agent/chris-marasco [active]

# Continue as if you never left
(chat with your exact tone...)
```

**Day 3 (Different AI tool — Cursor):**
```bash
# Same setup (persona already synced)
# Same skill commands work
/waldo mood pissed
/waldo learn
```

Same voice everywhere.

---

## How It Works (Under the Hood)

**Persona JSON:**
```json
{
  "tone": {
    "formality": 0.0–1.0,     ← how formal you sound
    "directness": 0.0–1.0,    ← straight to point vs roundabout
    "humor": 0.0–1.0,         ← dry vs witty
    "hedging": 0.0–1.0,       ← confident vs qualified
    "warmth": 0.0–1.0         ← cold vs enthusiastic
  },
  "verbosity": {
    "response_length": "concise | adaptive | verbose",
    "reading_level": "casual | professional | technical",
    "format_preference": "prose | bullets | adaptive"
  },
  "voice": {
    "custom_phrases": ["yea breh", "here's the deal"],
    "avoid_words": ["certainly", "leverage"],
    "prefer_words": ["use", "fix", "run"]
  }
}
```

**Injection (UserPromptSubmit hook):**
```
Claude sees: [PERSONA: my-voice]
Tone: direct, casual, witty
Voice: custom phrases, avoid corporate jargon

Applies this context before responding ↓

Response is now in your style.
```

**Learning (Session deltas):**
```
Session 1: "You were more sarcastic than usual"
Suggestion: humor 0.55 → 0.68

Session 2: "You led with conclusions 92% of the time"
Suggestion: directness 0.80 → 0.87

User runs: /waldo learn --accumulate
Merge with decay weighting ↓

Persona evolves gradually.
```

**Sync (S3):**
```
Machine A updates persona
Auto-pushes to s3://my-personas/personas/

Machine B starts session
SessionStart hook pulls from S3 ↓

Latest persona loaded.
```

---

## Edge Cases Handled

✓ **Mood safety:** Blocks slurs, explicit content, hate speech
✓ **AWS not configured:** Still works locally (S3 is optional)
✓ **Persona lost:** Backups at `~/.claude/personas/agent/my-voice.json.backup.TIMESTAMP`
✓ **S3 bucket missing:** Logs warning, doesn't block
✓ **First time:** Full persona context (~200 tokens)
✓ **Repeat sessions:** Fingerprint shorthand (~10 tokens, 95% reduction)

---

## One-Line Summary

Your voice in your code. Everywhere. Always learning.

---

**Get started:** https://github.com/caboose-mcp/waldo/tree/demo/dual-domain
**Setup script:** `curl -fsSL ... | bash`
**Time to first persona:** 5 minutes
