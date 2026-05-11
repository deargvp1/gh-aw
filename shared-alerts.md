# Shared Alerts — 2026-05-11T05:45Z

## P0 (Critical)
- **APM Unpack Systemic Failure** (#30252 OPEN): apm-default.tar.gz exits code 1. Last updated 2026-05-05. 3 workflows blocked.

## P0 — RESOLVED (Do Not Re-File)
- Smoke CI + Gemini (#29666): CLOSED ✅ 2026-05-11
- Daily Model Inventory Checker (#30043): CLOSED ✅ 2026-05-11
- config.models (#30307): CLOSED ✅ 2026-05-11

## P1 (High)
- **Daily Fact About gh-aw**: 15+ push-time parse failures today. Was P2, escalated to P1. Issue created this run.
- **Smoke macOS ARM64**: 100% failure since 2026-02-20 (81 days). Issue filed 2026-05-07 ✅
- **CI regression on main**: TestStrictModePermissions failing. Issue filed 2026-05-06.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression in Validation** (#30180): 82.1% slower.

## P2 (Watch)
- **PR-review cluster** (Q, cloclo, Archie, Scout, Grumpy, Security Review, PR Nitpick, PR Code Quality): ~78 wasted run-attempts per 100 runs. Trigger gate fix or consolidation needed. HIGHEST WASTE item.
- **on.labels push-time failures**: PR #31411 open fix. Merge unblocks systemic issue.
- **engine.max-runs migration**: PR #31418 open. Watch for compilation regressions after merge.
- **Deployment Incident Monitor**: zombie pattern — 8x skipped in 100 runs; consider deprecation.
- **Resource Summarizer Agent**: chronic skips, zero outputs.
- **Doc Build - Deploy**: action_required persistent (deployment stalled).
- **Node.js 20 deprecation** in CI: deadline Sep 16, 2026.

## Resolved (Do Not Re-File)
- #29863 Smoke Copilot regression → RECOVERED ✅
- #30205 Auto-Triage Issues → CLOSED ✅
- #30188 Documentation Unbloat → CLOSED ✅
