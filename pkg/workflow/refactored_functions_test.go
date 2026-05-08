//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── buildNativeLabelFilterSections ──────────────────────────────────────────

func TestBuildNativeLabelFilterSections(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter map[string]any
		wantKeys    []string
		dontWant    []string
	}{
		{
			name: "sections with native label filter marker are included",
			frontmatter: map[string]any{
				"on": map[string]any{
					"issues":       map[string]any{"__gh_aw_native_label_filter__": true},
					"pull_request": map[string]any{"types": []any{"opened"}},
				},
			},
			wantKeys: []string{"issues"},
			dontWant: []string{"pull_request"},
		},
		{
			name: "sections with false marker are excluded",
			frontmatter: map[string]any{
				"on": map[string]any{
					"issues": map[string]any{"__gh_aw_native_label_filter__": false},
				},
			},
			dontWant: []string{"issues"},
		},
		{
			name: "all four supported sections can be marked as native",
			frontmatter: map[string]any{
				"on": map[string]any{
					"issues":        map[string]any{"__gh_aw_native_label_filter__": true},
					"pull_request":  map[string]any{"__gh_aw_native_label_filter__": true},
					"discussion":    map[string]any{"__gh_aw_native_label_filter__": true},
					"issue_comment": map[string]any{"__gh_aw_native_label_filter__": true},
				},
			},
			wantKeys: []string{"issues", "pull_request", "discussion", "issue_comment"},
		},
		{
			name:        "empty frontmatter returns empty map",
			frontmatter: map[string]any{},
			dontWant:    []string{"issues", "pull_request"},
		},
		{
			name:        "nil on value is handled gracefully",
			frontmatter: map[string]any{"on": nil},
			dontWant:    []string{"issues"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNativeLabelFilterSections(tt.frontmatter)
			require.NotNil(t, result, "result should never be nil")
			for _, key := range tt.wantKeys {
				assert.True(t, result[key], "expected section %q to be in native filter set", key)
			}
			for _, key := range tt.dontWant {
				assert.False(t, result[key], "section %q should NOT be in native filter set", key)
			}
		})
	}
}

// ─── onSectionParseState.updateSectionEntry ──────────────────────────────────

func TestOnSectionParseState_UpdateSectionEntry(t *testing.T) {
	tests := []struct {
		name               string
		line               string
		wantHandled        bool
		wantSection        string
		wantInPullRequest  bool
		wantInIssues       bool
		wantInDiscussion   bool
		wantInIssueComment bool
	}{
		{
			name:              "pull_request: line is handled",
			line:              "  pull_request:",
			wantHandled:       true,
			wantSection:       "pull_request",
			wantInPullRequest: true,
		},
		{
			name:         "issues: line is handled",
			line:         "  issues:",
			wantHandled:  true,
			wantSection:  "issues",
			wantInIssues: true,
		},
		{
			name:             "discussion: line is handled",
			line:             "  discussion:",
			wantHandled:      true,
			wantSection:      "discussion",
			wantInDiscussion: true,
		},
		{
			name:               "issue_comment: line is handled",
			line:               "  issue_comment:",
			wantHandled:        true,
			wantSection:        "issue_comment",
			wantInIssueComment: true,
		},
		{
			name:        "regular YAML line is not handled",
			line:        "    types: [opened]",
			wantHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &onSectionParseState{}
			handled := state.updateSectionEntry(tt.line)

			assert.Equal(t, tt.wantHandled, handled, "handled mismatch")
			if tt.wantHandled {
				assert.Equal(t, tt.wantSection, state.currentSection, "currentSection mismatch")
				assert.Equal(t, tt.wantInPullRequest, state.inPullRequest, "inPullRequest mismatch")
				assert.Equal(t, tt.wantInIssues, state.inIssues, "inIssues mismatch")
				assert.Equal(t, tt.wantInDiscussion, state.inDiscussion, "inDiscussion mismatch")
				assert.Equal(t, tt.wantInIssueComment, state.inIssueComment, "inIssueComment mismatch")
			}
		})
	}
}

