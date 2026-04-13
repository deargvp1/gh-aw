---
title: GitHub Integrity Filtering
description: How integrity filtering restricts agent access to GitHub content based on author trust and merge status, and how filtered events appear in logs.
sidebar:
  order: 680
---

Integrity filtering (`tools.github.min-integrity`) controls which GitHub content an agent can access during a workflow run. Rather than filtering by permissions, it filters by **trust**: the author association of an issue, pull request, or comment, and whether that content has been merged into the main branch.

## How It Works

The MCP gateway intercepts tool calls and removes items below the configured integrity minimum before the agent sees them. Filtered items are logged as `DIFC_FILTERED` events for later inspection.

## Configuration

Set `min-integrity` under `tools.github` in your workflow frontmatter:

```aw wrap
tools:
  github:
    min-integrity: approved
```

`min-integrity` can be specified alone. When `allowed-repos` is omitted, it defaults to `"all"`. If `allowed-repos` is also specified, both fields must be present.

```aw wrap
tools:
  github:
    allowed-repos: "myorg/*"
    min-integrity: approved
```

## Configuration Reference

All integrity-filtering inputs are specified under `tools.github` in your workflow frontmatter. The table below summarizes every available field:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `min-integrity` | string | Yes (when any guard policy field is used) | `approved` for public repos; none for private | Minimum integrity level: `merged`, `approved`, `unapproved`, or `none` |
| `allowed-repos` | string or array | No | `"all"` | Repository scope: `"all"`, `"public"`, or an array of patterns like `["myorg/*", "partner/repo"]` |
| `blocked-users` | array or expression | No | `[]` | GitHub usernames whose content is unconditionally denied |
| `trusted-users` | array or expression | No | `[]` | GitHub usernames elevated to `approved` integrity regardless of author association |
| `approval-labels` | array or expression | No | `[]` | GitHub label names that promote items to `approved` integrity |
| `integrity-proxy` | boolean | No | `true` | Whether to run the DIFC proxy for pre-agent `gh` CLI calls. Set to `false` to disable |

> [!NOTE]
> `repos` is a deprecated alias for `allowed-repos`. Use `allowed-repos` in new workflows. Run `gh aw fix` to migrate existing workflows automatically.

## Integrity Levels

The full integrity hierarchy, from highest to lowest:

```text
merged > approved > unapproved > none > blocked
```

| Level | What qualifies at this level |
|-------|------------------------------|
| `merged` | Pull requests that have been merged, and commits reachable from the default branch (any author) |
| `approved` | Objects authored by `OWNER`, `MEMBER`, or `COLLABORATOR`; non-fork PRs on public repos; all items in private repos; trusted platform bots (e.g., dependabot); users listed in `trusted-users` |
| `unapproved` | Objects authored by `CONTRIBUTOR` or `FIRST_TIME_CONTRIBUTOR` |
| `none` | All objects, including `FIRST_TIMER` and users with no association (`NONE`) |
| `blocked` | Items authored by users in `blocked-users` — always denied, cannot be promoted |

The four configurable levels (`merged`, `approved`, `unapproved`, `none`) are cumulative and ordered from most restrictive to least. Setting `min-integrity: approved` means only items at `approved` level **or higher** (`merged`) reach the agent. Items at `unapproved` or `none` are filtered out.

`blocked` is not a configurable `min-integrity` value — it is assigned automatically to items from users in the `blocked-users` list and is always denied regardless of the configured threshold.

**`merged`**: PRs merged into the target branch, and commits reachable from the default branch.

**`approved`**: Owners, members, and collaborators; items in private repositories; platform bots (dependabot, github-actions); and users in `trusted-users`.

**`unapproved`**: Contributors and first-time contributors (`CONTRIBUTOR`, `FIRST_TIME_CONTRIBUTOR`). Appropriate for community workflows where outputs are reviewed before being applied.

**`none`**: All content, including anonymous users. Use deliberately for workflows designed to process untrusted input.

