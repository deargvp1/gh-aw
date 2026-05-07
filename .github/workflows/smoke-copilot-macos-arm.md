---
description: Smoke Copilot macOS ARM64
on:
  workflow_dispatch:
  pull_request:
    types: [labeled]
    names: ["water"]
  reaction: "eyes"
  status-comment: true
permissions:
  contents: read
  pull-requests: read
  issues: read
  discussions: read
  actions: read
name: Smoke macOS ARM64
engine: copilot
runs-on: macos-latest
imports:
  - shared/gh.md
  - shared/reporting-otlp.md
  - shared/github-queries-mcp-script.md
network:
  allowed:
    - defaults
    - node
    - github
tools:
  cli-proxy: true
  agentic-workflows:
  cache-memory: true
  edit:
  bash:
    - "*"
  github:
  web-fetch:
runtimes:
  go:
    version: "1.25"
safe-outputs:
    allowed-domains: [default-safe-outputs]
    add-comment:
      allowed-repos: ["github/gh-aw"]
      hide-older-comments: true
      max: 2
    create-issue:
      expires: 2h
      group: true
      close-older-issues: true
      close-older-key: "smoke-copilot-macos-arm"
      labels: [automation, testing]
    add-labels:
      allowed: [smoke-copilot-macos-arm]
      allowed-repos: ["github/gh-aw"]
    remove-labels:
      allowed: [smoke]
    messages:
      append-only-comments: true
      footer: "> 🍎 *Report filed by [{workflow_name}]({run_url})*{effective_tokens_suffix}{history_link}"
      run-started: "🍎 [{workflow_name}]({run_url}) is now running on macOS ARM64..."
      run-success: "🍎 [{workflow_name}]({run_url}) completed. macOS ARM64 is operational. ✅"
      run-failure: "🍎 [{workflow_name}]({run_url}) reports {status} on macOS ARM64. ❌"
timeout-minutes: 30
strict: false
---

# Smoke Test: Copilot Engine Validation (macOS ARM64)

**IMPORTANT: Keep all outputs extremely short and concise. Use single-line responses where possible. No verbose explanations.**

**PURPOSE**: This smoke test validates that the Copilot engine, AWF firewall, MCP servers, and safe outputs work correctly on macOS ARM64 (macos-latest / Apple Silicon) runners. Docker is provided via Colima (a lightweight macOS Docker runtime). This is critical for ensuring macOS ARM64 support.

## Test Requirements

1. **Architecture Verification**: Run `uname -m` to confirm you are running on an ARM64 (arm64) host. Report the architecture.
2. **macOS Version**: Run `sw_vers` to report the macOS version and build.
3. **Docker Verification**: Run `docker info` to confirm Docker is available (installed via Colima). If Docker is not available, mark this test as ❌ and continue with remaining tests.
4. **GitHub MCP Testing**: Review the last 2 merged pull requests in ${{ github.repository }}
5. **MCP Scripts GH CLI Testing**: Use the `mcpscripts-gh` tool to query 2 pull requests from ${{ github.repository }} (use args: "pr list --repo ${{ github.repository }} --limit 2 --json number,title,author")
6. **File Writing Testing**: Create a test file `/tmp/gh-aw/agent/smoke-test-copilot-macos-arm-${{ github.run_id }}.txt` with content "Smoke test passed for Copilot macOS ARM64 at $(date)" (create the directory if it doesn't exist)
7. **Bash Tool Testing**: Execute bash commands to verify file creation was successful (use `cat` to read the file back)
8. **Build gh-aw**: Run `GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/go-mod make build` to verify the agent can successfully build the gh-aw project on macOS ARM64 (both caches must be set to /tmp because the default cache locations are not writable). If the command fails, mark this test as ❌ and report the failure.

## Output

1. **Create an issue** with a summary of the smoke test run:
   - Title: "Smoke Test: Copilot macOS ARM64 - ${{ github.run_id }}"
   - Body should include:
     - Host architecture (arm64)
     - macOS version from sw_vers
     - Docker availability status
     - Test results (✅ or ❌ for each test)
     - Overall status: PASS or FAIL
     - Run URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}
     - Timestamp

2. Add a **very brief** comment (max 5-10 lines) to the current pull request with:
   - Architecture confirmation (arm64/Apple Silicon)
   - macOS version
   - ✅ or ❌ for each test result
   - Overall status: PASS or FAIL

If all tests pass:
- Use the `add_labels` safe-output tool to add the label `smoke-copilot-macos-arm` to the pull request
- Use the `remove_labels` safe-output tool to remove the label `smoke` from the pull request

{{#runtime-import shared/noop-reminder.md}}