func TestOnSectionParseState_UpdateSectionEntry_SkipsWhenInPermissionsOrSteps(t *testing.T) {
	t.Run("issues: inside on.permissions is not treated as event trigger", func(t *testing.T) {
		state := &onSectionParseState{inOnPermissions: true}
		handled := state.updateSectionEntry("    issues: read")
		assert.False(t, handled, "should not treat 'issues: read' inside on.permissions as event trigger")
		assert.False(t, state.inIssues, "inIssues should remain false")
	})

	t.Run("pull_request: inside on.steps is not treated as event trigger", func(t *testing.T) {
		state := &onSectionParseState{inOnSteps: true}
		handled := state.updateSectionEntry("  pull_request: something")
		assert.False(t, handled, "should not treat line inside on.steps as event trigger")
		assert.False(t, state.inPullRequest, "inPullRequest should remain false")
	})
}

// ─── onSectionParseState.classifyForCommentOut ───────────────────────────────

func TestOnSectionParseState_ClassifyForCommentOut_TopLevelFields(t *testing.T) {
	tests := []struct {
		name           string
		trimmedLine    string
		state          onSectionParseState
		wantComment    bool
		wantReasonPart string
	}{
		{
			name:           "manual-approval is commented with reason",
			trimmedLine:    "manual-approval: true",
			wantComment:    true,
			wantReasonPart: "activation job",
		},
		{
			name:           "stop-after is commented with reason",
			trimmedLine:    "stop-after: 2024-01-01",
			wantComment:    true,
			wantReasonPart: "pre-activation job",
		},
		{
			name:           "skip-if-match is commented with reason",
			trimmedLine:    "skip-if-match:",
			wantComment:    true,
			wantReasonPart: "pre-activation job",
		},
		{
			name:           "reaction is commented with reason",
			trimmedLine:    "reaction: eyes",
			wantComment:    true,
			wantReasonPart: "activation",
		},
		{
			name:           "stale-check is commented with reason",
			trimmedLine:    "stale-check: true",
			wantComment:    true,
			wantReasonPart: "activation job",
		},
		{
			name:           "permissions is commented when not in event sections",
			trimmedLine:    "permissions:",
			wantComment:    true,
			wantReasonPart: "pre-activation job",
		},
		{
			name:        "types: field is NOT commented (allowed field)",
			trimmedLine: "types: [opened]",
			wantComment: false,
		},
		{
			name:        "names: NOT commented when inside pull_request section with native filter (passes via nativeLabelFilterSections arg)",
			trimmedLine: "names:",
			state: onSectionParseState{
				inPullRequest:  true,
				currentSection: "pull_request",
			},
			// Without passing nativeLabelFilterSections, it WILL comment names: since currentSection isn't in the map.
			// The actual native filter exclusion is tested in TestBuildConclusionJobCondition tests below.
			wantComment: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.state
			shouldComment, reason := state.classifyForCommentOut(
				tt.trimmedLine, tt.trimmedLine,
				map[string]bool{},
				nil,
			)
			assert.Equal(t, tt.wantComment, shouldComment, "shouldComment mismatch")
			if tt.wantComment && tt.wantReasonPart != "" {
				assert.True(t, strings.Contains(reason, tt.wantReasonPart),
					"reason %q should contain %q", reason, tt.wantReasonPart)
			}
		})
	}
}

func TestOnSectionParseState_ClassifyForCommentOut_NamesNotCommentedWithNativeFilter(t *testing.T) {
	state := onSectionParseState{
		inIssues:       true,
		currentSection: "issues",
	}
	// With native label filter enabled for "issues", "names:" should NOT be commented
	shouldComment, _ := state.classifyForCommentOut(
		"names:", "names:",
		map[string]bool{"issues": true},
		nil,
	)
	assert.False(t, shouldComment, "names: should NOT be commented when native label filter is active")
}

func TestOnSectionParseState_ClassifyForCommentOut_NamesCommentedWithoutNativeFilter(t *testing.T) {
	state := onSectionParseState{
		inIssues:       true,
		currentSection: "issues",
	}
	// Without native label filter, "names:" SHOULD be commented
	shouldComment, reason := state.classifyForCommentOut(
		"names:", "names:",
		map[string]bool{},
		nil,
	)
	assert.True(t, shouldComment, "names: should be commented when no native label filter")
	assert.Contains(t, reason, "Label filtering", "reason should mention label filtering")
}

// ─── onSectionParseState.updateObjectExit ────────────────────────────────────

