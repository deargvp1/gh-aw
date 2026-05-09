# Workflow Health — 2026-05-09T05:31Z

Score: 61/100 (→ stable). 218 workflows (+1). Run: §25592941547

## KEY FINDINGS

### Compilation Status
- 218/218 lock files present ✅ (+1 new workflow added)
- 0 missing lock files ✅
- Compile validation unavailable (gh aw extension not in PATH this run)

### Data Availability
- GitHub API unavailable (connection refused / 403)
- Analysis based on prior memory state + local git

### New Workflow
- PR #31159 merged: `include label_command labels in create_labels` — 1 new .md + .lock.yml added

### P0 Issues (Active — from prior runs)
- **Smoke Gemini** (#29666 OPEN): 100% failure, 35+ days. #30175 fix ineffective.
- **Smoke CI** (#29666 OPEN): CGO/EROFS persistent
- **Daily Model Inventory Checker** (#30043 OPEN): Copilot CLI silent startup crash
- **APM Unpack** (#30252 OPEN): apm-default.tar.gz exit code 1
- **config.models** (#30307 OPEN): unsupported AWF config field

### P1 Issues (Active)
- **Smoke macOS ARM64**: Issue filed 2026-05-07 ✅
- **CI regression** TestStrictModePermissions: Issue filed 2026-05-06
- **MCP gateway session timeout** (#23153 OPEN)
- **Performance Regression** (#30180): 82.1% slower
- **CJS**: Mixed results (3 action_required + 2 success) — monitor

### Actions Taken This Run
- Updated shared memory files
- Added comment to dashboard issue #29109
- No new issues created (API unavailable, no new data)

### Trends
- Score: 61/100 (→ stable, day 3)
- 218 workflows (+1 from yesterday)
- No new data available (API offline)
