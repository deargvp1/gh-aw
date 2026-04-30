//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseBashToolExpression verifies that parseBashTool handles expression strings.
func TestParseBashToolExpression(t *testing.T) {
	tests := []struct {
		name                string
		input               any
		wantExpr            string
		wantAllowedCommands []string
		wantNil             bool
		wantUnrestricted    bool // true = AllowedCommands is nil but BashToolConfig is not nil
	}{
		{
			name:     "expression string",
			input:    "${{ inputs.bash-allowlist }}",
			wantExpr: "${{ inputs.bash-allowlist }}",
		},
		{
			name:     "expression with whitespace around braces",
			input:    "${{ inputs.bash_tools }}",
			wantExpr: "${{ inputs.bash_tools }}",
		},
		{
			name:             "bash: true (unrestricted)",
			input:            true,
			wantUnrestricted: true,
		},
		{
			name:                "bash: false (disabled)",
			input:               false,
			wantAllowedCommands: []string{},
		},
		{
			name:                "bash: [git, npm]",
			input:               []any{"git", "npm"},
			wantAllowedCommands: []string{"git", "npm"},
		},
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:    "non-expression plain string is rejected",
			input:   "git",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBashTool(tt.input)

			if tt.wantNil {
				assert.Nil(t, result, "expected nil BashToolConfig")
				return
			}
			require.NotNil(t, result, "expected non-nil BashToolConfig")

			if tt.wantExpr != "" {
				assert.Equal(t, tt.wantExpr, result.AllowedCommandsExpr, "AllowedCommandsExpr mismatch")
				assert.Nil(t, result.AllowedCommands, "AllowedCommands should be nil for expressions")
			} else if tt.wantUnrestricted {
				assert.Nil(t, result.AllowedCommands, "AllowedCommands should be nil for unrestricted bash")
				assert.Empty(t, result.AllowedCommandsExpr, "AllowedCommandsExpr should be empty")
			} else {
				assert.Equal(t, tt.wantAllowedCommands, result.AllowedCommands, "AllowedCommands mismatch")
				assert.Empty(t, result.AllowedCommandsExpr, "AllowedCommandsExpr should be empty for literal arrays")
			}
		})
	}
}

// TestGetGitHubToolsetsExpression verifies that getGitHubToolsets passes expressions through.
func TestGetGitHubToolsetsExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name: "Expression string is passed through",
			input: map[string]any{
				"toolsets": "${{ inputs.github-toolsets }}",
			},
			expected: "${{ inputs.github-toolsets }}",
		},
		{
			name: "Complex expression with fallback is passed through",
			input: map[string]any{
				"toolsets": "${{ inputs.toolsets || 'repos,issues' }}",
			},
			expected: "${{ inputs.toolsets || 'repos,issues' }}",
		},
		{
			name: "Literal array still works",
			input: map[string]any{
				"toolsets": []any{"repos", "issues"},
			},
			expected: "repos,issues",
		},
		{
			name:     "No toolsets uses action-friendly defaults",
			input:    map[string]any{},
			expected: "context,repos,issues,pull_requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGitHubToolsets(tt.input)
			assert.Equal(t, tt.expected, result, "getGitHubToolsets mismatch")
		})
	}
}

// TestGetToolsetsWithExpression verifies GitHubToolConfig.GetToolsets skips validation for expressions.
func TestGetToolsetsWithExpression(t *testing.T) {
	t.Run("expression returns empty (skips compile-time validation)", func(t *testing.T) {
		config := &GitHubToolConfig{
			ToolsetExpr: "${{ inputs.github-toolsets }}",
		}
		result := config.GetToolsets()
		assert.Empty(t, result, "GetToolsets should return empty string for expressions to skip compile-time validation")
	})

	t.Run("literal toolsets returns expanded string", func(t *testing.T) {
		config := &GitHubToolConfig{
			Toolset: GitHubToolsets{GitHubToolset("repos"), GitHubToolset("issues")},
		}
		result := config.GetToolsets()
		assert.Equal(t, "repos,issues", result, "GetToolsets should return comma-separated toolsets")
	})
}

