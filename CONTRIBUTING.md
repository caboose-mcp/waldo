# Contributing to waldo

## Local Development Setup

### Install lefthook

waldo uses **lefthook** for pre-commit checks (lint, validate, shellcheck).

```bash
# Install lefthook (if not already installed)
go install github.com/evilmartians/lefthook@latest

# Install git hooks
lefthook install

# Run checks manually
lefthook run pre-commit
```

### What lefthook checks

- **MEML validation** — `meml validate *.meml`
- **Shell linting** — `shellcheck *.sh`
- **No old naming refs** — blocks commits with legacy system name

### Running checks manually

```bash
# All pre-commit checks
lefthook run pre-commit

# Specific check
lefthook run meml
lefthook run shellcheck
```

## CI/CD Pipeline

GitHub Actions runs on every push to `main` and on PRs:

1. **Lint** — MEML, shell validation
2. **Build** — shell script syntax check
3. **Test** — MEML parsing, status line hook, mood overlay
4. **Coverage** — docs coverage report
5. **Release** — creates release on tagged commits (`v*`)

See `.github/workflows/ci.yml` for details.

## Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Make changes and run lefthook:**
   ```bash
   lefthook run pre-commit
   ```

3. **Commit:**
   ```bash
   git commit -m "feat: your message"
   ```

4. **Push and open PR:**
   ```bash
   git push -u origin feature/your-feature
   ```

## Adding New Personas

Personas are stored as `.meml` files in `~/.claude/personas/agent/`:

```meml
[🪪 meta]
name        = "my-voice"
description = "Description"
version     = "0.1.0"

[🎭 tone]
formality   = 0.5
directness  = 0.8
humor       = 0.6
hedging     = 0.1
warmth      = 0.5
```

Validate before committing:
```bash
meml validate ~/.claude/personas/agent/my-voice.meml
```

## Testing

Smoke tests run in CI. To test locally:

```bash
export HOME=/tmp/test-home
mkdir -p "$HOME/.claude/personas/agent"
mkdir -p "$HOME/.config/waldo"
echo "agent/default" > "$HOME/.config/waldo/.active"
bash .claude/hooks/waldo/status-line.sh
```

## Documentation

- [README.md](./README.md) — Overview and quick start
- [MEML.md](./MEML.md) — MEML format reference
- [waldo-SKILL-v5.md](./waldo-SKILL-v5.md) — Full skill reference

Update docs when adding features.

## Release Process

1. Create tag: `git tag v0.2.0`
2. Push: `git push origin v0.2.0`
3. GitHub Actions creates the release automatically

---

Questions? Open an issue.
