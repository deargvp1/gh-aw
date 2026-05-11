---
title: AWF Config Canonical Sources Specification
description: Canonical AWF configuration specification and schema sources that gh-aw agents MUST consult
sidebar:
  order: 1002
---

# AWF Config Canonical Sources Specification

**Version**: 0.1.0  
**Status**: Working Draft  
**Date**: 2026-05-10  
**Last Updated**: 2026-05-10  
**Editors**: GitHub gh-aw Team

---

## 1. Purpose

This document defines the canonical AWF configuration references in `github/gh-aw-firewall` that gh-aw agents and schema reconciliation workflows MUST use when generating or validating AWF config behavior.

## 2. Canonical sources (gh-aw-firewall)

The following documents are authoritative and MUST be consulted together:

### 2.1 Normative specification

- `docs/awf-config-spec.md` — processing model, precedence, CLI mapping, env merge semantics, credential isolation

### 2.2 JSON schemas

- `docs/awf-config.schema.json` — published schema for `.awf.json` / `.awf.yml`
- `src/awf-config-schema.json` — runtime schema source used by AWF CLI
- `schemas/audit.schema.json` — schema for firewall audit output
- `schemas/token-usage.schema.json` — schema for token usage output

### 2.3 Supporting docs

- `docs/environment.md` — environment variable configuration behavior
- `docs/authentication-architecture.md` — credential isolation architecture
- `schemas/README.md` — schema directory overview

## 3. Required coverage checks

When updating AWF config generation, schema sync, or validation in gh-aw, agents MUST verify:

1. Every relevant property in `docs/awf-config.schema.json` is represented in gh-aw logic.
2. CLI mapping behavior in `docs/awf-config-spec.md` is reconciled with schema-defined properties.
3. Config-only fields (without CLI flags) are still modeled where required by runtime behavior.

## 4. Known drift example (apiProxy)

The following fields previously existed in schema but were missed in spec CLI mapping checks:

| Config path | CLI flag |
|---|---|
| `apiProxy.anthropicAutoCache` | `--anthropic-auto-cache` |
| `apiProxy.anthropicCacheTailTtl` | `--anthropic-cache-tail-ttl` |
| `apiProxy.models` | config-only (model alias rewriting) |
| `apiProxy.modelMultipliers` | config-only (effective-token accounting) |
| `apiProxy.maxRuns` | config-only (LLM invocation hard cap) |
| `apiProxy.auth.*` | config-only (maps to `AWF_AUTH_*` env vars) |
| `container.dockerHostPathPrefix` | `--docker-host-path-prefix` |

Agents SHOULD treat this class of mismatch as a regression signal and open a corrective PR when detected.

---

## 3. Conformance Requirements

The key words **MUST**, **MUST NOT**, **SHOULD**, and **MAY** in this section are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

**CR-01**: Agents and schema reconciliation workflows MUST consult **both** the normative specification (`docs/awf-config-spec.md`) and the published JSON schema (`docs/awf-config.schema.json`) before generating or validating AWF config behavior. Consulting only one source is insufficient.

**CR-02**: When a property exists in the JSON schema but has no corresponding entry in the normative spec CLI mapping table, agents MUST treat this as a drift condition and flag it for corrective action.

**CR-03**: Agents MUST NOT generate AWF config fields that are absent from both the normative spec and all JSON schemas. Undocumented fields are out of scope and may be silently ignored or rejected by the AWF CLI.

**CR-04**: Schema reconciliation workflows SHOULD verify coverage of all top-level properties in `docs/awf-config.schema.json` against the CLI mapping table in `docs/awf-config-spec.md` on every run.

**CR-05**: When drift is detected, the detecting agent or workflow SHOULD open a corrective pull request with specific field paths and suggested remediation.

---

## 4. Drift Detection Procedure

This section describes the concrete steps for detecting schema drift between `gh-aw-firewall` and `gh-aw`.

### 4.1 When to Run

Drift detection MUST be triggered when:

1. A pull request modifies `docs/awf-config.schema.json`, `src/awf-config-schema.json`, or `docs/awf-config-spec.md` in `github/gh-aw-firewall`.
2. A scheduled workflow runs the reconciliation check (RECOMMENDED: daily or weekly).
3. An agent is asked to generate or validate AWF config behavior.

### 4.2 Step-by-Step Procedure

1. **Fetch the canonical sources** from `github/gh-aw-firewall`:
   - `docs/awf-config.schema.json` — published schema
   - `src/awf-config-schema.json` — runtime schema
   - `docs/awf-config-spec.md` — normative specification

2. **Extract the property inventory** from both schema files:
   - List all top-level and nested property keys.
   - Note which properties have corresponding CLI flags (as documented in `docs/awf-config-spec.md`).
   - Note which properties are config-only (no CLI flag).

3. **Compare against gh-aw implementation**:
   - For each schema property, check whether `pkg/workflow/` or `actions/setup/` in `github/gh-aw` references it.
   - For each CLI-mapped property, check whether the CLI flag is tested in `pkg/workflow/` tests.

