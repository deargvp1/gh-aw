---
tools:
  bash:
    - "jq *"
    - "${RUNNER_TEMP}/gh-aw/actions/jqschema.sh"
    - "git"
---

## jqschema - JSON Schema Discovery

A utility script is available at `${RUNNER_TEMP}/gh-aw/actions/jqschema.sh` to help you discover the structure of complex JSON responses. The script is pre-installed by the gh-aw setup action.

### Purpose

Generate a compact structural schema (keys + types) from JSON input. This is particularly useful when:
- Analyzing tool outputs from GitHub search (search_code, search_issues, search_repositories)
- Exploring API responses with large payloads
- Understanding the structure of unfamiliar data without verbose output
- Planning queries before fetching full data

### Usage

```bash
# Analyze a file
cat data.json | ${RUNNER_TEMP}/gh-aw/actions/jqschema.sh

# Analyze command output
echo '{"name": "test", "count": 42, "items": [{"id": 1}]}' | ${RUNNER_TEMP}/gh-aw/actions/jqschema.sh

# Analyze GitHub search results
gh api search/repositories?q=language:go | ${RUNNER_TEMP}/gh-aw/actions/jqschema.sh
```
