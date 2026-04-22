# Workflow Health — 2026-04-22T12:11Z

Score: 69/100 (↓-1 from 70 Apr 21). 197 workflows. Run: §24777438453

## KEY FINDINGS

### Compilation Status
- 197/197 lock files present ✅
- **23 stale lock files** ⚠️ (up from 15 yesterday — still needs `make recompile`)
  - copilot-agent-analysis, copilot-pr-merged-report, daily-astrostylelite-markdown-spellcheck
  - daily-cli-tools-tester, daily-doc-updater, daily-mcp-concurrency-analysis, daily-news
  - daily-regulatory, daily-semgrep-scan, daily-workflow-updater, dependabot-go-checker
  - developer-docs-consolidator, example-workflow-analyzer, go-logger, issue-monster
  - org-health-report, q, repo-audit-analyzer, semantic-function-refactor
  - smoke-claude, spec-enforcer, typist, video-analyzer

### P0 Issues
- None today

### P1 Issues (Active)
- **Stale lock files / Codex 401** (#27724 + #27731 OPEN): 23 workflows stale, Codex 401 loop continues
  - Duplicate Code Detector (#27816): Codex 401 from stale lock (12:00Z)
  - Daily Fact About gh-aw (#27810): Codex+mempalace MCP CLI server failure (11:32Z)
- **Design Decision Gate push bundle failure** (#27756 OPEN): `push_to_pull_request_branch: Failed to apply bundle`
- **Design Decision Gate max_turns=5** (#27470 OPEN): structurally impossible ADR generation
- **Safe outputs "session not found" at 37min** (#27755 OPEN + #23153): MCP server session expires at ~37min
- **Smoke Claude** (#27030 OPEN): Ongoing
- **Smoke Copilot** (#27028 OPEN): Ongoing
- **node: command not found on aw-gpu-runner-T4** (#27534 OPEN): Recurring
- **GitHub App rate limit exhaustion** (#27251 OPEN): Co-scheduled at 23:44 UTC
- **CODEX_HOME variable collision** (#27512 OPEN): cp same-file error

### P2 Issues
- **Daily Documentation Updater protected files** (#27801 OPEN, today): Tried to modify .github/aw/ agent instruction files
  - Fix: add `protected-files: fallback-to-issue` to daily-doc-updater frontmatter
- **Safe Outputs SEC-004** (#27235 OPEN): 4 handler files
- **Performance regressions** (#27280/#27279/#27278 OPEN)
- **Copilot reviewer fan-out** (#27130 OPEN)
- **MCP gateway long-running drops** (#23153 OPEN)

### Today's Run Stats (sample: 30 scheduled runs)
- Success: 27 (90%)
- Failures: 3 (10%)
  - Duplicate Code Detector: Codex 401 → #27816
  - Daily Fact About gh-aw: Codex+mempalace → #27810
  - Daily Documentation Updater: Protected files → #27801

## Open Issues (workflow-health related)
- #27816 Duplicate Code Detector Codex 401 (auto, today)
- #27810 Daily Fact About gh-aw failed (auto, today)
- #27801 Daily Documentation Updater protected files (auto, today)
- #27731 Recompile lock files Codex 401 (P1)
- #27724 Agentic workflows out of sync / stale lock files (P1)
- #27756 Design Decision Gate push bundle failure (P1)
- #27755 Safe outputs session not found 37min (P1)
- #27512 CODEX_HOME variable collision (P1)
- #27470 Design Decision Gate max_turns=5 (P1)
- #27534 Daily Issues Report GPU/node (P1)
- #27251 Rate limit exhaustion co-scheduled (P1)
- #27030 Smoke Claude (P1)
- #27028 Smoke Copilot (P1)
- #27235 Safe Outputs SEC-004 (P2)
- #23153 MCP gateway session drops (P2)

## Engine/Tool Status
- Copilot: mostly ✅ (Smoke Copilot ongoing #27028)
- Claude: mostly ✅ (Smoke Claude #27030 ongoing)
- Codex: ❌ Blocked by stale lock files (Codex 401 at api.openai.com until recompile)
- MCP Gateway: ⚠️ session timeout at ~37min (#27755)

## Actions This Run
- No new issues created (all failures have existing trackers)
- Added comment to #27724 with updated 23-file stale list
- Memory files updated

Last updated: 2026-04-22T12:11Z by workflow-health-manager
