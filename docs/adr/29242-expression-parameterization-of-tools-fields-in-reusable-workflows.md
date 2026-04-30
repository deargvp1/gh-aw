# ADR-29242: Expression Parameterization of tools.bash, tools.edit, and tools.github.toolsets in Reusable Workflows

**Date**: 2026-04-30
**Status**: Draft
**Deciders**: pelikhan, Copilot

---

## Part 1 — Narrative (Human-Friendly)

### Context

`workflow_call` reusable workflows in gh-aw previously required static, compile-time tool configurations for `tools.bash` (bash allowlist), `tools.edit` (edit tool enablement), and `tools.github.toolsets` (GitHub MCP toolset selection). Because these fields only accepted literal values, teams that wanted different tool policies for different callers were forced to maintain separate workflow files for each permission combination — one per allowlist, one with edit enabled, one without, and so on. The pattern of one workflow file per tool configuration undercuts the entire purpose of `workflow_call` reuse and was identified as a hard blocker for workflow composition.

### Decision

We will extend `tools.bash`, `tools.edit`, and `tools.github.toolsets` to accept GitHub Actions expression strings (e.g., `${{ inputs.bash-allowlist }}`) in addition to their existing literal types. At compile time, the parser detects expression strings via `isExpression()` and stores them in new dedicated fields (`AllowedCommandsExpr`, `EnabledExpr`, `ToolsetExpr`) rather than the existing literal fields. The compiler then omits static `--allow-tool` CLI arguments when an expression is present and instead injects a bash preamble (`buildCopilotDynamicToolArgsPreamble`) that reads `GH_AW_BASH_ALLOWLIST` and `GH_AW_EDIT_ENABLED` environment variables at runner runtime to construct the tool argument arrays. GitHub MCP toolset expressions are passed through verbatim as the `GITHUB_TOOLSETS` environment variable for runtime evaluation. All three fields remain fail-closed: empty or unresolved expressions grant no permissions.

### Alternatives Considered

#### Alternative 1: Separate Workflow File per Tool Configuration (Status Quo)

Teams maintain one workflow file per unique combination of tool permissions. This was rejected because it directly causes the workflow proliferation problem this PR addresses: a single reusable workflow covering bash-enabled and bash-disabled callers would require two files. Scaling to three fields with multiple values produces a combinatorial explosion.

#### Alternative 2: New Sibling Expression Fields (e.g., `tools.bash-expr`)

Introduce dedicated expression-only sibling fields alongside the existing literal fields, leaving the existing schema untouched. This was rejected because it doubles the schema surface for tool configuration, creates confusing dual-field semantics (which field wins when both are set?), and requires callers to learn a parallel naming convention for every parameterizable field — the same pattern already rejected in ADR-29212 for safe-output list constraints.

#### Alternative 3: Compile-Time Resolution of workflow_call Input Defaults

Evaluate `workflow_call` input defaults at compile time and fold them into the static tool configuration. This was rejected because `workflow_call` inputs are caller-supplied and only known at runtime; default values cover only the no-argument case and cannot represent the full range of values callers may provide.

### Consequences

#### Positive
- A single reusable `workflow_call` workflow can now expose `tools.bash`, `tools.edit`, and `tools.github.toolsets` as caller-controlled inputs, eliminating the need for separate workflow files per tool policy.
- All three fields remain fail-closed: unset or empty expressions grant no bash access and keep edit disabled, matching the principle of least privilege.
- All existing literal tool configurations continue to work unchanged; no migration is required for current workflows.
- `hasBashWildcardInTools` treats expressions as restricted (not wildcards), keeping Claude in `acceptEdits` mode rather than `bypassPermissions`.

#### Negative
- Compile-time toolset validation and permissions checks are bypassed when `tools.github.toolsets` is an expression, because the actual toolset names are unknown until runtime. Workflow authors must manually ensure the workflow's `permissions:` block covers the broadest toolset callers may supply.
- Expression correctness (e.g., whether the referenced input name exists, whether the resolved value is a valid command list) is validated only at runtime, not at compile time, pushing a class of errors from authoring time to execution time.
- The bash preamble (`buildCopilotDynamicToolArgsPreamble`) adds a hidden runtime step to copilot invocations that use expression fields; this preamble must be kept in sync if the tool argument interface changes.
- TOML-based engine configs (Codex MCP gateway) fall back to engine defaults when expressions are used, since TOML has no expression syntax — callers using those engines cannot benefit from runtime parameterization.

