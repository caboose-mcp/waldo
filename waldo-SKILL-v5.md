---
name: waldo
description: Manage Claude response personas — agent (tone, voice) and code (style, conventions). Subcommands: use, list, edit, export, import, slack-import, mood, learn, sync (agent); code-scan, code-style (code). Use when switching personas, tweaking voice/tone, analyzing Slack, applying mood overlays, learning from session, or syncing to/from S3 (agent), or scanning/viewing code style conventions (code).
user_invocable: true
---

# waldo

Manage response personas that shape Claude's tone, verbosity, and voice style.

## Persona file location

All personal personas live at `~/.claude/personas/agent/<name>.json` (and optionally `<name>.meml` alongside it). The active persona name is stored in `~/.claude/personas/.active` as plain text. A `UserPromptSubmit` hook reads `.active` on every prompt and injects the persona as context.

**Important:** Personal personas are always written to `~/.claude/personas/agent/` — local machine storage — and may be synced to S3 for cross-machine access when that option is configured. They are never committed to a project repo. MEML files in a project repo (e.g. `ask-a-dev/personas/bot.meml`) are service-level personas baked into a specific application, not personal Claude personas.

## Subcommands

### `/waldo use <name>`

Switch the active persona.

1. Read `~/.claude/personas/.active` to see the current persona.
2. Verify `~/.claude/personas/<name>.json` exists. If not, list available personas and ask the user to choose one or create a new one.
3. Write the name (and nothing else) to `~/.claude/personas/.active`:
   ```bash
   printf '%s' "<name>" > ~/.claude/personas/.active
   ```
4. Confirm: "Persona switched to <name>. It will apply starting from your next message."

### `/waldo list`

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

### `/waldo new <name>`

Create a new persona interactively.

1. Ask the user for each section if they don't provide it upfront:
   - Description (free text)
   - Tone: formality, directness, humor, hedging, warmth — each as 0.0–1.0. Offer named presets: "casual" (0.2), "professional" (0.6), "high" (0.8).
   - Verbosity: response_length (concise/adaptive/verbose), reading_level (casual/professional/technical), format_preference (prose/bullets/adaptive)
   - Voice: words to avoid, words to prefer, custom phrases
   - VTT: whether to enable caption mode
2. Build the JSON structure following the schema. Optionally, offer to create a MEML version for human-readable authoring.
3. Write the JSON to `~/.claude/personas/agent/<name>.json` and, if created, the MEML file to `~/.claude/personas/agent/<name>.meml`. Never write to the project repo.
4. Ask if the user wants to activate it now.

### `/waldo edit <name>`

Edit an existing persona.

1. Read the current persona file:
   ```bash
   cat ~/.claude/personas/<name>.json
   ```
2. Ask the user which fields they want to change.
3. Apply changes with the Edit tool (surgical field updates, not full rewrites).
4. Confirm what changed.

### `/waldo export <name>`

Export a persona to a shareable JSON snippet.

1. Read `~/.claude/personas/<name>.json`.
2. Print the full JSON contents in a code block so the user can copy it.
3. Note: "You can share this with others. They can import it with `/waldo import`."

### `/waldo import`

Import a persona from JSON the user pastes.

1. Ask the user to paste the persona JSON.
2. Validate required fields: `meta.name`, `meta.version`, `tone`, `verbosity`, `voice`.
3. Sanitize the name: only letters, numbers, underscores, hyphens. Reject path traversal attempts.
4. If a persona with that name already exists, ask whether to overwrite or rename.
5. Write to `~/.claude/personas/<name>.json`.
6. Confirm and ask if they want to activate it.

### `/waldo slack-import`

Generate a persona from Slack message samples.

See the "Slack Import Flow" section below for full instructions.

### `/waldo mood <description>`

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
4. Confirm: "Mood overlay applied: <description>. It's session-only — run `/waldo mood save` to make it permanent, or `/waldo mood reset` to clear it."

### `/waldo mood reset`

Clear the active mood overlay.

1. Delete `~/.claude/personas/.mood` if it exists.
2. Confirm: "Mood overlay cleared. Back to base persona."

### `/waldo mood save`

