# Final Two Items: Accessibility + Cross-Tool Integration
## Implementation Plan with Hanko Stubs

**Status:** Planning phase
**Priority:** High (ship before v0.2)
**Effort estimate:** 3-4 weeks
**Team:** 1 engineer (parallelizable with external community)

---

## I. Item 1: Accessibility Pass

### Goal
Make waldo fully usable without emoji; support screen readers, CI/CD environments, terminal variations.

### Current Issues
1. **Emoji hard-coded** in status bar, persona sections, TUI
   - `./bin/waldo-status` outputs: `🚨 main ● ⬆ 1`
   - Screen readers cannot interpret emoji
   - Some CI systems (Buildkite, GitHub Actions runners) strip emoji
   - Some terminals (vt100, SSH) misalign emoji rendering

2. **No fallback syntax** in MEML
   - Emoji annotations are required: `🔑 token = "x"`
   - Tools cannot query by ASCII alternative: `@secret token = "x"`

3. **TUI uses raw terminal mode**
   - Works perfectly on modern terminals (iTerm2, Hyper, Alacritty)
   - Fails silently on dumb terminals or SCREEN sessions

### A. Implement ASCII Fallback Mode

#### A.1 Go CLI Flags
```go
// internal/cli/flags.go
type CLIFlags struct {
    NoEmoji     bool   // --no-emoji: suppress all emoji
    AsciiMode   bool   // --ascii: use ASCII symbols only
    NoColor     bool   // --no-color: strip ANSI color
    Verbose     bool   // -v: detailed output
}

// Update all commands to respect flags
```

#### A.2 waldo-status ASCII Mode
```bash
# Current (emoji + color)
$ ./bin/waldo-status
🚨 main ● ⬆ 1 | ⏱ | 🎭 agent/default

# With --no-emoji
$ ./bin/waldo-status --no-emoji
[PROD] main [DIRTY] [AHEAD 1] | [IDLE] | [PERSONA] agent/default

# With --ascii (minimal)
$ ./bin/waldo-status --ascii
*PROD* main * +1 | * | # agent/default
```

#### A.3 waldo-tui ASCII Menu
```go
// internal/tui/menu.go - new method
func (m *Menu) RenderASCII() {
    // Use ASCII box drawing instead of ANSI colors
    // └─ for selections instead of > marker
    // ─── for dividers instead of colored lines
    // [123/456] instead of emoji page indicator
}

// internal/tui/renderer.go - add flag-aware rendering
type Renderer struct {
    UseEmoji bool
    UseColor bool
}

func (r *Renderer) SelectMarker() string {
    if r.UseEmoji {
        return "> "
    }
    return "→ "
}
```

#### A.4 MEML Annotation Fallbacks
```go
// internal/meml_ext/annotations.go - new module
// Mapping emoji ↔ ASCII aliases

var EmojiToASCII = map[string]string{
    "🔑": "@secret",
    "🌍": "@url",
    "📁": "@path",
    "📋": "@list",
    "⚠️": "@deprecated",
}

var ASCIIToEmoji = map[string]string{
    "@secret":     "🔑",
    "@url":        "🌍",
    "@path":       "📁",
    "@list":       "📋",
    "@deprecated": "⚠️",
}

// Parser accepts both: "🔑 token" and "@secret token"
// Persists in original form; display can toggle
```

#### A.5 Update MEML Parser
```go
// github.com/caboose-mcp/meml/parser/parser.go
// Add new parsing mode: emoji OR ASCII

func Parse(src []byte, opts ...ParseOption) (*Document, error) {
    p := newParser(src)
    for _, opt := range opts {
        opt.apply(p)
    }
    // If UseASCIIAnnotations is true, accept @secret as well as 🔑
    return p.parse()
}

type ParseOption interface {
    apply(*parser)
}

// Used by: meml validate --accept-ascii persona.meml
```

