# Copilot CLI Research Notes

## Analysis History

### 2026-04-18 (Run 24613840109)
- 195 total MD workflows, 86 using engine: copilot
- Key persistent gaps: version pinning (0%), api-target (0%), token-weights (0%), block-domains (0%)
- Improvements: cache-memory adoption growing (21%→30%), max-continuations up (2→3 workflows)
- 11 custom agent files available; ~8 still unused

### 2026-04-17 (Run 24586698669)
- 85 copilot workflows tracked
- Notable: engine.args/env at 0%, mcp-gateway at 0%

### 2026-04-16 (Run 24534029243)
- 192 total, 90 explicit copilot + 26 default = 116 effective
- playwright regression: 20→12 (-40%)
- strict_mode: 111→126 (+13%)

## Persistent Opportunities (Not Addressed in 3+ Runs)

1. **engine.version**: Never used → stability risk
2. **engine.api-target**: Never used → GHEC/GHES teams can't use this
3. **token-weights**: Never used → no custom cost modeling
4. **block-domains**: Never used → missed defense-in-depth
5. **mcp-scripts**: 1 workflow → underutilized dynamic MCP capability
6. **max-continuations**: 2-3 only → Copilot-unique autopilot underused

## Recommendations Tracking

| Recommendation | Status | Date Added |
|---|---|---|
| Use engine.version pinning for reproducibility | ⏳ Pending | 2026-04-16 |
| Expand max-continuations for complex workflows | ⏳ Pending | 2026-04-16 |
| Use bare:true for simple/creative workflows | ⏳ Pending | 2026-04-16 |
| Add network.blocked for defense-in-depth | ⏳ Pending | 2026-04-17 |
| Activate unused agent files | ⏳ Pending | 2026-04-16 |
| Model override for cost optimization | ⏳ Pending | 2026-04-17 |
| Cache-memory adoption growing | ✅ Improving | 2026-04-16 |
| Custom agent adoption | ✅ Improving | 2026-04-17 |