func TestOnSectionParseState_UpdateObjectExit(t *testing.T) {
	t.Run("exits skip-if-match when sibling field appears at indent 2", func(t *testing.T) {
		state := &onSectionParseState{inSkipIfMatch: true}
		state.updateObjectExit("  reaction:", "reaction:")
		assert.False(t, state.inSkipIfMatch, "should exit skip-if-match")
	})

	t.Run("does not exit skip-if-match for skip-if-match own line", func(t *testing.T) {
		state := &onSectionParseState{inSkipIfMatch: true}
		state.updateObjectExit("  skip-if-match:", "skip-if-match:")
		assert.True(t, state.inSkipIfMatch, "should remain in skip-if-match")
	})

	t.Run("exits skip-roles when sibling appears at indent 2", func(t *testing.T) {
		state := &onSectionParseState{inSkipRolesArray: true}
		state.updateObjectExit("  names:", "names:")
		assert.False(t, state.inSkipRolesArray, "should exit skip-roles array")
	})

	t.Run("does not exit forks when inside pull_request array items", func(t *testing.T) {
		state := &onSectionParseState{inForksArray: true, inPullRequest: true}
		state.updateObjectExit("    - public", "- public") // indent 4, starts with -
		assert.True(t, state.inForksArray, "should stay in forks array for list items")
	})

	t.Run("exits on.permissions at indent 2 when non-comment sibling found", func(t *testing.T) {
		state := &onSectionParseState{inOnPermissions: true}
		state.updateObjectExit("  reaction:", "reaction:")
		assert.False(t, state.inOnPermissions, "should exit on.permissions")
	})
}

// ─── buildConclusionJobCondition ─────────────────────────────────────────────

func TestBuildConclusionJobCondition(t *testing.T) {
	t.Run("condition includes always() and agent-not-skipped checks", func(t *testing.T) {
		condition := buildConclusionJobCondition("agent", []string{"safe_outputs"})
		rendered := RenderCondition(condition)
		assert.Contains(t, rendered, "always()", "condition should include always()")
		assert.Contains(t, rendered, "agent", "condition should reference the agent job")
	})

	t.Run("condition includes add_comment check when add_comment is in safe output jobs", func(t *testing.T) {
		condition := buildConclusionJobCondition("agent", []string{"add_comment", "safe_outputs"})
		rendered := RenderCondition(condition)
		assert.Contains(t, rendered, "add_comment", "condition should include add_comment output check")
	})

	t.Run("condition does NOT include add_comment check when not in safe output jobs", func(t *testing.T) {
		condition := buildConclusionJobCondition("agent", []string{"safe_outputs"})
		rendered := RenderCondition(condition)
		assert.NotContains(t, rendered, "add_comment", "condition should not reference add_comment when not in jobs")
	})

	t.Run("condition includes lockdown_check_failed and stale_lock_file_failed", func(t *testing.T) {
		condition := buildConclusionJobCondition("agent", nil)
		rendered := RenderCondition(condition)
		assert.Contains(t, rendered, "lockdown_check_failed", "condition should check lockdown failure")
		assert.Contains(t, rendered, "stale_lock_file_failed", "condition should check stale lock file failure")
	})
}

// ─── buildMaintenanceWorkflowOnSection ───────────────────────────────────────

func TestBuildMaintenanceWorkflowOnSection(t *testing.T) {
	t.Run("contains cron schedule and description", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:   "37 0 * * *",
			scheduleDesc:   "Daily",
			minExpiresDays: 7,
		}
		result := buildMaintenanceWorkflowOnSection(p)
		assert.Contains(t, result, `"37 0 * * *"`, "should contain the cron schedule")
		assert.Contains(t, result, "Daily", "should contain the schedule description")
		assert.Contains(t, result, "7 days", "should mention minExpiresDays")
	})

	t.Run("includes push trigger and compile-workflows in dev mode", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:   "37 0 * * *",
			scheduleDesc:   "Daily",
			minExpiresDays: 7,
			actionMode:     ActionModeDev,
			defaultBranch:  "main",
		}
		result := buildMaintenanceWorkflowOnSection(p)
		assert.Contains(t, result, "push:", "dev mode should add push trigger")
		assert.Contains(t, result, "main", "push trigger should reference the default branch")
	})

	t.Run("excludes push trigger in release mode", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:   "37 0 * * *",
			scheduleDesc:   "Daily",
			minExpiresDays: 7,
			actionMode:     ActionModeRelease,
		}
		result := buildMaintenanceWorkflowOnSection(p)
		assert.NotContains(t, result, "push:", "release mode should not add push trigger")
	})

	t.Run("includes label trigger when enabled", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:        "37 0 * * *",
			scheduleDesc:        "Daily",
			minExpiresDays:      7,
			disableLabelTrigger: false,
		}
		result := buildMaintenanceWorkflowOnSection(p)
		assert.Contains(t, result, "issues:", "should include issues label trigger")
		assert.Contains(t, result, "[labeled]", "should include labeled event type")
	})

	t.Run("excludes label trigger when disabled", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:        "37 0 * * *",
			scheduleDesc:        "Daily",
			minExpiresDays:      7,
			disableLabelTrigger: true,
		}
		result := buildMaintenanceWorkflowOnSection(p)
		// Should contain workflow_dispatch and workflow_call but not the "issues: labeled" trigger
		assert.Contains(t, result, "workflow_dispatch:", "should still include workflow_dispatch")
		assert.NotContains(t, result, "[labeled]", "label trigger should be absent")
	})

	t.Run("always contains permissions and jobs header", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			cronSchedule:   "37 0 * * *",
			scheduleDesc:   "Daily",
			minExpiresDays: 1,
		}
		result := buildMaintenanceWorkflowOnSection(p)
		assert.Contains(t, result, "permissions: {}", "should include permissions: {}")
		assert.Contains(t, result, "\njobs:\n", "should end with jobs header")
	})
}

