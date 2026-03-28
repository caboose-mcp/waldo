# ask-a-dev TUI Setup — Plan

## Context

User wants a **secure terminal UI** to configure ask-a-dev service. This replaces manual editing of MEML configs and setup steps. The TUI should guide users through:

1. Discord bot creation / token input (secure prompt)
2. n8n credentials (API keys, with validation)
3. GitHub PAT setup
4. Channel ID discovery (query Discord API)
5. Schedule configuration
6. Initial MEML config generation
7. Credential storage (encrypted in ~/.config/waldo/)
8. Preview + confirmation before committing

**Why a TUI:** Following waldo's existing pattern (waldo-tui for persona picker). This is the second waldo TUI tool.

**Security requirements:**
- Never log credentials to stdout or files
- Use secure password prompts (no echo)
- Store credentials in secure location (0600 perms)
- Validate all inputs before use
- Clear sensitive data from memory after use (where feasible)

---

## Implementation Plan

### Part 1: Create New Binary

**New file:** `cmd/waldo-setup-ask-a-dev/main.go`

- Entry point for TUI
- Uses lipgloss (already in go.mod) for terminal rendering
- Uses `golang.org/x/term` for secure password input
- Interactive step-by-step flow

### Part 2: New Internal Package

**New file:** `internal/ask-a-dev/setup.go`

- `NewSetup()` — Initialize configuration state
- `PromptDiscordToken()` — Secure password prompt + validate format
- `PromptClaudeKey()` — Secure prompt + validate format
- `PromptGitHubPAT()` — Secure prompt + validate format (optional test call)
- `QueryDiscordChannels()` — Call Discord API to list channels (requires bot token)
- `PromptChannelSelection()` — Menu to select question channel + admin channel
- `PromptSchedule()` — Input question_time, summary_time, timezone
- `BuildMEMLConfigs()` — Generate setup.meml, feature-flags.meml from prompts
- `SaveSecurely()` — Write encrypted credentials to ~/.config/waldo/ask-a-dev/credentials.json
- `ValidateAndPreview()` — Show what will be written, ask for confirmation

### Part 3: Credential Storage

**File:** `~/.config/waldo/ask-a-dev/credentials.json` (mode 0600)

```json
{
  "discord_bot_token": "xxx",
  "anthropic_api_key": "xxx",
  "github_pat": "xxx",
  "n8n_instance_url": "https://your-n8n-instance.n8n.cloud",
  "n8n_api_key": "xxx",
  "created_at": "2026-03-27T12:00:00Z",
  "updated_at": "2026-03-27T12:00:00Z"
}
```

**Security:**
- File is chmod 0600 (owner read/write only)
- Read with permission check: if perms != 0600, warn user
- Never print credentials to terminal after entry
- Clear from memory after used (defer zero-out)

### Part 4: TUI Flow

```
┌─────────────────────────────────────────────────┐
│  🎭 ask-a-dev Setup                             │
│  Configure your transparent Q&A service         │
└─────────────────────────────────────────────────┘

[1/6] Discord Bot Token
  ► Enter your Discord bot token (hidden input)
  ✓ Token validated (format correct)

[2/6] Anthropic Claude API Key
  ► Enter your Claude API key (hidden input)
  ✓ Key validated (format correct, optional: test call)

[3/6] GitHub Fine-Grained PAT
  ► Enter your GitHub PAT (hidden input)
  ✓ PAT validated (repo-scoped)

[4/6] Select Discord Channels
  Fetching channels from your server...

  Question Channel:
    ▶ #general
    • #random
    • #dev
    [ESC to skip, use arrows to select]

  Admin Channel:
    ▶ #admin-approvals
    • #mods

  ✓ Channels selected

[5/6] Configure Schedule
  Question Time (UTC, 24-hour): [09:00]
  Summary Time (UTC, 24-hour):  [22:00]
  Timezone (e.g., UTC):         [UTC]
  ✓ Schedule configured

[6/6] Review Configuration
  ────────────────────────────────────────
  Discord Bot Token:     ••••••••••••••••
  Claude API Key:        ••••••••••••••••
  GitHub PAT:            ••••••••••••••••
  Question Channel:      #general
  Admin Channel:         #admin-approvals
  Question Time:         09:00 UTC
  Summary Time:          22:00 UTC
  ────────────────────────────────────────

  Write config to ~/.config/waldo/ask-a-dev/? [Y/n]

  ✓ Configuration saved securely

  Next steps:
    1. Create n8n workflows (see SETUP.md for templates)
    2. Configure n8n credentials with tokens from above
    3. Activate workflows
    4. Run: waldo ask-a-dev test

  Docs: https://github.com/caboose-mcp/waldo/tree/main/ask-a-dev
```

