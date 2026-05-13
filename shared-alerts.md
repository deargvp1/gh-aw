# Shared Alerts — 2026-05-11T13:36Z

## P0 (Critical)
- **APM Unpack Systemic Failure** (#30252 OPEN): apm-default.tar.gz exits code 1. Last updated 2026-05-05. 3 workflows blocked.

## P0 — RESOLVED (Do Not Re-File)
- Smoke CI + Gemini (#29666): CLOSED ✅ 2026-05-11
- Daily Model Inventory Checker (#30043): CLOSED ✅ 2026-05-11
- config.models (#30307): CLOSED ✅ 2026-05-11

## P1 (High)
- **Daily Fact About gh-aw**: 15+ push-time parse failures. Issue created 2026-05-11. PR #31411 merge should help.
- **Smoke macOS ARM64**: 100% failure since 2026-02-20 (81+ days). Issue filed 2026-05-07 ✅
- **CI regression on main**: TestStrictModePermissions failing. Issue filed 2026-05-06.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression in Validation** (#30180): 82.1% slower.

## P2 (Watch)
- **PR-review cluster** (Q, cloclo, Archie, Scout, Grumpy, Security Review, PR Nitpick, PR Code Quality): ~272 wasted run-attempts/day. HIGHEST WASTE. Trigger gate fix or consolidation needed.
- **on.labels push-time failures**: PR #31411 open fix. Merge unblocks systemic issue.
- **engine.max-runs migration**: PR #31418 open. Watch for compilation regressions after merge.
- **Deployment Incident Monitor**: zombie pattern — 8x skipped per 100 runs; consider deprecation.
- **Resource Summarizer Agent**: chronic skips, zero outputs.
- **Doc Build - Deploy**: action_required persistent (deployment stalled).
- **Node.js 20 deprecation** in CI: deadline Sep 16, 2026.
- **Quality/Effectiveness plateau**: 10 days flat (Q:74, E:71) — structural bottleneck suspected.

## Resolved (Do Not Re-File)
- #29863 Smoke Copilot regression → RECOVERED ✅
- #30205 Auto-Triage Issues → CLOSED ✅
- #30188 Documentation Unbloat → CLOSED ✅

---
## Update — 2026-05-12T05:39Z (Workflow Health Manager)

### RESOLVED
- APM Unpack #30252: CLOSED ✅
- PR #31411 (on.labels fix): MERGED ✅
- PR #31418 (engine.max-runs): MERGED ✅

### NEW/ESCALATED
- **Daily Fact parse failures**: still occurring post-PR#31411 merge; issues #31432 #31524 open.
- **Smoke Gemini fetch failed**: issue #31575 open — 100% failure.
- **Firewall reporting broken** (#31607, #31620): no safe outputs from agent job.
- **Multiple workflow failures** today: Design Decision Gate #31626, Go Logger Enhancement #31628, Step Name Alignment #31636, jsweep #31637.

---
## Update — 2026-05-12T13:19Z (Agent Performance Manager)

### RESOLVED (since May 11)
- APM Unpack #30252: CLOSED ✅
- PR #31411 (on.labels fix): MERGED ✅
- PR #31418 (engine.max-runs): MERGED ✅

### NEW/ESCALATED
- **4 same-day workflow failures** (2026-05-12): Design Decision Gate #31626, Go Logger Enhancement #31628, Step Name Alignment #31636, jsweep #31637 — possible shared root cause (engine availability or PR #31418 side-effect). Needs investigation.
- **Daily Fact still failing** post-PR#31411 merge (#31432, #31524 still open)
- **Quality/Effectiveness plateau**: Day 11 flat at Q:74/E:71 — structural bottleneck suspected (PR-review cluster waste dragging averages)
- **PR-review cluster waste escalated**: ~272 wasted run-attempts/day confirmed — highest waste in ecosystem; trigger gate fix is highest-ROI action

---
## Update — 2026-05-13T05:45Z (Workflow Health Manager)

### NEW (since May 12)
- **CI integration test failure**: Fix failing "Integration: Workflow Misc Part 2" (#31860) — CGO failing 3/4 runs. New regression.
- **Semantic Function Refactoring** (#31827): new agentic failure
- **Daily Security Red Team Agent** (#31817): new failure
- **Scout** (#31811): failed
- **Daily Cache Strategy Analyzer** (#31773): new failure
- **4 new workflows added** (219→223): no compilation issues

### RESOLVED
- Daily Firewall Logs Collector: auto-close ran ✅
- Smoke Copilot dispatch_workflow: auto-close ran ✅

### WATCH
- CI failures trending up (scheduled CI + CGO): possible integration test regression around #31860
- Deep-report triage issue #31729: 18 stale [aw]-failed issues — recommend bulk close
