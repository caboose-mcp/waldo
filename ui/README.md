# waldo Web UI

Two interactive tools for working with personas:

1. **Playground** — Interactive sandbox to experiment with tone
2. **Config UI** — Visual editor for detailed persona customization

## Playground (Quick Start)

**Try without installation:** https://caboose-mcp.github.io/waldo/ui/playground.html

Interactive MEML editor with real-time tone slider controls:
- Adjust formality, directness, humor, hedging, warmth
- See JSON, ChatGPT prompt, and Gemini prompt update live
- Copy prompts to clipboard or download MEML file
- Pre-loaded examples: default, direct, formal, warm, technical
- Works entirely in browser (no backend needed)

### Local

```bash
# Open playground
open ui/playground.html
# or
firefox ui/playground.html
```

---

## Config UI (Advanced)

Full visual editor for tone, verbosity, voice customization.

### Local

```bash
# Open in browser
open ui/index.html
# or
firefox ui/index.html
```

> **Note:** Vue 3 is loaded from a CDN (unpkg), so an internet connection is required.
> Some features like `navigator.clipboard` may not work under a `file://` origin in all browsers;
> if clipboard copy fails, select the JSON from the preview panel and copy it manually.
> Serving the file over `http://localhost` (e.g. `python3 -m http.server`) avoids this.

### GitHub Pages

The UI is served at: `https://caboose-mcp.github.io/waldo/ui/`

## Features

- **Visual editor** for tone sliders (formality, directness, humor, etc.)
- **Word management** for avoid/prefer/custom phrases (add/remove tags)
- **Live preview** — JSON output updates in real-time
- **Copy to clipboard** — JSON ready to import
- **Download MEML** — Export as `.meml` file for the repo
- **GitHub sync** — Optional token for GitHub API features (not required for local editing)

## How It Works

```
User edits tone sliders & word lists
         ↓
Vue.js updates persona object in real-time
         ↓
Preview panel shows live JSON
         ↓
Copy JSON or download MEML
```

## Requirements

- Modern browser (Chrome, Firefox, Safari, Edge)
- Internet connection (Vue 3 is loaded from CDN)
- No server needed — runs entirely in browser
- GitHub token is **optional** — the editor works fully offline without one; a token is only needed for GitHub API sync features

## Setup

1. Open `ui/index.html` in any browser
2. Adjust tone sliders, add/remove words
3. Copy JSON or download MEML
4. Save to `~/.config/waldo/personas/agent/my-voice.meml`
5. Validate: `meml validate my-voice.meml`

## Default Persona

The UI loads with a default persona:

```
Name: my-voice
Formality: 0.3 (casual)
Directness: 0.8 (very direct)
Humor: 0.6 (moderate)
Hedging: 0.1 (confident)
Warmth: 0.5 (friendly)
```

Adjust and export.

### GitHub token security

If you use a GitHub token with this UI, it is stored in your browser's **session storage** so it can be reused across page loads within the same tab. It is automatically cleared when you close the tab or browser.

- Prefer [fine-grained tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token) with minimal scopes — only grant access to the repositories you need.
- Avoid using a high-privilege classic PAT (especially on shared or untrusted machines).
- To clear a stored token immediately, use the **Disconnect** button in the GitHub Sync section of the UI, or close the browser tab.

## Playground Features

| Feature | Status |
|---------|--------|
| Edit MEML in real-time | ✅ |
| Tone sliders (5 dimensions) | ✅ |
| Live JSON preview | ✅ |
| ChatGPT prompt generator | ✅ |
| Gemini prompt generator | ✅ |
| Copy to clipboard | ✅ |
| Download MEML file | ✅ |
| Pre-loaded examples | ✅ |
| Responsive design | ✅ |
| Zero installation | ✅ |

## Future

**Config UI:**
- GitHub API integration: save personas directly to repo
- Load personas from `.meml` files
- Slack import UI
- Mood overlay builder

**Playground:**
- WASM meml parser (faster validation)
- Load from S3 buckets
- Save personas to personal account (v1.1)
- Docker version with full API (optional)

---

**Tech stack:**
- Vue 3 (CDN, pinned to 3.4.21, no build)
- Vanilla CSS
- Single-file HTML
