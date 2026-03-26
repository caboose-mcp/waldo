#!/bin/bash
# waldo — Cross-machine persona sync setup
# One-liner: curl -fsSL https://raw.githubusercontent.com/caboose-mcp/waldo/demo/dual-domain/setup-waldo.sh | bash

set -e

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
  read -p "Overwrite? (y/n) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
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

# Try to get from local repo first, then fallback to download
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_HOOKS="$SCRIPT_DIR/.claude/hooks/waldo"

if [ -d "$REPO_HOOKS" ]; then
  echo "  Using local repo hooks..."
  cp "$REPO_HOOKS"/*.sh "$HOOKS_DIR/" 2>/dev/null || true
else
  echo "  Downloading hooks from GitHub..."
  REPO_URL="https://raw.githubusercontent.com/caboose-mcp/waldo/demo/dual-domain/.claude/hooks/waldo"

  HOOKS=(
    "inject-persona.sh"
    "session-counter.sh"
    "s3-sync.sh"
    "accumulate-deltas.sh"
    "scan-code-style.sh"
    "fingerprint-cache.sh"
  )

  for hook in "${HOOKS[@]}"; do
    if curl -fsSL "$REPO_URL/$hook" -o "$HOOKS_DIR/$hook" 2>/dev/null; then
      chmod +x "$HOOKS_DIR/$hook"
    fi
  done
fi

# Verify at least s3-sync exists (core dependency)
if [ -f "$HOOKS_DIR/s3-sync.sh" ]; then
  echo -e "  ${GREEN}✓${NC} Hook scripts installed"
else
  echo -e "  ${YELLOW}⚠${NC} Some hooks missing (manual setup needed)"
fi
echo

# 7. S3 bucket menu
echo -e "${YELLOW}S3 Cross-Machine Sync (Optional)${NC}"
read -p "Setup S3 sync? (y/n) " -n 1 -r SETUP_S3
echo

if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
  echo -e "${BLUE}S3 Bucket Configuration${NC}"
  echo

  # Check existing buckets
  EXISTING=$(aws s3api list-buckets --query 'Buckets[].Name' --output text 2>/dev/null || echo "")

  if [ -n "$EXISTING" ]; then
    echo "Your S3 buckets:"
    select BUCKET in $EXISTING "Create new bucket"; do
      if [ "$BUCKET" = "Create new bucket" ]; then
        read -p "  New bucket name: " NEW_BUCKET
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
    done
  else
    read -p "  New bucket name (e.g., my-personas): " BUCKET
    if aws s3api create-bucket --bucket "$BUCKET" --region us-east-1 2>/dev/null; then
      echo -e "  ${GREEN}✓ Bucket created: ${BUCKET}${NC}"
    else
      echo -e "  ${RED}✗ Failed to create bucket${NC}"
      SETUP_S3=
    fi
  fi

  if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
    echo
    read -p "  AWS profile (default: default): " AWS_PROFILE
    AWS_PROFILE="${AWS_PROFILE:-default}"

    # Update settings.json
    SETTINGS_FILE="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"
    if [ -f "$SETTINGS_FILE" ]; then
      echo -e "${YELLOW}Updating ${SETTINGS_FILE}...${NC}"

      # Backup
      cp "$SETTINGS_FILE" "${SETTINGS_FILE}.backup.$(date +%s)"

      # Update env vars
      jq ".env.AWS_PROFILE = \"$AWS_PROFILE\" | .env.AWS_REGION = \"us-east-1\"" "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp" && mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"

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

# 8. Initialize deltas
echo -e "${YELLOW}Initializing learning deltas...${NC}"
echo "[]" > "$PERSONAS_DIR/agent/.deltas"
echo "[]" > "$PERSONAS_DIR/code/.deltas"
echo -e "${GREEN}✓ Delta files initialized${NC}"
echo

# 9. Summary
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Setup Complete!                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo

echo -e "${GREEN}What's Ready:${NC}"
echo "  ✓ Personas directory: ${PERSONAS_DIR}"
echo "  ✓ Default persona active"
echo "  ✓ Hook scripts: ${HOOKS_DIR}"
if [[ $SETUP_S3 =~ ^[Yy]$ ]]; then
  echo "  ✓ S3 sync configured"
fi
echo

echo -e "${BLUE}Quick Start:${NC}"
echo "  1. Import waldo skill:"
echo "     /waldo import"
echo "     (paste from: https://github.com/caboose-mcp/waldo)"
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
