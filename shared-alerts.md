# Shared Alerts — 2026-05-05T13:05Z

## P0 (Critical)
- **Smoke Gemini** (#30175, #29666, #30242 OPEN): 100% failure, proxy architecture blocks all agent traffic. 31+ days unresolved.
- **Smoke CI** (#29666 OPEN): CGO/EROFS persistent, 100% action_required.
- **Daily Model Inventory Checker** (#30043 OPEN): Copilot CLI silent startup crash.
- **APM Unpack Systemic Failure** (#30252 OPEN): apm-default.tar.gz exits code 1, affects multiple workflows. Both firewall v0.25.35 and v0.25.38 affected.

## P1 (High)
- **config.models unsupported field** (#30307 OPEN): NEW May 5 — blocked 10 smoke runs in config.models sweep.
- **Smoke macOS ARM64**: 100% failure since 2026-02-20. NO ISSUE FILED — needs one urgently.
- **Smoke Claude** (#30241 OPEN): APM unpack failure
- **GitHub MCP Structural Analysis** (#30347, #30144 OPEN): claude engine auth crash, recurring
- **Auto-Triage Issues** (#30205 OPEN)
- **MCP gateway session timeout** (#23153 OPEN): Long-running workflows at risk.
- **Performance Regression in Validation** (#30180): 82.1% slower
- **Safeoutputs omission** (#30102 OPEN): Schema Consistency Checker + Multi-Device Docs Tester ($3-4/day wasted)

## P2 (Watch)
- **Node.js 20 deprecation** in CI: deadline Sep 16, 2026. Migrate to Node.js 22.
- **6 PR-review agents** on same triggers — 100+ action_required runs/day (Scout, Archie, /cloclo, Q, AI Moderator, Content Moderation, Grumpy, Security, PR Nitpick)
- **Safe Outputs Conformance** (#30085, #30086, #30087 OPEN): SEC-002, SEC-003, SEC-005

## Resolved (Do Not Re-File)
- #29863 Smoke Copilot regression → RECOVERED ✅
- #29088 Codex crash → CLOSED
- #28659 Doc Unbloat claude auth → CLOSED
- #27965 GitHub Remote MCP Auth → CLOSED

## Trends
- 213 workflows, 0 missing lock files
- Health: 63/100 (↓-2, APM regression)
- Quality: 74/100 stable (4-day plateau)
- Effectiveness: 71/100 stable (4-day plateau)
- Copilot coding agent active: 5 PRs opened May 5