#### Neutral
- The PR number (29242) is used as the ADR number, consistent with the project's ADR-by-PR-number convention.
- Three new struct fields (`AllowedCommandsExpr string`, `EnabledExpr string`, `ToolsetExpr string`) are added to the parser types; downstream consumers that switch on tool config must handle the new expression case.
- The JSON schema for `tools.bash`, `tools.edit`, and `tools.github.toolsets` each gain a `{ "type": "string", "pattern": "^\\$\\{\\{.*\\}\\}$" }` variant in their `oneOf`, which IDE tooling will surface as an additional accepted type.

---

## Part 2 — Normative Specification (RFC 2119)

> The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this section are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

### Expression Detection in Tool Fields

1. Implementations **MUST** accept a GitHub Actions expression string matching `^\$\{\{.*\}\}$` as the value of `tools.bash`, `tools.edit`, and `tools.github.toolsets` in addition to each field's existing literal types.
2. Implementations **MUST** detect expression strings at parse time using `isExpression()` and store them in dedicated expression fields (`AllowedCommandsExpr`, `EnabledExpr`, `ToolsetExpr`) rather than the existing literal fields.
3. Implementations **MUST NOT** attempt to evaluate, validate, or resolve expression content at compile time.
4. Implementations **MUST NOT** accept a bare non-expression string for `tools.bash` or `tools.github.toolsets`; such values **MUST** produce a descriptive parse error.

### Compile-Time Behavior

1. When `tools.bash` or `tools.edit` contains an expression, implementations **MUST** omit static `--allow-tool` CLI arguments from the compiled copilot invocation command.
2. When `tools.bash` or `tools.edit` contains an expression, implementations **MUST** inject the `buildCopilotDynamicToolArgsPreamble` bash preamble into the compiled workflow so that tool argument arrays are constructed from `GH_AW_BASH_ALLOWLIST` and `GH_AW_EDIT_ENABLED` environment variables at runner runtime.
3. When `tools.github.toolsets` contains an expression, implementations **MUST** pass the expression string verbatim as the `GITHUB_TOOLSETS` environment variable for runtime evaluation by the GitHub MCP server.
4. When `tools.github.toolsets` contains an expression, implementations **MUST** skip compile-time toolset validation and permissions checks.
5. `hasBashWildcardInTools` **MUST** return `false` (restricted mode) when `tools.bash` is an expression, ensuring Claude runs in `acceptEdits` mode rather than `bypassPermissions` mode.

### Runtime Fail-Closed Behavior

1. Implementations **MUST** grant no bash access when `GH_AW_BASH_ALLOWLIST` is empty or unset at runtime.
2. Implementations **MUST** grant write permission for the edit tool only when `GH_AW_EDIT_ENABLED` resolves exactly to the string `"true"` at runtime; all other values (including empty) **MUST** keep the edit tool disabled.

### JSON Schema

1. The JSON Schema entry for `tools.bash` **MUST** include a `{ "type": "string", "pattern": "^\\$\\{\\{.*\\}\\}$" }` variant in its `oneOf` alongside the existing array form.
2. The JSON Schema entry for `tools.github.toolsets` **MUST** include a `{ "type": "string", "pattern": "^\\$\\{\\{.*\\}\\}$" }` variant in its `oneOf` alongside the existing array form.
3. The JSON Schema entry for `tools.edit` **MUST** include `{ "type": "boolean" }` and `{ "type": "string", "pattern": "^\\$\\{\\{.*\\}\\}$" }` variants alongside the existing object/null forms.

### Conformance

An implementation is considered conformant with this ADR if it satisfies all **MUST** and **MUST NOT** requirements above. Failure to meet any **MUST** or **MUST NOT** requirement constitutes non-conformance.

---

*This is a DRAFT ADR generated by the [Design Decision Gate](https://github.com/github/gh-aw/actions/runs/25148307558) workflow. The PR author must review, complete, and finalize this document before the PR can merge.*