#### A.6 Git Status Accessibility
```go
// internal/git/status.go - update format
func (s *BranchStatus) ASCIIFormat() string {
    var parts []string

    if s.IsDetached {
        parts = append(parts, "[DETACHED]")
    } else if s.IsProd {
        parts = append(parts, fmt.Sprintf("[PROD] %s", s.Branch))
    } else {
        parts = append(parts, s.Branch)
    }

    if s.Dirty {
        parts = append(parts, "[DIRTY]")
    }

    if s.AheadBy > 0 {
        parts = append(parts, fmt.Sprintf("[+%d]", s.AheadBy))
    }

    if s.BehindBy > 0 {
        parts = append(parts, fmt.Sprintf("[-%d]", s.BehindBy))
    }

    return strings.Join(parts, " ")
}

// Auto-detect terminal type and choose output
func (s *BranchStatus) Format(ctx context.Context) string {
    if ctx.NoEmoji {
        return s.ASCIIFormat()
    }
    return s.StatuslineFormat()
}
```

#### A.7 Unit Tests
```go
// internal/cli/accessibility_test.go
func TestASCIIFallbacks(t *testing.T) {
    tests := []struct {
        name     string
        emoji    string
        wantASCII string
    }{
        {"detached", "🔴 DETACHED", "[DETACHED]"},
        {"prod", "🚨 main", "[PROD] main"},
        {"dirty", "●", "[DIRTY]"},
        {"ahead", "⬆ 5", "[+5]"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ToASCII(tt.emoji)
            if got != tt.wantASCII {
                t.Errorf("ToASCII(%q) = %q, want %q", tt.emoji, got, tt.wantASCII)
            }
        })
    }
}
```

### B. Detection + Auto-Fallback

```go
// internal/terminal/detect.go - new module
package terminal

import (
    "os"
    "strings"
)

type Capability struct {
    Emoji     bool
    Color     bool
    RawMode   bool
    UTF8      bool
}

func Detect() *Capability {
    term := os.Getenv("TERM")
    lang := os.Getenv("LANG")

    cap := &Capability{
        Emoji:   true,  // default
        Color:   true,
        RawMode: true,
        UTF8:    strings.Contains(lang, "UTF-8"),
    }

    // Downgrade for known limitations
    switch term {
    case "dumb", "vt100", "vt220":
        cap.Emoji = false
        cap.Color = false
        cap.RawMode = false
    case "screen", "screen-256color":
        if os.Getenv("TMUX") != "" {
            cap.RawMode = false  // tmux + raw mode = issues
        }
    }

    // CI/Docker downgrade
    if _, inDocker := os.LookupEnv("DOCKER_CONTAINER"); inDocker {
        cap.Emoji = false
    }

    if _, inGitHub := os.LookupEnv("GITHUB_ACTIONS"); inGitHub {
        cap.Emoji = false
    }

    // Command-line override
    if _, noEmoji := os.LookupEnv("WALDO_NO_EMOJI"); noEmoji {
        cap.Emoji = false
    }

    return cap
}
```

### C. Documentation

- [ ] Add `docs/ACCESSIBILITY.md` with:
  - Supported terminal types
  - Environment variables (`WALDO_NO_EMOJI`, `NO_COLOR`)
  - Flags: `--no-emoji`, `--ascii`, `--no-color`
  - MEML annotation fallbacks: `@secret` vs `🔑`
  - Screen reader compatibility notes

---

## II. Item 2: Cross-Tool Integration

### Goal
Make waldo usable in Cursor, ChatGPT, Gemini with minimal setup.

### Current Support
- ✅ **Claude Code** — Native hooks (best integration)
- 🔲 **Cursor** — Workspace rules (needs connector)
- 🔲 **ChatGPT** — Manual copy-paste (needs UI helper)
- 🔲 **Gemini** — Manual copy-paste (needs UI helper)

### A. Cursor Integration

#### A.1 Workspace Rules Sync
```go
// cmd/waldo-cursor-sync/main.go - NEW BINARY
// Watches ~/.config/waldo/personas/.active
// Writes active persona to .cursorrules in project root

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/caboose-mcp/waldo/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to load waldo config: %v\n", err)
        os.Exit(1)
    }

    // Read active persona
    personaFile := filepath.Join(cfg.Root, cfg.ActivePersona+".meml")
    personaContent, err := os.ReadFile(personaFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Persona not found: %s\n", personaFile)
        os.Exit(1)
    }

    // Convert MEML to Claude Code rules format
    rules := memlToCursorRules(personaContent)

    // Write to .cursorrules in current project
    cursorRulesFile := ".cursorrules"
    if err := os.WriteFile(cursorRulesFile, []byte(rules), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to write .cursorrules: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✓ Synced persona to %s\n", cursorRulesFile)
}

func memlToCursorRules(memlContent []byte) string {
    // Parse MEML persona
    // Convert to Cursor workspace rules format
    // Return markdown-like format

    // Example output:
    // # waldo: agent/default
    // ## Tone
    // - Formality: 0.5 (moderate)
    // - Directness: 0.8 (direct)
    // ...

    return ""  // stub
}
```

