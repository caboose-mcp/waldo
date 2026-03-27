# Contribution Workflow + Persona-Driven Project Templates
## Plan for caboose-mcp Ecosystem

**Status:** Planning phase
**Scope:** meml, waldo, peon-ping, caboose-ai, and related caboose-mcp projects
**Priority:** High (enables community adoption)

---

## I. Contribution Workflow

### A. Public Repo Contributions (Standardized)

**Public repos (encourage contributions):**
- meml
- waldo
- peon-ping
- caboose-ai
- vscode-caboose-mcp

**Workflow:**

```bash
# 1. Fork + Clone
git clone https://github.com/YOUR_USERNAME/waldo.git
cd waldo

# 2. Create feature branch
git checkout -b feature/my-feature

# 3. Run pre-commit checks (left hook)
left run pre-commit

# 4. Commit with conventional format
git commit -m "feat: add new feature

Description of change and why.

Fixes #123"

# 5. Push + Open PR
git push origin feature/my-feature
gh pr create --title "Add new feature"
```

#### A.1 CONTRIBUTING.md Template

```markdown
# Contributing to {project}

We welcome contributions! This project uses:

## Setup

1. Fork and clone
2. Install left hook: `left install`
3. Run checks: `left run pre-commit`

## Making Changes

1. Create feature branch: `git checkout -b feature/name`
2. Make changes
3. Run checks: `left run pre-commit`
4. Commit with conventional messages
5. Push and open PR

## Commit Style

Use conventional commits:
- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation
- `test:` tests
- `chore:` maintenance
- `refactor:` code cleanup

Example:
```
feat: add fuzzy search to S3 picker

Implements substring-based ranking with prefix bonus.
Helps users quickly find S3 buckets by name.

Fixes #45
```

## Tests

Run: `go test ./...`

## Code Style

- Format: `gofmt`
- Vet: `go vet ./...`
- Lint: Use local linter

## License

By contributing, you agree that your contributions are licensed under MIT.
```

### B. Private Repo Access Request Flow

**Private repos (private development, selective access):**
- caboose-cursor-rules
- caboose-ai-private (hypothetical)
- Internal tools, demos, work-in-progress

**Access request workflow:**

#### B.1 Access Request Form (GitHub Discussions)

```markdown
# GitHub Discussion: Access Request

## Title: Access Request: [Your Name] → [Project]

## Form

**Your GitHub username:** @yourname

**Project you want access to:** (e.g., caboose-cursor-rules)

**Why you need access:**
- [ ] Contributing a fix
- [ ] Collaborating on feature
- [ ] Security research / audit
- [ ] Learning from code
- [x] Other (explain below)

**Explanation:**
I want to [understand how X works / contribute fix for Y / collaborate on Z]

**Time-sensitive?** Yes / No

---

Reply will be within 48 hours via @mention.
```

#### B.2 Invite Process

```bash
# Repository owner/maintainer:
gh repo invite --repo caboose-mcp/private-repo --user YOUR_USERNAME --permission maintain

# Contributor receives email invite
# Accepts, gets repo access

# Owner documents reason in PRIVATE_ACCESS_LOG.md (local file, not in repo)
```

#### B.3 Private Repo Template

**Structure for private repos:**

```
private-repo/
  README.md (includes: access request instructions)
  PRIVATE_ACCESS_LOG.md (not committed, local only)
    # Private access log (for maintainers)
    - @alice: 2026-03-26, security audit access
    - @bob: 2026-03-26, contributing multi-user feature
  .github/ISSUE_TEMPLATE/
    - access-request.md
```

---

## II. Persona-Driven Project Templates

### A. Template System Overview

**Goal:** When a dev creates a new project, their active waldo persona influences:
- Suggested config file formats (MEML, YAML, JSON)
- Directory structure templates
- Coding style conventions
- Default tooling recommendations
- Theme/appearance suggestions

### B. Persona Config Extension

**Add to waldo persona MEML:**

```meml
[💾 dev-preferences]
config_format    = "meml"      # Preferred format: meml, yaml, json, toml
verbosity        = "detailed"  # detailed, concise, minimal
theme            = "dark"      # Editor theme: dark, light, auto
structure        = "flat"      # Project structure: flat, grouped, modular
testing          = "unit-focused"  # unit-focused, integration-heavy
documentation   = "code-first" # code-first, prose-first, balanced

[🎯 templates]
# Project templates this dev prefers
go_project      = "cli"        # cli, library, service, tool
node_project    = "node-esm"   # node-esm, node-cjs, deno, bun
template_style  = "minimal"    # minimal, comprehensive, opinionated
```

