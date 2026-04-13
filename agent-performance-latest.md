# Agent Performance - 2026-04-13
Run: §24326084430 | Q:73↑3 E:65↑5

Top: CLI Version Checker (Q:90 E:92), Copilot Coding Agent (Q:85 E:88, 20 PRs Apr 12-13), Issue Monster (Q:87 E:90 - 5/5), Agentic Maintenance (Q:83 E:82 - added cache cleanup), Smoke Copilot (Q:82 E:80 - RECOVERED to scheduled pass)

Watch: Smoke Claude (Q:40 E:35 - fails on schedule, passes on PR runs - environment-specific), Smoke Gemini (Q:10 E:10 - 100% fail #25216), Smoke Cross-Repo PR Create/Update (#25221, #25217 still open), GitHub Remote MCP Auth Test (persistent fail), Documentation Unbloat (50% success - inconsistent), Daily Semgrep Scan (new failure today)

Smoke Tests Apr 13: Copilot ✅ (sched) | Claude ❌ (sched) / ✅ (PR) | Codex ✅ | Multi PR ✅ | Gemini ❌ | Cross-Repo ❌/❌

Notable Apr 12-13: 20 PRs merged by Copilot bot (OTel spans, SEC-004 sanitization, cache cleanup, allow multiple agent assignments). CLI bump issue #25978 (Copilot 1.0.24, Claude Code 2.1.104, Codex 0.120.0, MCP v0.2.18). Contribution Check now passing (was report_incomplete).

Note: Shared alerts reference "#25548 DDG (Design Decision Gate)" - but #25548 is actually "AWF Diagnostic Logs feat request". Shared alerts metadata may be stale/mismatched.

Issues this cycle: ~15 stale smoke test failures from Apr 8-11 still open (#25371, #25372, #25374, #25395, #25415, #25727, etc.)
PRs merged: 20 (Apr 12-13) | Engine split: copilot:~124wf, claude:~41wf, codex:~18wf

Actions: Weekly discussion created. No new improvement issues (existing tracking in place).
