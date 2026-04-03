#!/bin/bash
# waldo — Cross-machine persona sync setup
# Install: curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/main/setup-waldo.sh -o setup-waldo.sh && bash setup-waldo.sh

set -eu

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  waldo — Persona Sync Setup      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo

# 1. Check dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

if ! command -v aws &>/dev/null; then
  echo -e "${RED}✗ AWS CLI not found${NC}"
  echo "Install: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
  exit 1
fi

if ! command -v jq &>/dev/null; then
  echo -e "${RED}✗ jq not found${NC}"
  echo "Install: brew install jq (macOS) or apt install jq (Linux)"
  exit 1
fi

if ! command -v git &>/dev/null; then
  echo -e "${RED}✗ git not found${NC}"
  exit 1
fi

echo -e "${GREEN}✓ AWS CLI, jq, git all present${NC}"
echo

# 2. Check AWS credentials
echo -e "${YELLOW}Checking AWS credentials...${NC}"

if ! aws sts get-caller-identity &>/dev/null; then
  echo -e "${RED}✗ AWS credentials not configured${NC}"
  echo "Run: aws configure"
  exit 1
fi

AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
AWS_USER=$(aws sts get-caller-identity --query Arn --output text | sed 's/.*\///')
echo -e "${GREEN}✓ AWS authenticated as ${AWS_USER} (Account: ${AWS_ACCOUNT})${NC}"
echo

# 3. Setup personas directory
echo -e "${YELLOW}Setting up personas directory...${NC}"
PERSONAS_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/personas"

if [ -d "$PERSONAS_DIR" ]; then
  echo -e "${YELLOW}Found existing: ${PERSONAS_DIR}${NC}"
  read -p "Overwrite? (y/n) " -n 1 -r < /dev/tty
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    [[ -z "$PERSONAS_DIR" ]] && { echo "PERSONAS_DIR is empty, aborting"; exit 1; }
    rm -rf "$PERSONAS_DIR"
  else
    echo "Using existing directory"
  fi
fi

mkdir -p "$PERSONAS_DIR"/{agent,code}
echo -e "${GREEN}✓ Created: ${PERSONAS_DIR}/{agent,code}${NC}"
echo

# 4. Create default persona
echo -e "${YELLOW}Creating default persona...${NC}"

cat > "$PERSONAS_DIR/agent/default.json" << 'EOF'
{
  "meta": {
    "name": "default",
    "description": "Neutral baseline persona",
    "version": "1.0.0",
    "created_at": "2026-03-26T00:00:00Z"
  },
  "tone": {
    "formality": 0.5,
    "directness": 0.5,
    "humor": 0.5,
    "hedging": 0.5,
    "warmth": 0.5
  },
  "verbosity": {
    "response_length": "adaptive",
    "reading_level": "professional",
    "format_preference": "adaptive",
    "bullet_threshold": "multi_step_only"
  },
  "voice": {
    "custom_phrases": [],
    "avoid_words": [],
    "prefer_words": [],
    "sign_off": null
  }
}
EOF

echo -e "${GREEN}✓ Default persona created${NC}"
echo -e "  Location: ${PERSONAS_DIR}/agent/default.json"
echo

# 5. Set active persona
echo -e "${YELLOW}Setting active persona...${NC}"
CONFIG_ROOT="${XDG_CONFIG_HOME:-$HOME/.config}/waldo"
mkdir -p "$CONFIG_ROOT"
printf '%s' "agent/default" > "$PERSONAS_DIR/.active"
printf '%s' "agent/default" > "$CONFIG_ROOT/.active"
echo -e "${GREEN}✓ Active persona: agent/default${NC}"
echo

# 6. Setup hook scripts
echo -e "${YELLOW}Setting up hook scripts...${NC}"

HOOKS_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/hooks/waldo"
mkdir -p "$HOOKS_DIR"

# Try to get from local repo first, then ensure all required hooks are present (download any missing)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_HOOKS="$SCRIPT_DIR/.claude/hooks/waldo"
REPO_URL="https://raw.githubusercontent.com/caboose-mcp/waldo/demo/dual-domain/.claude/hooks/waldo"