### Part 5: Input Validation

**Discord Bot Token:**
- Format: `^[A-Za-z0-9_-]{60,68}$` (approximately)
- Optional: Make a simple Discord API call (GET `/users/@me`) to validate

**Claude API Key:**
- Format: `^sk-ant-[A-Za-z0-9_-]{48,}$` (starts with `sk-ant-`)
- Optional: Call Claude API test endpoint

**GitHub PAT:**
- Format: `^ghp_[A-Za-z0-9_]{36,}$` (fine-grained) or `^github_pat_[A-Za-z0-9_]{82}$`
- Call GitHub API to verify scopes: `GET /repos/{owner}/{repo}` (test write access)

**Channel IDs:**
- Numeric strings, must exist in server

**Schedule:**
- Time: regex `^\d{2}:\d{2}$` (00:00–23:59)
- Timezone: validate against IANA list or allow any (TZ env compatibility)

---

## Critical Files

**Create:**
- `cmd/waldo-setup-ask-a-dev/main.go`
- `internal/ask-a-dev/setup.go` (or `internal/cli/setup_ask_a_dev.go`)

**Modify:**
- `go.mod` (no new deps — use existing lipgloss + golang.org/x/term)
- `.leftfile.toml` (add `*.go` linting for new files)
- `.github/workflows/ci.yml` (build the new binary)
- `Makefile` or build script to compile `waldo-setup-ask-a-dev`

**Reference:**
- `cmd/waldo-tui/main.go` — existing TUI patterns
- `internal/terminal/detect.go` — terminal capability detection
- `internal/github/fetch.go` — example of secure HTTP with timeout

---

## Verification Checklist

- [ ] TUI renders without crashes
- [ ] Secure prompt (no echo) works for tokens
- [ ] Validation rejects malformed input with clear error
- [ ] Discord channel query succeeds (needs valid token)
- [ ] Menu navigation works (arrows, enter, esc)
- [ ] Configuration file written with 0600 perms
- [ ] Credentials readable by n8n workflows (via GitHub Actions)
- [ ] Credentials NOT printed to terminal
- [ ] Sensitive data cleared from memory (zeroed)
- [ ] Help/docs link shown at end
- [ ] Works on macOS, Linux, Windows (if possible)

---

## Implementation Order

1. **core setup logic** (`internal/ask-a-dev/setup.go`) — validators, prompts, builders
2. **TUI entry point** (`cmd/waldo-setup-ask-a-dev/main.go`) — menu flow, lipgloss rendering
3. **build integration** — add to Makefile, CI
4. **testing** — manual test on macOS + Linux

---

## Open Questions for User

1. Should credentials be encrypted at rest, or just file perms (0600)?
2. Should TUI offer to test each credential (API calls) or just validate format?
3. Should we auto-discover n8n instance URL or ask user?
4. Should we generate n8n workflow templates as part of setup (or keep separate)?

---

## Success Criteria

✓ User runs `waldo-setup-ask-a-dev`
✓ TUI guides through 6 steps
✓ Credentials saved securely
✓ Next steps clearly documented
✓ No credentials leaked to terminal or logs
