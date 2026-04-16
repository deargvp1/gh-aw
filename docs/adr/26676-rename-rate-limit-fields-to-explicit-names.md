# ADR-26676: Rename rate-limit Fields to Explicit Names (`max-runs` / `max-runs-window`)

**Date**: 2026-04-16
**Status**: Draft
**Deciders**: pelikhan, Copilot

---

## Part 1 — Narrative (Human-Friendly)

### Context

The `rate-limit` frontmatter feature was originally shipped with abbreviated field names: `max` (maximum allowed runs) and `window` (time window in minutes). While concise, these names were ambiguous — `max` could mean "max concurrent runs", "max total runs", or "max cost", and `window` is a generic term with no obvious domain context. As the rate-limit feature graduates from experimental status and sees broader adoption, the naming ambiguity creates confusion for workflow authors reading documentation or debugging configurations. The project already has a `gh aw fix` codemod framework that can automate migration of deprecated fields, making a breaking rename practical at this stage.

### Decision

We will rename `rate-limit.max` to `rate-limit.max-runs` and `rate-limit.window` to `rate-limit.max-runs-window` to make the fields self-documenting. The old field names will be retained as deprecated aliases with fallback parsing (new name takes precedence when both are present) to preserve backward compatibility for existing workflows. A codemod (`rate-limit-fields-migration`) will be added to the `gh aw fix` command to automate migration of legacy field names. We will also remove the experimental warning for the `rate-limit` feature, treating it as stable.

### Alternatives Considered

#### Alternative 1: Keep the original abbreviated names (`max` / `window`)

The existing names are shorter and already documented. Keeping them avoids any migration burden. This was rejected because the names are genuinely ambiguous (`max` does not convey "maximum runs"; `window` does not convey "time window in minutes"), and the rate-limit feature is stable enough that renaming before wide adoption is lower cost than living with confusing names indefinitely.

#### Alternative 2: Use a nested sub-object (e.g., `rate-limit.runs.max` / `rate-limit.runs.window`)

A hierarchical structure could group related fields more explicitly. This was rejected because it adds nesting depth without meaningful benefit — the `rate-limit` object is already scoped, and two fields do not warrant a sub-object. Flatter structures are easier to read in YAML.

#### Alternative 3: Rename without backward-compatible fallback (hard break)

A clean rename with no legacy support would simplify the parser. This was rejected because existing workflows using `max` / `window` would silently stop applying rate limits (the fields would be ignored), which is a safety regression. The codemod migration path provides a better user experience.

### Consequences

#### Positive
- Field names are now self-documenting: `max-runs` and `max-runs-window` unambiguously describe their purpose.
- Removes the experimental warning, signaling that `rate-limit` is a stable, production-ready feature.
- The `gh aw fix` codemod allows users to migrate automatically without manual search-and-replace.
- JSON schema now explicitly marks `max` and `window` as deprecated with descriptions pointing to the new names.

#### Negative
- Any workflow using the old field names that is not run through `gh aw fix` will continue to work silently via the fallback, which means un-migrated configurations persist without visible warning.
- The parser must now maintain two code paths for the same logical fields, increasing maintenance surface until the deprecated aliases are removed.
- Documentation and examples must be updated in multiple locations (reference docs, guides, test fixtures).

#### Neutral
- The JSON struct tags on `RateLimitConfig` now map to `max-runs` / `max-runs-window`, so serialization round-trips through the struct will always emit the new names — legacy names only survive in the raw-YAML parsing layer.
- The schema now uses `anyOf` to allow either `max-runs` or `max` as the required field, which correctly models the backward-compatible state during the migration period.

---

## Part 2 — Normative Specification (RFC 2119)

> The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this section are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

### Rate-Limit Field Naming

1. New workflow files **MUST** use `max-runs` as the field name for the maximum run count within the `rate-limit` block.
2. New workflow files **MUST** use `max-runs-window` as the field name for the time window in minutes within the `rate-limit` block.
3. The deprecated field names `max` and `window` **MUST NOT** be used in newly authored workflow files.
4. The parser **MUST** accept `max` and `window` as fallback aliases for backward compatibility, reading them only when `max-runs` / `max-runs-window` are absent respectively.
5. When both a deprecated alias and its canonical replacement are present in the same `rate-limit` block, the canonical field (`max-runs` / `max-runs-window`) **MUST** take precedence.

### Schema and Validation

1. The JSON schema **MUST** document `max` and `window` with `"deprecated": true` and a description referring users to the canonical field names.
2. The JSON schema **MUST** use `anyOf` to accept either `max-runs` or `max` as satisfying the required-field constraint during the migration period.
3. The JSON schema **MUST NOT** require both `max` and `max-runs` simultaneously.

### Migration Tooling

1. The `gh aw fix` codemod registry **MUST** include the `rate-limit-fields-migration` codemod.
2. The `rate-limit-fields-migration` codemod **MUST** rename `max` to `max-runs` and `window` to `max-runs-window` within `rate-limit` blocks.
3. The codemod **MUST NOT** rename `max` or `window` fields that appear outside a `rate-limit` block.
4. The codemod **SHOULD** be idempotent: re-running on an already-migrated file **MUST NOT** modify the file.

### Experimental Warning

1. The compiler **MUST NOT** emit an experimental warning when a workflow uses the `rate-limit` feature.
2. The `rate-limit` feature **SHALL** be treated as stable and production-ready.

### Conformance

An implementation is considered conformant with this ADR if it satisfies all **MUST** and **MUST NOT** requirements above. Failure to meet any **MUST** or **MUST NOT** requirement constitutes non-conformance.

---

*This is a DRAFT ADR generated by the [Design Decision Gate](https://github.com/github/gh-aw/actions/runs/24519943082) workflow. The PR author must review, complete, and finalize this document before the PR can merge.*
