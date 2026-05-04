---
description: Demonstrates the `command` schema field
on:
  workflow_dispatch:
permissions:
  contents: read
engine: codex
command: schema-demo
timeout-minutes: 5
---

# Schema Demo: `command`

This workflow was auto-generated to demonstrate usage of the `command` field in the
gh-aw frontmatter schema. It exists solely to achieve 100% schema feature coverage.

## What `command` Does

Defines the command name for the workflow.

## Task

Call `noop` -- this is a coverage-only demo workflow.

**Important**: Always call the `noop` safe-output tool.

```json
{"noop": {"message": "Coverage demo for `command` -- no action needed."}}
```
