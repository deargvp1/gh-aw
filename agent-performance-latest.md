# Agent Performance - 2026-04-10
Run: §24226598383 | Q:65↓3 E:66↓1

Top: AI Moderator (Q:88 E:92 - Codex, stable), Smoke Claude (Q:90 E:87), CLI Version Checker (Q:86 E:88 ↑2), Issue Monster (Q:78 E:80)
Watch: Documentation Unbloat ($4.97/run × 11 runs = ~$55/week, 0 safe outputs), Design Decision Gate (2/3 failures, issue #25548), Contribution Check (report_incomplete every run)

Systemic: Copilot engine crash (Apr 8-9, v1.0.21 regression) caused 100+ zero-output runs. Fix: v1.0.22 (CLI Version Checker tracked in PR #25577). Recovery expected Apr 10+.

Engine dist (recent runs): copilot:19 runs/15wf, claude:7/5, codex:7/3, gemini:1/1
Cost risk: Documentation Unbloat ~$55/wk. Agent Persona Explorer 4.5M tokens but stable at 14 turns.

Stats: 187 compiled wfs. 150 runs 7d. 17 safe items. $6.75 total. Crash window heavily distorted metrics.
Actions: Weekly discussion created. No new issues (existing tracking: #25548, #25215+).
