# Copilot CLI Research Notes

### 2026-05-06 (Run 25416993511) — Run #16
- 214 total MD workflows; 93 Copilot (simple form)
- **mcp-session-timeout**: 0 (16th consecutive run — CRITICAL persistent gap)
- **mcp-tool-timeout**: 0 (16th run — persistent gap)
- **engine.api-target**: 0 (persistent)
- **engine.harness**: 0 (persistent)
- **max-continuations**: 0 in copilot wfs (test-quality-sentinel+smoke-copilot are the only users, neither filtered as copilot simple form)
- **engine.model**: 12 workflows using model overrides (gpt-5.4-mini×3, claude-sonnet×2, claude-haiku×2, gpt-5, gpt-5-mini, gpt-4.1, copilot/gpt-5)
- **engine.agent (custom)**: 7 copilot workflows using custom agent files
- **engine.bare**: 0 in copilot wfs (9 total across engines)
- **cache-memory**: 29/93 copilot wfs (31%)
- **repo-memory**: 18/93 (19%)
- **web-fetch**: 8/93 (9%)
- **sandbox AWF**: 11/93 (12%); **SRT**: 0
- **strict: true**: 56/93 (60%); **strict: false**: 4/93 (4%); not set: 33/93 (35%)
- **network config**: 45/93 (48%)
- **copilot-requests**: 38/93 (41%)
- **tracker-id**: 37/93 (40%)
- **timeout-minutes**: 91/93 (98%)
- 5 unused agent files: grumpy-reviewer, w3c-specification-writer, create-safe-output-type, custom-engine-implementation, interactive-agent-designer
- Discussion created: "Copilot CLI Deep Research - 2026-05-06"

### 2026-05-04 (Run 25301640113)
- 211 total MD workflows; 95 Copilot (simple form)
- **startup-timeout**: 0 (14th consecutive run — CRITICAL persistent gap)
- **tool-timeout**: 0 (14th run — persistent gap)
- **engine.api-target**: 0 (persistent gap)
- **engine.harness**: 0 (persistent gap)
- **max-continuations**: 2 (test-quality-sentinel:40, smoke-copilot:2)
- **engine.version pinning**: 6 (slight uptick vs 0 in older runs)
- **engine.model**: 4 (gpt-5-mini for auto-triage-issues)
- **engine.bare**: 2 (smoke-copilot + 1 more)
- **cache-memory**: 30 (solid adoption)
- **mcp-scripts**: 1 (daily-performance-summary only)
- **engine.agent (custom)**: 7 unique workflows
- **sandbox AWF**: 11; **network config**: 45
- 5/11 custom agent files still unused (grumpy-reviewer, w3c-spec-writer, create-safe-output-type, custom-engine-implementation, interactive-agent-designer)
- Discussion created: "Copilot CLI Deep Research - 2026-05-04"



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

### 2026-05-05 (Run 25358259379) — Run #15
- 213 total MD workflows; 92 Copilot (simple form)
- **startup-timeout**: 0 (15th consecutive run — CRITICAL persistent gap)
- **tool-timeout**: 0 (15th run — persistent gap)
- **engine.version pinning**: 0 (15th run)
- **engine.api-target**: 0 (persistent)
- **engine.harness**: 0 (persistent)
- **max-continuations**: 2 (test-quality-sentinel:40, smoke-copilot:2) — stable
- **engine.model**: 20 workflows (↑↑↑ from 4 — surge in model diversity)
- **engine.agent**: 34 workflows (↑↑↑ from 7 — custom agent adoption surging)
- **cache-memory**: 88 workflows (near-universal, methodology improved)
- **sandbox AWF**: 18 (↑ from 11)
- **strict mode**: 126 (expanded scope measurement)
- **web-fetch**: 23; **web-search**: 1 (ci-doctor only)
- **mcp-scripts**: 3 (↑ from 1)
- **GitHub MCP full-access**: 13 prod workflows without toolsets restriction
- 5 custom agent files still unused (unchanged since tracking began)
- Discussion created: "Copilot CLI Deep Research - 2026-05-05"
