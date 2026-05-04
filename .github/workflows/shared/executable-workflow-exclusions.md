## Important Note: Workflow Exclusions

`/tmp/gh-aw/agent/workflow-list.txt` is the source of truth for executable workflow inventory.

It already excludes:
- `.github/workflows/shared/` include files
- `smoke-*.md` workflows
- `test-*.md` workflows
- `example*.md` workflow files

Do not report excluded files as missing lock files.