### C. Template Generation

#### C.1 New Command: `waldo project`

```bash
# Create new project using persona templates
waldo project new my-project --type cli

# Output:
# Creating project: my-project
# Type: CLI tool
# Using persona: agent/default
#   - Config format: MEML
#   - Structure: flat
#   - Testing: unit-focused
# ✓ Project created at ./my-project/
#   - go.mod, go.sum
#   - cmd/my-project/main.go
#   - internal/...
#   - Makefile (formatted for your style)
#   - example.meml (template config)
```

#### C.2 Template Library

```
internal/templates/
  go/
    cli/
      main.go
      Makefile
      example.meml
      README.md (template)
    library/
      lib.go
      ...
  node/
    esm/
      package.json
      src/index.js
      example.meml
    cjs/
      ...
  python/
    script/
    library/
    service/
  ...
```

#### C.3 Implementation

```go
// cmd/waldo-project/main.go - NEW BINARY
package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/caboose-mcp/waldo/internal/config"
    "github.com/caboose-mcp/waldo/internal/project"
)

func main() {
    typeFlag := flag.String("type", "cli", "Project type: cli, library, service, tool")
    langFlag := flag.String("lang", "go", "Language: go, node, python, rust")
    flag.Parse()

    args := flag.Args()
    if len(args) < 2 || args[0] != "new" {
        usage()
        os.Exit(1)
    }

    projectName := args[1]

    // Load active persona
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load waldo config: %v\n", err)
        os.Exit(1)
    }

    // Create project
    creator := project.NewCreator(
        projectName,
        *langFlag,
        *typeFlag,
        cfg.ActivePersona,  // Pass persona for template customization
    )

    if err := creator.Create(); err != nil {
        fmt.Fprintf(os.Stderr, "❌ Failed to create project: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("✓ Project '%s' created\n", projectName)
    fmt.Printf("  Language: %s\n", *langFlag)
    fmt.Printf("  Type: %s\n", *typeFlag)
    fmt.Printf("  Persona: %s\n", cfg.ActivePersona)
}

func usage() {
    fmt.Fprintf(os.Stderr, `waldo project new <name> [options]

Options:
  -type go|node|python|rust  Project type (default: cli)
  -lang go|node|python|rust  Language (default: go)

Examples:
  waldo project new my-cli --type cli
  waldo project new my-lib --type library --lang go
`)
}
```

#### C.4 Template Customization via Persona

```go
// internal/project/templates.go
type TemplateContext struct {
    ProjectName    string
    PersonaName    string
    ConfigFormat   string  // From persona: meml, yaml, json
    DocStyle       string  // From persona: code-first, prose-first
    TestingStyle   string  // From persona: unit-focused, integration-heavy
    Theme          string  // From persona: dark, light
}

// Render template with persona context
func RenderTemplate(templateFile string, ctx *TemplateContext) (string, error) {
    // Read template
    // Substitute {{.ProjectName}}, {{.ConfigFormat}}, etc.
    // Return rendered content
    return "", nil
}
```

**Example: Makefile template**
```makefile
# Makefile — auto-generated by waldo

PROJECT = {{.ProjectName}}
PERSONA = {{.PersonaName}}  # {{.DocStyle}} | {{.TestingStyle}}

# Format based on persona preference
{{if eq .TestingStyle "unit-focused"}}
test:
	go test ./... -v
{{else}}
test:
	go test ./... -v
	go test ./... -race
{{end}}

build:
	go build -o bin/{{.ProjectName}} ./cmd/{{.ProjectName}}/
```

### D. Config File Suggestions

**When creating new config, suggest based on persona:**

```bash
# User's persona prefers MEML
$ waldo config create app-config

📋 Config File Suggestions:
  • MEML (your preference) — emoji + semantic annotations
  • YAML — familiar, good for configs
  • JSON — verbose but portable

Recommended: MEML (matches your persona style)

$ waldo config create app-config --format meml
✓ Created: app-config.meml

Sample structure:
  [⚙️ app]
  name        = "my-app"
  version     = "0.1.0"
  [🔧 server]
  host        = "localhost"
  port        = 8080
```

### E. Theme Recommendations

**Based on persona tone:**

