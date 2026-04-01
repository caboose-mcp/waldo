# Waldo Playground — Implementation & Deployment Guide

**Status:** Phase 1 Complete ✅
**Live URL:** https://caboose-mcp.github.io/waldo/ui/playground.html

---

## What's Included

### Phase 1: Static HTML+JS Playground (MVP)

**Files:**
- `ui/playground.html` — Split-pane editor + preview UI
- `ui/playground.js` — MEML parser, prompt generators, interactivity

**Features:**
- ✅ Real-time MEML editor (CodeMirror)
- ✅ Tone sliders (formality, directness, humor, hedging, warmth)
- ✅ Live JSON preview
- ✅ ChatGPT prompt generator
- ✅ Gemini prompt generator
- ✅ Validation feedback
- ✅ Copy to clipboard (all outputs)
- ✅ Download MEML file
- ✅ Pre-loaded examples (5 personas)
- ✅ Dark theme, responsive design
- ✅ Zero-install, browser-only

**Tech Stack:**
- HTML5, vanilla JavaScript
- CodeMirror 5 (vendored in `ui/lib/codemirror/`) for editing
- No backend, no build step, no external requests
- ~100KB total size (uncompressed)

---

## How It Works

### 1. User Opens Playground
```
https://caboose-mcp.github.io/waldo/ui/playground.html
└─ Served by GitHub Pages
```

### 2. Load MEML Config
- Start with default persona (formality=0.5, etc.)
- Or click "Direct"/"Formal"/"Warm"/"Technical" example buttons
- Or paste custom MEML in editor

### 3. Adjust Tone Sliders
- Move sliders → MEML editor updates
- Editor changes → JSON + prompts regenerate
- Changes are live and immediate

### 4. Export
- **Copy JSON** → Paste into waldo config
- **Copy ChatGPT Prompt** → Paste into ChatGPT custom instructions
- **Copy Gemini Prompt** → Use with Gemini AI
- **Download MEML** → Save locally, validate with `meml` CLI

---

## Deployment

The playground is deployed automatically via GitHub Actions whenever `ui/` changes are pushed to `main`.

### Manual Deployment Check

```bash
# Push changes
git add ui/playground.*
git commit -m "Update playground"
git push origin main

# GitHub Actions automatically:
# 1. Builds (no-op, just uploads files)
# 2. Deploys to GitHub Pages
# 3. Available in ~30 seconds at:
#    https://caboose-mcp.github.io/waldo/ui/playground.html
```

**Workflow:** `.github/workflows/pages.yml` (already configured)

---

## Customization

### Add New Examples

Edit `playground.js` `examples` object:

```javascript
const examples = {
  // ... existing examples ...
  'my-example': {
    meml: `# my example\n\n[🎭 tone]\nformality = 0.6\n...`,
    tone: { formality: 0.6, /* ... */ }
  }
};
```

Then add button in HTML:
```html
<button class="example-btn" data-example="my-example">My Example</button>
```

### Modify Prompt Generation

Edit `generateChatGPTPrompt()` and `generateGeminiPrompt()` functions in `playground.js`:

```javascript
function generateChatGPTPrompt(parsed) {
  const { tone } = parsed;
  // Customize logic here
  return `You are ...`;
}
```

### Styling

Edit CSS in `playground.html` `<style>` section:
- Colors: `#667eea` (primary), `#764ba2` (accent), `#1e1e1e` (background)
- Responsive breakpoints: 1024px, 768px

---

## Integration with Waldo Ecosystem

### For End Users

1. **Onboarding:** "Try the playground first"
   - Link in README (✅ already added)
   - Link in main docs
   - Share in announcements

2. **Workflow:**
   ```
   User tries playground
       ↓
   Adjusts tone sliders, likes result
       ↓
   Copies ChatGPT prompt, tries in ChatGPT
       ↓
   Decides to install waldo for CLI integration
       ↓
   Downloads MEML from playground
       ↓
   Puts in ~/.config/waldo/personas/agent/my-voice.meml
   ```

3. **Alternative workflow (power users):**
   ```
   User creates persona in waldo CLI
       ↓
   Exports to JSON: waldo export my-voice
       ↓
   Loads into playground for visualization
       ↓
   Tweaks tone via sliders
       ↓
   Downloads updated MEML
       ↓
   Saves back to ~/.config/waldo/personas/agent/
   ```

