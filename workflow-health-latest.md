# Workflow Health — 2026-05-11T05:45Z

Score: 62/100 (↑ +1). 218 workflows. Run: §25652578492

## KEY FINDINGS

### Compilation Status
- 218/218 lock files present ✅ (no change)
- 0 missing lock files ✅
- Compile validation unavailable (gh aw extension not in PATH)

### Recent Run Observations (last ~100 runs)
- 1 failure: Daily Fact About gh-aw (push event — parse validation failure, 15+ consecutive today)
- 78 action_required: PR-review cluster (Q, Scout, Archie, cloclo, Grumpy, Security Review, PR Nitpick, PR Code Quality) — expected skip behavior
- 8 skipped: Deployment Incident Monitor (zombie pattern continues)
- 10 success: CI, Issue Monster, Safe Output Health Monitor + others
- PR #31411 OPEN: systemic `on.labels` push-time parse failure fix (waiting for merge)
- PR #31418 OPEN: engine.max-runs to top-level migration

### P0 Issues (Active)
- **APM Unpack** (#30252 OPEN): apm-default.tar.gz exit code 1. Last updated 2026-05-05.

### P0 Issues (Resolved since last run)
- Smoke CI + Gemini (#29666): CLOSED ✅
- Daily Model Inventory Checker (#30043): CLOSED ✅
- config.models (#30307): CLOSED ✅

### P1 Issues (Active — escalated this run)
- **Daily Fact About gh-aw**: 15+ consecutive failures today (push-time parse validation). Was P2 watch, escalated to P1. No issue filed yet — create issue.
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression** (#30180): 82.1% slower.

### P2 Issues (Watch)
- **PR-review cluster** (Q, cloclo, Archie, Scout, Grumpy etc.): ~78 action_required in last 100 runs. Highest waste item. Trigger gate fix needed.
- **Deployment Incident Monitor**: zombie pattern, 8 skips in last 100 runs.
- **PR #31411 (open)**: on.labels push-time failures systemic fix waiting for merge.
- **Doc Build - Deploy**: action_required persistent.

### Actions Taken This Run
- Updated shared memory files
- Added comment to dashboard issue #29109
- Created issue for Daily Fact About gh-aw P1 escalation

### Trends
- Score: 62/100 (↑ from 61)
- 3 P0 issues resolved (Smoke CI/Gemini, Model Inventory, config.models)
- 218 workflows stable
- PR #31411 and #31418 pending merge — watch for regressions
