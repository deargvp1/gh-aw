# Agent Performance — 2026-05-01
Run: §25207518271 | Q:74→74 E:71→71

## Ecosystem Overview (May 1)
- Overall quality: 74/100 (→ stable day 5), effectiveness: 71/100 (→ stable)
- 50 sampled runs: 17 success, 0 failure, 16 skipped, 17 in-progress/other
- NEW P0: Gemini API_KEY_INVALID (issues #29421, #29422, #29423, #29459)
- Engines today: mix of copilot, claude (codex still broken)

## Top Performers
1. **Test Quality Sentinel** (Q:90 E:92) — consistent high quality
2. **Smoke CI** (Q:88 E:87) — infrastructure gatekeeper
3. **Daily Caveman Optimizer** (Q:85 E:85) — code quality
4. **Static Analysis Report** (Q:83 E:82) — 100% today, 4 security issues created
5. **[aw] Failure Investigator** (Q:82 E:80) — 100% today

## Active Failures (May 1)
- **Gemini API_KEY_INVALID** (P0 NEW): Smoke Gemini/Crush/OpenCode failing — #29459
- **Codex binary missing** (P0 ongoing): Daily Fact — #29088
- **GitHub Remote MCP Auth** (P1 day 10+): #27965
- **CI integration tests**: Ongoing (50% fail rate Apr 30)
- **Safe outputs batch**: 7 workflows failed yesterday (agent crash pattern)
- **Daily Cross-Repo Compile Check**: Hung 43+ min (potential MCP timeout)

## 7-day Trends
- Quality: 72→73→74→74→74→74→74 (→ stable)
- Effectiveness: 68→69→70→71→71→71→71 (→ stable)
- Success rate: 93%→94%→95%→57%→73%→85%→~85%
- P1 backlog: 13 open items (stagnant 5+ days)

## Discussion/Issues This Run
- Discussion posted (performance report May 1)
- No new issues (existing cover all active failures)

Last updated: 2026-05-01T08:08Z by agent-performance-manager
