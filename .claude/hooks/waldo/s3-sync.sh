#!/bin/bash
# waldo s3-sync.sh — Cross-machine persona sync via AWS S3
#
# Pushes/pulls ~/.claude/personas/ to/from an S3 bucket.
# Excludes .deltas and .cache (local-only; never synced).
#
# Usage:
#   s3-sync.sh push [bucket] [profile]   # Upload local personas to S3
#   s3-sync.sh pull [bucket] [profile]   # Download personas from S3
#
# Args override env vars. Env vars:
#   WALDO_S3_BUCKET   — bucket name (required if not passed as arg)
#   AWS_PROFILE       — AWS CLI profile (default: "default")
#   AWS_REGION        — AWS region (default: "us-east-1")
#
# Called automatically by Claude Code hooks:
#   SessionStart  → pull (populates latest personas at session open)
#   After /waldo learn --accumulate → push (propagates updated persona)
#
# Designed to be fire-and-forget: all errors are logged and the script
# always exits 0 so hook failures never block a Claude Code session.

set -euo pipefail

LOG_FILE="${WALDO_LOG:-$HOME/.waldo/s3-sync.log}"
PERSONAS_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/personas"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# ---------------------------------------------------------------------------
# Logging (append-only, capped at 500 lines to prevent unbounded growth)
# ---------------------------------------------------------------------------

log() {
  mkdir -p "$(dirname "$LOG_FILE")"
  printf '%s [waldo/s3-sync] %s\n' "$TIMESTAMP" "$*" >> "$LOG_FILE"
  # Trim to last 500 lines (in-place, POSIX-safe)
  if [[ -f "$LOG_FILE" ]]; then
    local line_count
    line_count=$(wc -l < "$LOG_FILE")
    if (( line_count > 500 )); then
      tail -n 500 "$LOG_FILE" > "${LOG_FILE}.tmp" && mv "${LOG_FILE}.tmp" "$LOG_FILE"
    fi
  fi
}

bail() {
  log "ERROR: $*"
  exit 0  # Always exit 0 — sync failures must not block Claude Code sessions
}

# ---------------------------------------------------------------------------
# Parse args
# ---------------------------------------------------------------------------

MODE="${1:-}"
if [[ -z "$MODE" ]]; then
  echo "Usage: s3-sync.sh push|pull [bucket] [profile]" >&2
  exit 1
fi

# Args take precedence over env vars
BUCKET="${2:-${WALDO_S3_BUCKET:-}}"
PROFILE="${3:-${AWS_PROFILE:-default}}"
REGION="${AWS_REGION:-us-east-1}"

# ---------------------------------------------------------------------------
# Guardrails: AWS must be configured before we attempt anything
# ---------------------------------------------------------------------------

# 1. aws CLI must be present
if ! command -v aws &>/dev/null; then
  bail "aws CLI not found — skipping sync (install from https://aws.amazon.com/cli/)"
fi

# 2. Bucket must be known
if [[ -z "$BUCKET" ]]; then
  bail "No S3 bucket configured. Set WALDO_S3_BUCKET in ~/.claude/settings.json env or pass as arg."
fi

# 3. AWS credentials must be valid — quick, cheap check (avoids confusing S3 errors)
if ! aws --profile "$PROFILE" sts get-caller-identity &>/dev/null 2>&1; then
  bail "AWS credentials not valid for profile '$PROFILE' — skipping sync. Run: aws --profile '$PROFILE' configure"
fi

# 4. Personas directory must exist locally
if [[ ! -d "$PERSONAS_DIR" ]]; then
  bail "Personas directory not found: $PERSONAS_DIR — run setup-waldo.sh first"
fi

# ---------------------------------------------------------------------------
# Build exclude list (files never synced to/from S3)
# ---------------------------------------------------------------------------

S3_PREFIX="personas"
S3_URI="s3://${BUCKET}/${S3_PREFIX}"

EXCLUDES=(
  "--exclude" ".deltas"
  "--exclude" ".cache/*"
  "--exclude" "*.backup.*"
  "--exclude" "*.tmp"
)

# ---------------------------------------------------------------------------
# Sync operations
# ---------------------------------------------------------------------------

do_push() {
  log "push → $S3_URI (profile: $PROFILE)"

  local output
  if output=$(
    aws --profile "$PROFILE" --region "$REGION" \
      s3 sync "$PERSONAS_DIR" "$S3_URI" \
      "${EXCLUDES[@]}" \
      --delete \
      --quiet \
      2>&1
  ); then
    log "push OK"
  else
    bail "push failed: $output"
  fi
}

do_pull() {
  log "pull ← $S3_URI (profile: $PROFILE)"

  # Verify the bucket/prefix is reachable before wiping any local state
  if ! aws --profile "$PROFILE" --region "$REGION" \
      s3 ls "$S3_URI/" &>/dev/null 2>&1; then
    bail "Cannot reach $S3_URI — bucket may be empty or missing. Skipping pull."
  fi

  local output
  if output=$(
    aws --profile "$PROFILE" --region "$REGION" \
      s3 sync "$S3_URI" "$PERSONAS_DIR" \
      "${EXCLUDES[@]}" \
      --quiet \
      2>&1
  ); then
    log "pull OK"
  else
    bail "pull failed: $output"
  fi
}

# ---------------------------------------------------------------------------
# Dispatch
# ---------------------------------------------------------------------------

case "$MODE" in
  push) do_push ;;
  pull) do_pull ;;
  *)
    echo "Unknown mode: $MODE (expected push or pull)" >&2
    exit 1
    ;;
esac