#### A.2 Hook for Cursor
```bash
# .claude/hooks/waldo/cursor-sync.sh
# Called on SessionStart or manually via: waldo-cursor-sync

#!/bin/bash
set -e

PERSONA_DIR="${HOME}/.config/waldo/personas"
ACTIVE_FILE="${PERSONA_DIR}/.active"

if [ ! -f "$ACTIVE_FILE" ]; then
    exit 0  # No active persona, skip
fi

PERSONA_NAME=$(cat "$ACTIVE_FILE")
PERSONA_FILE="${PERSONA_DIR}/${PERSONA_NAME}.meml"

if [ ! -f "$PERSONA_FILE" ]; then
    echo "❌ Persona not found: $PERSONA_FILE" >&2
    exit 1
fi

# Convert MEML to .cursorrules format
waldo-cursor-sync "$PERSONA_FILE" > .cursorrules

echo "✓ Synced waldo persona to .cursorrules"
```

#### A.3 Install Instructions
```bash
# In README.md, add to Cursor section:

curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/scripts/cursor-setup.sh | bash

# What it does:
# 1. Builds waldo-cursor-sync binary
# 2. Symlinks to ~/.local/bin/waldo-cursor-sync
# 3. Prints: "Now run 'waldo-cursor-sync' in your project to sync"
```

### B. ChatGPT / Gemini Integration

#### B.1 Web UI Export
```go
// internal/export/chatgpt.go - NEW MODULE
package export

import (
    "fmt"
)

type ChatGPTExport struct {
    Persona *config.PersonaConfig
}

// GenerateSystemPrompt creates a ChatGPT-compatible system prompt
func (ce *ChatGPTExport) GenerateSystemPrompt() string {
    p := ce.Persona

    prompt := `You are an AI assistant designed to match this personality profile:

## Tone
- Formality: %0.2f (0=casual, 1=formal)
- Directness: %0.2f (0=roundabout, 1=blunt)
- Humor: %0.2f (0=dry, 1=frequent)
- Hedging: %0.2f (0=confident, 1=qualified)
- Warmth: %0.2f (0=cold, 1=enthusiastic)

## Verbosity
- Response length: %s
- Reading level: %s
- Format preference: %s

## Voice
- Avoid words: %v
- Prefer words: %v
- Custom phrases: %v

Apply these traits to all your responses.`

    return fmt.Sprintf(prompt,
        p.Tone.Formality,
        p.Tone.Directness,
        p.Tone.Humor,
        p.Tone.Hedging,
        p.Tone.Warmth,
        p.Verbosity.ResponseLength,
        p.Verbosity.ReadingLevel,
        p.Verbosity.FormatPreference,
        p.Voice.AvoidWords,
        p.Voice.PreferWords,
        p.Voice.CustomPhrases,
    )
}

// GenerateMarkdown creates a ChatGPT-friendly markdown export
func (ce *ChatGPTExport) GenerateMarkdown() string {
    // Similar format but with markdown headers, code blocks, etc.
    return ""  // stub
}
```

#### B.2 CLI Command
```go
// internal/cli/export.go
func RunExport(format string, name string) {
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
        os.Exit(1)
    }

    // Load persona
    personaFile := filepath.Join(cfg.Root, "agent", name+".meml")
    // ... parse persona ...

    switch format {
    case "chatgpt":
        exporter := export.ChatGPTExport{Persona: persona}
        fmt.Println(exporter.GenerateSystemPrompt())
        fmt.Println("\n## Quick Copy for ChatGPT Custom Instructions:")
        fmt.Println("```")
        fmt.Println(exporter.GenerateSystemPrompt())
        fmt.Println("```")

    case "gemini":
        // Similar to ChatGPT

    case "json":
        // Raw JSON export

    case "clipboard":
        // Copy to clipboard (macOS: pbcopy, Linux: xclip)
    }
}
```

#### B.3 Web UI Enhancement
```html
<!-- ui/index.html - add Export section -->
<div class="export-section">
  <h2>Export for Other Tools</h2>

  <div class="tabs">
    <button @click="exportTab = 'chatgpt'">ChatGPT</button>
    <button @click="exportTab = 'gemini'">Gemini</button>
    <button @click="exportTab = 'claude-api'">Claude API</button>
  </div>

  <div v-if="exportTab === 'chatgpt'" class="export-content">
    <p>Copy this system prompt into ChatGPT Custom Instructions:</p>
    <button @click="copyExport('chatgpt')">Copy ChatGPT Prompt</button>
    <pre>{{ generateChatGPTPrompt() }}</pre>
  </div>

  <div v-if="exportTab === 'gemini'" class="export-content">
    <p>Copy this into Gemini's System Instructions:</p>
    <button @click="copyExport('gemini')">Copy Gemini Prompt</button>
    <pre>{{ generateGeminiPrompt() }}</pre>
  </div>
