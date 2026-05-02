# Workflow Health — 2026-05-02T05:30Z

Score: 68/100 (↓4 from 72). 207 workflows. Run: §25244683354

## KEY FINDINGS

### Compilation Status
- 207/207 lock files present ✅
- 0 missing lock files ✅

### P0 Issues (Active)
- **Smoke Engine Failures** (#29459 OPEN): Gemini API_KEY_INVALID + Crush EROFS + Claude timeout (newly added today). 3/7 smoke engines broken.

### P1 Issues (Active)
- **CI/CGO/CJS build failures** (NEW): push `refactor: flatten nested if in validateStrictSandboxCustomization` broke May 2 builds. 50% CI fail rate.
- **MCP gateway session timeout** (#23153 OPEN): Ongoing structural risk for long workflows.

### P2 Issues
- Node.js 20 deprecation in CI (deadline Sep 16, 2026)

### Issues Resolved Since Apr 30 ✅ (MAJOR IMPROVEMENT)
- #29088 Codex binary missing (Daily Fact) → CLOSED
- #28659 Documentation Unbloat claude auth → CLOSED
- #27965 GitHub Remote MCP Auth Test → CLOSED
- #27888 awf-api-proxy sidecar unhealthy → CLOSED
- #27251 GitHub App rate limit exhaustion → CLOSED
- #27512 CODEX_HOME variable collision → CLOSED

### Actions Taken This Run
- Created dashboard issue: [aw] Workflow Health Dashboard — 2026-05-02
- Added comment to #29459 with Claude smoke failure data

### Trends
- P1 backlog reduced from 13 → 2 active items (massive improvement)
- New: CI/CGO/CJS build broken today from recent refactor
- Smoke test reliability declining (3/7 engines broken)
