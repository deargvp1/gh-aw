---
title: Tools
description: Configure GitHub API tools, browser automation, and AI capabilities available to your agentic workflows, including GitHub tools and custom MCP servers.
sidebar:
  order: 700
---

[Tools](/gh-aw/reference/glossary/#tools) are defined in the frontmatter to specify which GitHub API calls, browser automation, and AI capabilities are available to your workflow:

```yaml wrap
tools:
  edit:
  bash: true
```

Some tools are available by default. All tools declared in imported components are merged into the final workflow.

## Built-in Tools

### Edit Tool (`edit:`)

Allows file editing in the GitHub Actions workspace.

```yaml wrap
tools:
  edit:
```

The edit tool also accepts a boolean or a GitHub Actions expression string (for `workflow_call` reusable workflows):

```yaml wrap
tools:
  edit: true                      # Always enabled
  edit: false                     # Explicitly disabled
  edit: ${{ inputs.enable-edit }} # Enabled at runtime by caller
```

> [!NOTE]
> When `tools.edit` is an expression, file-path access (`--allow-all-paths`) is still granted at compile time. Only write permission itself is gated at runtime when the expression resolves to `"true"`.

### GitHub Tools (`github:`)

Configure GitHub API operations including toolsets, remote/local modes, and authentication.

```yaml wrap
tools:
  github:
    toolsets: [repos, issues]
```

The `toolsets` field also accepts a GitHub Actions expression string, enabling reusable workflows to expose toolset selection to callers:

```yaml wrap
tools:
  github:
    toolsets: ${{ inputs.github-toolsets }}  # Comma-separated list at runtime
```

> [!NOTE]
> When `toolsets` is an expression, compile-time toolset validation and required-permissions checks are skipped. Ensure the compiled workflow has sufficient `permissions:` for the toolsets that callers may provide. The GitHub MCP server enforces toolset restrictions at runtime.

See **[GitHub Tools Reference](/gh-aw/reference/github-tools/)** for complete configuration options.

### Bash Tool (`bash:`)

Enables shell command execution in the workspace. Defaults to safe commands (`echo`, `ls`, `pwd`, `cat`, `head`, `tail`, `grep`, `wc`, `sort`, `uniq`, `date`).

```yaml wrap
tools:
  bash:                              # Default safe commands
  bash: []                           # Disable all commands
  bash: ["echo", "ls", "git status"] # Specific commands only
  bash: [":*"]                       # All commands (use with caution)
```

Use wildcards like `git:*` for command families or `:*` for unrestricted access.

The `bash` field also accepts a GitHub Actions expression string for reusable workflow parameterization:

```yaml wrap
tools:
  bash: ${{ inputs.bash-allowlist }}  # Comma-separated commands at runtime
```

> [!NOTE]
> When `bash` is an expression, the resolved value must be a comma-separated list of command names (e.g., `"git,npm,echo"`). Shell access is **fail-closed**: if the expression resolves to an empty string or is unset, no shell commands are permitted. Newline-separated command lists are also accepted.

### Web Tools

Enable web content fetching and search capabilities:

```yaml wrap
tools:
  web-fetch:   # Fetch web content
  web-search:  # Search the web (engine-dependent)
```

**Note:** Some engines require third-party Model Context Protocol (MCP) servers for web search. See [Using Web Search](/gh-aw/guides/web-search/).

For the **Codex** engine, `web-search:` is disabled by default. Web search is only enabled when `web-search:` is explicitly declared in the `tools:` block. Without this declaration, Codex runs with `-c web_search="disabled"` and cannot access the web.

### Playwright Tool (`playwright:`)

Configure Playwright for browser automation and testing:

```yaml wrap
tools:
  playwright:
    version: "1.56.1"  # Optional: specify version
```

See **[Playwright Reference](/gh-aw/reference/playwright/)** for complete configuration options, network access, browser support, and example workflows.

### Cache Memory (`cache-memory:`)

Persistent memory storage across workflow runs for trends and historical data.

```yaml wrap
tools:
  cache-memory:
```

See **[Cache Memory Reference](/gh-aw/reference/cache-memory/)** for complete configuration options and usage examples.

### Repo Memory (`repo-memory:`)

Repository-specific memory storage for maintaining context across executions.

```yaml wrap
tools:
  repo-memory:
```

See **[Repo Memory Reference](/gh-aw/reference/repo-memory/)** for complete configuration options and usage examples.

### QMD Documentation Search (`qmd:`) — Experimental

Build a local vector search index over documentation files and expose it as an MCP search tool. The index is built in a dedicated indexing job (no `contents: read` needed in the agent job):

```yaml wrap
tools:
  qmd:
    checkouts:
      - pattern: "docs/**/*.md"
```

See **[QMD Reference](/gh-aw/reference/qmd/)** for complete configuration options, checkout support, GitHub search integration, and cache key usage.

### Introspection on Agentic Workflows (`agentic-workflows:`)

Provides workflow introspection, log analysis, and debugging tools. Requires `actions: read` permission:

```yaml wrap
permissions:
  actions: read
tools:
  agentic-workflows:
```

See [GH-AW as an MCP Server](/gh-aw/reference/gh-aw-as-mcp-server/) for available operations.

## Tool Timeout Configuration

### Tool Operation Timeout (`tools.timeout`)

Sets the per-operation timeout in seconds for tool and MCP server calls. Applies to all tools and MCP servers when supported by the engine. Defaults vary by engine (Claude: 60 s, Codex: 120 s).

```yaml wrap
tools:
  timeout: 120   # seconds
```

### MCP Server Startup Timeout (`tools.startup-timeout`)

Sets the timeout in seconds for MCP server initialization. Default is 120 seconds.

```yaml wrap
tools:
  startup-timeout: 60   # seconds
```

Both fields accept either an integer or a GitHub Actions expression string, enabling `workflow_call` reusable workflows to parameterize these values:

```yaml wrap
tools:
  timeout: ${{ inputs.tool-timeout }}
  startup-timeout: ${{ inputs.startup-timeout }}
```

> [!NOTE]
> Expression values are passed through environment variables in the compiled workflow. TOML-based engine configs (Codex MCP gateway) fall back to engine defaults when an expression is used, since TOML has no expression syntax.

## Parameterizing Tool Configuration in Reusable Workflows

For `workflow_call` reusable workflows, you can expose selected tool policies as inputs. This allows callers to choose their bash allowlist, GitHub MCP toolsets, or edit tool availability without maintaining separate workflow files.

### Supported Expression Fields

The following tool fields accept GitHub Actions expression strings in addition to their literal types:

| Field | Expression Example | Resolved Format |
|---|---|---|
| `tools.bash` | `${{ inputs.bash-allowlist }}` | Comma- or newline-separated command list |
| `tools.edit` | `${{ inputs.enable-edit }}` | `"true"` to enable, anything else to disable |
| `tools.github.toolsets` | `${{ inputs.github-toolsets }}` | Comma-separated toolset names |

### Example: Reusable Workflow with Tool Inputs

```yaml wrap
# Reusable workflow: .github/workflows/my-workflow.md
on:
  workflow_call:
    inputs:
      bash-allowlist:
        type: string
        description: Comma-separated list of allowed bash commands (e.g. "git,npm,echo")
      github-toolsets:
        type: string
        description: Comma-separated list of GitHub MCP toolsets (e.g. "repos,issues")
      enable-edit:
        type: boolean
        description: Whether to enable the edit tool
        default: false
permissions:
  contents: read
engine: copilot
tools:
  bash: ${{ inputs.bash-allowlist }}
  github:
    toolsets: ${{ inputs.github-toolsets }}
  edit: ${{ inputs.enable-edit }}
```

### Security Model

Expression-based tool configuration maintains fail-closed behavior:

- **`tools.bash`**: If the expression resolves to empty or is not provided, no shell commands are permitted. If the expression resolves to a list, only those commands are allowed.
- **`tools.edit`**: Only enabled when the expression resolves exactly to `"true"`. Any other value (including empty) keeps the edit tool disabled.
- **`tools.github.toolsets`**: The GitHub MCP server enforces toolset restrictions at runtime. Compile-time permission validation is skipped for expressions; ensure the workflow has sufficient `permissions:` for the toolsets callers may provide.

### Literal Configs Remain Valid

All existing literal tool configurations continue to work unchanged:

```yaml wrap
tools:
  bash: ["git", "npm"]          # Literal allowlist
  edit:                          # Always enabled
  github:
    toolsets: [repos, issues]   # Literal toolsets
```

## Custom MCP Servers (`mcp-servers:`)

Integrate custom Model Context Protocol servers for third-party services:

```yaml wrap
mcp-servers:
  slack:
    command: "npx"
    args: ["-y", "@slack/mcp-server"]
    env:
      SLACK_BOT_TOKEN: "${{ secrets.SLACK_BOT_TOKEN }}"
    allowed: ["send_message", "get_channel_history"]
```

**Options**: `command` + `args` (process-based), `container` (Docker image), `url` + `headers` (HTTP endpoint), `registry` (MCP registry URI), `env` (environment variables), `allowed` (tool restrictions). See [MCPs Guide](/gh-aw/guides/mcps/) for setup.

### Registry Field

The `registry` field specifies the source URI of an MCP server in a registry. It is informational — useful for documenting server origin and enabling registry-aware tooling — and does not affect execution. gh-aw does not enforce registry usage. Works with both stdio and HTTP servers:

```yaml wrap
mcp-servers:
  filesystem:
    registry: "https://api.mcp.github.com/v0/servers/modelcontextprotocol/filesystem"
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem"]
```

## Related Documentation

- [GitHub Tools](/gh-aw/reference/github-tools/) - GitHub API operations, toolsets, and modes
- [Playwright](/gh-aw/reference/playwright/) - Browser automation and testing configuration
- [Cache Memory](/gh-aw/reference/cache-memory/) - Persistent memory across workflow runs
- [Repo Memory](/gh-aw/reference/repo-memory/) - Repository-specific memory storage
- [QMD Documentation Search](/gh-aw/reference/qmd/) - Vector similarity search over documentation files
- [MCP Scripts](/gh-aw/reference/mcp-scripts/) - Define custom inline tools with JavaScript or shell scripts
- [Frontmatter](/gh-aw/reference/frontmatter/) - All frontmatter configuration options
- [Network Permissions](/gh-aw/reference/network/) - Network access control for AI engines
- [MCPs](/gh-aw/guides/mcps/) - Complete Model Context Protocol setup and usage
