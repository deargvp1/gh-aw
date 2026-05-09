---
title: AWF container.maxRuns Specification
description: Specification for adding engine-agnostic multi-run orchestration via container.maxRuns in the AWF config
---

# AWF `container.maxRuns` Specification

## Status

**Pending** — the gh-aw plumbing is in place (version-gated on `AWFMaxRunsMinVersion = "v0.26.0"`),
but `gh-aw-firewall` does not yet implement the field. This spec describes what `gh-aw-firewall`
needs to add before the version gate is lifted.

This spec is tracked by the issue created on the gh-aw repository.

## Background

Currently, `engine.max-continuations` in the workflow frontmatter is a Copilot-specific feature
that maps to `--autopilot --max-autopilot-continues N` CLI flags passed to the Copilot CLI. This
limits multi-run mode to the Copilot engine only.

By adding `container.maxRuns` to the AWF config schema, AWF can orchestrate multi-run execution
engine-agnostically — without relying on engine-specific CLI flags — making the feature available
to all engines.

## Proposed AWF Config Schema Addition

Add the following field to the `container` object in `awf-config.schema.json`:

```json
"container": {
  "maxRuns": {
    "type": "integer",
    "minimum": 1,
    "description": "Maximum number of times the agent command may be re-launched within a single AWF container execution. Values greater than 1 enable multi-run mode where AWF restarts the agent after each completed run, up to this limit. This is the AWF-level alternative to engine-specific continuation flags (e.g. --max-autopilot-continues for Copilot). Requires AWF v0.26.0+."
  }
}
```

## gh-aw Mapping

Once `container.maxRuns` is supported by AWF, `gh-aw` will:

1. Populate `container.maxRuns` in the generated `awf-config.json` from `engine.max-continuations`
   in the workflow frontmatter when the effective AWF version is `>= AWFMaxRunsMinVersion`.
2. Skip the Copilot-specific `--autopilot --max-autopilot-continues N` flags (AWF takes over
   multi-run orchestration).
3. Allow `engine.max-continuations` for **any** engine (not just Copilot), since AWF handles
   the multi-run loop engine-agnostically.

The gh-aw plumbing is already implemented in:
- `pkg/workflow/awf_config.go` — `AWFContainerConfig.MaxRuns` field and `BuildAWFConfigJSON` population
- `pkg/workflow/awf_helpers.go` — `awfSupportsMaxRuns()` version gate helper
- `pkg/workflow/copilot_engine_execution.go` — fallback to `--max-autopilot-continues` when AWF version < min
- `pkg/workflow/agent_validation.go` — relaxed engine check when AWF handles multi-run

The version gate constant is `constants.AWFMaxRunsMinVersion = "v0.26.0"`. Update this constant
when `gh-aw-firewall` ships the `container.maxRuns` feature, then run `make build && make recompile && make recompile`.

## Expected AWF Behavior

When `container.maxRuns: N` is set in the AWF config:

1. AWF starts the agent command for the first run.
2. On agent exit (any non-fatal exit), AWF checks if the completed run count is below `maxRuns`.
3. If so, AWF re-launches the agent command for the next run.
4. AWF stops after `maxRuns` total launches, or when the agent exits with a fatal/unrecoverable error.

Context propagation between runs (e.g., conversation history, shared volumes) is to be defined
by the AWF implementation.

## Acceptance Criteria for gh-aw-firewall

- [ ] `container.maxRuns` added to `docs/awf-config.schema.json` (published) and `src/awf-config-schema.json` (runtime).
- [ ] AWF re-launches the agent command up to `container.maxRuns` times within a single container execution.
- [ ] `container.maxRuns: 1` is a no-op (single run, same as the default behavior).
- [ ] The field is documented in `docs/awf-config-spec.md` and `docs/environment.md`.
- [ ] Integration tests verify multi-run behavior for at least one engine.

## Follow-up in gh-aw

After gh-aw-firewall ships `container.maxRuns`:

1. Update `pkg/constants/version_constants.go`: set `AWFMaxRunsMinVersion` to the released version.
2. Update `pkg/workflow/schemas/awf-config.schema.json`: sync from the published schema.
3. Run `make build && make recompile && make recompile` to regenerate all lock files.
4. Consider broadening the `engine.max-continuations` frontmatter schema to note that it is now
   supported by all engines (not just Copilot) when AWF >= `AWFMaxRunsMinVersion`.

## Related Files

- `pkg/constants/version_constants.go` — `AWFMaxRunsMinVersion` constant
- `pkg/workflow/awf_config.go` — `AWFContainerConfig` struct and config builder
- `pkg/workflow/awf_helpers.go` — `awfSupportsMaxRuns()` version gate
- `pkg/workflow/copilot_engine_execution.go` — fallback for old AWF versions
- `pkg/workflow/agent_validation.go` — relaxed engine validation
- `pkg/workflow/schemas/awf-config.schema.json` — embedded AWF config schema (already updated)
- `specs/awf-config-sources-spec.md` — canonical AWF config source references
