# Agent Performance - 2026-04-18
Run: §24596972251 | Q:74↓3 E:73↓2

## Ecosystem Overview (Apr 17-18)
- Overall success rate: ~83% (stable from Apr 17)
- 194 workflows total (stable)
- Infrastructure regressions causing cascade failures (-3Q, -2E)

## Top Performers
1. **[aw] Failure Investigator** (Q:90 E:85) - Outstanding RCA quality: #26876, #26970
2. **Agentic Maintenance** (Q:87 E:100) - 9/9 runs, 100% success
3. **DeepReport** (Q:87 E:100) - High-quality strategic issues #26895/#26894/#26893
4. **Issue Monster** (Q:88 E:95) - 20/21 runs, 95% success
5. **Architecture Guardian** (Q:82 E:100) - 1/1 run, 100%
6. **Workflow Health Manager** (Q:82 E:100) - 1/1 run, 100%, excellent reports

## Watch List
- **Agent Persona Explorer** (Q:40 E:0) - NEW: 100% failure today; 1.68M tokens wasted
- **AI Moderator** (Q:45 E:33) - Codex 401 (#26929) + Copilot failures (#26911)
- **Test Quality Sentinel** (Q:65 E:85) - Turn drift 4–18 (avg 8.7); unstable prompts
- **Daily Community Attribution** (Q:55 E:50) - 50% failure rate (#26848)
- **Daily Fact About gh-aw** (Q:40 E:0) - 10+ consecutive Codex failures (#26852)

## P0 Infrastructure Issues (NEW)
- **Node.js binary not found** (#26876) - AWF v0.25.23 regression; GH_AW_NODE_BIN path missing in container
- **Copilot CLI shell permission blocks safeoutputs** (#26970) - Agents timeout after full execution; MCP tool needed

## Tool Upgrade Available
- Copilot CLI 1.0.21 → 1.0.32 (#26977 open — 11 versions behind!)
- Claude Code 2.1.112 → 2.1.114 (#26977 same PR)

## Issues Created This Run
- No new issues created (all patterns already tracked in #26876, #26970, #26977)

## Key Findings
- 2 infrastructure bugs explain ~8 of Apr 17 cascade failures (not agent quality issues)
- Codex engine broadly unreliable: 401 auth errors affecting multiple workflows
- 0/13 runs today used write-capable safe outputs (low-actuation morning session)
- Claude engine: 0 errors, exploratory style; Copilot: 2 errors, permission issues

Last updated: 2026-04-18T04:35Z by agent-performance-manager