Bake the active mood overlay permanently into the persona JSON.

1. Read `~/.claude/personas/.mood` to get the overrides.
2. Load the active persona from `~/.claude/personas/.active`.
3. Merge mood overrides into the persona's tone and verbosity fields (update, don't replace).
4. Write updated persona back to its JSON file.
5. Delete `~/.claude/personas/.mood`.
6. Confirm: "Mood saved to <persona-name>. It's now permanent (until you edit it again)."

### `/waldo learn`

Analyze this session's patterns and record incremental deltas.

1. **Review session context**: Look back at the user's prompts and your responses. Identify:
   - Average message length (terse vs. long-form)
   - Use of slang, technical terms, humor, formality
   - Directness level (leading with conclusions vs. hedging)
   - Emoji or punctuation patterns
   - Topics and depth
2. **Compare against active persona**: Load the active persona JSON and calculate deltas.
3. **Record deltas** (don't update yet) to `~/.claude/personas/agent/.deltas` as JSONL:
   ```json
   {"timestamp": 1711428000, "type": "tone", "field": "humor", "old_value": 0.65, "new_value": 0.75, "reason": "more playful/sarcastic this session", "confidence": 0.8}
   {"timestamp": 1711428000, "type": "tone", "field": "directness", "old_value": 0.85, "new_value": 0.9, "reason": "consistently leading with conclusions", "confidence": 0.9}
   ```
4. Show suggested updates with confidence scores:
   ```
   Recorded deltas for waldo:
   - humor: 0.65 → 0.75  (0.8 confidence: more playful/sarcastic this session)
   - directness: 0.85 → 0.9  (0.9 confidence: consistently leading with conclusions)
   - response_length: concise → adaptive  (0.7 confidence: you asked several detailed questions)
   ```
5. Ask: "Review deltas? (y/n) to accumulate and merge into persona. Older deltas decay over 30 days."
6. On yes, run `/waldo learn --accumulate` to merge deltas into the base persona with decay weighting.

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
    "sign_off":       "string or null",
    "attribution":    "boolean — include AI agent attribution (default: false)"
  },
  "workflow": {
    "pre_pr": {
      "required": "boolean — whether the checklist is enforced before creating/updating a PR",
      "steps":    "array of strings — ordered steps to run (e.g. owasp_top10_review, lint, test, build)",
      "note":     "string or null — optional context shown alongside the checklist"
    },
    "git": {
      "branch_from":          "string — base branch to always cut new branches from (e.g. main)",
      "sync_base_before_branch": "boolean — pull/update base branch before cutting a new branch",
      "no_direct_push":       "array of strings — branches that must never be pushed to directly (e.g. [main])",
      "commit_convention":    "conventional | none — enforced commit message format",
      "ai_attribution":       "boolean — include Co-Authored-By AI footer in commits (default: false)",
      "pre_push_doc_check":   "boolean — review in-code comments and README for needed updates before pushing"
    }
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
echo '{"session_id":"test","prompt":"hello"}' | bash ~/.claude/hooks/waldo/inject-persona.sh
```

Expected output shape:
```json
{"continue": true, "hookSpecificOutput": {"hookEventName": "UserPromptSubmit", "additionalContext": "..."}}
```

---

## Code Domain (Beta)

Manage and evolve coding style profiles alongside agent personas.

### `/waldo code-scan <repo-path>`

Auto-scan a repository for coding conventions.

1. Run the quick code scanner: `bash ~/.claude/hooks/waldo/scan-code-style.sh <repo-path>`
2. The scanner:
   - Finds 10–20 code files (TS, JS, Python, Go, Rust, etc.)
   - Respects `.gitignore` and skips node_modules/dist/build
   - Extracts: indentation (spaces/tabs), naming conventions (camelCase/snake_case), line length, imports style, comment patterns, error handling, type hints
3. Output saved to `~/.claude/personas/code/coding-style.json`
4. Confirm: "Code style scanned and saved. Review with `/waldo code-style`"

### `/waldo code-style`

View or edit the current code style profile.

1. Read `~/.claude/personas/code/coding-style.json`
2. Pretty-print the profile — indentation, naming, line length, imports, comments, functions, error handling, types
3. Ask: "Want to edit any of these? (run `/waldo edit code-style` to modify)"

### `/waldo code-learn`

Analyze code in this session and record incremental deltas.

Same process as agent learn: capture deltas to `~/.claude/personas/code/.deltas`, record confidence scores, merge with decay.

### `/waldo learn --accumulate` or `/waldo code --accumulate`

Merge all recorded deltas into the active persona.

1. Run the accumulate script: `bash ~/.claude/hooks/waldo/accumulate-deltas.sh agent waldo 30`
2. Applies weighted averaging to tone/verbosity fields:
   - Recent deltas (< 30 days) have higher weight
   - Older deltas decay to 0 weight after 30 days
   - Confidence scores influence the merge strength
3. Backs up the original persona before merging
4. Clears the deltas file after successful merge
5. Confirm: "Persona updated. X deltas merged with decay weighting."
6. If S3 sync is configured (`WALDO_S3_BUCKET` is set), push the updated persona:
   ```
   bash ~/.claude/hooks/waldo/s3-sync.sh push
   ```
   Confirm: "Pushed to S3." (or log the error and continue if sync fails)

---

### `/waldo sync`

Manually push or pull personas to/from S3.

**Usage:**
- `/waldo sync push` — upload local personas to S3
- `/waldo sync pull` — download personas from S3
- `/waldo sync status` — show sync config (bucket, profile, last sync log tail)

**Steps for each subcommand:**

`/waldo sync push`:
1. Check `WALDO_S3_BUCKET` is set; if not, reply: "S3 sync not configured. Run `setup-waldo.sh` or add `WALDO_S3_BUCKET` to `~/.claude/settings.json` env."
2. Run: `bash ~/.claude/hooks/waldo/s3-sync.sh push`
3. Confirm: "Pushed personas to s3://$WALDO_S3_BUCKET/personas/" or surface the error from the log.

`/waldo sync pull`:
1. Check `WALDO_S3_BUCKET` is set; if not, show same setup message.
2. Run: `bash ~/.claude/hooks/waldo/s3-sync.sh pull`
3. Confirm: "Pulled personas from s3://$WALDO_S3_BUCKET/personas/" or surface the error.

`/waldo sync status`:
1. Read `WALDO_S3_BUCKET`, `AWS_PROFILE`, `AWS_REGION` from env.
2. If not configured, say so and show setup instructions.
3. Show last 10 lines of `~/.waldo/s3-sync.log` (if it exists).
4. Example output:
   ```
   S3 sync: configured
     Bucket:  my-personas
     Profile: default
     Region:  us-east-1

   Last sync log:
     2026-04-05T14:32:01Z [waldo/s3-sync] pull OK
   ```

---

## Cross-Machine Sync (S3 Backend)

Personas automatically sync across machines via S3. No manual export/import needed.

### Setup

The easiest path is `setup-waldo.sh` — it handles bucket selection, writes env vars, and wires the `SessionStart` hook automatically.

**Manual setup** (if skipping the installer):

1. **Create an S3 bucket** (or reuse existing one):
   ```bash
   aws s3api create-bucket --bucket my-personas --region us-east-1
   ```

2. **Add env vars and hook** to `~/.claude/settings.json`:
   ```json
   {
     "env": {
       "AWS_PROFILE": "default",
       "AWS_REGION": "us-east-1",
       "WALDO_S3_BUCKET": "my-personas"
     },
     "hooks": {
       "SessionStart": ["Bash(timeout 15 bash ~/.claude/hooks/waldo/s3-sync.sh pull &)"]
     }
   }
   ```
   `WALDO_S3_BUCKET` is read by `s3-sync.sh` so the hook needs no arguments.
   The `&` makes it fire-and-forget — it never delays session start.

3. **Guardrails** — before syncing, the hook validates:
   - `aws` CLI is installed
   - `WALDO_S3_BUCKET` is set
   - AWS credentials are valid for the configured profile
   - The local personas directory exists

   Any check failure logs to `~/.waldo/s3-sync.log` and exits cleanly — sync errors never block a session.

### Manual Sync

Push current personas to S3:
```bash
bash ~/.claude/hooks/waldo/s3-sync.sh push
# or with explicit overrides:
bash ~/.claude/hooks/waldo/s3-sync.sh push my-personas my-profile
```

Pull latest from S3:
```bash
bash ~/.claude/hooks/waldo/s3-sync.sh pull
# or with explicit overrides:
bash ~/.claude/hooks/waldo/s3-sync.sh pull my-personas my-profile
```

**Note:** `.deltas` and `.cache` are excluded from sync (local-only learning history). Only base personas and mood overlays are synced.

### Workflow

1. Machine A: Edit persona, run `/waldo learn --accumulate` → call `s3-sync.sh push` to propagate
2. Machine B: Start new session → `SessionStart` hook pulls from S3 → you have the latest persona
3. Cross-machine consistency — pull is automatic; push is explicit after accumulation

### Error Handling

- **Missing AWS credentials**: Logs warning to `~/.waldo/s3-sync.log`, continues without syncing
- **`WALDO_S3_BUCKET` not set**: Logs and skips — no sync attempted
- **Bucket unreachable or empty**: Pull is skipped gracefully; local personas are untouched
- **Permission denied**: Check IAM policy grants `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` on the bucket
- **All errors**: Hook always exits 0 — sync issues never block Claude Code

---

## Fingerprint Caching (Token Optimization)

Persona context is cached with SHA256 fingerprints to reduce token overhead on repeat sessions.

**How it works:**
- First session: Full persona JSON injected (~200 tokens)
- Hook generates fingerprint: `nk:waldo:1.0:abc123def456`
- Subsequent sessions: Only fingerprint + short human summary (~10 tokens)
- Model learns shorthand from first exposure

**Benefit:** ~95% reduction in persona context tokens after initial session.

---

## Delta Accumulation & Learning Decay

When you run `/waldo learn --accumulate`, persona updates are merged with time-decay weighting:

- Recent deltas (< 30 days) have 100% influence
- Older deltas decay linearly to 0% at 30 days
- Each delta has a confidence score (0.0–1.0) that weights its influence
- Originals are backed up before merge (`persona.json.backup.TIMESTAMP`)

**Example:**
```
Recorded deltas for waldo:
- humor: 0.65 → 0.75  (0.8 confidence: more playful this session)
- directness: 0.85 → 0.9  (0.9 confidence: leading with conclusions)

Apply these? (y/n)
→ yes
Persona updated. 2 deltas merged with decay weighting.
Automatically pushing to S3...
```

---

## Session Counter & Learning Nudge

A counter tracks messages per session. At 50 messages (configurable), you'll see a nudge:

> "You've sent 50 messages this session — want me to analyze your patterns? Run `/waldo learn`"

This is optional — learning only happens on demand. Counter resets at SessionStart.

---

## Status Line Integration (Agent Agnostic)

The `status-line.sh` hook displays your active persona in editor status bars — works across all tools.

**Supported:**
- **Claude Code** — via `~/.claude/settings.json` hook
- **Cursor** — via workspace status bar
- **VS Code** — via status bar extension API
- **Terminal** — via custom PS1 prompt

**Output format:** `waldo: chris-marasco [mood]`

Shows:
- Current active persona name (short form, no `agent/` prefix)
- `[mood]` badge if session mood overlay is active
- Updates in real-time as you switch personas

**Setup in Claude Code:**

Add to `~/.claude/settings.json`:
```json
{
  "hooks": {
    "SessionStart": "Bash(bash ~/.claude/hooks/waldo/status-line.sh > /tmp/waldo-status.txt)"
  }
}
```

Then display in your prompt or tool output.

---

## Emoji Configuration (Optional)

Personas can include sparse, configurable emoji use:

```meml
[😊 emoji]
enabled           = ✅           # Enable emoji in responses
frequency         = "sparse"     # sparse | moderate | frequent
contexts          = ["emphasis", "mood", "transitions"]
```

- **sparse** — 0–2 emoji per response, only for emphasis
- **moderate** — 2–5 emoji, for mood + transitions
- **frequent** — 5+ emoji, liberal use

Hook respects the config and injects hint into persona context. Claude decides emoji use based on context and persona intent.

