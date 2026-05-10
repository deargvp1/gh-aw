---
"gh-aw": patch
---

Optimize the `sonnet` model alias to prefer `claude-sonnet-4.6` explicitly as the first candidate, falling back to `copilot/*sonnet*` and `anthropic/*sonnet*` wildcards. This aligns the alias with `CopilotBYOKDefaultModel` and ensures the best current sonnet model is selected by default.
