# ask-a-dev Setup TUI

Secure terminal UI for configuring the ask-a-dev service.

## Quick Start

```bash
# Build
go build -o bin/waldo-setup-ask-a-dev ./cmd/waldo-setup-ask-a-dev/

# Run
./bin/waldo-setup-ask-a-dev
```

## What It Does

The TUI guides you through 5 steps:

1. **Discord Bot Token** (secure input)
   - Validates format: 60-68 alphanumeric characters
   - Masked input (no echo)

2. **Claude API Key** (secure input)
   - Validates format: must start with `sk-ant-`
   - Masked input

3. **GitHub Fine-Grained PAT** (secure input)
   - Validates format: starts with `ghp_` or `github_pat_`
   - Masked input

4. **Schedule Configuration** (visible input)
   - Question time: `HH:MM` (UTC)
   - Summary time: `HH:MM` (UTC)
   - Timezone: IANA format (e.g., `UTC`, `America/New_York`)

5. **Optional: n8n Configuration** (secure input)
   - n8n instance URL
   - n8n API key

All inputs are validated in real-time. Invalid inputs are rejected with clear error messages.

## Output

Credentials are saved to:

```
~/.config/waldo/ask-a-dev/credentials.json
```

With permissions `0600` (readable only by you).

Example:

```json
{
  "discord_bot_token": "MjI4MzAyNjE1OTA0MDE4Njcz...",
  "anthropic_api_key": "sk-ant-abc123...",
  "github_pat": "ghp_xxxx...",
  "question_time": "09:00",
  "summary_time": "22:00",
  "timezone": "UTC",
  "created_at": "2026-03-27T12:00:00Z",
  "updated_at": "2026-03-27T12:00:00Z"
}
```

## Security

### Secure Prompts

Tokens are entered without echo (masked). The terminal doesn't show what you type.

### File Permissions

The credentials file is written with `0600` permissions:
- Owner can read and write
- Group has no access
- Others have no access

The TUI verifies permissions after writing:

```
✓ Verification: 0600 correct
```

### Memory Clearing

Sensitive strings are cleared from memory after use (best-effort). This prevents them from being accidentally dumped.

### No Logging

Credentials are never printed to stdout or logs. The preview shows masked values:

```
Discord Bot Token:     xxxx•xxx
Claude API Key:        xxxx•xxx
GitHub PAT:            xxxx•xxx
```

## Validation Rules

| Token | Format | Example |
|-------|--------|---------|
| Discord Bot | 60-68 alphanumeric | `MjI4MzAyNjE1OTA0MDE4Njcz.G...` |
| Claude API Key | starts with `sk-ant-` | `sk-ant-abc123...` |
| GitHub PAT | starts with `ghp_` or `github_pat_` | `ghp_xxxx...` |
| Question Time | `HH:MM`, `00:00`–`23:59` | `09:00` |
| Timezone | IANA name | `UTC`, `America/New_York` |

## Next Steps

After setup, follow these steps:

1. Create n8n workflows (see `SETUP.md` for templates)
2. Configure n8n credentials with tokens from above
3. Activate workflows
4. Test the service

## Troubleshooting

### "Invalid Discord Bot Token"

Tokens must be 60-68 alphanumeric characters. Check:
- Copy entire token from Discord Developer Portal
- No leading/trailing spaces
- No line breaks

### "Invalid Claude API Key"

Claude keys must start with `sk-ant-`. Check:
- Copy from https://console.anthropic.com/
- Key starts with `sk-ant-`
- No extra characters

### "Invalid GitHub PAT"

GitHub PATs must start with `ghp_` (fine-grained) or `github_pat_` (new format). Check:
- Create at https://github.com/settings/tokens
- Select fine-grained personal access token
- Scoped to your repo

### "Failed to save: permission denied"

The TUI tried to write to `~/.config/waldo/ask-a-dev/` but couldn't. Check:
- Directory exists and is writable: `chmod 700 ~/.config/waldo/ask-a-dev/`
- Disk space available
- SELinux/AppArmor not blocking write

### "Credentials file has wrong permissions"

The credentials file wasn't written with `0600` permissions. This can happen if umask is set too permissively. Check:
- `ls -la ~/.config/waldo/ask-a-dev/credentials.json`
- Should show: `-rw-------`
- Manually fix: `chmod 600 ~/.config/waldo/ask-a-dev/credentials.json`

## Advanced: Loading Credentials Programmatically

If you're building tooling that uses the credentials:

```go
import "github.com/caboose-mcp/waldo/internal/ask-a-dev"

cfg, err := askadev.LoadSecurely()
if err != nil {
  log.Fatal(err)
}

// cfg.DiscordBotToken, cfg.ClaudeAPIKey, cfg.GitHubPAT, etc.
```

The loader verifies file permissions and returns an error if they're wrong:

```
credentials file has wrong permissions: 0644 (expected 0600)
```

## Design

### Why Secure Prompts?

Tokens in plaintext on the terminal can be:
- Seen over your shoulder
- Recorded by terminal history
- Captured by screen recordings
- Visible in tmux session listings

Masked input prevents this.

### Why 0600 Permissions?

This is the standard Unix permission for sensitive files. It prevents:
- Other users on the system from reading your tokens
- Accidental over-sharing (copy-paste mistakes)
- Backup systems from including credentials

### Why Masked in Preview?

Even in your own terminal, you might:
- Share a screenshot
- Record a demo
- Accidentally push credentials

Masked tokens let you verify the config without exposing secrets.

## Code

Core functions in `internal/ask-a-dev/setup.go`:

- `PromptSecure()` — Read without echo using `golang.org/x/term`
- `ValidateDiscordToken()` — Format validation with regex
- `SaveSecurely()` — Write with 0600 perms + verify
- `LoadSecurely()` — Read with permission check
- `PrintPreview()` — Display with masked values

TUI logic in `cmd/waldo-setup-ask-a-dev/main.go`:

- 5-step flow with lipgloss styling
- Error handling for each step
- Confirmation before saving
- Next steps documentation

---

**🎭 Secure setup complete!**
