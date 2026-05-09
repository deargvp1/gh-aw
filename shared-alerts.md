# Shared Alerts — 2026-05-09T05:31Z

## P0 (Critical)
- **Smoke Gemini** (#29666 OPEN): 100% failure, proxy/API-key blocks. 35+ days. #30175 closed ineffective.
- **Smoke CI** (#29666 OPEN): CGO/EROFS persistent, 100% action_required.
- **Daily Model Inventory Checker** (#30043 OPEN): Copilot CLI silent startup crash.
- **APM Unpack Systemic Failure** (#30252 OPEN): apm-default.tar.gz exits code 1.
- **config.models** (#30307 OPEN): unsupported AWF config field, blocks smoke runs.

## P1 (High)
- **Smoke macOS ARM64**: 100% failure since 2026-02-20 (78 days). Issue filed 2026-05-07 ✅
- **CI regression on main**: TestStrictModePermissions failing. Issue filed 2026-05-06.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression in Validation** (#30180): 82.1% slower.
- **CJS test**: Mixed results — monitor.

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
- #29109 Dashboard issue (active, updated periodically)

## Trends
- 218 workflows (+1), 0 missing lock files
- Health: 61/100 (→ stable, day 3)
- API unavailable for May 9 run — no new execution data
- Gemini still broken despite #30175 closure — needs re-investigation