// ─── buildMaintenanceCloseExpiredEntitiesJob ─────────────────────────────────

func TestBuildMaintenanceCloseExpiredEntitiesJob(t *testing.T) {
	t.Run("job has correct name and permissions", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			runsOnValue: "ubuntu-latest",
		}
		result := buildMaintenanceCloseExpiredEntitiesJob(p)
		assert.Contains(t, result, "close-expired-entities:", "should contain job name")
		assert.Contains(t, result, "discussions: write", "should have discussions write permission")
		assert.Contains(t, result, "issues: write", "should have issues write permission")
		assert.Contains(t, result, "pull-requests: write", "should have pull-requests write permission")
	})

	t.Run("includes checkout step in dev mode", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			runsOnValue:    "ubuntu-latest",
			actionMode:     ActionModeDev,
			setupActionRef: "./actions/setup",
		}
		result := buildMaintenanceCloseExpiredEntitiesJob(p)
		assert.Contains(t, result, "Checkout actions folder", "dev mode should include checkout step")
	})

	t.Run("excludes checkout step in release mode", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			runsOnValue:    "ubuntu-latest",
			actionMode:     ActionModeRelease,
			setupActionRef: "github/gh-aw/actions/setup@v1",
		}
		result := buildMaintenanceCloseExpiredEntitiesJob(p)
		assert.NotContains(t, result, "Checkout actions folder", "release mode should not checkout")
	})

	t.Run("includes all three close scripts", func(t *testing.T) {
		p := maintenanceWorkflowParams{
			runsOnValue: "ubuntu-latest",
		}
		result := buildMaintenanceCloseExpiredEntitiesJob(p)
		assert.Contains(t, result, "close_expired_discussions.cjs", "should include discussions script")
		assert.Contains(t, result, "close_expired_issues.cjs", "should include issues script")
		assert.Contains(t, result, "close_expired_pull_requests.cjs", "should include PRs script")
	})

	t.Run("job condition excludes push and dispatch events (schedule-only)", func(t *testing.T) {
		p := maintenanceWorkflowParams{runsOnValue: "ubuntu-latest"}
		result := buildMaintenanceCloseExpiredEntitiesJob(p)
		// The condition uses buildNotForkAndScheduleOnly() which renders as an if expression
		// excluding push, workflow_dispatch, and workflow_call events.
		assert.Contains(t, result, "github.event_name != 'push'", "job condition should exclude push events")
		assert.Contains(t, result, "workflow_dispatch", "job condition should reference workflow_dispatch exclusion")
	})
}

// ─── extractSafeOutputsDefaultHandlers ───────────────────────────────────────