// TestBashToolConfigToMap verifies that ToMap correctly serializes bash expression configs.
func TestBashToolConfigToMap(t *testing.T) {
	t.Run("expression bash serializes to expression string", func(t *testing.T) {
		tools := &ToolsConfig{
			Bash: &BashToolConfig{AllowedCommandsExpr: "${{ inputs.bash-allowlist }}"},
		}
		m := tools.ToMap()
		assert.Equal(t, "${{ inputs.bash-allowlist }}", m["bash"], "bash should serialize as expression string")
	})

	t.Run("literal commands serializes to slice", func(t *testing.T) {
		tools := &ToolsConfig{
			Bash: &BashToolConfig{AllowedCommands: []string{"git", "npm"}},
		}
		m := tools.ToMap()
		assert.Equal(t, []string{"git", "npm"}, m["bash"], "bash should serialize as string slice")
	})
}

// TestHasBashWildcardWithExpression verifies that expressions are not treated as wildcards.
func TestHasBashWildcardWithExpression(t *testing.T) {
	t.Run("expression is not treated as wildcard", func(t *testing.T) {
		tools := map[string]any{
			"bash": "${{ inputs.bash-allowlist }}",
		}
		result := hasBashWildcardInTools(tools)
		assert.False(t, result, "expression should not be treated as wildcard bash")
	})

	t.Run("wildcard array still detected", func(t *testing.T) {
		tools := map[string]any{
			"bash": []any{"*"},
		}
		result := hasBashWildcardInTools(tools)
		assert.True(t, result, "wildcard array should be detected")
	})

	t.Run("bash: true still treated as wildcard", func(t *testing.T) {
		tools := map[string]any{
			"bash": true,
		}
		result := hasBashWildcardInTools(tools)
		assert.True(t, result, "bash: true should be treated as wildcard")
	})
}

// TestBuildCopilotDynamicToolArgsPreamble verifies the preamble generator.
func TestBuildCopilotDynamicToolArgsPreamble(t *testing.T) {
	t.Run("no expressions returns empty preamble", func(t *testing.T) {
		result := buildCopilotDynamicToolArgsPreamble(false, false)
		assert.Empty(t, result, "should return empty string when no expressions")
	})

	t.Run("bash expression includes _bash_tool_args setup", func(t *testing.T) {
		result := buildCopilotDynamicToolArgsPreamble(true, false)
		assert.Contains(t, result, "_bash_tool_args=()", "should initialize _bash_tool_args array")
		assert.Contains(t, result, "GH_AW_BASH_ALLOWLIST", "should reference GH_AW_BASH_ALLOWLIST")
		assert.Contains(t, result, "--allow-tool", "should include --allow-tool in preamble")
		assert.NotContains(t, result, "_edit_tool_args", "should not include edit args when hasEditExpr is false")
	})

	t.Run("edit expression includes _edit_tool_args setup", func(t *testing.T) {
		result := buildCopilotDynamicToolArgsPreamble(false, true)
		assert.Contains(t, result, "_edit_tool_args=()", "should initialize _edit_tool_args array")
		assert.Contains(t, result, "GH_AW_EDIT_ENABLED", "should reference GH_AW_EDIT_ENABLED")
		assert.NotContains(t, result, "_bash_tool_args", "should not include bash args when hasBashExpr is false")
	})

	t.Run("both expressions includes both setups", func(t *testing.T) {
		result := buildCopilotDynamicToolArgsPreamble(true, true)
		assert.Contains(t, result, "_bash_tool_args=()", "should initialize _bash_tool_args")
		assert.Contains(t, result, "_edit_tool_args=()", "should initialize _edit_tool_args")
	})
}