**`blocked`**: Unconditionally denied — even `min-integrity: none` does not allow them through. See [Blocking specific users](#blocking-specific-users) below.

## Scoping to Repositories

`allowed-repos` defines which repositories the guard policy applies to. It accepts three forms:

- **`"all"`** — All repositories the token can access (default when omitted).
- **`"public"`** — Only public repositories.
- **An array of patterns** — Specific repositories or owner wildcards.

```aw wrap
tools:
  github:
    allowed-repos:
      - "myorg/*"
      - "partner/shared-repo"
    min-integrity: approved
```

Repository patterns must be lowercase and follow one of these formats:

| Pattern | Meaning |
|---------|---------|
| `owner/*` | All repositories under `owner` |
| `owner/prefix*` | Repositories under `owner` whose name starts with `prefix` |
| `owner/repo` | A single specific repository |

## Adjusting Integrity Per-Item

Beyond setting a minimum level, you can override integrity for specific authors or labels.

### Blocking specific users

`blocked-users` unconditionally blocks content from listed GitHub usernames, regardless of `min-integrity`, `trusted-users`, or any labels. Blocked items receive an effective integrity of `blocked` (below `none`) and are always denied.

```aw wrap
tools:
  github:
    min-integrity: none
    blocked-users:
      - "spam-bot"
      - "compromised-account"
```

### Trusting specific users

`trusted-users` elevates content from listed GitHub usernames to `approved` integrity, regardless of their author association. This is useful for contractors, partner developers, or external contributors who should be treated as trusted even though GitHub classifies them as `CONTRIBUTOR` or `FIRST_TIME_CONTRIBUTOR`.

```aw wrap
tools:
  github:
    min-integrity: approved
    trusted-users:
      - "contractor-1"
      - "partner-dev"
```

Trust elevation only raises integrity — it never lowers it. A user already at `merged` stays at `merged`. `blocked-users` always takes precedence: if a user appears in both `blocked-users` and `trusted-users`, they are blocked.

`trusted-users` requires `min-integrity` to be set.

### Promoting items via labels

`approval-labels` promotes items bearing any listed GitHub label to `approved` integrity, enabling human-review workflows where a trusted reviewer labels content to signal it is safe for the agent.

```aw wrap
tools:
  github:
    min-integrity: approved
    approval-labels:
      - "human-reviewed"
      - "safe-for-agent"
```

Promotion only raises integrity — it never lowers it. An item already at `merged` stays at `merged`. Blocked-user exclusion always takes precedence: a blocked user's items remain blocked even if they carry an approval label.

### Using GitHub Actions expressions

`blocked-users`, `trusted-users`, and `approval-labels` can each accept a GitHub Actions expression instead of a literal array. The expression is evaluated at runtime and should resolve to a comma- or newline-separated list of values.

```aw wrap
tools:
  github:
    min-integrity: approved
    blocked-users: ${{ vars.BLOCKED_USERS }}
    trusted-users: ${{ vars.TRUSTED_USERS }}
    approval-labels: ${{ vars.APPROVAL_LABELS }}
```

### Effective integrity computation

The gateway computes each item's effective integrity in this order:

1. **Start** with the base integrity level from GitHub metadata (author association, merge status, repo visibility).
2. **If the author is in `blocked-users`**: effective integrity → `blocked` (always denied).
3. **Else if the author is in `trusted-users`**: effective integrity → max(base, `approved`).
4. **Else if the item has a label in `approval-labels`**: effective integrity → max(base, `approved`).
5. **Else**: effective integrity → base.

## Centralized Management via GitHub Variables

Each per-item list can also be extended with a GitHub repository or organization variable — the runtime unions the two:

| Workflow field | GitHub variable |
|---------------|----------------|
| `blocked-users` | `GH_AW_GITHUB_BLOCKED_USERS` |
| `trusted-users` | `GH_AW_GITHUB_TRUSTED_USERS` |
| `approval-labels` | `GH_AW_GITHUB_APPROVAL_LABELS` |

For example, if a workflow declares `blocked-users: ["spam-bot"]` and the organization variable `GH_AW_GITHUB_BLOCKED_USERS` is set to `compromised-acct,old-bot`, the effective blocked-users list at runtime is `["spam-bot", "compromised-acct", "old-bot"]`.

Variables are split on commas and newlines, trimmed, and deduplicated. Set these as repository or organization-level variables to apply them across all workflows.

## Default Behavior

For **public repositories**, the runtime automatically applies `min-integrity: approved` when none is configured, protecting public workflows even without additional authentication.

For **private and internal repositories**, no guard policy is applied and all content is accessible by default.

## Pre-Agent Integrity Proxy

When a guard policy is configured (`min-integrity` is set), the compiler injects a DIFC proxy that filters `gh` CLI calls in pre-agent setup steps. This ensures that custom steps running before the agent see the same integrity-filtered API responses that the agent itself operates under.

The proxy routes `gh` CLI calls through the same MCP gateway container, applying static fields (`min-integrity` and `allowed-repos`) available at compile time. Runtime fields (`blocked-users`, `trusted-users`, `approval-labels`) are not applied. The proxy starts before custom steps and stops before the MCP gateway to avoid double-filtering.

### Disabling the proxy

The proxy is enabled by default whenever a guard policy is configured. To disable it, set `integrity-proxy: false`:

```aw wrap
tools:
  github:
    min-integrity: approved
    integrity-proxy: false
```

This is an opt-out escape hatch for workflows where pre-agent steps should not be filtered — for example, when custom steps need unfiltered API access for setup purposes.

> [!NOTE]
> Disabling the proxy only affects pre-agent `gh` CLI calls. The agent itself always operates under the configured guard policy via the MCP gateway.

## Choosing a Level

- **Workflows that automate code review or apply changes**: `merged` or `approved` — only act on trusted content.
- **Workflows that respond to maintainers and trusted contributors**: `approved` — a common, safe default for most workflows.
- **Community triage or planning workflows**: `unapproved` — allow contributor input while excluding anonymous or first-time interactions.
- **Public-data workflows or spam detection**: `none` — see all activity, but ensure the workflow's outputs are not directly applied without review.

> [!NOTE]
> Setting `min-integrity: none` on a public repository disables the automatic protection. Only use it when the workflow is designed to handle untrusted input.

## Examples

**Allow only merged content:**

```aw wrap
tools:
  github:
    allowed-repos: "all"
    min-integrity: merged
```

**Trusted contributors only (typical for a public repository workflow):**

```aw wrap
tools:
  github:
    min-integrity: approved
```

**Allow all community contributions (for a triage workflow):**

```aw wrap
tools:
  github:
    min-integrity: unapproved
```

**Explicitly disable filtering on a public repository, apart from blocked users:**

```aw wrap
tools:
  github:
    min-integrity: none
```

**Combined: blocking, trusting, and labeling:**

```aw wrap
tools:
  github:
    allowed-repos: "all"
    min-integrity: approved
    blocked-users:
      - "known-spam-bot"
    trusted-users:
      - "contractor-1"
    approval-labels:
      - "agent-approved"
```

## In Logs and Reports

When an item is filtered by the integrity check, the MCP gateway records a `DIFC_FILTERED` event in the run's `gateway.jsonl` log. Each event includes:

- **Server**: the MCP server that returned the filtered content
- **Tool**: the tool call that produced it (e.g., `list_issues`, `get_pull_request`)
- **User**: the login of the content's author
- **Reason**: a description such as `"Resource has lower integrity than agent requires."`
- **Integrity tags**: the tags assigned to the item that caused it to be filtered
- **Author association**: the GitHub author association (`CONTRIBUTOR`, `FIRST_TIMER`, etc.)

When gateway metrics are displayed, filtered events appear in a **DIFC Filtered Events** table alongside the standard server usage table:

```text
┌────────────────────────────────────────────────────────────────────────────────────┐
│ DIFC Filtered Events                                                               │
├────────────────┬───────────────┬───────────────┬──────────────────────────────────-┤
│ Server         │ Tool          │ User          │ Reason                            │
├────────────────┼───────────────┼───────────────┼───────────────────────────────────┤
│ github         │ list_issues   │ new-user      │ Resource has lower integrity than │
│                │               │               │ agent requires.                   │
└────────────────┴───────────────┴───────────────┴───────────────────────────────────┘
```

The `Total DIFC Filtered` count in the summary line shows how many items were suppressed during the run.

### Filtering Logs by Integrity Events

To download only runs that had integrity-filtered content, use the `--filtered-integrity` flag with the `logs` command:

```bash
gh aw logs --filtered-integrity
```

This is useful when investigating whether your `min-integrity` configuration is filtering expected content or when tuning the level after observing real traffic patterns.

## Related Documentation

- [GitHub Tools Reference](/gh-aw/reference/github-tools/) — Full `tools.github` configuration
- [MCP Gateway](/gh-aw/reference/mcp-gateway/) — Gateway architecture and log format
- [CLI Reference: logs](/gh-aw/setup/cli/#logs) — Downloading and analyzing workflow run logs
