# Workflow Health — 2026-05-08T05:20Z

Score: 61/100 (→ stable). 217 workflows. Run: §25538277149

## KEY FINDINGS

### Compilation Status
- 217/217 lock files present ✅ (unchanged)
- 0 missing lock files ✅

### P0 Issues (Active)
- **Smoke Gemini** (#29666 OPEN, #30175 CLOSED but fix ineffective): Still 100% failure on May 7. 34+ days.
- **Smoke CI** (#29666 OPEN): CGO/EROFS persistent, action_required
- **Daily Model Inventory Checker** (#30043 OPEN): Copilot CLI silent startup crash
- **APM Unpack** (#30252 OPEN): apm-default.tar.gz exit code 1
- **config.models** (#30307 OPEN): unsupported AWF config field

### P1 Issues (Active)
- **Smoke macOS ARM64**: Issue FILED 2026-05-07 ✅ (77 days)
- **CI regression** TestStrictModePermissions: Filed 2026-05-06
- **MCP gateway session timeout** (#23153 OPEN)
- **Performance Regression** (#30180): 82.1% slower
- **CJS**: Mixed results today (3 action_required + 2 success) — PR-approval pattern

### Actions Taken This Run
- Added comment to #29109 dashboard
- Updated shared memory

### Trends
- Score: 61/100 (→ stable, day 2)
- 217 workflows (→ unchanged)
- No new regressions
- #30175 Gemini issue closed but fix appears ineffective
