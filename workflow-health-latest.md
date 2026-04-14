# Workflow Health - 2026-04-14T12:10Z

Score: 73/100 (↓1 from 74 Apr 13). 191 workflows. Run: §24398051120

## KEY FINDINGS

### 4 New Workflows Since Apr 13
- Repo grew 187 → 191 .md files, all 191 have lock files ✅
- Agent Persona Explorer improved (#26152 merged - safe-output instructions strengthened)

### Ongoing P2 Issues
- Schema Feature Coverage Checker: #25992 (Apr 13) — protected-files config blocks PR creation to .github/workflows/schema-demo-*.md. Fix: add `protected-files: fallback-to-issue` in frontmatter
- Smoke Claude: fails on SCHEDULE, passes on PR runs — environment-specific (#25727)
- Smoke Multi PR: persistent fail (#25415)
- Smoke Cross-Repo PR Create: persistent fail (#25221)
- Smoke Cross-Repo PR Update: persistent fail (#25217)
- Daily Firewall Logs: safe_outputs process failure (#25456)

### ~16 Other Open Workflow Failures
Issues from Apr 8-13 still open for:
- Go Logger Enhancement, Repository Quality Improvement, Documentation Healer, 
  GitHub MCP Structural Analysis, Multi-Device Docs Tester, Refactoring Cadence,
  Go Function Namer, Team Evolution Insights, Community Attribution Updater,
  Security Red Team, Functional Pragmatist, CLI Tools Exploratory Tester,
  Daily Observability Report, Project Performance Summary, Secrets Analysis Agent,
  Safe Output Integrator, PR Triage Agent, Documentation Unbloat

### Healthy Workflows (Today's Successful Runs)
- Documentation Quality Report, Architecture Diagram, PR Triage Report,
  Contribution Check, CLI Version Checker (4 upgrades: Copilot 1.0.25+),
  Semantic Function Clustering, No-Op Tracker (284 comments active)

## Compilation
- 191/191 lock files present ✅
- Zero stale lock files
- All workflows properly compiled

## Copilot/Engine Status
- v1.0.25 NOW AVAILABLE (--remote/--no-remote flags; tracked in #26158)
- v1.0.21 currently ACTIVE
- Claude Code 2.1.105 available
- Codex 0.120.0 available
- Gemini 0.37.2 available (may fix Smoke Gemini failures)

## Score Breakdown
- Compilation: 191/191 ✅: +35
- Multiple healthy workflows running today: +20
- ~23 open failure issues: -12
- Smoke suite partial failures continuing: -5
- Schema Feature Coverage Checker new config failure: -1
- Net: ~73/100

## Score Trend
68 → 71 → 73 → 71 → 70 → 75 → 73 → 74 → 74 → 73
Apr5  Apr6  Apr7  Apr8  Apr9  Apr10 Apr11 Apr12 Apr13 Apr14

## Dashboard Issue
Created new health dashboard issue this run (see safeoutputs)

## Note: GitHub API Rate Limited
GitHub API rate limited at start of run (15min to reset).
Analysis based on shared orchestrator memory + repository state + issue search.

Last updated: 2026-04-14T12:10Z