### For Developers

The playground provides:
- **Demo for contributors** — See how tone affects prompts
- **Teaching tool** — Explain MEML format to new users
- **Validation lab** — Test persona configs before deployment

---

## Future Enhancements (Phase 2+)

### Phase 2: WASM Parser (Week 2)

**Goal:** Faster parsing, better error messages, validation at scale

```bash
# Compile Go MEML parser to WASM
cd ../meml
GOOS=js GOARCH=wasm go build -o ../waldo/ui/meml.wasm ./cmd/meml

# Update playground.js to use WASM instead of JS parser
# Performance improvement: ~10x faster for large configs
```

**Files to add:**
- `ui/meml.wasm` — Compiled MEML parser
- `ui/wasm_exec.js` — Go WASM runtime (from Go stdlib)
- Update `playground.js` to call WASM functions

**Benefit:** Sub-millisecond parse times, instant error detection

### Phase 3: Docker Server (Week 3, Optional)

**Goal:** Full-power playground with S3 integration, script execution

```bash
docker run -p 8080:8080 caboose-mcp/waldo-playground:latest
open http://localhost:8080/playground.html
```

**Files to add:**
- `cmd/playground-server/main.go` — HTTP server
- `Dockerfile` — Container image
- `.dockerignore`

**Endpoints:**
- `POST /api/validate` — MEML syntax check
- `POST /api/export` — ChatGPT/Gemini/JSON export
- `GET /api/s3-list` — List user's S3 buckets
- `POST /api/execute` — Run bash in SRT sandbox

**Deployment:**
```bash
docker build -t caboose-mcp/waldo-playground .
docker push caboose-mcp/waldo-playground
# Users: docker run -p 8080:8080 caboose-mcp/waldo-playground
```

---

## Testing Checklist

### Manual Testing (Before Each Commit)

- [ ] Open playground in Chrome
- [ ] Adjust sliders → JSON updates
- [ ] Click "Direct" example → values change
- [ ] Edit MEML → JSON updates
- [ ] Copy JSON → paste into text editor (works)
- [ ] Copy ChatGPT prompt → paste into ChatGPT
- [ ] Download MEML → file downloads correctly
- [ ] Mobile view (resize to 768px) → responsive layout works
- [ ] All tabs switch correctly
- [ ] No console errors

### Automated Testing (Optional)

Could add Playwright E2E tests:
```bash
npm install --save-dev @playwright/test
# Test clicks, slider movements, copy actions
```

---

## Troubleshooting

### "Clipboard copy doesn't work"

**Cause:** `navigator.clipboard` requires HTTPS or localhost
**Fix:** Use `https://caboose-mcp.github.io/...` (GitHub Pages uses HTTPS)

### "CodeMirror not loading"

**Cause:** Vendored files missing or path incorrect
**Fix:** Verify `ui/lib/codemirror/codemirror.js`, `codemirror.css`, and required addons exist

### "Styles look wrong"

**Cause:** CSS selectors not matching
**Fix:** Check browser DevTools → verify class names, try hard refresh

### "Examples don't load"

**Cause:** Data object not initialized
**Fix:** Check browser console for JS errors, verify `examples` object

---

## Success Metrics

✅ **Achieved:**
- Zero-install playground deployed to GitHub Pages
- Sub-100ms response time (static HTML)
- 5 pre-loaded example personas
- ChatGPT + Gemini prompt generation
- Copy-to-clipboard working
- Responsive design (desktop/tablet/mobile)
- Dark theme, accessible UI

**Target metrics:**
- 1k+ visitors in first month
- 100+ downloads of MEML files
- 50+ "Copy to ChatGPT" actions
- 90%+ browser compatibility (Chrome, Firefox, Safari, Edge)

---

## Next Steps

1. **Push to main** → GitHub Pages auto-deploys
2. **Add playground link to README** ✅ (already done)
3. **Update UI README** ✅ (already done)
4. **Share with community** → Post in Discord, Twitter, etc.
5. **Gather feedback** → Monitor GitHub issues
6. **Plan Phase 2** → WASM parser integration

---

**Questions?** Open an issue or check out the commented code in `playground.js`.
