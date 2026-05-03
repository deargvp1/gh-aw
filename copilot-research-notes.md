# Copilot CLI Research Notes

## Analysis History

### 2026-04-29 (Run 25134300030)
- 205 total MD workflows; 110 Copilot (improved counting: simple form 89 + object form 21)
- **startup-timeout**: 0 (11th consecutive run — persistent gap — CRITICAL)
- **tool-timeout**: 0 (persistent)
- **engine.version**: 0 (persistent)
- **sandbox AWF**: 17 (up from 11 prior run) — slowly growing
- **cache-memory**: 79 (all forms counted — large jump due to methodology fix)
- **repo-memory**: 23 (new metric tracked)
- **mcp-scripts**: 6 (stable)
- **engine.agent**: 22 (custom agents up significantly — includes awf + custom files)
- **engine.model**: 10 (stable — gpt-5-mini, claude-haiku-4.5, etc.)
- **max-continuations**: 2 (stable — persistent gap)
- **web-search**: 2, **web-fetch**: 19 (stable)
- 5 unused custom agent files: grumpy-reviewer, w3c-specification-writer, create-safe-output-type, custom-engine-implementation, interactive-agent-designer
- Discussion created: "Copilot CLI Deep Research - 2026-04-29"

### Earlier Runs (Apr 16–26): Persistent gaps startup-timeout (0%), tool-timeout (0%), engine.version (0%), api-target (0%), max-continuations (~2). AWF sandbox grew ~10→15. mcp-scripts 0→6 (may be non-Copilot). 5/11 custom agent files unused throughout.

### 2026-04-25 (Run 24940623939) [CONDENSED]
- 202 total; 91 copilot; max-continuations: 2; startup-timeout: 0 (10th run); mcp-scripts: 6

### Apr 21–28 [CONDENSED]
- Persistent gaps confirmed (7-11th runs): startup-timeout 0%, tool-timeout 0%, engine.version 0%, api-target 0%, max-continuations ~2
- model overrides 0-10 (methodology dep); AWF sandbox 11-17; mcp-scripts 0-6 (non-Copilot scope)
- 5/11 custom agent files unused across all runs

### 2026-05-01 (Run 25213682014)
- 205 total MD workflows; 110 Copilot (89 simple form + 21 object form)
- **startup-timeout**: 0 (12th consecutive run — CRITICAL persistent gap)
- **tool-timeout**: 0 (12th run — persistent gap)
- **engine.version pinning (Copilot)**: 0 (runtimes pin node/python/etc versions, not engine)
- **bare mode**: 8 workflows (smoke-copilot, daily-*, hippo, poem-bot, constraint-solving)
- **max-continuations**: 2 workflows (test-quality-sentinel:40, smoke-copilot:2)
- **sandbox AWF**: ~17 workflows
- **cache-memory**: 62 workflows
- **web-fetch**: 19; **web-search**: 2
- **mcp-scripts**: 1
- **safe-outputs**: 162 occurrences
- **github MCP tool**: 144 workflows (dominant pattern)
- **playwright**: 13 workflows
- **engine.agent**: 11 actual custom agent files used (not counting `agent: awf`)
  - Used: adr-writer(1), agentic-workflows(2), ci-cleaner(1), contribution-checker(1), developer.instructions(1), technical-doc-writer(2)
  - UNUSED: grumpy-reviewer, w3c-specification-writer, create-safe-output-type, custom-engine-implementation, interactive-agent-designer
- **engine.model**: 6 workflows using model overrides
- **network config**: 104 workflows (good adoption)

### 2026-05-03 (Run 25270216421)
- 209 total MD workflows; 91 Copilot (simple form only)
- **startup-timeout**: 0 (13th consecutive run — CRITICAL persistent gap)
- **tool-timeout**: 0 (13th run)
- **engine.version**: 0 (persistent)
- **engine.model**: 11 workflows (gpt-5, gpt-5-mini, gpt-4.1-mini, gpt-5.4-mini, claude variants)
- **bare mode**: 9 workflows (up from 8)
- **max-continuations**: 2 workflows (test-quality-sentinel:40, smoke-copilot:2)
- **sandbox AWF**: 15 workflows (up from 11, trending up)
- **sandbox SRT**: 0 (never used)
- **cache-memory**: 19 copilot workflows
- **repo-memory**: 14 copilot workflows
- **web-fetch**: 7, **web-search**: 0 in copilot workflows
- **mcp-scripts**: 1
- **gh-proxy mode**: 105 across all workflows
- **specific toolsets**: ~10 workflows
- **imports (agentic-workflows)**: 40 workflows
- 5 unused custom agent files (same 5 since April)
- Discussion created: "Copilot CLI Deep Research - 2026-05-03"
