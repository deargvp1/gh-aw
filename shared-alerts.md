# Shared Alerts — 2026-04-14T12:10Z

## P2 (High)
- **Smoke Claude schedule failure** (ongoing, #25727): Still failing on daily schedule, passes on PR runs. Environment-specific — schedule vs PR context divergence.
- **Smoke Cross-Repo PR Create/Update** (#25221, #25217, Apr 8): STALE 6 days. No fix applied. Needs escalation.
- **Schema Feature Coverage Checker** (#25992, Apr 13): Protected-files config blocks PR creation to .github/workflows/schema-demo-*.md. Fix: add `protected-files: fallback-to-issue` to frontmatter.
- **Documentation Unbloat inconsistent** (ongoing): ~$55/week Claude, 50% success. Cost gate needed.
- **Daily Firewall Logs** (#25456): safe_outputs process failure.
- **Smoke Multi PR** (#25415): persistent failure.
- **GitHub Remote MCP Auth Test**: 100% failure — #24829 closed not_planned. Test still failing.

## P3 (Watch)
- **Smoke Gemini**: 100% failure. Gemini 0.37.2 now available (#26158) — may fix after CLI bump.
- **~16 other daily workflow failures**: From Apr 8-13, mostly engine/copilot crashes.
- **Daily Issues Report recurring failure** (#25265, #25503): Copilot agent crash pattern.

## Copilot Version Status
- v1.0.25 NOW AVAILABLE (new --remote/--no-remote flags; see #26158)
- v1.0.21 ACTIVE (current in production as of Apr 14)
- Claude Code 2.1.105 available (#26158)
- Codex 0.120.0 available (#26158)
- Gemini 0.37.2 available (#26158) — may fix Smoke Gemini

## Recoveries (Apr 11-14)
- ✅ Smoke Copilot: RECOVERED
- ✅ Contribution Check: RECOVERED
- ✅ Agent Persona Explorer: IMPROVED (safe-output instructions strengthened, #26152)
- ✅ 20 PRs merged by Copilot bot (OTel, security, workflow fixes)

## Ecosystem State
- 191 compiled workflows (+4 since Apr 13). Health: 73/100 (↓1 Apr 14 12:10Z)
- Engine split: ~124 copilot, ~41 claude, ~18 codex, ~4 others
- v1.0.21 currently active, v1.0.25 available for upgrade

## Orchestrator Summaries (Apr 14)
- Agent Performance (Apr 14 04:37Z): Q:74↑1 E:66↑1. CLI Version Checker standout (4 upgrades).
- Workflow Health (Apr 14 12:10Z): Score 73/100 (↓1). 23 open failure issues. 191 workflows all compiled.
- Campaign Manager (last known: Mar 16 17:41Z): Status unknown — no recent update

Last updated: 2026-04-14T12:10Z by workflow-health-manager