HOOKS=(
  "inject-persona.sh"
  "session-counter.sh"
  "s3-sync.sh"
  "accumulate-deltas.sh"
  "scan-code-style.sh"
  "fingerprint-cache.sh"
  "status-line.sh"
)

if [ -d "$REPO_HOOKS" ]; then
  echo "  Using local repo hooks..."
  cp "$REPO_HOOKS"/*.sh "$HOOKS_DIR/" 2>/dev/null || true
fi

echo "  Ensuring required hooks are installed..."
for hook in "${HOOKS[@]}"; do
  if [ ! -f "$HOOKS_DIR/$hook" ]; then
    if curl -fsSL "$REPO_URL/$hook" -o "$HOOKS_DIR/$hook" 2>/dev/null; then
      chmod +x "$HOOKS_DIR/$hook"
    fi
  fi
done

# Verify at least s3-sync exists (core dependency)
if [ -f "$HOOKS_DIR/s3-sync.sh" ]; then
  echo -e "  ${GREEN}✓${NC} Hook scripts installed"
else
  echo -e "  ${YELLOW}⚠${NC} Some hooks missing (manual setup needed)"
fi
echo

# 7. Install waldo skill
echo -e "${YELLOW}Installing /waldo skill...${NC}"

SKILLS_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills/waldo"
mkdir -p "$SKILLS_DIR"

SKILL_SRC="$SCRIPT_DIR/waldo-SKILL-v5.md"
SKILL_DEST="$SKILLS_DIR/SKILL.md"

if [ -f "$SKILL_SRC" ]; then
  # Symlink so edits to the repo are reflected immediately
  ln -sf "$SKILL_SRC" "$SKILL_DEST"
  echo -e "${GREEN}✓ Skill installed (symlinked): ${SKILL_DEST}${NC}"
else
  SKILL_URL="https://raw.githubusercontent.com/caboose-mcp/waldo/main/waldo-SKILL-v5.md"
  if curl -fsSL "$SKILL_URL" -o "$SKILL_DEST" 2>/dev/null; then
    echo -e "${GREEN}✓ Skill downloaded: ${SKILL_DEST}${NC}"
  else
    echo -e "${YELLOW}⚠ Could not install skill — copy manually:${NC}"
    echo "  cp <waldo-repo>/waldo-SKILL-v5.md ~/.claude/skills/waldo/SKILL.md"
  fi
fi
echo

# 8. S3 bucket menu
echo -e "${YELLOW}S3 Cross-Machine Sync (Optional)${NC}"
read -p "Setup S3 sync? (y/n) " -n 1 -r SETUP_S3 < /dev/tty
echo

if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
  echo -e "${BLUE}S3 Bucket Configuration${NC}"
  echo

  # Check existing buckets
  EXISTING=$(aws s3api list-buckets --query 'Buckets[].Name' --output text 2>/dev/null || echo "")

  if [ -n "$EXISTING" ]; then
    echo "Your S3 buckets:"
    # $EXISTING is intentionally unquoted to word-split bucket names into select options
    select BUCKET in $EXISTING "Create new bucket"; do
      if [ "$BUCKET" = "Create new bucket" ]; then
        read -rp "  New bucket name: " NEW_BUCKET < /dev/tty
        # AWS rules: lowercase alphanum/hyphens/dots, no consecutive dots, no dot-hyphen, no IP format
        if [[ ! "$NEW_BUCKET" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$ ]] \
           || [[ ${#NEW_BUCKET} -lt 3 || ${#NEW_BUCKET} -gt 63 ]] \
           || [[ "$NEW_BUCKET" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
          echo -e "  ${RED}✗ Invalid bucket name (lowercase alphanumeric, dots, hyphens only)${NC}"
          SETUP_S3=
          break
        fi
        if aws s3api create-bucket --bucket "$NEW_BUCKET" --region us-east-1 2>/dev/null; then
          BUCKET="$NEW_BUCKET"
          echo -e "  ${GREEN}✓ Bucket created: ${BUCKET}${NC}"
        else
          echo -e "  ${RED}✗ Failed to create bucket${NC}"
          SETUP_S3=
          break
        fi
      fi
      if [ -n "$BUCKET" ]; then
        break
      fi
    done < /dev/tty
  else
    read -rp "  New bucket name (e.g., my-personas): " BUCKET < /dev/tty
    # AWS rules: lowercase alphanum/hyphens/dots, no consecutive dots, no dot-hyphen, no IP format
    if [[ ! "$BUCKET" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$ ]] \
       || [[ ${#BUCKET} -lt 3 || ${#BUCKET} -gt 63 ]] \
       || [[ "$BUCKET" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo -e "  ${RED}✗ Invalid bucket name (lowercase alphanumeric, dots, hyphens only)${NC}"
      SETUP_S3=
    elif aws s3api create-bucket --bucket "$BUCKET" --region us-east-1 2>/dev/null; then
      echo -e "  ${GREEN}✓ Bucket created: ${BUCKET}${NC}"
    else
      echo -e "  ${RED}✗ Failed to create bucket${NC}"
      SETUP_S3=
    fi
  fi

  if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
    echo
    read -rp "  AWS profile (default: default): " AWS_PROFILE < /dev/tty
    AWS_PROFILE="${AWS_PROFILE:-default}"

    # Update settings.json
    SETTINGS_FILE="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"
    if [ -f "$SETTINGS_FILE" ]; then
      echo -e "${YELLOW}Updating ${SETTINGS_FILE}...${NC}"

      # Backup (immediately restrict permissions in case original was world-readable)
      BACKUP_FILE="${SETTINGS_FILE}.backup.$(date +%s)"
      cp "$SETTINGS_FILE" "$BACKUP_FILE"
      chmod 600 "$BACKUP_FILE"

      # Update env vars
      jq --arg profile "$AWS_PROFILE" --arg region "us-east-1" \
        '.env.AWS_PROFILE = $profile | .env.AWS_REGION = $region' \
        "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp" && mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"
      chmod 600 "$SETTINGS_FILE"

      echo -e "  ${GREEN}✓ env.AWS_PROFILE = ${AWS_PROFILE}${NC}"
      echo -e "  ${GREEN}✓ env.AWS_REGION = us-east-1${NC}"
    fi

    # Test S3 access
    echo -e "${YELLOW}Testing S3 access...${NC}"
    if aws --profile "$AWS_PROFILE" s3 ls "s3://$BUCKET" &>/dev/null; then
      echo -e "  ${GREEN}✓ Can access s3://${BUCKET}${NC}"
    else
      echo -e "  ${YELLOW}⚠ Cannot access bucket (check IAM permissions)${NC}"
    fi

    echo
    echo -e "${GREEN}✓ S3 sync configured${NC}"
    echo "  Bucket: s3://$BUCKET"
    echo "  Profile: $AWS_PROFILE"
  fi
  echo
fi

# 9. Initialize deltas
echo -e "${YELLOW}Initializing learning deltas...${NC}"
echo "[]" > "$PERSONAS_DIR/agent/.deltas"
echo "[]" > "$PERSONAS_DIR/code/.deltas"
echo -e "${GREEN}✓ Delta files initialized${NC}"
echo

# 10. Summary
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Setup Complete!                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo

echo -e "${GREEN}What's Ready:${NC}"
echo "  ✓ Personas directory: ${PERSONAS_DIR}"
echo "  ✓ Default persona active"
echo "  ✓ Hook scripts: ${HOOKS_DIR}"
echo "  ✓ /waldo skill: ${SKILLS_DIR}/SKILL.md"
if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
  echo "  ✓ S3 sync configured"
fi
echo

echo -e "${BLUE}Quick Start:${NC}"
echo "  1. Restart Claude Code (skills load at startup)"
echo
echo "  2. Create a persona:"
echo "     /waldo new my-voice"
echo
echo "  3. Switch persona:"
echo "     /waldo use agent/my-voice"
echo
echo "  4. Learn from conversation:"
echo "     /waldo learn"
echo
echo -e "${BLUE}Next:${NC}"
echo "  • Full docs: ${PERSONAS_DIR}/../../../dev/waldo/WALDO-SETUP.md"
echo "  • GitHub: https://github.com/caboose-mcp/waldo"
echo
