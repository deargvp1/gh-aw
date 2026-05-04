## Important Note: Workflow Exclusions

`/tmp/gh-aw/agent/workflow-list.txt` is the source of truth for executable workflow inventory.

It is built from top-level `.github/workflows/*.md` files only, so `.github/workflows/shared/` include files are excluded automatically.

It also excludes:
- `smoke-*.md` workflows
- `test-*.md` workflows
- `example*.md` workflow files

Do not report excluded files as missing lock files.
