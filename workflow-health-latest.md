# Workflow Health — 2026-05-13T05:45Z

Score: 63/100 (↓ -1). 223 workflows. Run: §25780678654

## KEY FINDINGS

### Compilation Status
- 223/223 lock files present ✅ (4 new workflows since last run: 219→223)
- 0 missing lock files ✅
- Compile validation unavailable (gh aw extension not in PATH)

### New Failures Since Last Run (2026-05-13)
- **Daily Observability Report** (#31828): failed — ongoing (was #31607)
- **Semantic Function Refactoring** (#31827): failed — NEW
- **Daily Security Red Team Agent** (#31817): failed — NEW
- **Scout** (#31811): failed — NEW
- **Daily Cache Strategy Analyzer** (#31773): failed — NEW
- **Design Decision Gate** (#31780): failed — ongoing
- **Smoke Codex** (#31778): failed — ongoing
- **Smoke Pi** (#31765): failed — ongoing
- **CI** (scheduled): failed — NEW (integration test regression)
- **CGO**: 3/4 failed on PRs — regression (Fix failing Integration tests)

### Resolved Since Last Run
- **Daily Firewall Logs Collector**: auto-close issued ✅
- **Step Name Alignment**: auto-close issued (single run)
- **Smoke Copilot dispatch_workflow**: auto-close issued ✅

### Active Issues (P1)
- Daily Fact parse failures (#31432, #31524): ongoing
- Smoke Gemini (#31778): fetch failed
- MCP gateway session timeout (#23153): open
- Performance Regression (#30180): 82.1% slower
- CI integration test failure: new (Fix failing Integration tests - #31860)

### P2 Issues (Watch)
- PR-review cluster (~272 wasted/day): #31724 filed
- Security findings: #31708, #31704 open
- Deep-report triage: #31729 (18 stale [aw] failed issues)
- Node.js 20 deprecation: Sep 16, 2026

### Actions Taken This Run
- Updated shared memory files
- Added comment to dashboard issue #29109

### Trends
- Score: 63/100 (↓ -1 from 64)
- P0 count: 0 ✅
- 223 workflows (was 219 — 4 new workflows)
- CI integration test new failure: CGO + CI failing on main
- Multiple agentic workflow failures: 8 open [aw] failed issues
