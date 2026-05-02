---
title: A/B Experiments
description: Run A/B experiments in GitHub Agentic Workflows to test prompt variants and measure the effect of different instructions across runs.
sidebar:
  order: 7
---

The `experiments` section of the workflow frontmatter enables statistical A/B testing by defining named experiments, each with a set of variant values. At runtime the activation job selects one variant per experiment using a balanced counter and exposes the selection to the workflow prompt.

## Declaring experiments

Add an `experiments` map to the workflow frontmatter. Each key names an experiment; the value is either a bare array of variant strings (simple form) or an object with a `variants` field and optional metadata (object form).

**Simple form** — a list of two or more variant strings:

```aw wrap
---
on:
  issues:
    types: [opened]
engine: copilot

experiments:
  style: [concise, detailed]
---

Summarize this issue in a **${{ experiments.style }}** way.
```

**Object form** — variants plus optional metadata for tracking and scheduling:

```aw wrap
---
experiments:
  prompt_style:
    variants: [concise, verbose]
    description: Test whether concise vs verbose prompts reduce token consumption
    metric: effective_tokens
    weight: [70, 30]
    issue: 1234
    start_date: 2026-05-01
    end_date: 2026-06-15
---
```

> [!NOTE]
> Experiment names must be valid identifiers: start with a letter or underscore, followed by letters, digits, or underscores (e.g. `style`, `feature_1`). Names that do not match this pattern are ignored.

## Using variants in the prompt

Reference a variant with `${{ experiments.<name> }}`. At runtime this is substituted with the selected variant string (e.g. `concise`).

Use the `{{#if experiments.<name> }}` block syntax for conditional prompt sections. A variant value of `no` is treated as falsy, enabling yes/no flag experiments:

```aw wrap
---
experiments:
  caveman: [yes, no]
---

{{#if experiments.caveman }}
Talk like a caveman in all your responses. Me test. You run.
{{/if}}

Address the issue described above.
```

## Statistical balancing

The activation job maintains a per-variant invocation counter in an `actions/cache` entry keyed by workflow ID. The variant with the lowest cumulative count is selected on each run; ties are broken by variant order. Over N runs every variant is used approximately N/K times (K = variant count), providing basic A/B balance with no configuration.

When `weight` is provided (object form only), the counter uses weighted round-robin instead of equal balancing. Weights are relative integers — `weight: [70, 30]` allocates roughly 70% of runs to the first variant and 30% to the second.

The counter persists across workflow runs via the GitHub Actions cache. A fresh repository starts from zero counts.

When `start_date` or `end_date` is set, the experiment is inactive outside the specified window and the control variant (the first in the `variants` array) is used instead.

## Accessing assignments downstream

Each experiment exposes its selected variant as an activation job output:

| Expression | Description |
|---|---|
| `needs.activation.outputs.<name>` | Selected variant for experiment `<name>` |
| `needs.activation.outputs.experiments` | All assignments as a JSON object |

Use these expressions in downstream jobs defined in the `jobs:` frontmatter section.

## Analyzing results

The activation job uploads the counter state as an `experiment` artifact. Download and inspect it with the `gh aw` CLI:

```bash
# Display experiment assignments in the audit report
gh aw audit <run-id>

# Download only the experiment artifact for a specific run
gh aw audit <run-id> --artifacts experiment

# Filter audit output to runs that include a specific experiment
gh aw audit <run-id> --experiment style

# Filter to runs where a specific variant was selected
gh aw audit <run-id> --experiment style --variant concise
```

The `🧪 A/B Experiments` section of the audit report shows the variant chosen on the most recent run and the cumulative counts across all runs:

```
🧪 A/B Experiments
  • caveman = yes (cumulative: no:4, yes:5)
  • style = concise (cumulative: concise:5, detailed:4)
```

## Frontmatter reference

### Simple form

| Field | Type | Description |
|---|---|---|
| `experiments` | `object` | Map of experiment name → variant array |
| `experiments.<name>` | `string[]` | Array of two or more variant strings |

### Object form

| Field | Type | Required | Description |
|---|---|---|---|
| `variants` | `string[]` | Yes | Array of two or more variant strings |
| `description` | `string` | No | Human-readable description of what this experiment tests |
| `metric` | `string` | No | Primary metric to observe (e.g. `effective_tokens`) |
| `weight` | `integer[]` | No | Per-variant probability weights (relative, need not sum to 100). For example, `[70, 30]` assigns ~70% of runs to the first variant and ~30% to the second. Length must equal the number of variants |
| `issue` | `integer` | No | GitHub issue number tracking this experiment |
| `start_date` | `string` | No | ISO-8601 date (`YYYY-MM-DD`). Experiment is active on and after this date; the control variant (first in `variants`) is used before it |
| `end_date` | `string` | No | ISO-8601 date (`YYYY-MM-DD`). Experiment is active on and before this date; the control variant is used after it |
