# Agent Performance — 2026-05-05
Run: §25377869054 | Q:74→74 E:71→71

## Ecosystem Overview (May 5)
- Overall quality: 74/100 (→ stable plateau, day 3), effectiveness: 71/100 (→ stable)
- 213 workflows (+2 from May 4), health: 63/100 (↓-2 from APM unpack regression)
- Engines active: copilot, claude, codex; gemini 100% broken ongoing (31+ days)
- Today PR-review cluster: 100% action_required (expected — awaiting copilot agent)
- Today failures: GitHub MCP Structural Analysis, Draft PR Cleanup, Daily Documentation Updater, Daily News, Dev, Design Decision Gate, Test Quality Sentinel, Agent Container Smoke Test, Smoke Copilot (early run)

## Top Performers (May 5)
1. **Agentic Maintenance** (Q:90 E:92) — 3/3 success ✅
2. **License Compliance Check** (Q:89 E:88) — 3/3 success ✅
3. **Copilot code review** (Q:87 E:85) — 3/3 success ✅
4. **Auto-Close Parent Issues** (Q:85 E:83) — 2/2 success ✅
5. **Daily File Diet** (Q:82 E:80) — success ✅

## Key Patterns Detected (May 5)
- `action_required`: PR-review cluster (Scout 18 runs, /cloclo 22, Q 19, Grumpy 14, Security 14, PR Nitpick 13) — all action_required, large volume (expected for PR-review trigger)
- `inconsistency` + `under-creation`: Draft PR Cleanup, Daily Documentation Updater, Daily News — recurring failures, noop or safeoutputs not called
- `failure`: aw-failures agent actively filing issue reports — functioning well (P0 issues #30307, #30306)
- `success`: Copilot PRs (30354, 30352, 30350, 30316, 30315) — Copilot coding agent performing well on fix tasks
- NEW: config.models smoke sweep failure — 10 smoke runs blocked by unsupported config field (#30307 filed)

## Active Issues (May 5)
- **P0 ongoing**: Smoke Gemini 100% fail (30+ days) — #30175, #29852
- **P0 ongoing**: Smoke CI CGO/EROFS — #29666
- **P0 ongoing**: APM unpack systemic — #30252
- **P0 ongoing**: Daily Model Inventory Checker — #30043
- **P1 ongoing**: Smoke macOS ARM64 (NO ISSUE — needs filing)
- **P1 new**: config.models unsupported field — #30307
- **P1 ongoing**: Schema Consistency Checker + Multi-Device Docs Tester safeoutputs omission — #30102
- **P1 ongoing**: GitHub MCP Structural Analysis claude crash — #30347 (new failure today)
- **P1 ongoing**: Auto-Triage Issues — #30205
- **P1**: Documentation Unbloat recurring failure — #30188

## 7-day Quality Trend
- Quality:      72→73→74→74→74→74→74 (→ stable plateau, day 4)
- Effectiveness: 68→69→70→71→71→71→71 (→ stable plateau, day 4)

## Actions This Run
- Discussion created: Agent Performance Report — Week of 2026-05-05
- No new issues filed (all tracked in existing issues)
- Top recommendation: Fix config.models unsupported field (#30307) — blocking 10+ smoke runs

Last updated: 2026-05-05T13:05Z by agent-performance-manager
