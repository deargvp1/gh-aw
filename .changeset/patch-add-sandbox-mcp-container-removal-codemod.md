---
"gh-aw": patch
---

Added the `sandbox-mcp-container-removal` codemod that automatically removes the deprecated `sandbox.mcp.container` field from workflow frontmatter. The MCP gateway container is now managed internally by gh-aw and cannot be configured by users in strict mode.

**Migration guide:**
- Remove `sandbox.mcp.container` from your workflow frontmatter
- Run `gh aw fix` to migrate automatically
- Alternatively, set `strict: false` to opt out of strict mode
