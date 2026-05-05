---
name: Daily Testify Uber Super Expert
description: Daily expert that analyzes one test file and creates an issue with testify-based improvements
on:
  schedule: daily
  workflow_dispatch:
  skip-if-match: 'is:issue is:open in:title "[testify-expert]"'

permissions:
  contents: read
  issues: read
  pull-requests: read

tracker-id: daily-testify-uber-super-expert
engine: copilot

imports:
  - uses: shared/daily-issue-base.md
    with:
      title-prefix: "[testify-expert] "
      expires: "2d"
      labels: [testing, code-quality, automated-analysis, cookie]
  - shared/go-source-analysis.md
  - shared/safe-output-app.md
  - shared/observability-otlp.md

tools:
  cli-proxy: true
  repo-memory:
    branch-name: memory/testify-expert
    description: "Tracks processed test files to avoid duplicates"
    file-glob: ["*.json", "*.txt"]
    max-file-size: 51200  # 50KB
  github:
    mode: gh-proxy
    toolsets: [default]
  bash:
    - "find . -name '*_test.go' -type f"
    - "cat **/*_test.go"
    - "grep -r 'func Test' . --include='*_test.go'"
    - "wc -l **/*_test.go"

timeout-minutes: 20
strict: true
features:
  copilot-requests: true
  inline-agents: true
---

{{#runtime-import? .github/shared-instructions.md}}

# Daily Testify Uber Super Expert 🧪✨

Analyze one Go test file per run and create a focused issue with actionable testify improvements.

## Workflow

1. **Load cache** from `/tmp/gh-aw/repo-memory/default/memory/testify-expert/processed_files.txt`.
   - Format per line: `path|YYYY-MM-DD`
   - If missing, treat as first run.

2. **Select one target file**:
   - List all `*_test.go` files.
   - Prefer files not seen in the last 30 days.
   - If all files were processed recently, stop and output:
     - `✅ All test files have been analyzed in the last 30 days.`

3. **Run sub-agent analysis**:
   - Invoke `test-analyzer` with the selected test file path.
   - If source file exists (replace `_test.go` with `.go`), include it in context.
   - Use the returned JSON as the source of truth for recommendations.

4. **Create one issue** using `create_issue`:
   - Title: `Improve Test Quality: <test file path>`
   - Include:
     - Current state (file, source file, test count, LOC)
     - 2-3 strengths
     - Prioritized improvements with concrete examples
     - Missing coverage opportunities
     - Acceptance checklist
   - Keep content specific to the selected file.

5. **Update cache**:
   - Append `<selected file>|<today>`
   - Deduplicate by file path, keeping newest date.

## Output

If analyzed:
- Selected file
- Issue number and title
- Count of strengths and improvement areas
- Confirmation cache was updated

If skipped:
- Total test files counted
- Cache location

## Guardrails

- Analyze exactly one file per run.
- Do not run full test suite; this workflow is read-only analysis.
- Prefer specific examples over generic advice.
- Follow repository testing guidance in `scratchpad/testing.md`.

{{#runtime-import shared/noop-reminder.md}}

## agent: `test-analyzer`
---
model: claude-haiku-4.5
description: Performs structured testify-focused analysis for one Go test file and returns JSON
---
You are a Go test-quality analyzer.

Input:
- `test_file`: path to one `*_test.go` file
- `source_file`: optional path to matching `.go` file

Read the files and return only JSON with this shape:
```json
{
  "test_file": "",
  "source_file": "",
  "test_function_count": 0,
  "loc": 0,
  "strengths": ["", ""],
  "improvements": [
    {
      "title": "",
      "priority": "high|medium|low",
      "reason": "",
      "examples": ["", ""]
    }
  ],
  "coverage_gaps": [""],
  "recommended_tests": [""],
  "acceptance_criteria": [""]
}
```

Rules:
- Focus on testify usage, table-driven structure, and coverage gaps.
- Provide concrete, file-specific examples.
- Keep `strengths` to 2-3 items.
- Keep `improvements` to 3-5 items.
