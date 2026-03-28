# ask-a-dev Setup Guide

**Live URL:** `https://caboose-mcp.github.io/waldo/ui/ask-a-dev/`

A transparent, AI-moderated daily Q&A for developers. Each day a new question is posted to Discord, developers reply anonymously, and an AI summarizes the answers with full transparency about how it works.

---

## Architecture Overview

```
n8n Cloud (Orchestration)
  ├── Daily Question Generation (cron 9am UTC)
  ├── Admin Approval Form
  ├── End-of-Day Summary (cron 10pm UTC)
  └── Reply Moderation Webhook

    ↓ (GitHub API commits)

GitHub Repo (Data + Configs)
  ├── ask-a-dev/config/ (MEML configs)
  ├── ask-a-dev/personas/ (AI personas)
  └── ui/ask-a-dev/data/ (JSON data files)

    ↓ (served via GitHub Pages)

Public UI (GitHub Pages)
  ├── index.html (main Q&A page)
  └── maintenance.html (down page)
```

All logic runs in n8n Cloud. The website is entirely static and served by GitHub Pages. No backend needed.

---

## Prerequisites

- Discord server + bot application
- n8n Cloud account (free tier is fine)
- Anthropic Claude API key
- GitHub Fine-Grained Personal Access Token (PAT)
- Google OAuth credentials (optional, for admin panel)

---

