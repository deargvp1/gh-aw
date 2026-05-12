# Workflow Health — 2026-05-12T05:39Z

Score: 64/100 (↑ +2). 219 workflows. Run: §25715699041

## KEY FINDINGS

### Compilation Status
- 219/219 lock files present ✅ (1 new workflow added since last run)
- 0 missing lock files ✅
- Compile validation unavailable (gh aw extension not in PATH)

### POSITIVES SINCE LAST RUN
- PR #31411 (on.labels push-time fix) MERGED ✅ 2026-05-11T10:55Z
- PR #31418 (engine.max-runs migration) MERGED ✅
- APM issue #30252 CLOSED ✅ (P0 resolved)

### Active Failures (last 100 runs)
- **daily-fact.lock.yml**: 10+ failures on 2026-05-12 (event=push). Issue #31432 and #31524 open.
- **Smoke Gemini** (#31565, #31575): `fetch failed` — 100% failure
- **Smoke Pi** (#31563): failing
- **Smoke Codex** (#31567): failing
- **Daily Firewall Logs Collector** (#31620): no safe outputs — agent job bug
- **Daily Observability Report** (#31607): failing
- **Design Decision Gate** (#31626): failed
- **Go Logger Enhancement** (#31628): failed
- **Step Name Alignment** (#31636): failed
- **jsweep** (#31637): failed
- ~74 action_required: PR-review cluster (expected skip behavior)

### P0 Issues (Active)
- None (APM #30252 closed ✅)

### P1 Issues (Active)
- **Daily Fact About gh-aw** (#31432, #31524): push-time parse failures continue post-PR#31411 merge. Still failing 2026-05-12.
- **Smoke Gemini** (#31575): 100% failure — `fetch failed` to Gemini.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression** (#30180): 82.1% slower.

### P2 Issues (Watch)
- **PR-review cluster**: ~74 action_required in 100 runs. Waste issue.
- **Firewall reporting**: #31607, #31620 — no safe outputs in agent job.
- **Multiple workflow failures**: Design Decision Gate, Go Logger Enhancement, Step Name Alignment, jsweep — need investigation.
- Node.js 20 deprecation deadline Sep 16, 2026.

### Actions Taken This Run
- Updated shared memory files
- Added comment to dashboard issue #29109

### Trends
- Score: 64/100 (↑ +2 from 62)
- P0 count: 0 (was 1 — APM resolved) ✅
- 219 workflows (was 218 — 1 new workflow)
- PRs #31411, #31418 merged — compilation regressions not seen
- Smoke tests failing across multiple engines (Pi, Gemini, Codex)