func TestExtractSafeOutputsDefaultHandlers_DefaultsEnabled(t *testing.T) {
	t.Run("missing-tool enabled by default when key is absent", func(t *testing.T) {
		c := &Compiler{}
		config := &SafeOutputsConfig{}
		c.extractSafeOutputsDefaultHandlers(map[string]any{}, config)
		require.NotNil(t, config.MissingTool, "missing-tool should be set by default")
		assert.Equal(t, "true", *config.MissingTool.CreateIssue, "create-issue should be true by default")
	})

	t.Run("noop enabled by default when key is absent", func(t *testing.T) {
		c := &Compiler{}
		config := &SafeOutputsConfig{}
		c.extractSafeOutputsDefaultHandlers(map[string]any{}, config)
		require.NotNil(t, config.NoOp, "noop should be set by default")
		require.NotNil(t, config.NoOp.ReportAsIssue, "noop.report-as-issue should be set")
		assert.Equal(t, "true", *config.NoOp.ReportAsIssue, "noop.report-as-issue should be true by default")
	})

	t.Run("missing-data enabled by default when key is absent", func(t *testing.T) {
		c := &Compiler{}
		config := &SafeOutputsConfig{}
		c.extractSafeOutputsDefaultHandlers(map[string]any{}, config)
		require.NotNil(t, config.MissingData, "missing-data should be set by default")
	})

	t.Run("report-incomplete enabled by default when key is absent", func(t *testing.T) {
		c := &Compiler{}
		config := &SafeOutputsConfig{}
		c.extractSafeOutputsDefaultHandlers(map[string]any{}, config)
		require.NotNil(t, config.ReportIncomplete, "report-incomplete should be set by default")
	})

	t.Run("missing-tool not set when key is present but disabled", func(t *testing.T) {
		c := &Compiler{}
		config := &SafeOutputsConfig{}
		// Passing "missing-tool: false" means the key exists in outputMap but parseMissingToolConfig returns nil
		// Here we test the branch where the key exists → no default is applied
		outputMap := map[string]any{"missing-tool": false}
		c.extractSafeOutputsDefaultHandlers(outputMap, config)
		assert.Nil(t, config.MissingTool, "missing-tool should remain nil when explicitly disabled")
	})
}

// ─── buildConclusionJobNoOpSteps / buildConclusionJobDetectionRunsSteps ───────

func TestBuildConclusionJobNoOpSteps_NilWhenNoOpNotConfigured(t *testing.T) {
	c := newTestCompiler(t)
	data := &WorkflowData{SafeOutputs: &SafeOutputsConfig{}}
	steps := c.buildConclusionJobNoOpSteps(data, "agent")
	assert.Nil(t, steps, "should return nil when NoOp is not configured")
}

func TestBuildConclusionJobNoOpSteps_ReturnsStepsWhenNoOpConfigured(t *testing.T) {
	c := newTestCompiler(t)
	trueVal := "true"
	data := &WorkflowData{
		SafeOutputs: &SafeOutputsConfig{
			NoOp: &NoOpConfig{ReportAsIssue: &trueVal},
		},
	}
	steps := c.buildConclusionJobNoOpSteps(data, "agent")
	require.NotNil(t, steps, "should return steps when NoOp is configured")
	combined := strings.Join(steps, "\n")
	assert.Contains(t, combined, "handle_noop_message.cjs", "steps should include the noop script")
}

func TestBuildConclusionJobDetectionRunsSteps_NilWhenNoDetection(t *testing.T) {
	c := newTestCompiler(t)
	data := &WorkflowData{SafeOutputs: &SafeOutputsConfig{}}
	steps := c.buildConclusionJobDetectionRunsSteps(data, "agent")
	assert.Nil(t, steps, "should return nil when threat detection is not enabled")
}

func TestBuildConclusionJobMissingToolSteps_NilWhenNotConfigured(t *testing.T) {
	c := newTestCompiler(t)
	data := &WorkflowData{SafeOutputs: &SafeOutputsConfig{}}
	steps := c.buildConclusionJobMissingToolSteps(data, "agent")
	assert.Nil(t, steps, "should return nil when missing-tool is not configured")
}

func TestBuildConclusionJobMissingToolSteps_ContainsScript(t *testing.T) {
	c := newTestCompiler(t)
	trueVal := "true"
	data := &WorkflowData{
		SafeOutputs: &SafeOutputsConfig{
			MissingTool: &MissingToolConfig{CreateIssue: &trueVal},
		},
	}
	steps := c.buildConclusionJobMissingToolSteps(data, "agent")
	require.NotNil(t, steps, "should return steps when missing-tool is configured")
	combined := strings.Join(steps, "\n")
	assert.Contains(t, combined, "missing_tool.cjs", "steps should include the missing_tool script")
}

func TestBuildConclusionJobReportIncompleteSteps_NilWhenNotConfigured(t *testing.T) {
	c := newTestCompiler(t)
	data := &WorkflowData{SafeOutputs: &SafeOutputsConfig{}}
	steps := c.buildConclusionJobReportIncompleteSteps(data, "agent")
	assert.Nil(t, steps, "should return nil when report-incomplete is not configured")
}