4. **Identify drift categories**:
   - **Missing in gh-aw**: Property exists in schema but `gh-aw` has no coverage.
   - **Missing in schema**: `gh-aw` generates a field not present in either schema.
   - **Spec mismatch**: CLI mapping in `gh-aw` disagrees with the normative spec description.

5. **Produce a drift report** listing:
   - Each drifted property path (e.g., `apiProxy.anthropicAutoCache`).
   - Drift category (missing in gh-aw / missing in schema / spec mismatch).
   - Suggested corrective action (add coverage, open PR, update spec).

6. **Open a corrective PR** when any drift of category "missing in gh-aw" or "spec mismatch" is found. The PR description MUST include the drift report and reference this procedure.

### 4.3 Example Drift Check (CLI)

```bash
# Fetch both schema files
gh api repos/github/gh-aw-firewall/contents/docs/awf-config.schema.json \
  --jq '.content' | base64 -d > /tmp/published-schema.json

gh api repos/github/gh-aw-firewall/contents/src/awf-config-schema.json \
  --jq '.content' | base64 -d > /tmp/runtime-schema.json

# Extract all property keys
jq '[.. | objects | keys[]] | unique | sort' /tmp/published-schema.json > /tmp/schema-keys.txt

# Compare against gh-aw source references
grep -rh '"apiProxy\|"network\|"model\|"auth' pkg/workflow/ | sort -u > /tmp/ghaw-refs.txt

# Review diff for drift
diff /tmp/schema-keys.txt /tmp/ghaw-refs.txt
```

### 4.4 Automation

A scheduled GitHub Actions workflow in `github/gh-aw` SHOULD automate this procedure. The workflow SHOULD:

- Run on a weekly schedule and on pull requests that touch AWF config handling.
- Fail the check (non-zero exit) when any "missing in gh-aw" drift is found.
- Post a summary comment on PRs with the drift report.
- Create a tracking issue when drift is detected on the scheduled run.

The workflow implementing this procedure is located at `.github/workflows/awf-config-drift.yml` in this repository. It runs weekly (Mondays at 08:00 UTC) and on pull requests that modify AWF config handling paths.

---

## 5. Safeguards

This section describes what `gh-aw` does when AWF config drift is detected. Safeguards apply both in CI (automated runs) and when agents execute drift detection interactively.

### 5.1 CI Failure on Detected Drift

**Trigger:** The drift detection workflow (`.github/workflows/awf-config-drift.yml`) finds a property of category "missing in gh-aw" or "spec mismatch".

**Response:**
- The workflow step exits with a non-zero code, causing the GitHub Actions job to fail.
- A summary of the detected drift is written to `$GITHUB_STEP_SUMMARY`, identifying each drifted property path, its category, and the suggested corrective action.
- The failing job blocks merging on pull requests where the check is required.

### 5.2 Tracking Issue Creation

**Trigger:** The scheduled (weekly) drift detection run finds any drift.

**Response:**
- The workflow opens a new GitHub issue in `github/gh-aw` with:
  - Title: `AWF Config Drift Detected — <date>`
  - Body: Full drift report (property paths, categories, suggested actions)
  - Label: `awf-config-drift`
- If an open issue with label `awf-config-drift` already exists, the workflow appends a comment to the existing issue instead of opening a duplicate.

### 5.3 Agent Corrective PR

**Trigger:** An agent running drift detection (interactively or in CI) finds drift of category "missing in gh-aw" or "spec mismatch" (per CR-05).

**Response:**
- The agent SHOULD open a corrective pull request to `github/gh-aw` that:
  - Adds coverage for the missing property in `pkg/workflow/` or `actions/setup/`.
  - References the drifted property path(s) and this spec in the PR description.
  - Assigns the `awf-config-drift` label.
- The PR description MUST include the drift report and reference Section 4.2 of this specification.

---

## 6. Spec Maintenance

This section tracks when the canonical sources in Section 2 were last verified against the `gh-aw` implementation.

| Canonical Source | Last Verified | Verified By | Notes |
|-----------------|---------------|-------------|-------|
| `docs/awf-config.schema.json` | 2026-05-10 | Daily SPDD agent | Drift example in Section 4 identified; apiProxy fields added |
| `src/awf-config-schema.json` | 2026-05-10 | Daily SPDD agent | Cross-referenced with published schema |
| `docs/awf-config-spec.md` | 2026-05-10 | Daily SPDD agent | CLI mapping table in Section 4 added |
| `docs/environment.md` | — | — | Not yet verified |
| `docs/authentication-architecture.md` | — | — | Not yet verified |
| `schemas/audit.schema.json` | — | — | Not yet verified |
| `schemas/token-usage.schema.json` | — | — | Not yet verified |
| `schemas/README.md` | — | — | Not yet verified |

**Next scheduled verification:** Week of 2026-05-18 (automated via `.github/workflows/awf-config-drift.yml`)
