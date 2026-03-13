#!/bin/bash
set -e

PROXY_URL="https://cc-workshop-proxy.grafana.fun"
PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
SETTINGS_FILE="$HOME/.claude/settings.json"

# Check if already set up
if [ -f "$SETTINGS_FILE" ] && [ -f ~/.claude.json ]; then
  echo "Setup has already been run. To re-run, remove ~/.claude/ and ~/.claude.json first."
  echo "You can run: claude"
  exit 0
fi

# Check if claude is installed
if ! command -v claude &> /dev/null; then
  echo "Error: claude is not installed."
  echo "This should not happen with the devcontainer setup."
  exit 1
fi

# Check dependencies
if ! command -v jq &> /dev/null; then
  echo "Error: jq is required but not installed."
  exit 1
fi

if ! command -v go &> /dev/null; then
  echo "Error: go is required but not installed."
  exit 1
fi

# Install mage
if ! command -v mage &> /dev/null; then
  echo "Installing mage..."
  go install github.com/magefile/mage@latest
fi

# Prompt for password with retry
MAX_ATTEMPTS=3
ATTEMPT=0
KEY=""

while [ -z "$KEY" ] && [ "$ATTEMPT" -lt "$MAX_ATTEMPTS" ]; do
  ATTEMPT=$((ATTEMPT + 1))
  read -s -p "Enter workshop password: " PASSWORD
  echo

  RESPONSE=$(curl -s -X POST "$PROXY_URL/workshop/key" -d "{\"password\":\"$PASSWORD\"}")
  KEY=$(echo "$RESPONSE" | jq -r '.key // empty')

  if [ -z "$KEY" ]; then
    REMAINING=$((MAX_ATTEMPTS - ATTEMPT))
    if [ "$REMAINING" -gt 0 ]; then
      echo "Invalid password. $REMAINING attempt(s) remaining."
    else
      echo "Too many failed attempts."
      exit 1
    fi
  fi
done

# Write ~/.claude/settings.json with base URL and API key
rm -rf "$HOME/.claude"
mkdir -p "$HOME/.claude"
jq -n \
  --arg base_url "$PROXY_URL" \
  --arg key "$KEY" \
  '{
    env: {
      ANTHROPIC_BASE_URL: $base_url,
      ANTHROPIC_API_KEY: $key
    }
  }' > "$SETTINGS_FILE"

# Write ~/.claude.json with onboarding completed and project trusted
KEY_SUFFIX="${KEY: -20}"
jq -n \
  --arg dir "$PROJECT_DIR" \
  --arg suffix "$KEY_SUFFIX" \
  '{
    hasCompletedOnboarding: true,
    customApiKeyResponses: {
      approved: [$suffix],
      rejected: []
    },
    projects: {
      ($dir): {
        hasTrustDialogAccepted: true
      }
    }
  }' > ~/.claude.json

echo "Done! You can now run: claude"