## Step 1: Create Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **New Application** → name it `ask-a-dev`
3. Go to **Bot** section → click **Add Bot**
4. Copy the **TOKEN** (you'll need this later)
5. Go to **OAuth2 → URL Generator**:
   - Scopes: `bot`
   - Permissions: `Send Messages`, `Embed Links`, `Read Message History`
   - Copy the generated URL and invite the bot to your server

6. Create a private Discord channel (e.g., `#admin-approvals`) for admin DMs

**Save these values:**
```
DISCORD_BOT_TOKEN=xxx
DISCORD_SERVER_ID=xxx
DISCORD_ADMIN_CHANNEL_ID=xxx      # DM channel or approvals channel
DISCORD_QUESTION_CHANNEL_ID=xxx   # Where daily Qs post
```

---

## Step 2: Set Up n8n Cloud

1. Go to [n8n Cloud](https://n8n.cloud) → sign up (free tier)
2. Create a new workspace
3. Go to **Credentials** (left sidebar)
4. Add three credentials:

### 2a. Discord Bot Token
- Type: `HTTP Header Auth`
- Name: `Discord Bot`
- Header: `Authorization`
- Value: `Bot YOUR_DISCORD_BOT_TOKEN`

### 2b. Anthropic Claude API Key
- Type: `Custom API`
- Name: `Anthropic API`
- Header: `x-api-key`
- Value: `YOUR_ANTHROPIC_API_KEY`
- Also add header: `anthropic-version` = `2023-06-01`

### 2c. GitHub PAT
- Type: `GitHub`
- Name: `GitHub PAT`
- Personal Access Token: `YOUR_GITHUB_PAT` (create at https://github.com/settings/tokens)
  - Scopes needed: `repo` (full control), `workflow` (to trigger Actions)

---

## Step 3: Configure GitHub

### 3a. Create Fine-Grained PAT

Go to [GitHub Settings → Developer Settings → Personal access tokens → Fine-grained tokens](https://github.com/settings/tokens?type=beta)

- **Token name:** `ask-a-dev-n8n`
- **Resource owner:** Select your account
- **Repository access:** `Only select repositories` → select `caboose-mcp/waldo`
- **Permissions:**
  - Contents: `Read and Write` (n8n commits data files)
  - Actions: `Read and Write` (optional, for workflow dispatch)
  - Workflows: `Read and Write` (optional)

**Copy the token** (you won't see it again).

### 3b. Create GitHub Secrets (optional, for Actions)

If you want GitHub Actions to trigger n8n workflows:

Go to `caboose-mcp/waldo` → Settings → Secrets and variables → Actions

```
N8N_WEBHOOK_URL = https://your-n8n-instance.n8n.cloud/webhook/...
N8N_API_KEY = your-n8n-api-key
```

---

## Step 4: Create n8n Workflows

The n8n workflows are exported as JSON files in `ask-a-dev/workflows/`. You'll import them into n8n:

### 4a. Daily Question Generation Workflow

1. In n8n, click **Import Workflow**
2. Select `ask-a-dev/workflows/daily-question.n8n.json`
3. Configure nodes:
   - Replace `{{ $env.DISCORD_QUESTION_CHANNEL_ID }}` with actual channel ID
   - Set cron trigger to **9:00 AM UTC** (or your preferred time)
   - Link Discord Bot credential
   - Link Anthropic API credential
   - Link GitHub PAT credential
4. **Activate** the workflow

### 4b. Admin Approval Form Workflow

1. Import `ask-a-dev/workflows/admin-approval.n8n.json`
2. This workflow is triggered when admin submits the approval form (n8n Form Trigger)
3. Configure Discord channel ID
4. **Activate**

### 4c. Reply Moderation Webhook

1. Import `ask-a-dev/workflows/reply-moderation.n8n.json`
2. This creates a webhook URL that the website POSTs to
3. Copy the webhook URL once deployed
4. Update `ui/ask-a-dev/data/feature-flags.json` with the webhook URL:
   ```json
   {
     "reply_webhook_url": "https://your-n8n-instance.n8n.cloud/webhook/ask-a-dev-replies"
   }
   ```
5. **Activate**

### 4d. Daily Summary Workflow

1. Import `ask-a-dev/workflows/daily-summary.n8n.json`
2. Set cron trigger to **10:00 PM UTC** (or your preferred time)
3. Link credentials
4. **Activate**

---

## Step 5: Configure MEML Files

Edit the MEML config files in `ask-a-dev/config/`:

### setup.meml

```meml
[🔧 discord]
server_id = "1234567890"
admin_dm_channel_id = "0987654321"
question_channel_id = "1111111111"
admin_user_id = "YOUR_DISCORD_ID"

[⏰ schedule]
question_time = "09:00"
summary_time = "22:00"
timezone = "UTC"
```

### feature-flags.meml

```meml
[🚦 service_status]
maintenance_mode = ❌
service_status = "up"

[🔍 feature_toggles]
show_ai_reasoning = ✅
show_moderation_status = ✅
```

Commit these changes to the `main` branch.

---

## Step 6: Update Data Files

The data files in `ui/ask-a-dev/data/` are auto-updated by n8n. But you need to initialize `feature-flags.json`:

Edit `ui/ask-a-dev/data/feature-flags.json`:

```json
{
  "maintenance_mode": false,
  "service_status": "up",
  "reply_webhook_url": "https://your-n8n-instance.n8n.cloud/webhook/ask-a-dev-replies",
  "google_client_id": "YOUR_GOOGLE_CLIENT_ID.apps.googleusercontent.com",
  "admin_emails": ["your-email@example.com"]
}
```

Commit to `main` → GitHub Pages deploys automatically.

---

## Step 7 (Optional): Google OAuth for Admin Panel

If you want admin features (override question, flip maintenance mode):

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Go to **APIs & Services → OAuth consent screen**
4. Configure consent screen
5. Go to **Credentials** → **Create Credentials → OAuth 2.0 Client ID**
   - Application type: `Web application`
   - Name: `ask-a-dev-admin`
   - Authorized JavaScript origins: `https://caboose-mcp.github.io`
   - Authorized redirect URIs: `https://caboose-mcp.github.io/waldo/ui/ask-a-dev/`

6. Copy **Client ID** and **Client Secret**
7. Update `feature-flags.json` with Client ID
8. Create n8n credential for Client Secret (for n8n to verify tokens)

---

## Testing

### Manual Test: Post a Question

1. Go to n8n → select **Daily Question Workflow**
2. Click **Execute Workflow** (manual trigger)
3. Check your Discord `#question-channel` for the posted question
4. Check GitHub for new commit to `ui/ask-a-dev/data/today.json`

### Manual Test: Submit Reply

1. Go to `https://caboose-mcp.github.io/waldo/ui/ask-a-dev/`
2. Type a test reply in the form
3. Click **Submit Reply**
4. Check n8n logs for the webhook execution
5. Check `today.json` for the appended reply

### Check Maintenance Mode

1. Edit `ui/ask-a-dev/data/feature-flags.json`:
   ```json
   { "maintenance_mode": true }
   ```
2. Commit and push to `main`
3. Refresh the page → should redirect to `maintenance.html`
4. Flip back to `false` to re-enable

---

## Daily Operation

### For End Users

- Visit `https://caboose-mcp.github.io/waldo/ui/ask-a-dev/` each day
- Read the question
- Submit an anonymous reply
- See AI summary of all replies at end of day
- Learn from other developers' approaches

### For Admin (@sgcaboose)

- **9:00 AM UTC:** n8n generates 5 questions
- **Approval Step:** n8n DMs you with a form link
  - Click link → see 5 choices
  - Answer learning questions (for future improvement)
  - Click **Approve**
- **Thereafter:** Question posted to Discord + GitHub Pages
- **Throughout day:** Replies collected, moderated
- **10:00 PM UTC:** AI generates summary, posts to Discord + website

All moderation decisions are logged and visible in the UI.

---

## Monitoring

### n8n Logs

Visit your n8n workspace → click on workflow → **Executions**:
- Green checkmark = success
- Red X = failed
- Click to see logs

Common failures:
- **Discord API 401:** Bot token expired or wrong
- **GitHub 401:** PAT expired or wrong
- **Anthropic 401:** API key invalid
- **Webhook 404:** n8n workflow not activated or wrong URL

### GitHub Commits

Check `https://github.com/caboose-mcp/waldo/commits/main` for:
- `ui/ask-a-dev/data/today.json` - updated when question/summary changes
- No manual commits should happen here (all n8n)

### Website

Visit `https://caboose-mcp.github.io/waldo/ui/ask-a-dev/` and check:
- Question displays correctly
- Replies appear throughout day
- Summary shows at end of day
- Moderation status badge updates
- Maintenance page works if flipped

---

## MEML Natural Usage

This is a real-world example of MEML in production:

1. **Persona configs** (`bot.meml`, `moderation.meml`) - AI behavior defined in MEML
2. **Feature flags** (`feature-flags.meml`) - Service status stored in MEML
3. **Question pool** (`question-pool.meml`) - Seed questions as MEML records
4. **Setup config** (`setup.meml`) - Service configuration in MEML with emoji hints

Why MEML works here:
- Human-readable configs that GitHub renders nicely
- Emoji hints (`🎭 tone`, `🛡️ moderation`) make sections self-documenting
- Can version control alongside code
- Tools can parse emoji to detect section type
- More expressive than YAML (supports inline tables, booleans as emoji)

---

## Troubleshooting

### "Webhook URL not configured"

The website POSTs to the webhook URL in `feature-flags.json`. Make sure:
1. You've copied the webhook URL from n8n
2. You've updated `feature-flags.json` with it
3. The n8n workflow is activated
4. You've committed and pushed (GitHub Pages auto-deploys)

### "Discord bot not sending DMs"

1. Make sure bot has permissions in the server
2. Check n8n logs for the Discord API call
3. Verify bot token in n8n credentials is correct
4. Try manual test (n8n Execute Workflow button)

### "Replies not appearing"

1. Check n8n webhook logs for incoming POSTs
2. Check moderation logs to see if reply was rejected
3. Check that the webhook URL in HTML matches the one in feature-flags.json
4. Try submitting a blank reply (should fail validation in n8n)

### "Maintenance page won't go away"

1. Check `feature-flags.json` - make sure `maintenance_mode` is `false`
2. Hard refresh the page (Cmd+Shift+R or Ctrl+Shift+R)
3. Wait for GitHub Pages cache to expire (~30 seconds)

---

## Future Enhancements

- [ ] Reddit-style voting on replies
- [ ] Tag/filter replies by technology
- [ ] Export daily Q&A as Markdown
- [ ] Archive search across past days
- [ ] Email digest of best answers
- [ ] Integration with dev.to, Hacker News
- [ ] Mobile app

---

## Support

Issues? Questions?
- Check n8n logs for errors
- File an issue on GitHub
- Review the MEML configs for typos

---

**Enjoy building a transparent, AI-driven dev community!**
