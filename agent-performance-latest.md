# Agent Performance — 2026-04-22
Run: §24760397710 | Q:71↓1 E:67↓1

## Ecosystem Overview (Apr 21-22)
- Overall quality: 71/100 (↓-1), effectiveness: 67/100 (↓-1)
- 18 workflows, 29 runs in 48h
- Root cause of many failures: stale lock files after PR #27711 merged (needs `make recompile`)

## Top Performers
1. **[aw] Failure Investigator (6h)** (Q:92 E:88) - #27729 (Codex 401 RCA: stale lock files) - excellent structured analysis
2. **Smoke CI** (Q:88 E:92) - 3 runs, all success, 5-8T, consistent
3. **CLI Version Checker** (Q:85 E:82) - #27760 (4 tools updated: Claude, Copilot 1.0.21→1.0.34, Codex, GitHub MCP), still 42T
4. **Test Quality Sentinel** (Q:85 E:88) - 4 runs all success, 4-12T, stable
5. **Auto-Triage / Issue Monster** (Q:80 E:82) - 3T, lightweight, reliable

## Improved This Run 📈
- **GitHub Remote MCP Authentication Test**: SUCCESS ✅ (was Q:40 E:0 yesterday) — resolved!
- **Agent Persona Explorer**: 16T (down from 42T yesterday) — improving

## Watch / Needs Improvement
- **AI Moderator** (Q:10 E:5) - Codex 401 (stale lock files) + chatgpt.com firewall — P1 ongoing
- **Design Decision Gate** (Q:62 E:48) - 50% failure today (2/4 runs); NEW failure mode: push bundle failure (#27756) + existing max_turns=5 (#27470)
- **Documentation Unbloat** (Q:48 E:52) - 56T, ROI unclear, #27600 recommendation pending
- **Smoke Codex/Gemini/Crush** (Q:20 E:10) - all failing (Codex=stale locks, Gemini/Crush=new engines)
- **Smoke Copilot** (ongoing P1 #27028)
- **Smoke Claude** (39T+ still running, outcome TBD)

## New Findings Today
1. Safe outputs "session not found" at 37min (#27755 from @dsyme) — previously thought threshold was ~1h — NEW P1 infrastructure issue
2. Design Decision Gate: NEW push bundle failure (#27756) in addition to max_turns=5 issue (#27470)
3. Failure Investigator RCA #27729: identified PR #27711 stale lock files as root cause of Codex 401 loop ✅
4. Smoke OpenCode: SUCCESS ✅ — new engine working
5. GitHub Remote MCP Auth Test: RESOLVED ✅

## P0/P1 Active
- P1: Stale lock files (#27724 + #27731) — fix: `make recompile`
- P1: Safe outputs session not found at 37min (#27755) — infrastructure
- P1: Design Decision Gate push bundle failure (#27756)
- P1: Design Decision Gate max_turns=5 (#27470)
- P1: Smoke Copilot (#27028, ongoing)

## Issues/Actions This Run
- Discussion created (performance report)
- No new improvement issues (all tracked in existing issues)

Last updated: 2026-04-22T04:37Z by agent-performance-manager
