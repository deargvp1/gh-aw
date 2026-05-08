package workflow

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/parser"
	"github.com/goccy/go-yaml"
)

var frontmatterLog = logger.New("workflow:frontmatter_extraction")

// indentYAMLLines adds indentation to all lines of a multi-line YAML string except the first
func (c *Compiler) indentYAMLLines(yamlContent, indent string) string {
	if yamlContent == "" {
		return yamlContent
	}

	lines := strings.Split(yamlContent, "\n")
	if len(lines) <= 1 {
		return yamlContent
	}

	// First line doesn't get additional indentation
	var result strings.Builder
	result.WriteString(lines[0])
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			result.WriteString("\n" + indent + lines[i])
		} else {
			result.WriteString("\n" + lines[i])
		}
	}

	return result.String()
}

// extractTopLevelYAMLSection extracts a top-level YAML section from frontmatter
func (c *Compiler) extractTopLevelYAMLSection(frontmatter map[string]any, key string) string {
	value, exists := frontmatter[key]
	if !exists {
		return ""
	}

	frontmatterLog.Printf("Extracting YAML section: %s", key)

	// Convert the value back to YAML format with field ordering
	var yamlBytes []byte
	var err error

	// Check if value is a map that we should order alphabetically
	if valueMap, ok := value.(map[string]any); ok {
		// Use OrderMapFields for alphabetical sorting (empty priority list = all alphabetical)
		orderedValue := OrderMapFields(valueMap, []string{})
		// Wrap the ordered value with the key using MapSlice
		wrappedData := yaml.MapSlice{{Key: key, Value: orderedValue}}
		yamlBytes, err = yaml.MarshalWithOptions(wrappedData, DefaultMarshalOptions...)
		if err != nil {
			return ""
		}
	} else {
		// Use standard marshaling for non-map types
		yamlBytes, err = yaml.Marshal(map[string]any{key: value})
		if err != nil {
			return ""
		}
	}

	yamlStr := string(yamlBytes)
	// Remove the trailing newline
	yamlStr = strings.TrimSuffix(yamlStr, "\n")

	// Post-process YAML to ensure cron expressions are quoted
	// The YAML library may drop quotes from cron expressions like "0 14 * * 1-5"
	// which causes validation errors since they start with numbers but contain spaces
	yamlStr = parser.QuoteCronExpressions(yamlStr)

	// Clean up null values - replace `: null` with `:` for cleaner output
	// GitHub Actions treats `workflow_dispatch:` and `workflow_dispatch: null` identically
	yamlStr = CleanYAMLNullValues(yamlStr)

	// Clean up quoted keys - replace "key": with key: at the start of a line
	// Don't unquote "on" key as it's a YAML boolean keyword and must remain quoted
	if key != "on" {
		yamlStr = UnquoteYAMLKey(yamlStr, key)
	}

	// Special handling for "on" section - comment out draft and fork fields from pull_request
	if key == "on" {
		yamlStr = c.commentOutProcessedFieldsInOnSection(yamlStr, frontmatter)
		// Add zizmor ignore comment if workflow_run trigger is present
		yamlStr = c.addZizmorIgnoreForWorkflowRun(yamlStr)
		// Add friendly format comments for schedule cron expressions
		yamlStr = c.addFriendlyScheduleComments(yamlStr, frontmatter)
	}

	return yamlStr
}

