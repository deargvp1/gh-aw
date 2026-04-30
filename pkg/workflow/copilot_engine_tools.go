// This file provides Copilot engine tool permission and error pattern logic.
//
// This file handles three key responsibilities:
//
//  1. Tool Permission Arguments (computeCopilotToolArguments):
//     Converts workflow tool configurations into --allow-tool flags for Copilot CLI.
//     Handles bash/shell tools, edit tools, safe outputs, mcp-scripts, and MCP servers.
//     Supports granular permissions (e.g., "github(get_file)") and server-level wildcards.
//
//  2. Tool Argument Comments (generateCopilotToolArgumentsComment):
//     Generates human-readable comments documenting which tool permissions are granted.
//     Used in compiled workflows for transparency and debugging.
//
//  3. Error Patterns (GetErrorPatterns):
//     Defines regex patterns for extracting error messages from Copilot CLI logs.
//     Includes timestamped log formats, command failures, module errors, and permission issues.
//     Used by log parsers to detect and categorize errors.
//
// These functions are grouped together because they all relate to tool configuration
// and error handling in the Copilot engine.

package workflow

import (
	"fmt"
	"sort"
	"strings"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var copilotEngineToolsLog = logger.New("workflow:copilot_engine_tools")

// computeCopilotToolArguments computes the --allow-tool arguments for Copilot CLI based on tool configurations.
// It handles bash/shell tools, edit tools, safe outputs, mcp-scripts, and MCP server tools.
// Returns a sorted list of arguments ready to be passed to the Copilot CLI.
//
// When tools.bash is a GitHub Actions expression, bash tool arguments are omitted from the
// returned list. The caller (GetExecutionSteps) is responsible for detecting this case via
// workflowData.ParsedTools.Bash.AllowedCommandsExpr and injecting a runtime preamble that
// dynamically builds shell() --allow-tool arguments from the GH_AW_BASH_ALLOWLIST env var.
//
// When tools.edit is a GitHub Actions expression, edit tool arguments are omitted from the
// returned list. The caller is responsible for detecting this and conditionally adding
// --allow-tool write via a runtime check on GH_AW_EDIT_ENABLED.
func (e *CopilotEngine) computeCopilotToolArguments(tools map[string]any, safeOutputs *SafeOutputsConfig, mcpScripts *MCPScriptsConfig, workflowData *WorkflowData) []string {
	copilotEngineToolsLog.Printf("Computing tool arguments: tools=%d", len(tools))
	if tools == nil {
		tools = make(map[string]any)
	}

	var args []string
	hasRestrictedBashAllowlist := false

	// Check if bash has wildcard - if so, use --allow-all-tools instead
	if bashConfig, hasBash := tools["bash"]; hasBash {
		if bashCommands, ok := bashConfig.([]any); ok {
			// Check for :* or * wildcard - if present, allow all tools
			for _, cmd := range bashCommands {
				if cmdStr, ok := cmd.(string); ok {
					if cmdStr == ":*" || cmdStr == "*" {
						// Use --allow-all-tools flag instead of individual tool permissions
						copilotEngineToolsLog.Print("Bash wildcard detected, using --allow-all-tools")
						return []string{"--allow-all-tools"}
					}
				}
			}
		}
	}

	// Handle bash/shell tools (when no wildcard)
	if bashConfig, hasBash := tools["bash"]; hasBash {
		if bashExpr, ok := bashConfig.(string); ok && isExpression(bashExpr) {
			// GitHub Actions expression: tool arguments are built dynamically at runtime
			// from GH_AW_BASH_ALLOWLIST env var. Treat as restricted (no static shell args).
			copilotEngineToolsLog.Printf("Bash tool is a runtime expression, deferring shell args to runtime: %s", bashExpr)
			hasRestrictedBashAllowlist = true
			// No static --allow-tool shell(...) args added here.
		} else if bashCommands, ok := bashConfig.([]any); ok {
			hasRestrictedBashAllowlist = true
			// Add specific shell commands
			for _, cmd := range bashCommands {
				if cmdStr, ok := cmd.(string); ok {
					// For stem commands (like dotnet, npm, cargo), Copilot CLI uses
					// subcommand matching. When the user specifies just the base command
					// (e.g., "dotnet"), append :* so "dotnet build", "dotnet test", etc.
					// are all permitted. Skip if the command already has a colon (explicit
					// matching) or a space (user already specified the subcommand).
					if !strings.Contains(cmdStr, ":") && !strings.Contains(cmdStr, " ") && constants.CopilotStemCommands[cmdStr] {
						args = append(args, "--allow-tool", fmt.Sprintf("shell(%s:*)", cmdStr))
					} else {
						args = append(args, "--allow-tool", fmt.Sprintf("shell(%s)", cmdStr))
					}
				}
			}
		} else {
			// Bash with no specific commands or null value - allow all shell
			args = append(args, "--allow-tool", "shell")
		}
	}

	// When MCP tools are mounted as CLI commands and bash uses a restricted allowlist,
	// ensure mounted MCP CLI commands are executable via shell(<server>:*).
	// This avoids Copilot CLI permission blocks for mounted commands such as safeoutputs.
	if hasRestrictedBashAllowlist {
		for _, serverName := range getMountedCLIServerNamesIfBashRestricted(workflowData, tools, safeOutputs, mcpScripts) {
			args = append(args, "--allow-tool", fmt.Sprintf("shell(%s:*)", serverName))
		}
	}

	// Handle edit tools requirement for file write access
	// Note: safe-outputs do not need write permission as they use MCP
	if editConfig, hasEdit := tools["edit"]; hasEdit {
		if editExpr, ok := editConfig.(string); ok && isExpression(editExpr) {
			// GitHub Actions expression: edit permission added dynamically at runtime
			// via GH_AW_EDIT_ENABLED env var. No static --allow-tool write added here.
			copilotEngineToolsLog.Printf("Edit tool is a runtime expression, deferring write permission to runtime: %s", editExpr)
		} else {
			copilotEngineToolsLog.Print("Edit tool enabled, adding write permission")
			args = append(args, "--allow-tool", "write")
		}
	}

	// Handle safe_outputs MCP server - allow all tools if safe outputs are enabled
	// This includes both safeOutputs config and safeOutputs.Jobs
	if HasSafeOutputsEnabled(safeOutputs) {
		copilotEngineToolsLog.Print("Safe-outputs enabled, adding MCP server permission")
		args = append(args, "--allow-tool", constants.SafeOutputsMCPServerID.String())
	}

	// Handle mcp_scripts MCP server - allow the server if mcp-scripts are configured and feature flag is enabled
	if IsMCPScriptsEnabled(mcpScripts) {
		args = append(args, "--allow-tool", constants.MCPScriptsMCPServerID.String())
	}

	// Handle web-fetch builtin tool (Copilot CLI uses web_fetch with underscore)
	if _, hasWebFetch := tools["web-fetch"]; hasWebFetch {
		copilotEngineToolsLog.Print("Web-fetch tool enabled, adding web_fetch permission")
		// web-fetch -> web_fetch
		args = append(args, "--allow-tool", "web_fetch")
	}

	// Built-in tool names that should be skipped when processing MCP servers
	// Note: GitHub is NOT included here because it needs MCP configuration in CLI mode
	// Note: web-fetch is NOT included here because it needs explicit --allow-tool argument
	builtInTools := map[string]bool{
		"bash":       true,
		"edit":       true,
		"web-search": true,
		"playwright": true,
	}

	// Handle MCP server tools
	for toolName, toolConfig := range tools {
		// Skip built-in tools we've already handled
		if builtInTools[toolName] {
			continue
		}

		// GitHub is a special case - it's an MCP server but doesn't have explicit MCP config in the workflow
		// It gets MCP configuration through the parser's processBuiltinMCPTool
		if toolName == "github" {
			if toolConfigMap, ok := toolConfig.(map[string]any); ok {
				if allowed, hasAllowed := toolConfigMap["allowed"]; hasAllowed {
					if allowedList, ok := allowed.([]any); ok {
						// Process allowed list in a single pass
						hasWildcard := false
						for _, allowedTool := range allowedList {
							if toolStr, ok := allowedTool.(string); ok {
								if toolStr == "*" {
									// Wildcard means allow entire GitHub MCP server
									hasWildcard = true
								} else {
									// Add individual tool permission
									args = append(args, "--allow-tool", fmt.Sprintf("github(%s)", toolStr))
								}
							}
						}

						// Add server-level permission only if wildcard was present
						if hasWildcard {
							args = append(args, "--allow-tool", "github")
						}
					}
				} else {
					// No allowed field specified - allow entire GitHub MCP server
					args = append(args, "--allow-tool", "github")
				}
			} else {
				// GitHub tool exists but is not a map (e.g., github: null) - allow entire server
				args = append(args, "--allow-tool", "github")
			}
			continue
		}

		// Check if this is an MCP server configuration
		if toolConfigMap, ok := toolConfig.(map[string]any); ok {
			if hasMcp, _ := hasMCPConfig(toolConfigMap); hasMcp {
				copilotEngineToolsLog.Printf("Adding custom MCP server permission: %s", toolName)
				// Allow the entire MCP server
				args = append(args, "--allow-tool", toolName)

				// If it has specific allowed tools, add them individually
				if allowed, hasAllowed := toolConfigMap["allowed"]; hasAllowed {
					if allowedList, ok := allowed.([]any); ok {
						for _, allowedTool := range allowedList {
							if toolStr, ok := allowedTool.(string); ok {
								args = append(args, "--allow-tool", fmt.Sprintf("%s(%s)", toolName, toolStr))
							}
						}
					}
				}
			}
		}
	}

	// Simple sort - extract values, sort them, and rebuild args
	if len(args) > 0 {
		var values []string
		for i := 1; i < len(args); i += 2 {
			values = append(values, args[i])
		}
		sort.Strings(values)

		// Rebuild args with sorted values
		newArgs := make([]string, 0, len(args))
		for _, value := range values {
			newArgs = append(newArgs, "--allow-tool", value)
		}
		args = newArgs
	}

	copilotEngineToolsLog.Printf("Computed %d tool arguments", len(args)/2)
	return args
}

// generateCopilotToolArgumentsComment generates a multi-line comment showing each tool argument.
// This is used to document which tool permissions are being granted in the compiled workflow.
func (e *CopilotEngine) generateCopilotToolArgumentsComment(tools map[string]any, safeOutputs *SafeOutputsConfig, mcpScripts *MCPScriptsConfig, workflowData *WorkflowData, indent string) string {
	toolArgs := e.computeCopilotToolArguments(tools, safeOutputs, mcpScripts, workflowData)
	if len(toolArgs) == 0 {
		return ""
	}

	var comment strings.Builder
	comment.WriteString(indent + "# Copilot CLI tool arguments (sorted):\n")

	// Group flag-value pairs for better readability
	for i := 0; i < len(toolArgs); i += 2 {
		if i+1 < len(toolArgs) {
			fmt.Fprintf(&comment, "%s# %s %s\n", indent, toolArgs[i], toolArgs[i+1])
		}
	}

	return comment.String()
}

// buildCopilotDynamicToolArgsPreamble generates a bash preamble script that builds
// --allow-tool arguments for parameterized bash and edit tools at runtime.
//
// For bash expressions: reads GH_AW_BASH_ALLOWLIST (comma-separated commands)
// and populates _bash_tool_args array with --allow-tool shell(<cmd>) entries.
//
// For edit expressions: reads GH_AW_EDIT_ENABLED ("true" to enable) and
// populates _edit_tool_args with --allow-tool write.
//
// The caller injects these arrays into the copilot command as:
//
//	copilot "${_bash_tool_args[@]}" "${_edit_tool_args[@]}" <other-args>
func buildCopilotDynamicToolArgsPreamble(hasBashExpr, hasEditExpr bool) string {
	if !hasBashExpr && !hasEditExpr {
		return ""
	}

	var preamble strings.Builder

	if hasBashExpr {
		// Build _bash_tool_args from GH_AW_BASH_ALLOWLIST (comma-separated commands).
		// Each command becomes --allow-tool shell(<cmd>). Empty or whitespace-only
		// entries are skipped. The array is empty when GH_AW_BASH_ALLOWLIST is unset
		// or empty, which results in no shell access being granted (fail-closed).
		preamble.WriteString("_bash_tool_args=()\n")
		preamble.WriteString("if [ -n \"${GH_AW_BASH_ALLOWLIST:-}\" ]; then\n")
		preamble.WriteString("  while IFS= read -r _cmd; do\n")
		// Trim leading and trailing whitespace from each command
		preamble.WriteString("    _cmd=\"${_cmd#\"${_cmd%%[![:space:]]*}\"}\"\n")
		preamble.WriteString("    _cmd=\"${_cmd%\"${_cmd##*[![:space:]]}\"}\"\n")
		preamble.WriteString("    [ -z \"$_cmd\" ] && continue\n")
		preamble.WriteString("    _bash_tool_args+=(\"--allow-tool\" \"shell($_cmd)\")\n")
		preamble.WriteString("  done < <(printf '%s\\n' \"${GH_AW_BASH_ALLOWLIST}\" | tr ',' '\\n')\n")
		preamble.WriteString("fi\n")
	}

	if hasEditExpr {
		// Conditionally add --allow-tool write based on GH_AW_EDIT_ENABLED env var.
		// The array is empty when the expression evaluates to anything other than "true".
		preamble.WriteString("_edit_tool_args=()\n")
		preamble.WriteString("if [ \"${GH_AW_EDIT_ENABLED:-}\" = \"true\" ]; then\n")
		preamble.WriteString("  _edit_tool_args=(\"--allow-tool\" \"write\")\n")
		preamble.WriteString("fi\n")
	}

	return preamble.String()
}