</div>

<style>
.export-section {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-top: 20px;
}

.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.tabs button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
  border-radius: 4px;
}

.tabs button.active {
  background: #0066cc;
  color: white;
}

pre {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
}
</style>
```

### C. Hanko Integration Stubs

#### C.1 Registry Scaffolding
```go
// cmd/waldo-registry/main.go - STUB
// Future: Web service for persona marketplace with Hanko auth

package main

import (
    "fmt"
)

func main() {
    fmt.Println(`
waldo-registry — Persona marketplace and registry

Features (planned v1.1):
  - User signup via Hanko (passwordless, passkeys)
  - Publish personas: waldo registry publish my-voice
  - Download personas: waldo registry install user/my-voice
  - Rate & review: ⭐⭐⭐⭐⭐
  - Full-text search

Environment:
  HANKO_URL          https://hanko.your-domain.com
  REGISTRY_DB        postgres://...
  REGISTRY_PORT      8080

Usage:
  waldo registry login              # Hanko SSO
  waldo registry publish ./my-voice.meml
  waldo registry search tone:direct
  waldo registry install caboose/default

TODO: Implement Hanko authentication
`)
}
```

#### C.2 Hanko Client Stub
```go
// internal/auth/hanko.go - STUB
package auth

import (
    "fmt"
)

type HankoClient struct {
    BaseURL string
    // TODO: github.com/teamhanko/hanko-sdk-go
}

func NewHankoClient(baseURL string) *HankoClient {
    return &HankoClient{BaseURL: baseURL}
}

// Login opens browser for passkey authentication
// TODO: Implement with hanko-sdk-go
func (h *HankoClient) Login() (string, error) {
    fmt.Println("TODO: Implement Hanko passwordless login")
    return "", nil
}

// Validate checks JWT token
// TODO: Implement JWT validation
func (h *HankoClient) ValidateToken(token string) error {
    fmt.Println("TODO: Implement token validation")
    return nil
}
```

#### C.3 Registry API Stub
```go
// cmd/waldo-registry/api/routes.go - STUB
package api

import (
    "fmt"
)

