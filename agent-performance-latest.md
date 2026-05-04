# Agent Performance — 2026-05-04
Run: §25320885374 | Q:74→74 E:71→71

## Ecosystem Overview (May 4)
- Overall quality: 74/100 (→ stable, 7-day plateau), effectiveness: 71/100 (→ stable)
- 211 workflows, health: 65/100
- Engines active: copilot, claude, codex; gemini 100% broken ongoing (30+ days)
- Today success rate: ~60% (PR-review cluster ar=normal); failures: GitHub MCP Structural Analysis, Schema Consistency Checker, Multi-Device Docs Tester

## Top Performers
1. **Test Quality Sentinel** (Q:90 E:92) — Copilot, 0 errors, ~5-6m
2. **Issue Monster** (Q:87 E:88) — 4/4 runs success, 100%
3. **Daily Go Function Namer** (Q:84 E:83) — Claude, 0 errors
4. **Draft PR Cleanup** (Q:82 E:80) — 0 errors
5. **Package Specification Enforcer** (Q:82 E:81) — Claude, note: historical variance

## Key Patterns Detected (May 4)
- `repetition` + `inconsistency`: PR-review cluster (Scout, Archie, Q, /cloclo, AI Moderator, Content Moderation) — 6 agents on same trigger, run counts diverge 5–8
- `under-creation`: Schema Consistency Checker + Multi-Device Docs Tester — completed analysis but never called safeoutputs ($3.84 wasted)
- `under-creation`: GitHub MCP Structural Analysis — claude engine crash, 0 successful outputs
- `inconsistency`: Copilot Prompt Clustering — 1/5 historical fail rate

## Active Issues (May 4)
- **P0 ongoing**: Smoke Gemini 100% fail — API_KEY_INVALID (#29852, #29816, #29459)
- **P0 ongoing**: Smoke CI — CGO/EROFS (#29666)
- **P0 ongoing**: Smoke macOS ARM64 — 100% fail since Feb 2026, NO ISSUE FILED
- **P0**: Daily Model Inventory Checker — Copilot CLI silent crash (#30043)
- **P1**: Claude safeoutputs omission — Schema Consistency Checker + Multi-Device Docs Tester (#30102)
- **P1**: GitHub MCP Structural Analysis — claude auth crash (#30144)
- **P1**: Q agent — 0→72 turn variance (prompt instability)
- **P1**: 01:49 UTC transient wave — 5 smokes failed simultaneously

## 7-day Quality Trend
- Quality:      72→73→74→74→74→74→74 (→ stable plateau)
- Effectiveness: 68→69→70→71→71→71→71 (→ stable plateau)

## Actions This Run
- Discussion created: Agent Performance Report — Week of 2026-05-04
- No new issues (active issues #30102, #30144, #30069 already cover failures)
- Top recommendation: Enforce safeoutputs in claude workflows (eliminates $3–4/day waste)

Last updated: 2026-05-04T13:12Z by agent-performance-manager
