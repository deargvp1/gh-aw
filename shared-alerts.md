# Shared Alerts — 2026-04-18T04:35Z

## P0 (Critical)
- **Copilot CLI shell permission blocks safeoutputs** (#26970, OPEN TODAY): Agents complete correctly but timeout emitting noop. Affects Daily Safe Output Integrator, Daily Project Performance. Fix: use MCP noop tool, not shell CLI.
- **Node.js binary not found in container** (#26876, OPEN Apr 17): AWF v0.25.23 regression; GH_AW_NODE_BIN toolcache path missing. Affects Daily Issues Report Generator, Daily News.

## P1 (High)
- **Copilot CLI 11 versions behind** (#26977 OPEN TODAY): v1.0.21 active; v1.0.32 available. Claude Code 2.1.114 also available. Upgrade needed.
- **AI Moderator codex 401 auth** (#26929, OPEN): OPENAI_API_KEY invalid. Also affects Daily Fact About gh-aw.
- **Agent Persona Explorer** (NEW WATCH): 100% failure today, 1.68M tokens, no output. Investigate.
- **Test Quality Sentinel turn drift** (NEW WATCH): 4–18 turns variance. Prompts unstable.
- **Daily Community Attribution** (#26848, OPEN): 50% failure, Copilot crash during README write.
- **Smoke Claude** (#26777, #26790 OPEN): ~60% failure. No recent schedule data.
- **Smoke Gemini** (#26980 OPEN TODAY): New auto-issue; API key context unclear.
- **Smoke Crush** (#26979 OPEN TODAY): New failure.

## P2 (Watch)
- **PR Triage Agent** (#26778 OPEN): 67% success rate.
- **Auto-Triage Issues** (#26364 OPEN): 67% success, intermittent.
- **MCP Rate Limit** (#26239 OPEN): Circuit breaker needed.
- **GitHub MCP get_me 403 errors** (#26458 OPEN): Auth/permission failures.

## Recoveries (Apr 17-18)
- ✅ Daily Issues Report Generator (#26393): Closed not_planned by pelikhan
- ✅ Smoke Gemini (#26351): Closed not_planned by pelikhan (API key invalid)
- ✅ Smoke Codex: Recovered
- ✅ Compilation: 0 stale lock files (was 16 on Apr 16)
- ✅ GitHub Remote MCP Auth Test: Recovered

## Engine/Tool Status
- Copilot v1.0.21 active / v1.0.32 available (#26977 open)
- Claude Code 2.1.114 available (#26977 same PR)
- Codex: 401 auth failures (OPENAI_API_KEY) — avoid for new workflows
- Gemini: API key invalid (closed not_planned)

## Ecosystem State
- 194 workflows total (stable)
- 0 stale lock files ✅
- Schedule success rate: ~83% (stable)
- P0 failures: 2 (new infra bugs; down from 2 on Apr 16 but different issues)
- Overall quality trend: ↓ temporarily (infrastructure, not agent quality)

## Orchestrator Summaries
- Agent Performance (Apr 18 04:35Z): Q:74 E:73. 83% success rate. 2 infra P0s. Copilot 1.0.32 upgrade critical.
- Workflow Health (Apr 17 12:10Z): Score 73/100. 194 workflows. 0 stale lock files.

Last updated: 2026-04-18T04:35Z by agent-performance-manager
