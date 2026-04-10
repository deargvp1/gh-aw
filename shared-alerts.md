# Shared Alerts — 2026-04-10T04:45Z

## P1 (Critical) - RECOVERING
- **Copilot Engine Crash** (v1.0.21 regression): Started Apr 8 01:02 UTC, 35h+ duration. CLI Version Checker identified fix: upgrade to v1.0.22 (PR #25577 in progress Apr 10). Recovery expected within 24h. Monitor: #25215, #25396, #25374, #25257, #25372.

## P2 (High)
- **Documentation Unbloat cost** (new, Apr 10): Claude workflow running 11x/week at $4.97/run ≈ $55/week. No safe outputs detected. Investigate ROI and add cost guards.
- **Design Decision Gate failures** (#25548): 2/3 runs failing with errors. Gates architecture decisions — failures block reviews.

## P3 (Watch)
- **Contribution Check report_incomplete**: Workflow completing with only `report_incomplete`. May have permission/network issues accessing PR data.

## Recent Fixes
- CLI proxy policy fix (#25419, f0b0d232, Apr 9): Adds default CLI_PROXY_POLICY. Deployed.
- #25022 AI Moderator missing_data: CLOSED not_planned Apr 9
- #24718 Duplicate Code Detector: CLOSED not_planned Apr 6
- #24829 GitHub Remote MCP Auth: CLOSED not_planned Apr 7
- Copilot CLI v1.0.22 fix: In progress (PR #25577, Apr 10)

## Active Failure Issues (20+)
Key open: #25215, #25396, #25374, #25290, #25261, #25260, #25257, #25276, #25372, #25447, #25440, #25415, #25398, #25395, #25384, #25305, #25315, #25312, #25259, #25236, #25287

## Ecosystem State
- 187 compiled workflows. Engine split: ~124 copilot, ~41 claude, ~18 codex, ~4 others
- Copilot crash distorted all 7d metrics — quality scores temporarily depressed
- Claude/Codex engines showed 100% resilience during crash window

Last updated: 2026-04-10T04:45Z by agent-performance-analyzer
