---
description: Daily report of token consumption across all agentic workflows using OTel telemetry stored in Sentry, with fallback to pre-downloaded workflow run logs
on:
  schedule: daily on weekdays
permissions:
  actions: read
  contents: read
  issues: read
  pull-requests: read
tracker-id: daily-token-consumption-report
engine: claude
strict: true
tools:
  bash: true
safe-outputs:
  mentions: false
  allowed-github-references: []
  create-issue:
    title-prefix: "[token-consumption] "
    labels: [automation, observability, telemetry]
    close-older-issues: true
    expires: 1d
    max: 1
timeout-minutes: 45
imports:
  - shared/mcp/sentry.md
  - shared/aw-logs-24h-fetch.md
  - uses: shared/daily-audit-base.md
    with:
      title-prefix: "[token-consumption] "
      expires: 1d
---

{{#runtime-import? .github/shared-instructions.md}}

# Daily Token Consumption Report (Sentry OTel)

You are an observability analyst. Generate a daily token consumption report across all agentic workflows in this repository. Use Sentry OTel telemetry as the primary data source. If Sentry is unavailable (HTTP 403 or missing tools), fall back to pre-downloaded workflow run logs.

## Context

- Repository: `${{ github.repository }}`
- Run ID: `${{ github.run_id }}`
- Time Window: last 24 hours

## Mission

1. Collect token usage data for the last 24 hours (Sentry first, then log fallback).
2. Aggregate token usage by workflow.
3. Identify top token consumers and anomalous usage.
4. Publish a concise daily GitHub issue report.

## Data Collection

### Primary Source: Sentry OTel

#### Step 1: Discover Sentry Context

1. Call `find_organizations` and select the org for this repository.
2. Call `find_projects` and select the project that corresponds to `${{ github.repository }}`.

If either call returns HTTP 403, skip to the **Fallback Data Source** section below.

#### Step 2: Fetch Telemetry Events

Call `search_events` using:
- `dataset: spans`
- query constrained to the selected project
- time range: last 24 hours
- include enough results to cover the day (use pagination as needed)

If `dataset: spans` returns no usable records, retry with `dataset: transactions`.

Treat "no usable records" as either:
- zero events returned after pagination, or
- events returned but none contain any recognized token fields.

#### Step 3: Extract Workflow + Token Fields

For each event/span, derive:

- **Workflow name** using first non-empty of likely fields:
  - `github.workflow`
  - `github.workflow_ref`
  - `workflow.name`
  - `gh_aw.workflow`
  - fallback: `"unknown-workflow"`
- **Run ID** using:
  - `github.run_id`
  - `gh_aw.run_id`
- **Token counts** with precedence to avoid double counting:
  - Prefer explicit totals first: `ai.total_tokens` → `gen_ai.usage.total_tokens` → `usage.total_tokens` → `total_tokens`.
  - For input tokens: `ai.input_tokens` → `gen_ai.usage.input_tokens` → `usage.input_tokens` → `prompt_tokens`.
  - For output tokens: `ai.output_tokens` → `gen_ai.usage.output_tokens` → `usage.output_tokens` → `completion_tokens`.
  - If only total is present and input/output are missing, keep input/output at `0` and use total.
  - If input and output are present but total is missing, set total = input + output.
  - Do not sum overlapping aliases for the same token type.
- Recognized token fields:
  - `ai.input_tokens`, `ai.output_tokens`, `ai.total_tokens`
  - `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.usage.total_tokens`
  - `usage.input_tokens`, `usage.output_tokens`, `usage.total_tokens`
  - `prompt_tokens`, `completion_tokens`, `total_tokens`

Normalize missing values to `0`.

### Fallback Data Source: Pre-Downloaded Workflow Logs

Use this fallback when Sentry is unavailable (HTTP 403, missing `search_events` tool, or zero usable records after exhausting Sentry options).

Workflow run logs from the last 24 hours have been pre-downloaded to `/tmp/gh-aw/aw-mcp/logs/`.

#### Collect Token Data from Logs

For each run directory under `/tmp/gh-aw/aw-mcp/logs/`:
1. Read `aw_info.json` for workflow metadata (`workflow_name`, `run_id`).
2. Read `agent_usage.json` for token counts (`input_tokens`, `output_tokens`, `cache_read_tokens`, `cache_write_tokens`).

Use bash to aggregate:

```bash
for dir in /tmp/gh-aw/aw-mcp/logs/run-*/; do
  aw_info="$dir/aw_info.json"
  usage="$dir/agent_usage.json"
  if [ -f "$aw_info" ] && [ -f "$usage" ]; then
    workflow=$(jq -r '.workflow_name // "unknown-workflow"' "$aw_info")
    run_id=$(jq -r '.run_id // "unknown"' "$aw_info")
    input=$(jq -r '.input_tokens // 0' "$usage")
    output=$(jq -r '.output_tokens // 0' "$usage")
    cache_read=$(jq -r '.cache_read_tokens // 0' "$usage")
    echo "{\"workflow\":\"$workflow\",\"run_id\":\"$run_id\",\"input\":$input,\"output\":$output,\"cache_read\":$cache_read}"
  fi
done
```

Total tokens = `input_tokens + output_tokens` (do not add cache tokens to avoid double counting). Note in the report that the data source is local workflow run logs rather than Sentry.

## Analysis Requirements

Calculate:

- `total_events_analyzed`
- `events_with_token_data`
- `events_missing_workflow`
- `total_input_tokens`
- `total_output_tokens`
- `total_tokens`
- `workflow_count` (unique workflows)
- `top_workflows_by_tokens` (top 10)
- `avg_tokens_per_event`
- `p95_tokens_per_event`

For each workflow include:
- workflow name
- event count
- input tokens
- output tokens
- total tokens
- average tokens/event
- highest-token event (with run id if available)

## Report Output

Create exactly one issue titled:

`[token-consumption] Daily Token Consumption Report - YYYY-MM-DD`

Use this body structure:

### Executive Summary
- Total tokens, workflow count, and high-level trend notes.

### Key Metrics
| Metric | Value |
|---|---|
| Events analyzed | ... |
| Events with token data | ... |
| Total input tokens | ... |
| Total output tokens | ... |
| Total tokens | ... |
| Unique workflows | ... |
| Avg tokens/event | ... |
| P95 tokens/event | ... |

### Top 10 Workflows by Token Consumption
| Workflow | Events | Input Tokens | Output Tokens | Total Tokens | Avg/Event |
|---|---:|---:|---:|---:|---:|
| ... |

<details>
<summary>Data Quality and Gaps</summary>

- Data source used: Sentry OTel or local workflow run logs (fallback)
- If fallback was used, explain why (e.g., Sentry HTTP 403 - token missing `org:read` scope)
- Events missing workflow identifiers
- Events missing token attributes
- Any assumptions or fallback fields used

</details>

### Recommendations
- 2-4 concrete actions to reduce token usage for the highest consumers.
- If Sentry was unavailable due to a 403 error, recommend granting the `SENTRY_ACCESS_TOKEN` the `org:read` scope to restore full telemetry.

### References
- Include up to three relevant links (Sentry query links and/or run links when available).
- If fallback data source was used, note that Sentry was unavailable.

## Guardrails

- Be explicit when telemetry fields are absent or ambiguous.
- Never invent token values.
- Keep the report concise and actionable.
- Use `###` or lower headers only.

## Completion Requirement

You must call one safe output tool before finishing:
- `create_issue` for normal reporting (use even when falling back to local log data).
- `noop` only if neither Sentry nor local workflow run logs contain any valid token data.

{{#runtime-import shared/noop-reminder.md}}
