# Shared Alerts — 2026-05-08T05:20Z

## P0 (Critical)
- **Smoke Gemini** (#29666 OPEN): 100% failure, proxy/API-key blocks traffic. 34+ days. ⚠️ Issue #30175 was closed May 6 as "fixed" but runs on May 7 still show 100% failure — fix was NOT effective.
- **Smoke CI** (#29666 OPEN): CGO/EROFS persistent, 100% action_required.
- **Daily Model Inventory Checker** (#30043 OPEN): Copilot CLI silent startup crash.
- **APM Unpack Systemic Failure** (#30252 OPEN): apm-default.tar.gz exits code 1, affects multiple workflows.

## P1 (High)
- **Smoke macOS ARM64**: 100% failure since 2026-02-20 (77 days). **Issue FILED 2026-05-07** ✅
- **CI regression on main** (May 6): `TestStrictModePermissions` failing. Issue filed.
- **config.models unsupported field** (#30307 OPEN): blocks smoke runs.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression in Validation** (#30180): 82.1% slower.
- **CJS test**: 3 action_required + 2 success today (May 8) — likely PR-triggered agent-approval runs, not a true failure. Monitor.

## P2 (Watch)
- **Node.js 20 deprecation** in CI: deadline Sep 16, 2026. Migrate to Node.js 22.
- **PR-review cluster** (Q, /cloclo, Archie, Scout): ~100+ action_required/day. Consolidation needed.
- **Doc Build - Deploy**: action_required persistent (deployment stalled).
- **Content Moderation + AI Moderator**: scope-creep on PR diff events.
- **Resource Summarizer Agent**: chronic skips, zero outputs.

## Resolved (Do Not Re-File)
- #29863 Smoke Copilot regression → RECOVERED ✅
- #30205 Auto-Triage Issues → CLOSED ✅
- #30188 Documentation Unbloat → CLOSED ✅
- #30233 Daily Documentation Healer → CLOSED ✅
- #30069 Step Name Alignment → CLOSED ✅
- #30241 Smoke Claude → CLOSED ✅
- #30244 Smoke Codex → CLOSED ✅
- #30347, #30144 GitHub MCP Structural Analysis → CLOSED ✅
- #30085, #30086, #30087 Safe Outputs Conformance → CLOSED ✅
- #30102 Schema Consistency Checker → CLOSED ✅

## Trends
- 217 workflows, 0 missing lock files
- Health: 61/100 (→ stable, day 2)
- Gemini still broken despite #30175 closure — needs re-investigation
