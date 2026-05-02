# Shared Alerts — 2026-05-02T05:30Z

## P0 (Critical)
- **Smoke Engine Failures** (#29459 OPEN): Gemini API_KEY_INVALID + Crush EROFS + Claude timeout. 3/7 smoke engines broken. Every PR sees red on these checks.

## P1 (High)
- **CI/CGO/CJS build broken** (NEW May 2): push `refactor: flatten nested if in validateStrictSandboxCustomization` broke CGO/CJS. 50% CI fail rate. Investigate immediately.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk of MCP disconnect.

## P2 (Watch)
- **Node.js 20 deprecation** in CI: deadline Sep 16, 2026. Migrate to Node.js 22.
- **Safe Outputs SEC-004** (#27235 OPEN).

## Resolved Since Apr 30 (Do Not Re-File)
- #29088 Codex crash → CLOSED
- #28659 Doc Unbloat claude auth → CLOSED
- #27965 GitHub Remote MCP Auth → CLOSED
- #27888 awf-api-proxy sidecar → CLOSED
- #27251 GitHub App rate limit → CLOSED
- #27512 CODEX_HOME collision → CLOSED

## Trends
- 207 workflows, 0 missing lock files
- P1 backlog reduced from 13 → 2 (major improvement)
- Smoke engines: 3 broken (Gemini, Crush, Claude)
- CI build broken from today's refactor push
