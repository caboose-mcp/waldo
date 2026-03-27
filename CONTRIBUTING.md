# Contributing to waldo

## Local Development Setup

### Install left hook

waldo uses **left** for pre-commit checks (lint, validate, test).

```bash
# Install left
curl https://raw.githubusercontent.com/evilmartians/left/master/install | bash

# Install git hooks
left install

# Run checks before commit
left run pre-commit
```

### What left checks

- **MEML validation** — `meml validate *.meml`
- **Shell linting** — `shellcheck *.sh`
- **No old naming refs** — blocks commits with legacy system name
- **Markdown lint** — warns on style issues

### Running checks manually

```bash
# All pre-commit checks
left run pre-commit

# Specific check
left run meml
left run shellcheck
```

## CI/CD Pipeline

GitHub Actions runs on every push to `main` and PR:

1. **Lint** — MEML, shell, markdown validation
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

2. **Make changes and run left:**
   ```bash
   left run pre-commit
   ```

3. **Commit:**
   ```bash
   git commit -m "Your message"
   ```

4. **Push and open PR:**
   ```bash
   git push -u origin feature/your-feature
   ```

## Adding New Personas

Personas are stored as `.meml` files in `~/.config/waldo/personas/agent/`:

```meml
[🪪 meta]
name        = "my-voice"
description = "Description"
version     = "0.1.0"

[🎭 tone]
formality   = 0.5
directness = 0.8
humor       = 0.6
hedging     = 0.1
warmth      = 0.5
```

Validate before committing:
```bash
meml validate ~/.config/waldo/personas/agent/my-voice.meml
```

## Testing

Smoke tests run in CI:

- MEML file parsing
- Status line hook output
- Mood overlay detection

To test locally:
```bash
export HOME=/tmp/test-home
mkdir -p $HOME/.config/waldo
echo "agent/default" > $HOME/.config/waldo/.active
bash .claude/hooks/waldo/status-line.sh
```

## Documentation

- [README.md](./README.md) — Overview, quick start
- [MEML.md](./MEML.md) — MEML format reference
- [WALDO.md](./WALDO.md) — System architecture
- [waldo-SKILL-v5.md](./waldo-SKILL-v5.md) — Full skill reference

Update docs when adding features.

## Release Process

1. Create tag: `git tag v0.2.0`
2. Push: `git push origin v0.2.0`
3. GitHub Actions creates release automatically
4. Add release notes to GitHub

---

Questions? Check [WALDO.md](./WALDO.md) or open an issue.
