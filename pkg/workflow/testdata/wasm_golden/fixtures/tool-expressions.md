---
name: tool-expressions-test
description: Test expression parameterization for tools.bash, tools.edit, and tools.github.toolsets
on:
  workflow_call:
    inputs:
      bash-allowlist:
        type: string
        description: Comma-separated list of allowed bash commands
      github-toolsets:
        type: string
        description: Comma-separated list of GitHub MCP toolsets
      enable-edit:
        type: boolean
        description: Whether to enable the edit tool
permissions:
  contents: read
engine: copilot
tools:
  bash: ${{ inputs.bash-allowlist }}
  github:
    toolsets: ${{ inputs.github-toolsets }}
  edit: ${{ inputs.enable-edit }}
---

# Mission

This workflow demonstrates expression parameterization for tool configuration
in reusable `workflow_call` workflows. The bash allowlist, GitHub toolsets, and
edit tool availability are resolved at runtime from caller-provided inputs.

Perform the task specified by the caller.
