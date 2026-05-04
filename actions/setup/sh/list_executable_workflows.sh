#!/usr/bin/env bash
set -euo pipefail

WORKFLOW_DIR="${1:-.github/workflows}"
OUTPUT_FILE="${2:-/tmp/gh-aw/agent/workflow-list.txt}"

mkdir -p "$(dirname "$OUTPUT_FILE")"

find "$WORKFLOW_DIR" -maxdepth 1 -type f -name '*.md' \
  ! -name 'smoke-*.md' \
  ! -name 'test-*.md' \
  ! -name 'example*.md' \
  | sort > "$OUTPUT_FILE"

echo "Inventory complete: $(wc -l < "$OUTPUT_FILE" | tr -d ' ') workflows found"
