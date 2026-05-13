---
name: Outcome Collector
description: Periodic evaluation of safe output outcomes to measure workflow value and acceptance rates
on:
  schedule:
    - cron: every 6 hours
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
  actions: read
  discussions: read
tracker-id: outcome-collector
engine:
  id: copilot
  model: claude-haiku-4.5
  bare: true
strict: true
timeout-minutes: 20
network:
  allowed:
    - defaults
    - github
tools:
  bash: true
  cache-memory: true
  github:
    mode: gh-proxy
    toolsets: [default]
safe-outputs:
  create-issue:
    title-prefix: "[Outcome Report]"
    labels: [automation, observability, outcomes]
    close-older-issues: true
    group-by-day: true
    expires: 7d
  noop:
  messages:
    footer: "> 📊 *Measured by [{workflow_name}]({run_url})*{effective_tokens_suffix}"
    run-started: "📊 [{workflow_name}]({run_url}) is evaluating safe output outcomes..."
    run-success: "📊 [{workflow_name}]({run_url}) outcome evaluation complete!"
    run-failure: "📊 [{workflow_name}]({run_url}) {status}"
imports:
  - shared/observability-otlp.md
pre-agent-steps:
  - name: Evaluate outcomes for recent runs
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    run: |
      echo "Evaluating safe output outcomes for recent workflow runs..."

      REPO="${GITHUB_REPOSITORY}"

      # Load previously evaluated run IDs from cache-memory to avoid re-processing
      SEEN_FILE="/tmp/gh-aw/cache-memory/outcome-collector/seen-runs.json"
      mkdir -p "$(dirname "$SEEN_FILE")"
      if [ ! -f "$SEEN_FILE" ]; then
        echo '[]' > "$SEEN_FILE"
      fi

      # Get recent successful workflow runs (wider window for better coverage)
      RUNS=$(gh run list --repo "$REPO" --limit 200 --json databaseId,conclusion,workflowName,event \
        --jq '[.[] | select(.conclusion == "success")] | .[0:150]' 2>/dev/null)

      if [ -z "$RUNS" ] || [ "$RUNS" = "[]" ] || [ "$RUNS" = "null" ]; then
        echo "No recent successful runs found"
        echo '{"runs_checked": 0, "total_outcomes": 0}' > /tmp/gh-aw/outcome-summary.json
        exit 0
      fi

      mkdir -p /tmp/gh-aw/outcomes

      CHECKED=0
      ACCEPTED=0
      REJECTED=0
      IGNORED=0
      PENDING=0
      TOTAL=0
      NOOP=0
      EVAL_JSONL="/tmp/gh-aw/outcome-evaluations.jsonl"
      > "$EVAL_JSONL"
      EVALUATED_IDS_FILE="${SEEN_FILE}.evaluated"
      echo '[]' > "$EVALUATED_IDS_FILE"

      for RUN_ID in $(echo "$RUNS" | jq -r '.[].databaseId'); do
        # Skip runs already evaluated in a previous collection pass
        if jq -e --argjson id "$RUN_ID" '. | index($id)' "$SEEN_FILE" > /dev/null 2>&1; then
          continue
        fi

        # Try to download safe-outputs-items artifact (skip runs without it)
        ITEM_DIR="/tmp/gh-aw/outcomes/run-${RUN_ID}"
        gh run download "$RUN_ID" --repo "$REPO" --name safe-outputs-items --dir "$ITEM_DIR" 2>/dev/null || continue

        MANIFEST="$ITEM_DIR/safe-output-items.jsonl"
        if [ ! -f "$MANIFEST" ]; then
          continue
        fi

        # Count all items and noops separately
        ALL_ITEMS=$(jq -r 'select(.type != null and .type != "") | .type' "$MANIFEST" 2>/dev/null | wc -l | tr -d ' ')
        RUN_NOOPS=$(jq -r 'select(.type == "noop" or .type == "missing_tool" or .type == "missing_data" or .type == "report_incomplete") | .type' "$MANIFEST" 2>/dev/null | wc -l | tr -d ' ')
        ITEMS=$((ALL_ITEMS - RUN_NOOPS))
        NOOP=$((NOOP + RUN_NOOPS))

        if [ "$ITEMS" = "0" ] && [ "$RUN_NOOPS" = "0" ]; then
          continue
        fi

        WF=$(echo "$RUNS" | jq -r ".[] | select(.databaseId == $RUN_ID) | .workflowName")
        EVENT=$(echo "$RUNS" | jq -r ".[] | select(.databaseId == $RUN_ID) | .event")
        echo "Run $RUN_ID ($WF): $ITEMS item(s), $RUN_NOOPS noop(s) [trigger: $EVENT]"

        CHECKED=$((CHECKED + 1))
        TOTAL=$((TOTAL + ITEMS))

        # Write noop entries to JSONL for noop-rate tracking
        if [ "$RUN_NOOPS" -gt 0 ]; then
          jq -c 'select(.type == "noop" or .type == "missing_tool" or .type == "missing_data" or .type == "report_incomplete")' "$MANIFEST" 2>/dev/null | while IFS= read -r noop_line; do
            NOOP_TYPE=$(echo "$noop_line" | jq -r '.type // empty')
            jq -n -c \
              --arg type "$NOOP_TYPE" \
              --arg url "" \
              --arg repo "$REPO" \
              --arg result "noop" \
              --arg detail "$NOOP_TYPE" \
              --arg workflow "$WF" \
              --argjson run_id "$RUN_ID" \
              --arg timestamp "" \
              --arg event "$EVENT" \
              '{type: $type, url: $url, repo: $repo, result: $result, detail: $detail, workflow: $workflow, run_id: $run_id, timestamp: $timestamp, event: $event}' \
              >> "$EVAL_JSONL"
          done
        fi

        if [ "$ITEMS" = "0" ]; then
          # Only noops in this run — still track it but skip actionable evaluation
          jq -n --arg wf "$WF" --argjson items 0 --argjson run_id "$RUN_ID" --argjson noops "$RUN_NOOPS" --arg event "$EVENT" \
            '{workflow: $wf, run_id: $run_id, items: $items, noops: $noops, event: $event}' \
            > "/tmp/gh-aw/outcomes/run-${RUN_ID}.json"
          jq --argjson id "$RUN_ID" '. + [$id]' "$EVALUATED_IDS_FILE" > "${EVALUATED_IDS_FILE}.tmp" \
            && mv "${EVALUATED_IDS_FILE}.tmp" "$EVALUATED_IDS_FILE"
          continue
        fi

        # Basic outcome evaluation per item using GitHub API
        while IFS= read -r line; do
          TYPE=$(echo "$line" | jq -r '.type // empty')
          case "$TYPE" in
            ""|noop|missing_tool|missing_data|report_incomplete) continue ;;
          esac

          URL=$(echo "$line" | jq -r '.url // empty')
          ITEM_REPO=$(echo "$line" | jq -r '.repo // empty')
          TIMESTAMP=$(echo "$line" | jq -r '.timestamp // empty')
          [ -z "$ITEM_REPO" ] && ITEM_REPO="$REPO"

          RESULT="pending"
          DETAIL=""

          RESOLUTION_SEC=""
          REVIEW_COMMENTS=""
          CHANGED_FILES=""
          ADDITIONS=""
          DELETIONS=""

          if [ -z "$URL" ]; then
            RESULT="pending"
            DETAIL="no url"
            PENDING=$((PENDING + 1))
          elif echo "$URL" | grep -qE '/issues/[0-9]+|/issuecomment-'; then
            NUM=$(echo "$URL" | grep -oE '/(issues|pull)/[0-9]+' | grep -oE '[0-9]+' | head -1)
            if [ -n "$NUM" ]; then
              ISSUE_JSON=$(gh api "repos/$ITEM_REPO/issues/$NUM" 2>/dev/null || echo "")
              STATE=$(echo "$ISSUE_JSON" | jq -r '.state // empty' 2>/dev/null)
              if [ "$STATE" = "open" ] || [ "$STATE" = "closed" ]; then
                RESULT="accepted"
                DETAIL="$STATE"
                ACCEPTED=$((ACCEPTED + 1))
                # Time-to-resolution: created_at → closed_at
                if [ "$STATE" = "closed" ]; then
                  CREATED_AT=$(echo "$ISSUE_JSON" | jq -r '.created_at // empty' 2>/dev/null)
                  CLOSED_AT=$(echo "$ISSUE_JSON" | jq -r '.closed_at // empty' 2>/dev/null)
                  if [ -n "$CREATED_AT" ] && [ -n "$CLOSED_AT" ]; then
                    RESOLUTION_SEC=$(( $(date -d "$CLOSED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$CLOSED_AT" +%s 2>/dev/null || echo 0) - $(date -d "$CREATED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$CREATED_AT" +%s 2>/dev/null || echo 0) ))
                  fi
                fi
              else
                RESULT="pending"
                DETAIL="api error"
                PENDING=$((PENDING + 1))
              fi
            else
              RESULT="pending"
              DETAIL="no number"
              PENDING=$((PENDING + 1))
            fi
          elif echo "$URL" | grep -qE '/pull/[0-9]+'; then
            NUM=$(echo "$URL" | grep -oE '/pull/[0-9]+' | grep -oE '[0-9]+')
            if [ -n "$NUM" ]; then
              PR_JSON=$(gh api "repos/$ITEM_REPO/pulls/$NUM" 2>/dev/null || echo "")
              MERGED=$(echo "$PR_JSON" | jq -r '.merged // empty' 2>/dev/null)
              STATE=$(echo "$PR_JSON" | jq -r '.state // empty' 2>/dev/null)
              REVIEW_COMMENTS=$(echo "$PR_JSON" | jq -r '.review_comments // 0' 2>/dev/null)
              CHANGED_FILES=$(echo "$PR_JSON" | jq -r '.changed_files // 0' 2>/dev/null)
              ADDITIONS=$(echo "$PR_JSON" | jq -r '.additions // 0' 2>/dev/null)
              DELETIONS=$(echo "$PR_JSON" | jq -r '.deletions // 0' 2>/dev/null)
              if [ "$MERGED" = "true" ]; then
                RESULT="accepted"
                DETAIL="merged"
                ACCEPTED=$((ACCEPTED + 1))
                # Time-to-resolution: created_at → merged_at
                CREATED_AT=$(echo "$PR_JSON" | jq -r '.created_at // empty' 2>/dev/null)
                MERGED_AT=$(echo "$PR_JSON" | jq -r '.merged_at // empty' 2>/dev/null)
                if [ -n "$CREATED_AT" ] && [ -n "$MERGED_AT" ]; then
                  RESOLUTION_SEC=$(( $(date -d "$MERGED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$MERGED_AT" +%s 2>/dev/null || echo 0) - $(date -d "$CREATED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$CREATED_AT" +%s 2>/dev/null || echo 0) ))
                fi
              elif [ "$STATE" = "closed" ]; then
                RESULT="rejected"
                DETAIL="closed"
                REJECTED=$((REJECTED + 1))
                CREATED_AT=$(echo "$PR_JSON" | jq -r '.created_at // empty' 2>/dev/null)
                CLOSED_AT=$(echo "$PR_JSON" | jq -r '.closed_at // empty' 2>/dev/null)
                if [ -n "$CREATED_AT" ] && [ -n "$CLOSED_AT" ]; then
                  RESOLUTION_SEC=$(( $(date -d "$CLOSED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$CLOSED_AT" +%s 2>/dev/null || echo 0) - $(date -d "$CREATED_AT" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$CREATED_AT" +%s 2>/dev/null || echo 0) ))
                fi
              elif [ "$STATE" = "open" ]; then
                RESULT="pending"
                DETAIL="open"
                PENDING=$((PENDING + 1))
              else
                RESULT="pending"
                DETAIL="api error"
                PENDING=$((PENDING + 1))
              fi
            else
              RESULT="pending"
              DETAIL="no number"
              PENDING=$((PENDING + 1))
            fi
          else
            # Comments, labels, etc. — if URL exists, the item was created
            RESULT="accepted"
            DETAIL="object exists"
            ACCEPTED=$((ACCEPTED + 1))
          fi

          # Compute pending age in seconds for unresolved items
          PENDING_AGE_SEC=""
          if [ "$RESULT" = "pending" ] && [ -n "$TIMESTAMP" ]; then
            NOW_EPOCH=$(date +%s)
            ITEM_EPOCH=$(date -d "$TIMESTAMP" +%s 2>/dev/null || date -jf "%Y-%m-%dT%H:%M:%SZ" "$TIMESTAMP" +%s 2>/dev/null || echo "")
            if [ -n "$ITEM_EPOCH" ] && [ "$ITEM_EPOCH" != "0" ]; then
              PENDING_AGE_SEC=$((NOW_EPOCH - ITEM_EPOCH))
            fi
          fi

          # Write per-item evaluation for OTEL export
          jq -n -c \
            --arg type "$TYPE" \
            --arg url "$URL" \
            --arg repo "$ITEM_REPO" \
            --arg result "$RESULT" \
            --arg detail "$DETAIL" \
            --arg workflow "$WF" \
            --argjson run_id "$RUN_ID" \
            --arg timestamp "$TIMESTAMP" \
            --arg event "$EVENT" \
            --arg resolution_sec "${RESOLUTION_SEC:-}" \
            --arg pending_age_sec "${PENDING_AGE_SEC:-}" \
            --arg review_comments "${REVIEW_COMMENTS:-}" \
            --arg changed_files "${CHANGED_FILES:-}" \
            --arg additions "${ADDITIONS:-}" \
            --arg deletions "${DELETIONS:-}" \
            '{type: $type, url: $url, repo: $repo, result: $result, detail: $detail, workflow: $workflow, run_id: $run_id, timestamp: $timestamp, event: $event, resolution_sec: (if $resolution_sec != "" then ($resolution_sec | tonumber) else null end), pending_age_sec: (if $pending_age_sec != "" then ($pending_age_sec | tonumber) else null end), review_comments: (if $review_comments != "" then ($review_comments | tonumber) else null end), changed_files: (if $changed_files != "" then ($changed_files | tonumber) else null end), additions: (if $additions != "" then ($additions | tonumber) else null end), deletions: (if $deletions != "" then ($deletions | tonumber) else null end)}' \
            >> "$EVAL_JSONL"
        done < "$MANIFEST"

        # Save per-run data
        jq -n --arg wf "$WF" --argjson items "$ITEMS" --argjson run_id "$RUN_ID" --argjson noops "$RUN_NOOPS" --arg event "$EVENT" \
          '{workflow: $wf, run_id: $run_id, items: $items, noops: $noops, event: $event}' \
          > "/tmp/gh-aw/outcomes/run-${RUN_ID}.json"

        # Track this run as successfully evaluated
        jq --argjson id "$RUN_ID" '. + [$id]' "$EVALUATED_IDS_FILE" > "${EVALUATED_IDS_FILE}.tmp" \
          && mv "${EVALUATED_IDS_FILE}.tmp" "$EVALUATED_IDS_FILE"
      done

      # Compute fleet summary
      RESOLVED=$((ACCEPTED + REJECTED))
      if [ "$RESOLVED" -gt 0 ]; then
        ACCEPTANCE_RATE=$(echo "scale=4; $ACCEPTED / $RESOLVED" | bc)
      else
        ACCEPTANCE_RATE="0"
      fi
      if [ "$TOTAL" -gt 0 ]; then
        WASTE_RATE=$(echo "scale=4; $REJECTED / $TOTAL" | bc)
      else
        WASTE_RATE="0"
      fi

      if [ "$TOTAL" -gt 0 ]; then
        NOOP_RATE=$(echo "scale=4; $NOOP / ($TOTAL + $NOOP)" | bc)
      else
        NOOP_RATE="0"
      fi

      jq -n \
        --argjson checked "$CHECKED" \
        --argjson total "$TOTAL" \
        --argjson accepted "$ACCEPTED" \
        --argjson rejected "$REJECTED" \
        --argjson ignored "$IGNORED" \
        --argjson pending "$PENDING" \
        --argjson noop "$NOOP" \
        --arg acceptance_rate "$ACCEPTANCE_RATE" \
        --arg waste_rate "$WASTE_RATE" \
        --arg noop_rate "$NOOP_RATE" \
        '{
          runs_checked: $checked,
          total_outcomes: $total,
          accepted: $accepted,
          rejected: $rejected,
          ignored: $ignored,
          pending: $pending,
          noop: $noop,
          acceptance_rate: ($acceptance_rate | tonumber),
          waste_rate: ($waste_rate | tonumber),
          noop_rate: ($noop_rate | tonumber),
          date: (now | strftime("%Y-%m-%d"))
        }' > /tmp/gh-aw/outcome-summary.json

      # Update seen-runs cache so subsequent passes skip these runs.
      # Keep only the last 500 run IDs to prevent unbounded growth.
      jq -s '.[0] + .[1] | unique | .[-500:]' "$SEEN_FILE" "$EVALUATED_IDS_FILE" > "${SEEN_FILE}.tmp" \
        && mv "${SEEN_FILE}.tmp" "$SEEN_FILE"
      rm -f "$EVALUATED_IDS_FILE"

      echo "✓ Checked $CHECKED runs, $TOTAL outcomes"
      echo "  Accepted: $ACCEPTED, Rejected: $REJECTED, Ignored: $IGNORED, Pending: $PENDING"
      echo "  Acceptance rate: $ACCEPTANCE_RATE"
      cat /tmp/gh-aw/outcome-summary.json
  - name: Export outcome telemetry
    run: |
      if [ -f /tmp/gh-aw/outcome-evaluations.jsonl ] && [ -s /tmp/gh-aw/outcome-evaluations.jsonl ]; then
        node "${RUNNER_TEMP}/gh-aw/actions/emit_outcome_spans.cjs"
      else
        echo "No outcome evaluations to export"
      fi
---

# Outcome Collector

You are the Outcome Collector. Your job is to create a concise report of safe output outcomes.

## Input

The pre-agent step has already evaluated outcomes for recent workflow runs. Results are in:

- `/tmp/gh-aw/outcome-summary.json` — fleet-wide summary
- `/tmp/gh-aw/outcomes/run-*.json` — per-run outcome details

## Task

1. Read `/tmp/gh-aw/outcome-summary.json`
2. If `total_outcomes` is 0, call `noop` with "No new safe output outcomes to report"
3. Otherwise, create a report issue with the summary

## Report Format

Create an issue with this structure:

```markdown
## Safe Output Outcomes — {date}

### Fleet Summary

| Metric | Value |
|--------|-------|
| Runs checked | {runs_checked} |
| Total outcomes | {total_outcomes} |
| Accepted | {accepted} |
| Rejected | {rejected} |
| Ignored | {ignored} |
| Pending | {pending} |
| **Acceptance rate** | **{acceptance_rate}%** |
| Waste rate | {waste_rate}% |

### Per-Workflow Breakdown

For each workflow with outcomes, show:
- Workflow name
- Outcomes: accepted / rejected / ignored
- Acceptance rate

### Key Observations

- Which workflows have the highest acceptance rate?
- Which workflows have the highest waste rate?
- Any workflows with all outcomes ignored (noise signal)?
```

## Guidelines

- Keep the report factual — numbers only, no speculation
- Do not re-evaluate outcomes — use the pre-computed data
- If no outcomes exist, use `noop`
- Stop immediately after creating the issue