```go
// internal/project/themes.go
type ThemeRecommendation struct {
    Primary   string
    Secondary string
    Accent    string
    Reasoning string
}

func RecommendTheme(persona *config.Persona) *ThemeRecommendation {
    // Analyze persona tone
    if persona.Tone.Formality > 0.7 {
        // Formal → cool colors, high contrast
        return &ThemeRecommendation{
            Primary:   "dark-blue",
            Secondary: "gray",
            Accent:    "cyan",
            Reasoning: "High formality suggests professional, cool-toned theme",
        }
    }

    if persona.Tone.Warmth > 0.6 {
        // Warm → warm colors, softer contrast
        return &ThemeRecommendation{
            Primary:   "dark-brown",
            Secondary: "orange",
            Accent:    "gold",
            Reasoning: "High warmth suggests inviting, warm-toned theme",
        }
    }

    // Default → balanced
    return &ThemeRecommendation{
        Primary:   "dark",
        Secondary: "gray",
        Accent:    "blue",
        Reasoning: "Neutral tone; balanced, adaptable theme",
    }
}
```

### F. Learning + Suggestions

**Persona learns config preferences over time:**

```go
// internal/persona/learning.go
type ConfigLearning struct {
    ProjectsCreated  int
    PreferredFormats map[string]int  // meml: 5, yaml: 2, json: 1
    TemplateUsage    map[string]int  // cli: 3, library: 2
    ThemesUsed       map[string]int  // dark: 10, light: 2
    AvgProjectSize   string          // small, medium, large
}

// Suggest based on learning
func (pl *ConfigLearning) SuggestFormat() string {
    // Find most-used format
    max := 0
    var bestFormat string
    for fmt, count := range pl.PreferredFormats {
        if count > max {
            max = count
            bestFormat = fmt
        }
    }
    return bestFormat  // e.g., "meml" (used 5 times)
}

// Store learning in deltas
func SaveLearning(learning *ConfigLearning) error {
    delta := map[string]interface{}{
        "config_format_preference": learning.PreferredFormats,
        "template_preference":      learning.TemplateUsage,
        "theme_preference":         learning.ThemesUsed,
    }
    // Save to ~/.config/waldo/personas/agent/.deltas
    return nil
}
```

---

## III. Implementation Roadmap

### Phase 1: Contribution Workflow (Week 1)
- [ ] Create standardized CONTRIBUTING.md template
- [ ] Set up access request flow (GitHub Discussions)
- [ ] Document in each public repo
- [ ] Create PRIVATE_ACCESS_LOG.md template

### Phase 2: Persona Extensions (Week 2)
- [ ] Extend MEML schema with `[💾 dev-preferences]` section
- [ ] Add parser support for new fields
- [ ] Update example personas with dev preferences

### Phase 3: Project Templates (Week 3)
- [ ] Implement `waldo project` command
- [ ] Build template library (Go CLI, library, service)
- [ ] Config suggestion system
- [ ] Theme recommendations

### Phase 4: Learning System (Week 4)
- [ ] Track config preferences over time
- [ ] Store in persona deltas
- [ ] Generate suggestions based on history
- [ ] Update persona based on learning

---

## IV. File Changes

### New Files
- [ ] `cmd/waldo-project/main.go` — Project creation command
- [ ] `internal/project/templates.go` — Template rendering
- [ ] `internal/project/learning.go` — Preference learning
- [ ] `CONTRIBUTING.md` (template)
- [ ] `.github/ISSUE_TEMPLATE/access-request.md`
- [ ] `docs/CONTRIBUTION-WORKFLOW.md`
- [ ] `internal/templates/go/cli/main.go`, `Makefile`, `README.md`

### Existing Files (Updates)
- [ ] Update `MEML.md` with `[💾 dev-preferences]` section
- [ ] Update example personas (example.meml)
- [ ] Update README.md (add contribution section)

---

## V. Example: Full Workflow

**User journey:**

```bash
# 1. Create new Go CLI project
waldo project new my-app --type cli

# Creates:
# my-app/
#   go.mod
#   cmd/my-app/main.go
#   internal/cli/commands.go
#   Makefile (formatted for user's style)
#   example.meml (config template)
#   README.md

# 2. User customizes config
cat my-app/example.meml
# Uses MEML (matches their persona)
# Includes emoji annotations, clean format

# 3. User learns preferences over time
# After creating 5 projects, waldo learns:
#   - User prefers MEML configs (5 of 5)
#   - User prefers flat structure
#   - User prefers unit-focused testing
#   - User likes dark theme

# 4. Next project inherits learning
waldo project new another-app
# Suggestions now more accurate
# Template rendering matches past preferences
```

---

## VI. Success Criteria

- ✅ CONTRIBUTING.md is standardized across all public repos
- ✅ Access request flow is documented and working
- ✅ `waldo project new` creates functional project skeleton
- ✅ Templates render with persona-specific customization
- ✅ Persona learns config preferences over time
- ✅ Suggestions improve with more project history

---

**End of Plan**