// TestCopilotEngineWithBashExpressionStep verifies compiled output when tools.bash is an expression.
func TestCopilotEngineWithBashExpressionStep(t *testing.T) {
	tools := map[string]any{
		"bash": "${{ inputs.bash-allowlist }}",
	}
	wd := &WorkflowData{
		Name:        "test-bash-expr",
		WorkflowID:  "test-bash-expr",
		On:          "issues",
		Tools:       tools,
		ParsedTools: NewTools(tools),
	}

	engine := NewCopilotEngine()
	steps := engine.GetExecutionSteps(wd, "/tmp/test.log")

	require.NotEmpty(t, steps, "should have execution steps")
	stepContent := strings.Join([]string(steps[0]), "\n")

	// Verify dynamic preamble is included
	assert.Contains(t, stepContent, "_bash_tool_args=()", "should include dynamic bash args preamble")
	assert.Contains(t, stepContent, "GH_AW_BASH_ALLOWLIST", "preamble should reference GH_AW_BASH_ALLOWLIST env var")
	assert.Contains(t, stepContent, `"${_bash_tool_args[@]}"`, "copilot command should include dynamic bash args")

	// Verify env var is set with the expression
	assert.Contains(t, stepContent, "GH_AW_BASH_ALLOWLIST: ${{ inputs.bash-allowlist }}", "env var should contain the expression")

	// Verify no static --allow-tool shell(...) is in the command
	assert.NotContains(t, stepContent, "--allow-tool shell(", "should not contain static bash tool args")
}

// TestCopilotEngineWithEditExpressionStep verifies compiled output when tools.edit is an expression.
func TestCopilotEngineWithEditExpressionStep(t *testing.T) {
	tools := map[string]any{
		"edit": "${{ inputs.enable-edit }}",
	}
	wd := &WorkflowData{
		Name:        "test-edit-expr",
		WorkflowID:  "test-edit-expr",
		On:          "issues",
		Tools:       tools,
		ParsedTools: NewTools(tools),
	}

	engine := NewCopilotEngine()
	steps := engine.GetExecutionSteps(wd, "/tmp/test.log")

	require.NotEmpty(t, steps, "should have execution steps")
	stepContent := strings.Join([]string(steps[0]), "\n")

	// Verify dynamic preamble is included
	assert.Contains(t, stepContent, "_edit_tool_args=()", "should include dynamic edit args preamble")
	assert.Contains(t, stepContent, "GH_AW_EDIT_ENABLED", "preamble should reference GH_AW_EDIT_ENABLED")
	assert.Contains(t, stepContent, `"${_edit_tool_args[@]}"`, "copilot command should include dynamic edit args")

	// Verify env var is set with the expression
	assert.Contains(t, stepContent, "GH_AW_EDIT_ENABLED: ${{ inputs.enable-edit }}", "env var should contain the expression")

	// --allow-all-paths should still be included (path access is less sensitive than write)
	assert.Contains(t, stepContent, "--allow-all-paths", "should include --allow-all-paths for edit")
}

// TestCopilotComputeToolArgsWithBashExpression verifies no static bash args when expression is used.
func TestCopilotComputeToolArgsWithBashExpression(t *testing.T) {
	engine := NewCopilotEngine()

	tools := map[string]any{
		"bash": "${{ inputs.bash-allowlist }}",
	}

	args := engine.computeCopilotToolArguments(tools, nil, nil, &WorkflowData{
		ParsedTools: NewTools(tools),
	})

	// Should not contain any shell() allow-tool args
	joinedArgs := strings.Join(args, " ")
	assert.NotContains(t, joinedArgs, "shell(", "should not contain static shell args for expression bash")
	// Should also not be --allow-all-tools
	assert.NotContains(t, joinedArgs, "--allow-all-tools", "should not use --allow-all-tools for expression bash")
}

// TestCopilotEngineWithGitHubToolsetsExpressionYAML verifies GITHUB_TOOLSETS value in MCP config
// when toolsets is an expression.
func TestCopilotEngineWithGitHubToolsetsExpressionYAML(t *testing.T) {
	githubTool := map[string]any{
		"toolsets": "${{ inputs.github-toolsets }}",
	}

	toolsets := getGitHubToolsets(githubTool)
	assert.Equal(t, "${{ inputs.github-toolsets }}", toolsets,
		"expression should be passed through as GITHUB_TOOLSETS value for runtime evaluation")
}
