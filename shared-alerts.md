# Shared Alerts — 2026-04-22T04:37Z

## P0 (Critical)
- None currently

## P1 (High)
- **Stale lock files — Codex 401 auth loop** (#27724 + #27731 OPEN, Apr 22): PR #27711 merged openai-proxy config but lock files not recompiled → Codex workflows use old config → 401 at api.openai.com. Fix: `make recompile`. Blocks: AI Moderator, Smoke Codex, Changeset Generator, Daily Obs Report.
- **Safe outputs "session not found" at 37min** (#27755 NEW, Apr 22, @dsyme): MCP server returning session not found at 37min (not just 1h+). All long-running workflows at risk.
- **Design Decision Gate push bundle failure** (#27756 OPEN, Apr 22): `push_to_pull_request_branch: Failed to apply bundle` — NEW failure mode on top of max_turns=5 issue. 50% success rate today (2/4 runs).
- **Design Decision Gate max_turns=5** (#27470 OPEN, Apr 21): ADR generation requires ≥6 turns; 5-turn hard limit makes it structurally impossible.
- **node: command not found on aw-gpu-runner-T4** (#27534 OPEN): Recurring. Node.js PATH not available in bash on GPU runners.
- **GitHub App rate limit exhaustion** (#27251 OPEN): Co-scheduled at 23:44 UTC.
- **CODEX_HOME variable collision** (#27512 OPEN, Apr 21): cp same-file error breaks Codex workflows with MCP config.
- **Smoke Claude** (#27030 OPEN): Ongoing; 39T+ run in progress today
- **Smoke Copilot** (#27028 OPEN): Ongoing since Apr 14

## P2 (Watch)
- **Documentation Unbloat cost drain** (#27600 OPEN): 56T run today, ROI unclear; cap recommendation pending
- **Agent Persona Explorer turn count** - improved 16T today (was 42T) — monitoring
- **Safe Outputs SEC-004** (#27235 OPEN): 4 handler files need sanitization
- **Performance regressions** (#27280/#27279/#27278 OPEN)
- **dev-hawk github-env** (#26933): High severity zizmor finding
- **PR Triage Agent** (#26778 OPEN): 67% success rate
- **MCP gateway long-running drops** (#23153 OPEN): Session not found after 30-45min (now confirmed shorter: #27755)
- **Copilot reviewer fan-out** (#27130 OPEN): 6 review workflows per Copilot PR push

## Resolved (Recent)
- GitHub Remote MCP Authentication Test ✅ RESOLVED Apr 22 (was persistent failure)
- Codex 401 auth root cause IDENTIFIED (#27729) — fix pending recompile
- Smoke OpenCode ✅ NEW engine working

## Ecosystem State
- 198+ workflows (new ones being added)
- Stale lock files: 15+ (from #27563 Apr 21, likely more after PR #27711)
- Schedule success rate: ~75% (stale locks + engine failures)
- P0 failures: 0
- P1 failures: stale locks (Codex), safe outputs session timeout, Design Decision Gate, Smoke Copilot/Claude
- Overall quality trend: Q:71 (↓-1 from 72)

## Orchestrator Summaries
- Agent Performance (Apr 22 04:37Z): Q:71 E:67. 18 workflows, 29 runs. Stale lock files P1. Safe outputs 37min threshold P1. Design Decision Gate push failure P1. GitHub Remote MCP Auth RESOLVED.
- Workflow Health (Apr 21 12:14Z): Score 70/100. 198 workflows. 15 stale locks. MCP Gateway P1 (codex+CLI servers). node not found GPU runner P1.
- Agent Performance (Apr 21 04:45Z): Q:72 E:68. 25 workflows, 31 runs. DDG max_turns P1. Docs Unbloat 0-output cost drain P2.
- Workflow Health (Apr 20 12:14Z): Score 73/100. 197 workflows. 0 stale locks. node not found on GPU runner P1.

Last updated: 2026-04-22T04:37Z by agent-performance-manager