// commentOutProcessedFieldsInOnSection comments out draft, fork, forks, names, manual-approval, stop-after, skip-if-match, skip-if-no-match, skip-roles, reaction, lock-for-agent, steps, permissions, and stale-check fields in the on section
// These fields are processed separately and should be commented for documentation
// Exception: names fields in sections with __gh_aw_native_label_filter__ marker in frontmatter are NOT commented out
func (c *Compiler) commentOutProcessedFieldsInOnSection(yamlStr string, frontmatter map[string]any) string {
	frontmatterLog.Print("Processing 'on' section to comment out processed fields")

	nativeLabelFilterSections := buildNativeLabelFilterSections(frontmatter)

	lines := strings.Split(yamlStr, "\n")
	var result []string
	state := &onSectionParseState{}

	for _, line := range lines {
		// Section entry detection — updates state and signals caller to append+continue.
		// Skip these checks when inside on.permissions or on.steps to avoid false matches
		// (e.g. `    issues: read` inside on.permissions previously triggered inIssues).
		if state.updateSectionEntry(line) {
			result = append(result, line)
			continue
		}

		// Update state for section and object exits
		state.updateSectionExit(line)

		trimmedLine := strings.TrimSpace(line)

		// Skip native-label-filter marker lines from the YAML output
		if (state.inPullRequest || state.inIssues || state.inDiscussion || state.inIssueComment) &&
			strings.Contains(trimmedLine, "__gh_aw_native_label_filter__:") {
			continue
		}

		// Update state for entering object/array blocks
		state.updateObjectEntry(trimmedLine)

		// Update state for exiting object/array blocks
		state.updateObjectExit(line, trimmedLine)

		// Classify line for comment-out
		shouldComment, commentReason := state.classifyForCommentOut(trimmedLine, line, nativeLabelFilterSections, result)

		if shouldComment {
			indentation := ""
			trimmed := strings.TrimLeft(line, " \t")
			if len(line) > len(trimmed) {
				indentation = line[:len(line)-len(trimmed)]
			}
			result = append(result, indentation+"# "+trimmed+commentReason)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// addZizmorIgnoreForWorkflowRun adds a zizmor ignore comment for workflow_run triggers
// The comment is added after the workflow_run: line to suppress dangerous-triggers warnings
// since the compiler adds proper role and fork validation to secure these triggers
func (c *Compiler) addZizmorIgnoreForWorkflowRun(yamlStr string) string {
	// Check if the YAML contains workflow_run trigger
	if !strings.Contains(yamlStr, "workflow_run:") {
		return yamlStr
	}
	frontmatterLog.Print("Adding zizmor ignore annotation for workflow_run trigger")

	lines := strings.Split(yamlStr, "\n")
	var result []string
	annotationAdded := false // Track if we've already added the annotation

	for _, line := range lines {
		result = append(result, line)

		// Skip if we've already added the annotation (prevents duplicates)
		if annotationAdded {
			continue
		}

		// Check if this is a non-comment workflow_run: key at the correct YAML level
		trimmedLine := strings.TrimSpace(line)

		// Skip if the line is a comment
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Match lines that are only 'workflow_run:' (possibly with trailing whitespace or a comment)
		// e.g., 'workflow_run:', 'workflow_run: # comment', '  workflow_run:'
		// But not 'someworkflow_run:', 'workflow_run: value', etc.
		if idx := strings.Index(trimmedLine, "workflow_run:"); idx == 0 {
			after := strings.TrimSpace(trimmedLine[len("workflow_run:"):])
			// Only allow if nothing or only a comment follows
			if after == "" || strings.HasPrefix(after, "#") {
				// Get the indentation of the workflow_run line
				indentation := ""
				if len(line) > len(trimmedLine) {
					indentation = line[:len(line)-len(trimmedLine)]
				}

				// Add zizmor ignore comment with proper indentation
				// The comment explains that the trigger is secured with role and fork validation
				comment := indentation + "  # zizmor: ignore[dangerous-triggers] - workflow_run trigger is secured with role and fork validation"
				result = append(result, comment)
				annotationAdded = true
			}
		}
	}

	return strings.Join(result, "\n")
}

// extractPermissions extracts permissions from frontmatter using the permission parser
func (c *Compiler) extractPermissions(frontmatter map[string]any) string {
	permissionsValue, exists := frontmatter["permissions"]
	if !exists {
		frontmatterLog.Print("No permissions field found in frontmatter")
		return ""
	}

	// Check if this is an "all: read" case by using the parser
	parser := NewPermissionsParserFromValue(permissionsValue)

	// If it's "all: read", use the parser to expand it
	if parser.hasAll && parser.allLevel == "read" {
		frontmatterLog.Print("Expanding 'all: read' permissions to individual scopes")
		permissions := parser.ToPermissions()
		yaml := permissions.RenderToYAML()

		// Adjust indentation from 6 spaces to 2 spaces for workflow-level permissions
		// RenderToYAML uses 6 spaces for job-level rendering
		lines := strings.Split(yaml, "\n")
		for i := 1; i < len(lines); i++ {
			if strings.HasPrefix(lines[i], "      ") {
				lines[i] = "  " + lines[i][6:]
			}
		}
		return strings.Join(lines, "\n")
	}

	// For all other cases, use standard extraction
	return c.extractTopLevelYAMLSection(frontmatter, "permissions")
}

// extractIfCondition extracts the if condition from frontmatter, returning just the expression
// without the "if: " prefix. Also merges any condition derived from on.deployment_status.state
// and on.workflow_run.conclusion.
func (c *Compiler) extractIfCondition(frontmatter map[string]any) (string, error) {
	var ifExpr string
	if value, exists := frontmatter["if"]; exists {
		if strValue, ok := value.(string); ok {
			// Strip "if: " prefix and ${{ }} wrapper to get a bare expression for safe merging
			ifExpr = stripExpressionWrapper(c.extractExpressionFromIfString(strValue))
			frontmatterLog.Printf("Extracted if condition from frontmatter: %s", ifExpr)
		}
	}

	// Merge any condition generated from on.deployment_status.state
	stateCondition := extractDeploymentStatusStateCondition(frontmatter)
	if stateCondition != "" {
		frontmatterLog.Printf("Merging deployment_status state condition: %s", stateCondition)
		if ifExpr != "" {
			ifExpr = "(" + ifExpr + ") && (" + stateCondition + ")"
		} else {
			ifExpr = stateCondition
		}
	}

	// Merge any condition generated from on.workflow_run.conclusion
	conclusionCondition, err := extractWorkflowRunConclusionCondition(frontmatter)
	if err != nil {
		return "", err
	}
	if conclusionCondition != "" {
		frontmatterLog.Printf("Merging workflow_run conclusion condition: %s", conclusionCondition)
		if ifExpr != "" {
			ifExpr = "(" + ifExpr + ") && (" + conclusionCondition + ")"
		} else {
			ifExpr = conclusionCondition
		}
	}

	return ifExpr, nil
}

// extractDeploymentStatusStateCondition reads on.deployment_status.state and converts it
// into a GitHub Actions expression string (without ${{ }} wrappers). Returns "" if not set.
func extractDeploymentStatusStateCondition(frontmatter map[string]any) string {
	onValue, ok := frontmatter["on"]
	if !ok {
		return ""
	}
	onMap, ok := onValue.(map[string]any)
	if !ok {
		return ""
	}
	dsValue, ok := onMap["deployment_status"]
	if !ok {
		return ""
	}
	dsMap, ok := dsValue.(map[string]any)
	if !ok {
		return ""
	}
	stateValue, ok := dsMap["state"]
	if !ok {
		return ""
	}

	// GitHub Actions allows state as a single string or an array
	var states []string
	if s, ok := stateValue.(string); ok {
		states = []string{s}
	} else {
		states = parseStringSliceAny(stateValue, nil)
	}

	if len(states) == 0 {
		return ""
	}

	parts := make([]string, 0, len(states))
	for _, s := range states {
		parts = append(parts, "github.event.deployment_status.state == '"+s+"'")
	}
	stateExpr := strings.Join(parts, " || ")

	// Guard the state check with an event_name test so the condition remains true
	// when the workflow is triggered by other events (e.g. workflow_dispatch).
	// Without the guard, a non-deployment_status event would see the state as
	// empty/undefined and the entire activation condition would evaluate to false.
	return "github.event_name != 'deployment_status' || (" + stateExpr + ")"
}

// validWorkflowRunConclusions is the exhaustive list of conclusion values that GitHub
// Actions emits for workflow_run events.  Values outside this set are rejected at
// compile time to prevent expression injection (a raw value is interpolated directly
// into a GitHub Actions expression string).
var validWorkflowRunConclusions = []string{
	"success",
	"failure",
	"neutral",
	"cancelled",
	"skipped",
	"timed_out",
	"action_required",
	"stale",
}

// isValidWorkflowRunConclusion reports whether v is a recognised conclusion value.
func isValidWorkflowRunConclusion(v string) bool {
	return slices.Contains(validWorkflowRunConclusions, v)
}

// extractWorkflowRunConclusionCondition reads on.workflow_run.conclusion and converts it
// into a GitHub Actions expression string (without ${{ }} wrappers). Returns "" if not set.
func extractWorkflowRunConclusionCondition(frontmatter map[string]any) (string, error) {
	onValue, ok := frontmatter["on"]
	if !ok {
		return "", nil
	}
	onMap, ok := onValue.(map[string]any)
	if !ok {
		return "", nil
	}
	wrValue, ok := onMap["workflow_run"]
	if !ok {
		return "", nil
	}
	wrMap, ok := wrValue.(map[string]any)
	if !ok {
		return "", nil
	}
	conclusionValue, ok := wrMap["conclusion"]
	if !ok {
		return "", nil
	}

	var conclusions []string
	switch v := conclusionValue.(type) {
	case string:
		conclusions = []string{v}
	case []any:
		for _, s := range v {
			if str, ok := s.(string); ok {
				conclusions = append(conclusions, str)
			}
		}
	}

	if len(conclusions) == 0 {
		return "", nil
	}

	for _, c := range conclusions {
		if !isValidWorkflowRunConclusion(c) {
			return "", fmt.Errorf("invalid on.workflow_run.conclusion value %q: must be one of %s",
				c, strings.Join(validWorkflowRunConclusions, ", "))
		}
	}

	parts := make([]string, 0, len(conclusions))
	for _, c := range conclusions {
		parts = append(parts, "github.event.workflow_run.conclusion == '"+c+"'")
	}
	conclusionExpr := strings.Join(parts, " || ")

	// Guard the conclusion check with an event_name test so the condition remains true
	// when the workflow is triggered by other events (e.g. workflow_dispatch).
	// Without the guard, a non-workflow_run event would see conclusion as
	// empty/undefined and the entire activation condition would evaluate to false.
	return "github.event_name != 'workflow_run' || (" + conclusionExpr + ")", nil
}

// extractExpressionFromIfString extracts the expression part from a string that might
// contain "if: expression" or just "expression", returning just the expression
func (c *Compiler) extractExpressionFromIfString(ifString string) string {
	if ifString == "" {
		return ""
	}

	// Check if the string starts with "if: " and strip it
	if strings.HasPrefix(ifString, "if: ") {
		expr := strings.TrimSpace(ifString[4:]) // Remove "if: " prefix
		frontmatterLog.Printf("Stripped 'if: ' prefix from if condition: %s", expr)
		return expr
	}

	// Return the string as-is (it's just the expression)
	return ifString
}

// extractCommandConfig extracts command configuration from frontmatter including name and events
func (c *Compiler) extractCommandConfig(frontmatter map[string]any) (commandNames []string, commandEvents []string) {
	frontmatterLog.Print("Extracting command configuration from frontmatter")
	// Check new format: on.slash_command or on.slash_command.name (preferred)
	// Also check legacy format: on.command or on.command.name (deprecated)
	if onValue, exists := frontmatter["on"]; exists {
		if onMap, ok := onValue.(map[string]any); ok {
			var commandValue any
			var hasCommand bool
			var isDeprecated bool

			// Check for slash_command first (preferred)
			if slashCommandValue, hasSlashCommand := onMap["slash_command"]; hasSlashCommand {
				commandValue = slashCommandValue
				hasCommand = true
				isDeprecated = false
			} else if legacyCommandValue, hasLegacyCommand := onMap["command"]; hasLegacyCommand {
				// Fall back to command (deprecated)
				commandValue = legacyCommandValue
				hasCommand = true
				isDeprecated = true
			}

			if hasCommand {
				// Show deprecation warning if using old field name
				if isDeprecated {
					fmt.Fprintln(os.Stderr, console.FormatWarningMessage("The 'command:' trigger field is deprecated. Please use 'slash_command:' instead."))
					c.IncrementWarningCount()
				}

				// Check if command is a string (shorthand format)
				if commandStr, ok := commandValue.(string); ok {
					frontmatterLog.Printf("Extracted command name (shorthand): %s", commandStr)
					return []string{commandStr}, nil // nil means default (all events)
				}
				// Check if command is a map with a name key (object format)
				if commandMap, ok := commandValue.(map[string]any); ok {
					var names []string
					var events []string

					if nameValue, hasName := commandMap["name"]; hasName {
						// Handle string or array of strings
						if nameStr, ok := nameValue.(string); ok {
							names = []string{nameStr}
						} else if nameArray, ok := nameValue.([]any); ok {
							for _, nameItem := range nameArray {
								if nameItemStr, ok := nameItem.(string); ok {
									names = append(names, nameItemStr)
								}
							}
						}
					}

					// Extract events field
					if eventsValue, hasEvents := commandMap["events"]; hasEvents {
						events = ParseCommandEvents(eventsValue)
					}

					frontmatterLog.Printf("Extracted command config: names=%v, events=%v", names, events)
					return names, events
				}
			}
		}
	}

	return nil, nil
}

// extractLabelCommandConfig extracts the label-command configuration from frontmatter
// including label name(s), the events field, and the remove_label flag.
// It reads on.label_command which can be:
//   - a string: label name directly (e.g. label_command: "deploy")
//   - a map with "name" or "names", optional "events", and optional "remove_label" fields
//
// Returns (labelNames, labelEvents, removeLabel) where labelEvents is nil for default (all events)
// and removeLabel defaults to true when not specified.
func (c *Compiler) extractLabelCommandConfig(frontmatter map[string]any) (labelNames []string, labelEvents []string, removeLabel bool) {
	frontmatterLog.Print("Extracting label-command configuration from frontmatter")
	onValue, exists := frontmatter["on"]
	if !exists {
		return nil, nil, true
	}
	onMap, ok := onValue.(map[string]any)
	if !ok {
		return nil, nil, true
	}
	labelCommandValue, hasLabelCommand := onMap["label_command"]
	if !hasLabelCommand {
		return nil, nil, true
	}

	// Simple string form: label_command: "my-label"
	if nameStr, ok := labelCommandValue.(string); ok {
		frontmatterLog.Printf("Extracted label-command name (shorthand): %s", nameStr)
		return []string{nameStr}, nil, true
	}

	// Map form: label_command: {name: "...", names: [...], events: [...], remove_label: bool}
	if lcMap, ok := labelCommandValue.(map[string]any); ok {
		var names []string
		var events []string
		removeLabelVal := true // default to true

		if nameVal, hasName := lcMap["name"]; hasName {
			if nameStr, ok := nameVal.(string); ok {
				names = []string{nameStr}
			} else if nameArray, ok := nameVal.([]any); ok {
				for _, item := range nameArray {
					if s, ok := item.(string); ok {
						names = append(names, s)
					}
				}
			}
		}
		if namesVal, hasNames := lcMap["names"]; hasNames {
			if namesArray, ok := namesVal.([]any); ok {
				for _, item := range namesArray {
					if s, ok := item.(string); ok {
						names = append(names, s)
					}
				}
			} else if namesStr, ok := namesVal.(string); ok {
				names = append(names, namesStr)
			}
		}

		if eventsVal, hasEvents := lcMap["events"]; hasEvents {
			events = ParseCommandEvents(eventsVal)
		}

		if removeLabelField, hasRemoveLabel := lcMap["remove_label"]; hasRemoveLabel {
			if b, ok := removeLabelField.(bool); ok {
				removeLabelVal = b
			}
		}

		frontmatterLog.Printf("Extracted label-command config: names=%v, events=%v, remove_label=%v", names, events, removeLabelVal)
		return names, events, removeLabelVal
	}

	return nil, nil, true
}

// isGitHubAppNestedField returns true if the trimmed YAML line represents a known
// nested field or array item inside an on.github-app object.
func isGitHubAppNestedField(trimmedLine string) bool {
	githubAppFields := []string{"app-id:", "client-id:", "private-key:", "owner:", "repositories:"}
	for _, field := range githubAppFields {
		if strings.HasPrefix(trimmedLine, field) {
			return true
		}
	}
	// Array items (repositories list)
	return strings.HasPrefix(trimmedLine, "-")
}