func RegisterRoutes(mux interface{}) {
    // TODO: Implement with Hanko auth middleware

    // GET /api/personas/{id} — public
    // GET /api/personas — search (public)
    // POST /api/personas — create (authenticated)
    // PUT /api/personas/{id} — update (authenticated, owner)
    // DELETE /api/personas/{id} — delete (authenticated, owner)
    // POST /api/personas/{id}/rate — rate (authenticated)

    fmt.Println("TODO: Implement registry API endpoints with Hanko middleware")
}
```

---

## III. Implementation Roadmap

### Phase 1: Accessibility (Week 1-2)
- [ ] Add CLI flags: `--no-emoji`, `--ascii`, `--no-color`
- [ ] Update waldo-status formatting (ASCII fallbacks)
- [ ] Update waldo-tui rendering (ASCII boxes, ASCII markers)
- [ ] Terminal detection + auto-fallback
- [ ] MEML parser: accept `@secret` as well as `🔑`
- [ ] Unit tests for ASCII conversions
- [ ] Documentation: ACCESSIBILITY.md

**Deliverable:** `waldo-status --no-emoji` produces:
```
[PROD] main [DIRTY] [+1] | [IDLE] | [PERSONA] agent/default
```

### Phase 2: Cross-Tool Integration (Week 2-3)
- [ ] Implement `waldo export` command (ChatGPT, Gemini, JSON)
- [ ] Build waldo-cursor-sync (Cursor workspace rules)
- [ ] Web UI: Export tab with copy-to-clipboard
- [ ] Cursor setup script (install waldo-cursor-sync)
- [ ] Integration tests: verify exports are valid prompts
- [ ] Documentation: INTEGRATIONS.md (Cursor, ChatGPT, Gemini)

**Deliverable:**
```bash
waldo export chatgpt agent/default | pbcopy  # Copy to clipboard
waldo-cursor-sync  # Write .cursorrules file
```

### Phase 3: Hanko Stubs (Week 3-4, Low Priority)
- [ ] Add scaffold: cmd/waldo-registry/
- [ ] Add client stub: internal/auth/hanko.go
- [ ] Add API stub: cmd/waldo-registry/api/
- [ ] Documentation: REGISTRY-ROADMAP.md
- [ ] Link to Hanko docs for future implementer

**Deliverable:** Skeleton code ready for full implementation in v1.1

---

## IV. Testing Strategy

### Unit Tests
```bash
go test ./internal/cli/... -v          # Accessibility flags
go test ./internal/export/... -v       # Export formats
go test ./internal/terminal/... -v     # Terminal detection
```

### Integration Tests
```bash
# Test waldo-status in CI environment (no emoji)
GITHUB_ACTIONS=true ./bin/waldo-status | grep -v "🚨"

# Test waldo export produces valid prompt
waldo export chatgpt agent/default | grep -q "Formality:"

# Test waldo-cursor-sync creates valid .cursorrules
waldo-cursor-sync && test -f .cursorrules
```

### Manual Testing Checklist
- [ ] Run in iTerm2 (emoji works)
- [ ] Run in SSH session (emoji + raw mode disabled)
- [ ] Run in GitHub Actions runner (emoji detected, ASCII used)
- [ ] Run in Docker container (CI detection works)
- [ ] Run in tmux (raw mode disabled if needed)
- [ ] Test `--no-emoji` flag in all tools
- [ ] Test ChatGPT export, paste into ChatGPT, verify it works
- [ ] Test Gemini export, paste into Gemini, verify it works
- [ ] Test Cursor sync, verify .cursorrules written

---

## V. Documentation Updates

### New Files
- [ ] `docs/ACCESSIBILITY.md` — Terminal support, environment variables, screen reader notes
- [ ] `docs/INTEGRATIONS.md` — Cursor, ChatGPT, Gemini setup + examples
- [ ] `docs/REGISTRY-ROADMAP.md` — Future persona marketplace with Hanko

### Updates to Existing
- [ ] `README.md` — Add "Export & Integration" section
- [ ] `CONTRIBUTING.md` — Update testing guidelines

---

## VI. Success Criteria

**Accessibility:**
- ✅ `waldo-status --no-emoji` produces readable ASCII output
- ✅ Runs without errors in GitHub Actions, Docker, CI/CD
- ✅ Screen reader friendly (no emoji in output)
- ✅ All tests pass in CI with `GITHUB_ACTIONS=true`

**Cross-Tool Integration:**
- ✅ `waldo export chatgpt` produces valid ChatGPT system prompt
- ✅ `waldo export gemini` produces valid Gemini system instruction
- ✅ `waldo-cursor-sync` writes valid `.cursorrules` file
- ✅ Users can manually copy-paste exports to ChatGPT/Gemini UI
- ✅ Cursor users can `waldo-cursor-sync` to auto-update workspace rules

**Hanko Stubs:**
- ✅ Skeleton code is in place
- ✅ Clear `// TODO` comments for implementer
- ✅ Roadmap documented for future v1.1 release

---

## VII. Open Questions

1. **MEML ASCII annotations:** Should old MEML files with emoji be auto-converted to ASCII?
   - Recommendation: Support both, no auto-conversion (backwards compatible)

2. **Cursor sync:** Should it auto-run on SessionStart hook or manual only?
   - Recommendation: Manual (`waldo-cursor-sync`); auto-sync is Cursor-specific setup

3. **Export formats:** Should we support Anthropic Claude API format too?
   - Recommendation: Yes (similar to ChatGPT, different field names)

4. **Registry scope:** Full v1.1 or defer to v2.0?
   - Recommendation: Stubs + roadmap in v0.2; full implementation in v1.1

---

**End of Plan**
