---
description: Detects inconsistencies between JSON schema, implementation code, and documentation
on:
  schedule: daily
  workflow_dispatch:
permissions:
  contents: read
  discussions: read
  issues: read
  pull-requests: read
engine:
  id: claude
  max-turns: 40
tools:
  edit:
  bash: ["*"]
  github:
    mode: gh-proxy
    toolsets: [default, discussions]
  cache-memory:
    key: schema-consistency-cache-${{ github.workflow }}
timeout-minutes: 30
checkout:
  - fetch-depth: 1
    current: true
imports:
  - uses: shared/daily-audit-base.md
    with:
      title-prefix: "[Schema Consistency] "
      expires: 1d
  - shared/observability-otlp.md
pre-agent-steps:
  - name: Precompute schema analysis data
    run: |
      set -e
      mkdir -p /tmp/gh-aw/agent

      echo "=== Extracting schema fields ==="

      # 1. All top-level fields in the main JSON schema
      SCHEMA_FIELDS=$(jq -r '.properties | keys[]' pkg/parser/schemas/main_workflow_schema.json 2>/dev/null | sort -u || echo "")

      # 2. yaml-tagged struct fields in pkg/parser/*.go
      PARSER_YAML_FIELDS=$(grep -rh 'yaml:"' pkg/parser/*.go 2>/dev/null \
        | grep -o 'yaml:"[^"]*"' \
        | sed 's/yaml:"//;s/"//' \
        | sed 's/,omitempty//' \
        | sed 's/,.*$//' \
        | grep -v '^-$' \
        | grep -v '^$' \
        | sort -u || echo "")

      # 3. yaml-tagged struct fields in pkg/workflow/*.go
      WORKFLOW_YAML_FIELDS=$(grep -rh 'yaml:"' pkg/workflow/*.go 2>/dev/null \
        | grep -o 'yaml:"[^"]*"' \
        | sed 's/yaml:"//;s/"//' \
        | sed 's/,omitempty//' \
        | sed 's/,.*$//' \
        | grep -v '^-$' \
        | grep -v '^$' \
        | sort -u || echo "")

      # 4. Top-level frontmatter keys actually used in workflow .md files
      USED_FIELDS=$(grep -rh '^[a-z][a-z0-9_-]*:' .github/workflows/*.md 2>/dev/null \
        | sed 's/:.*//' \
        | grep -v '^#' \
        | sort -u || echo "")

      # 5. Schema field types for all top-level fields
      FIELD_TYPES=$(jq -r '.properties | to_entries[] |
        "\(.key): \(.value.type // (.value.anyOf // .value.oneOf // [] | map(.type // "complex") | unique | join("|")) // "complex")"' \
        pkg/parser/schemas/main_workflow_schema.json 2>/dev/null | sort || echo "")

      # 6. Fields in schema but absent as yaml tags in parser structs
      IN_SCHEMA_NOT_PARSER=$(comm -23 \
        <(echo "$SCHEMA_FIELDS") \
        <(echo "$PARSER_YAML_FIELDS" | sort -u) 2>/dev/null || echo "")

      # 7. yaml tags in parser structs absent from schema
      IN_PARSER_NOT_SCHEMA=$(comm -23 \
        <(echo "$PARSER_YAML_FIELDS" | sort -u) \
        <(echo "$SCHEMA_FIELDS") 2>/dev/null || echo "")

      # 8. Fields in schema but absent from workflow compiler structs
      IN_SCHEMA_NOT_WORKFLOW=$(comm -23 \
        <(echo "$SCHEMA_FIELDS") \
        <(echo "$WORKFLOW_YAML_FIELDS" | sort -u) 2>/dev/null || echo "")

      # 9. Fields used in actual workflow .md files but not in schema
      IN_USED_NOT_SCHEMA=$(comm -23 \
        <(echo "$USED_FIELDS" | sort -u) \
        <(echo "$SCHEMA_FIELDS") 2>/dev/null || echo "")

      # Write JSON output
      jq -n \
        --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg schema_fields "$SCHEMA_FIELDS" \
        --arg parser_yaml_fields "$PARSER_YAML_FIELDS" \
        --arg workflow_yaml_fields "$WORKFLOW_YAML_FIELDS" \
        --arg used_in_workflows "$USED_FIELDS" \
        --arg field_types "$FIELD_TYPES" \
        --arg in_schema_not_parser "$IN_SCHEMA_NOT_PARSER" \
        --arg in_parser_not_schema "$IN_PARSER_NOT_SCHEMA" \
        --arg in_schema_not_workflow "$IN_SCHEMA_NOT_WORKFLOW" \
        --arg in_used_not_schema "$IN_USED_NOT_SCHEMA" \
        '{
          generated_at: $generated_at,
          schema_fields: ($schema_fields | split("\n") | map(select(. != ""))),
          parser_yaml_fields: ($parser_yaml_fields | split("\n") | map(select(. != ""))),
          workflow_yaml_fields: ($workflow_yaml_fields | split("\n") | map(select(. != ""))),
          used_in_workflows: ($used_in_workflows | split("\n") | map(select(. != ""))),
          field_types: ($field_types | split("\n") | map(select(. != ""))),
          field_gaps: {
            in_schema_not_parser: ($in_schema_not_parser | split("\n") | map(select(. != ""))),
            in_parser_not_schema: ($in_parser_not_schema | split("\n") | map(select(. != ""))),
            in_schema_not_workflow: ($in_schema_not_workflow | split("\n") | map(select(. != ""))),
            in_used_not_schema: ($in_used_not_schema | split("\n") | map(select(. != "")))
          }
        }' > /tmp/gh-aw/agent/schema-diff.json

      echo "✓ Schema diff written to /tmp/gh-aw/agent/schema-diff.json"
      echo "Summary:"
      jq '{
        schema_field_count: (.schema_fields | length),
        parser_yaml_field_count: (.parser_yaml_fields | length),
        workflow_yaml_field_count: (.workflow_yaml_fields | length),
        gaps: {
          in_schema_not_parser: (.field_gaps.in_schema_not_parser | length),
          in_parser_not_schema: (.field_gaps.in_parser_not_schema | length),
          in_schema_not_workflow: (.field_gaps.in_schema_not_workflow | length),
          in_used_not_schema: (.field_gaps.in_used_not_schema | length)
        }
      }' /tmp/gh-aw/agent/schema-diff.json

  - name: Determine analysis focus area for this run
    run: |
      set -e
      mkdir -p /tmp/gh-aw/agent

      # Rotate through the 4 analysis areas on each scheduled run so every
      # run stays within the 40-turn budget.
      FOCUS_AREAS=("schema-vs-parser" "schema-vs-docs" "schema-vs-workflows" "parser-vs-docs")
      AREA_NAMES=("Schema vs Parser Implementation" "Schema vs Documentation" "Schema vs Actual Workflows" "Parser vs Documentation")

      NEXT_INDEX=0
      if [ -f /tmp/gh-aw/cache-memory/focus-state.json ]; then
        NEXT_INDEX=$(jq -r '.next_area_index // 0' /tmp/gh-aw/cache-memory/focus-state.json 2>/dev/null || echo "0")
        # Guard against out-of-range values
        if ! [[ "$NEXT_INDEX" =~ ^[0-9]+$ ]] || [ "$NEXT_INDEX" -ge "${#FOCUS_AREAS[@]}" ]; then
          NEXT_INDEX=0
        fi
      fi

      FOCUS_AREA="${FOCUS_AREAS[$NEXT_INDEX]}"
      FOCUS_AREA_NAME="${AREA_NAMES[$NEXT_INDEX]}"
      NEXT_AREA_INDEX=$(( (NEXT_INDEX + 1) % ${#FOCUS_AREAS[@]} ))

      jq -n \
        --arg focus_area "$FOCUS_AREA" \
        --arg focus_area_name "$FOCUS_AREA_NAME" \
        --argjson next_area_index "$NEXT_AREA_INDEX" \
        --argjson current_area_index "$NEXT_INDEX" \
        '{
          focus_area: $focus_area,
          focus_area_name: $focus_area_name,
          current_area_index: $current_area_index,
          next_area_index: $next_area_index
        }' > /tmp/gh-aw/agent/focus-area.json

      echo "✓ Focus area for this run: $FOCUS_AREA_NAME ($FOCUS_AREA)"
      echo "  Next run will analyze: ${AREA_NAMES[$NEXT_AREA_INDEX]}"

---
# Schema Consistency Checker

You are an expert system that detects inconsistencies between:
- The main JSON schema of the frontmatter (`pkg/parser/schemas/main_workflow_schema.json`)
- The parser and compiler implementation (`pkg/parser/*.go` and `pkg/workflow/*.go`)
- The documentation (`docs/src/content/docs/**/*.md`)
- The workflows in the project (`.github/workflows/*.md`)

## ⚠️ Turn Budget (40 turns total)

You have a strict budget of **40 turns**. Manage your turns carefully:

- **By turn 10**: Read pre-computed data, load cache, begin targeted analysis
- **By turn 25**: Complete all primary analysis and record findings
- **By turn 30**: Start writing the discussion report
- **Turn 35 or later (hard stop)**: Stop all new analysis immediately. Write whatever findings you have so far to the discussion and call `create_discussion`. It is better to report partial-but-accurate findings than to run out of turns with no output.

If at any point you notice you are on turn 30 or beyond, **stop exploring and finalize your report immediately**.

## This Run: Single Focus Area

Each run is scoped to **one analysis area** to stay within the 40-turn budget. Read your focus area from the pre-computed file:

```bash
cat /tmp/gh-aw/agent/focus-area.json
```

This file contains `focus_area` (machine key) and `focus_area_name` (human-readable). Only analyze that one area this run. The 4 areas rotate across runs:

1. `schema-vs-parser` — Schema vs Parser Implementation
2. `schema-vs-docs` — Schema vs Documentation
3. `schema-vs-workflows` — Schema vs Actual Workflows
4. `parser-vs-docs` — Parser vs Documentation

## Mission

Analyze the repository for the **current run's focus area** and create a discussion report with actionable findings.

## Cache Memory Strategy Storage

Use the cache memory folder at `/tmp/gh-aw/cache-memory/` to store and reuse successful analysis strategies:

1. **Read Previous Strategies**: Check `/tmp/gh-aw/cache-memory/strategies.json` for previously successful detection methods
2. **Strategy Selection**:
   - 70% of the time: Use a proven strategy from the cache
   - 30% of the time: Try a radically different approach to discover new inconsistencies
   - Implementation: Use the day of year (e.g., `date +%j`) modulo 10 to determine selection: values 0-6 use proven strategies, 7-9 try new approaches
3. **Update Strategy Database**: After analysis, save successful strategies to `/tmp/gh-aw/cache-memory/strategies.json`

Strategy database structure:
```json
{
  "strategies": [
    {
      "id": "strategy-1",
      "name": "Schema field enumeration check",
      "description": "Compare schema enum values with parser constants",
      "success_count": 5,
      "last_used": "2024-01-15",
      "findings": 3
    }
  ],
  "last_updated": "2024-01-15"
}
```

## Analysis Areas

### 1. Schema vs Parser Implementation (`schema-vs-parser`)

**Check for:**
- Fields defined in schema but not handled in parser/compiler
- Fields handled in parser/compiler but missing from schema
- Type mismatches (schema says `string`, parser expects `object`)
- Enum values in schema not validated in parser/compiler
- Required fields not enforced
- Default values inconsistent between schema and parser/compiler

**Key files to analyze:**
- `pkg/parser/schemas/main_workflow_schema.json`
- `pkg/parser/schemas/mcp_config_schema.json`
- `pkg/parser/frontmatter.go` and `pkg/parser/*.go`
- `pkg/workflow/compiler.go` - main workflow compiler
- `pkg/workflow/tools.go` - tools configuration processing
- `pkg/workflow/safe_outputs.go` - safe-outputs configuration
- `pkg/workflow/cache.go` - cache and cache-memory configuration
- `pkg/workflow/permissions.go` - permissions processing
- `pkg/workflow/engine.go` - engine config and network permissions types
- `pkg/workflow/domains.go` - network domain allowlist functions
- `pkg/workflow/engine_network_hooks.go` - network hook generation
- `pkg/workflow/engine_firewall_support.go` - firewall support checking
- `pkg/workflow/strict_mode.go` - strict mode validation
- `pkg/workflow/stop_after.go` - stop-after processing
- `pkg/workflow/safe_jobs.go` - safe-jobs configuration (internal - accessed via safe-outputs.jobs)
- `pkg/workflow/runtime_setup.go` - runtime overrides
- `pkg/workflow/github_token.go` - github-token configuration
- `pkg/workflow/*.go` (all workflow processing files that use frontmatter)

### 2. Schema vs Documentation (`schema-vs-docs`)

**Check for:**
- Schema fields not documented
- Documented fields not in schema
- Type descriptions mismatch
- Example values that violate schema
- Missing or outdated examples
- Enum values documented but not in schema

**Key files to analyze:**
- `docs/src/content/docs/reference/frontmatter.md`
- `docs/src/content/docs/reference/frontmatter-full.md`
- `docs/src/content/docs/reference/*.md` (all reference docs)

### 3. Schema vs Actual Workflows (`schema-vs-workflows`)

**Check for:**
- Workflows using fields not in schema
- Workflows using deprecated fields
- Invalid field values according to schema
- Missing required fields
- Type violations in actual usage
- Undocumented field combinations

**Key files to analyze:**
- `.github/workflows/*.md` (all workflow files)
- `.github/workflows/shared/**/*.md` (shared components)

### 4. Parser vs Documentation (`parser-vs-docs`)

**Check for:**
- Parser/compiler features not documented
- Documented features not implemented in parser/compiler
- Error messages that don't match docs
- Validation rules not documented

**Focus on:**
- `pkg/parser/*.go` - frontmatter parsing
- `pkg/workflow/*.go` - workflow compilation and feature processing

## Detection Strategies

Here are proven strategies you can use or build upon:

### Strategy 1: Field Enumeration Diff
1. Extract all field names from schema
2. Extract all field names from parser code (look for YAML tags, map keys)
3. Extract all field names from documentation
4. Compare and find missing/extra fields

### Strategy 2: Type Analysis
1. For each field in schema, note its type
2. Search parser for how that field is processed
3. Check if types match
4. Report type mismatches

### Strategy 3: Enum Validation
1. Extract enum values from schema
2. Search for those enums in parser validation
3. Check if all enum values are handled
4. Find undocumented enum values

### Strategy 4: Example Validation
1. Extract code examples from documentation
2. Validate each example against the schema
3. Report examples that don't validate
4. Suggest corrections

### Strategy 5: Real-World Usage Analysis
1. Parse all workflow files in the repo
2. Extract frontmatter configurations
3. Check each against schema
4. Find patterns that work but aren't in schema (potential missing features)

### Strategy 6: Grep-Based Pattern Detection
1. Use bash/grep to find specific patterns
2. Example: `grep -r "type.*string" pkg/parser/schemas/ | grep engine`
3. Cross-reference with parser implementation

## Implementation Steps

### Step 0: Read Pre-Computed Data and Focus Area (Start Here)

Read both files before doing anything else:

```bash
cat /tmp/gh-aw/agent/focus-area.json
cat /tmp/gh-aw/agent/schema-diff.json
```

The `focus-area.json` tells you which analysis area to investigate this run.
The `schema-diff.json` contains pre-computed field gap data — use it as your primary starting point.

**Use the pre-computed data as your starting point.** Do NOT re-run the field enumeration commands from scratch — instead, refine and supplement with targeted follow-up queries.

### Step 1: Load Previous Strategies
```bash
# Check if strategies file exists
if [ -f /tmp/gh-aw/cache-memory/strategies.json ]; then
  cat /tmp/gh-aw/cache-memory/strategies.json
fi
```

### Step 2: Choose Analysis Focus

Focus **exclusively** on the area from `focus-area.json`. Using the pre-computed `field_gaps` from Step 0 plus the strategy cache from Step 1:
- If `field_gaps` show promising leads for your focus area, start there
- If cache has strategies, use a proven strategy 70% of the time; try a new approach 30% of the time

### Step 3: Execute Targeted Analysis

Use the pre-computed data as context and run **targeted** follow-up commands only when deeper inspection is needed for your **focus area only**.

**Example: Verify a gap from pre-computed data**
```bash
# Verify a specific field gap by searching implementation files
grep -r "fieldName" pkg/parser/ pkg/workflow/ 2>/dev/null | grep -v "_test.go"
```

**Example: Type checking for a specific field**
```bash
# Find schema field types (handles different JSON Schema patterns)
jq -r '
  (.properties // {}) | to_entries[] |
  "\(.key): \(.value.type // .value.oneOf // .value.anyOf // .value.allOf // "complex")"
' pkg/parser/schemas/main_workflow_schema.json 2>/dev/null || echo "Failed to parse schema"
```

### Step 4: Record Findings

Create a structured list of inconsistencies found in your focus area:

```markdown
## Inconsistencies Found

### [Focus Area Name] Mismatches
1. **Field `engine.version`**:
   - Schema: defines as string
   - Parser: not validated in frontmatter.go
   - Impact: Invalid values could pass through
```

### Step 5: Update Cache

Save successful strategy and focus-area rotation state to cache:

```bash
# Update strategies.json with results
cat > /tmp/gh-aw/cache-memory/strategies.json << 'EOF'
{
  "strategies": [...],
  "last_updated": "2024-XX-XX"
}
EOF

# IMPORTANT: Save the next focus area index so the next run picks up where this one left off.
# Read the next_area_index from focus-area.json (already computed by the pre-agent step).
NEXT_INDEX=$(jq -r '.next_area_index' /tmp/gh-aw/agent/focus-area.json)
cat > /tmp/gh-aw/cache-memory/focus-state.json << EOF
{
  "next_area_index": $NEXT_INDEX,
  "last_updated": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
```

### Step 6: Create Discussion

**⚠️ MANDATORY STEP**: After completing your analysis, you **MUST** call the `create_discussion` safe-output tool with your findings report. **DO NOT just write the report in your output text** — you MUST actually invoke the tool. The workflow will fail if you skip this step.

Use this discussion format for the content you pass to `create_discussion`:

## Discussion Report Format

Create a well-structured discussion report:

```markdown
# 🔍 Schema Consistency Check - [DATE]

## Summary

- **Focus Area**: [FOCUS AREA NAME]
- **Inconsistencies Found**: [NUMBER]
- **Strategy Used**: [STRATEGY NAME]
- **New Strategy**: [YES/NO]

## Critical Issues

[List high-priority inconsistencies that could cause bugs]

## Documentation Gaps

[List areas where docs don't match reality]

## Schema Improvements Needed

[List schema enhancements needed]

## Parser Updates Required

[List parser code that needs updates]

## Workflow Violations

[List workflows using invalid/undocumented features]

## Recommendations

1. [Specific actionable recommendation]
2. [Specific actionable recommendation]
3. [...]

## Strategy Performance

- **Strategy Used**: [NAME]
- **Findings**: [COUNT]
- **Effectiveness**: [HIGH/MEDIUM/LOW]
- **Should Reuse**: [YES/NO]

## Next Steps

- [ ] Fix schema definitions
- [ ] Update parser validation
- [ ] Update documentation
- [ ] Fix workflow files

## Coverage

This run analyzed: **[FOCUS AREA NAME]**. Remaining areas will be covered in subsequent runs.
```

## Important Guidelines

### Security
- Never execute untrusted code from workflows
- Validate all file paths before reading
- Sanitize all grep/bash commands
- Read-only access to schema, parser, and documentation files for analysis
- Only modify files in `/tmp/gh-aw/cache-memory/` (never modify source files)

### Quality
- Be thorough but focused on actionable findings
- Prioritize issues by severity (critical bugs vs documentation gaps)
- Provide specific file:line references when possible
- Include code snippets to illustrate issues
- Suggest concrete fixes

### Efficiency
- **Always start from the pre-computed files** — they eliminate the need to re-read all source files
- Use targeted bash commands to verify specific leads from the pre-computed data
- Cache results when re-analyzing same data
- Don't re-check things found in previous runs (check cache first)
- Focus on high-impact areas (field gaps with parser mismatches are usually most critical)
- **Stay within the 40-turn budget** — if you reach turn 30, finalize immediately

### Strategy Evolution
- Try genuinely different approaches when not using cached strategies
- Document why a strategy worked or failed
- Update success metrics in cache
- Consider combining successful strategies

## Tools Available

You have access to:
- **bash**: Any command (use grep, jq, find, cat, etc.)
- **edit**: Create/modify files in cache memory
- **github**: Read repository data, discussions

## Success Criteria

A successful run:
- ✅ Reads the focus area and only analyzes that area
- ✅ Completes within 40 turns
- ✅ Uses or creates an effective detection strategy
- ✅ Updates cache with strategy results **and** next focus area index
- ✅ Finds at least one category of inconsistencies OR confirms consistency
- ✅ Creates a detailed discussion report
- ✅ Provides actionable recommendations

Begin your analysis now. Read the focus area, check the cache, choose a strategy, execute it, and **call `create_discussion` with your findings** to complete the workflow.

{{#runtime-import shared/noop-reminder.md}}
