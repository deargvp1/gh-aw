# Shared Alerts — 2026-05-01T08:08Z

## P0 (Critical)
- **Gemini API_KEY_INVALID** (NEW TODAY): Smoke Gemini/Crush/OpenCode all fail. api-proxy not injecting Gemini key. Issues: #29459, #29421, #29422, #29423. Any engine=gemini workflow broken.
- **Daily Fact About gh-aw codex crash** (#29088 OPEN): `codex: command not found`. Recurring daily. Day 9+.

## P1 (High)
- **CI integration tests failing**: js-integration-live-api fetch failures; 50% fail rate Apr 30.
- **Documentation Unbloat claude auth failure** (#28659 OPEN): Claude OAuth token issue. Recurring.
- **GitHub Remote MCP Authentication Test** (#27965 OPEN): Day 10+ of model-not-supported error.
- **Safe outputs session not found** (#23153 OPEN): Long-running workflows at risk.
- **awf-api-proxy sidecar unhealthy** (#27888 OPEN): Docker compose failures.
- **GitHub App rate limit exhaustion** (#27251 OPEN).
- **CODEX_HOME variable collision** (#27512 OPEN).
- **P1 backlog stagnant**: 13 open items, 5+ days without resolution.

## P2 (Watch)
- **Daily Cross-Repo Compile Check** hang: 43+ min today (expected ~10 min). Potential MCP inactivity timeout.
- **Safe Outputs SEC-004** (#27235 OPEN).
- **Node.js 20 deprecation** in CI (removal Sep 16, 2026).
- **Performance regressions** (#27280/#27279/#27278 OPEN).

## Resolved (Recent)
- **THREAT_DETECTION_RESULT parse failure** (#28866): Appears resolved/intermittent.

## Trends
- 205 workflows, 0 missing lock files
- Gemini engine newly broken (May 1)
- 7-day quality stable at 74, effectiveness stable at 71
- Success rate recovering: 57%→73%→85%
