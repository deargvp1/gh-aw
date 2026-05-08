package workflow

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/github/gh-aw/pkg/constants"
)

// buildTitlePrefixEnvVar returns the env-var declaration line for a title-prefix field
// if the prefix is non-empty, otherwise returns nil.  Used by both missing-tool and
// report-incomplete step builders to avoid duplicated logic.
func buildTitlePrefixEnvVar(envVarName, prefix string) []string {
	if prefix == "" {
		return nil
	}
	return []string{fmt.Sprintf("          %s: %q\n", envVarName, prefix)}
}

// buildConclusionJobNoOpSteps builds the noop processing step for the conclusion job.
// Returns nil if noop is not configured in safe-outputs.
func (c *Compiler) buildConclusionJobNoOpSteps(data *WorkflowData, mainJobName string) []string {
	if data.SafeOutputs.NoOp == nil {
		return nil
	}

	var noopEnvVars []string
	noopEnvVars = append(noopEnvVars, buildTemplatableIntEnvVar("GH_AW_NOOP_MAX", data.SafeOutputs.NoOp.Max)...)
	noopEnvVars = append(noopEnvVars, buildWorkflowMetadataEnvVarsWithTrackerID(data.Name, data.Source, data.TrackerID)...)
	// Agent conclusion and run URL are used to decide whether to post to the runs issue
	noopEnvVars = append(noopEnvVars, "          GH_AW_RUN_URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}\n")
	noopEnvVars = append(noopEnvVars, fmt.Sprintf("          GH_AW_AGENT_CONCLUSION: ${{ needs.%s.result }}\n", mainJobName))
	noopEnvVars = append(noopEnvVars, buildTemplatableBoolEnvVar("GH_AW_NOOP_REPORT_AS_ISSUE", data.SafeOutputs.NoOp.ReportAsIssue)...)
	if data.SafeOutputs.NoOp.ReportAsIssue == nil {
		noopEnvVars = append(noopEnvVars, "          GH_AW_NOOP_REPORT_AS_ISSUE: \"true\"\n")
	}

	return c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
		StepName:      "Process no-op messages",
		StepID:        "noop",
		MainJobName:   mainJobName,
		CustomEnvVars: noopEnvVars,
		ScriptFile:    "handle_noop_message.cjs",
		CustomToken:   data.SafeOutputs.NoOp.GitHubToken,
	})
}

// buildConclusionJobDetectionRunsSteps builds the detection runs logging step for the conclusion job.
// This step posts a comment to the "[aw] Detection Runs" tracking issue whenever
// the detection job produces a warning or failure conclusion.
// Returns nil if threat detection is not enabled.
func (c *Compiler) buildConclusionJobDetectionRunsSteps(data *WorkflowData, mainJobName string) []string {
	if !IsDetectionJobEnabled(data.SafeOutputs) {
		return nil
	}

	var detectionRunsEnvVars []string
	detectionRunsEnvVars = append(detectionRunsEnvVars, buildWorkflowMetadataEnvVarsWithTrackerID(data.Name, data.Source, data.TrackerID)...)
	detectionRunsEnvVars = append(detectionRunsEnvVars, "          GH_AW_RUN_URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}\n")
	detectionRunsEnvVars = append(detectionRunsEnvVars, fmt.Sprintf("          GH_AW_DETECTION_CONCLUSION: ${{ needs.%s.outputs.detection_conclusion }}\n", constants.DetectionJobName))
	detectionRunsEnvVars = append(detectionRunsEnvVars, fmt.Sprintf("          GH_AW_DETECTION_REASON: ${{ needs.%s.outputs.detection_reason }}\n", constants.DetectionJobName))

	steps := c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
		StepName:      "Log detection run",
		StepID:        "detection_runs",
		MainJobName:   mainJobName,
		CustomEnvVars: detectionRunsEnvVars,
		ScriptFile:    "handle_detection_runs.cjs",
	})
	notifyCommentLog.Print("Added detection runs logging step to conclusion job")
	return steps
}

// buildConclusionJobMissingToolSteps builds the missing-tool processing step for the conclusion job.
// Returns nil if missing-tool is not configured in safe-outputs.
func (c *Compiler) buildConclusionJobMissingToolSteps(data *WorkflowData, mainJobName string) []string {
	if data.SafeOutputs.MissingTool == nil {
		return nil
	}

	var missingToolEnvVars []string
	missingToolEnvVars = append(missingToolEnvVars, buildTemplatableIntEnvVar("GH_AW_MISSING_TOOL_MAX", data.SafeOutputs.MissingTool.Max)...)
	missingToolEnvVars = append(missingToolEnvVars, buildTemplatableBoolEnvVar("GH_AW_MISSING_TOOL_CREATE_ISSUE", data.SafeOutputs.MissingTool.CreateIssue)...)
	missingToolEnvVars = append(missingToolEnvVars, buildTitlePrefixEnvVar("GH_AW_MISSING_TOOL_TITLE_PREFIX", data.SafeOutputs.MissingTool.TitlePrefix)...)
	if len(data.SafeOutputs.MissingTool.Labels) > 0 {
		labelsJSON, err := json.Marshal(data.SafeOutputs.MissingTool.Labels)
		if err == nil {
			missingToolEnvVars = append(missingToolEnvVars, fmt.Sprintf("          GH_AW_MISSING_TOOL_LABELS: %q\n", string(labelsJSON)))
		}
	}
	missingToolEnvVars = append(missingToolEnvVars, buildWorkflowMetadataEnvVarsWithTrackerID(data.Name, data.Source, data.TrackerID)...)

	return c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
		StepName:      "Record missing tool",
		StepID:        "missing_tool",
		MainJobName:   mainJobName,
		CustomEnvVars: missingToolEnvVars,
		Script:        "const { main } = require('${{ runner.temp }}/gh-aw/actions/missing_tool.cjs'); await main();",
		ScriptFile:    "missing_tool.cjs",
		CustomToken:   data.SafeOutputs.MissingTool.GitHubToken,
	})
}

// buildConclusionJobReportIncompleteSteps builds the report-incomplete processing step for the conclusion job.
// Returns nil if report-incomplete is not configured in safe-outputs.
func (c *Compiler) buildConclusionJobReportIncompleteSteps(data *WorkflowData, mainJobName string) []string {
	if data.SafeOutputs.ReportIncomplete == nil {
		return nil
	}

	var reportIncompleteEnvVars []string
	reportIncompleteEnvVars = append(reportIncompleteEnvVars, buildTemplatableIntEnvVar("GH_AW_REPORT_INCOMPLETE_MAX", data.SafeOutputs.ReportIncomplete.Max)...)
	reportIncompleteEnvVars = append(reportIncompleteEnvVars, buildTemplatableBoolEnvVar("GH_AW_REPORT_INCOMPLETE_CREATE_ISSUE", data.SafeOutputs.ReportIncomplete.CreateIssue)...)
	reportIncompleteEnvVars = append(reportIncompleteEnvVars, buildTitlePrefixEnvVar("GH_AW_REPORT_INCOMPLETE_TITLE_PREFIX", data.SafeOutputs.ReportIncomplete.TitlePrefix)...)
	if len(data.SafeOutputs.ReportIncomplete.Labels) > 0 {
		labelsJSON, err := json.Marshal(data.SafeOutputs.ReportIncomplete.Labels)
		if err == nil {
			reportIncompleteEnvVars = append(reportIncompleteEnvVars, fmt.Sprintf("          GH_AW_REPORT_INCOMPLETE_LABELS: %q\n", string(labelsJSON)))
		}
	}
	reportIncompleteEnvVars = append(reportIncompleteEnvVars, buildWorkflowMetadataEnvVarsWithTrackerID(data.Name, data.Source, data.TrackerID)...)

	return c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
		StepName:      "Record incomplete",
		StepID:        "report_incomplete",
		MainJobName:   mainJobName,
		CustomEnvVars: reportIncompleteEnvVars,
		Script:        "const { main } = require('${{ runner.temp }}/gh-aw/actions/report_incomplete_handler.cjs'); await main();",
		ScriptFile:    "report_incomplete_handler.cjs",
		CustomToken:   data.SafeOutputs.ReportIncomplete.GitHubToken,
	})
}

// buildConclusionJobAgentFailureEnvVars assembles all environment variables for the
// "Handle agent failure" step in the conclusion job.
func (c *Compiler) buildConclusionJobAgentFailureEnvVars(
	data *WorkflowData,
	mainJobName string,
	safeOutputJobNames []string,
	messagesJSON string,
) ([]string, error) {
	var agentFailureEnvVars []string
	agentFailureEnvVars = append(agentFailureEnvVars, buildWorkflowMetadataEnvVarsWithTrackerID(data.Name, data.Source, data.TrackerID)...)
	agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_RUN_URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}\n")
	agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_AGENT_CONCLUSION: ${{ needs.%s.result }}\n", mainJobName))
	agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_WORKFLOW_ID: %q\n", data.WorkflowID))

	actionFailureIssueExpiresHours := DefaultActionFailureIssueExpiresHours
	repoConfig, repoConfigErr := c.loadRepoConfig()
	if repoConfigErr != nil {
		notifyCommentLog.Printf(
			"Warning: failed to load repo config for action failure issue expiration (using default %d hours): %v. Check that %s exists and matches schema requirements",
			DefaultActionFailureIssueExpiresHours,
			repoConfigErr,
			RepoConfigFileName,
		)
	} else {
		actionFailureIssueExpiresHours = repoConfig.ActionFailureIssueExpiresHours()
	}
	agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_ACTION_FAILURE_ISSUE_EXPIRES_HOURS: %q\n", strconv.Itoa(actionFailureIssueExpiresHours)))

	// Pass the engine ID so the failure handler can surface which AI engine terminated
	if data.EngineConfig != nil && data.EngineConfig.ID != "" {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_ENGINE_ID: %q\n", data.EngineConfig.ID))
	}

	// Only add secret_verification_result if the engine provides a validate-secret step.
	engine, err := c.getAgenticEngine(data.AI)
	if err != nil {
		return nil, fmt.Errorf("failed to get agentic engine: %w", err)
	}
	if EngineHasValidateSecretStep(engine, data) {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_SECRET_VERIFICATION_RESULT: ${{ needs.%s.outputs.secret_verification_result }}\n", string(constants.ActivationJobName)))
	}

	// Add checkout_pr_success to detect PR checkout failures
	if ShouldGeneratePRCheckoutStep(data) {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_CHECKOUT_PR_SUCCESS: ${{ needs.%s.outputs.checkout_pr_success }}\n", mainJobName))
	}

	// Pass Copilot-engine-specific error detection outputs
	if _, ok := engine.(*CopilotEngine); ok {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_INFERENCE_ACCESS_ERROR: ${{ needs.%s.outputs.inference_access_error }}\n", mainJobName))
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_MCP_POLICY_ERROR: ${{ needs.%s.outputs.mcp_policy_error }}\n", mainJobName))
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_AGENTIC_ENGINE_TIMEOUT: ${{ needs.%s.outputs.agentic_engine_timeout }}\n", mainJobName))
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_MODEL_NOT_SUPPORTED_ERROR: ${{ needs.%s.outputs.model_not_supported_error }}\n", mainJobName))
	}

	// Pass the engine's primary AI inference API hosts
	if apiHosts := getEngineAPIHosts(data, engine); len(apiHosts) > 0 {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_ENGINE_API_HOSTS: %q\n", strings.Join(apiHosts, ",")))
	}

	// Pass assignment error outputs from safe_outputs job if assign-to-agent is configured
	if data.SafeOutputs != nil && data.SafeOutputs.AssignToAgent != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_ASSIGNMENT_ERRORS: ${{ needs.safe_outputs.outputs.assign_to_agent_assignment_errors }}\n")
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_ASSIGNMENT_ERROR_COUNT: ${{ needs.safe_outputs.outputs.assign_to_agent_assignment_error_count }}\n")
	}

	// Pass copilot assignment failure outputs if create-issue with copilot assignee is configured
	if data.SafeOutputs != nil && data.SafeOutputs.CreateIssues != nil && hasCopilotAssignee(data.SafeOutputs.CreateIssues.Assignees) {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_ASSIGN_COPILOT_FAILURE_COUNT: ${{ needs.safe_outputs.outputs.assign_copilot_failure_count }}\n")
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_ASSIGN_COPILOT_ERRORS: ${{ needs.safe_outputs.outputs.assign_copilot_errors }}\n")
	}

	// Pass create_discussion error outputs if create-discussions is configured
	if data.SafeOutputs != nil && data.SafeOutputs.CreateDiscussions != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CREATE_DISCUSSION_ERRORS: ${{ needs.safe_outputs.outputs.create_discussion_errors }}\n")
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CREATE_DISCUSSION_ERROR_COUNT: ${{ needs.safe_outputs.outputs.create_discussion_error_count }}\n")
	}

	// Pass code-push failure outputs if push-to-pull-request-branch or create-pull-request is configured
	if data.SafeOutputs != nil && (data.SafeOutputs.PushToPullRequestBranch != nil || data.SafeOutputs.CreatePullRequests != nil) {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CODE_PUSH_FAILURE_ERRORS: ${{ needs.safe_outputs.outputs.code_push_failure_errors }}\n")
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CODE_PUSH_FAILURE_COUNT: ${{ needs.safe_outputs.outputs.code_push_failure_count }}\n")
	}

	// Pass GitHub App token minting failure status
	if data.SafeOutputs != nil && data.SafeOutputs.GitHubApp != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_SAFE_OUTPUTS_APP_TOKEN_MINTING_FAILED: ${{ needs.safe_outputs.outputs.app_token_minting_failed }}\n")
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CONCLUSION_APP_TOKEN_MINTING_FAILED: ${{ steps.safe-outputs-app-token.outcome == 'failure' }}\n")
	}

	// Pass activation job GitHub App token minting failure status if configured
	if data.ActivationGitHubApp != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_ACTIVATION_APP_TOKEN_MINTING_FAILED: ${{ needs.%s.outputs.activation_app_token_minting_failed }}\n", string(constants.ActivationJobName)))
	}

	// Always pass lockdown check failure status
	agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_LOCKDOWN_CHECK_FAILED: ${{ needs.%s.outputs.lockdown_check_failed }}\n", string(constants.ActivationJobName)))

	// Pass stale lock file check failure status
	agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_STALE_LOCK_FILE_FAILED: ${{ needs.%s.outputs.stale_lock_file_failed }}\n", string(constants.ActivationJobName)))

	// Pass custom messages config if present
	if messagesJSON != "" {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_SAFE_OUTPUT_MESSAGES: %q\n", messagesJSON))
	}

	// Pass repo-memory failure outputs if repo-memory is configured
	if data.RepoMemoryConfig != nil && len(data.RepoMemoryConfig.Memories) > 0 {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_PUSH_REPO_MEMORY_RESULT: ${{ needs.push_repo_memory.result }}\n")
		for _, memory := range data.RepoMemoryConfig.Memories {
			agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_REPO_MEMORY_VALIDATION_FAILED_%s: ${{ needs.push_repo_memory.outputs.validation_failed_%s }}\n", memory.ID, memory.ID))
			agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_REPO_MEMORY_VALIDATION_ERROR_%s: ${{ needs.push_repo_memory.outputs.validation_error_%s }}\n", memory.ID, memory.ID))
			agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_REPO_MEMORY_PATCH_SIZE_EXCEEDED_%s: ${{ needs.push_repo_memory.outputs.patch_size_exceeded_%s }}\n", memory.ID, memory.ID))
		}
	}

	// Pass group-reports configuration flag
	if data.SafeOutputs != nil && data.SafeOutputs.GroupReports {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_GROUP_REPORTS: \"true\"\n")
	} else {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_GROUP_REPORTS: \"false\"\n")
	}

	// Pass report-failure-as-issue configuration flag (defaults to true)
	if data.SafeOutputs != nil && data.SafeOutputs.ReportFailureAsIssue != nil && !*data.SafeOutputs.ReportFailureAsIssue {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_FAILURE_REPORT_AS_ISSUE: \"false\"\n")
	} else {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_FAILURE_REPORT_AS_ISSUE: \"true\"\n")
	}

	// Pass missing-tool report-as-failure flag (defaults to true)
	if data.SafeOutputs != nil && data.SafeOutputs.MissingTool != nil && data.SafeOutputs.MissingTool.ReportAsFailure != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, buildTemplatableBoolEnvVar("GH_AW_MISSING_TOOL_REPORT_AS_FAILURE", data.SafeOutputs.MissingTool.ReportAsFailure)...)
	} else {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_MISSING_TOOL_REPORT_AS_FAILURE: \"true\"\n")
	}

	// Pass missing-data report-as-failure flag (defaults to true)
	if data.SafeOutputs != nil && data.SafeOutputs.MissingData != nil && data.SafeOutputs.MissingData.ReportAsFailure != nil {
		agentFailureEnvVars = append(agentFailureEnvVars, buildTemplatableBoolEnvVar("GH_AW_MISSING_DATA_REPORT_AS_FAILURE", data.SafeOutputs.MissingData.ReportAsFailure)...)
	} else {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_MISSING_DATA_REPORT_AS_FAILURE: \"true\"\n")
	}

	// Pass failure-issue-repo configuration (optional, defaults to current repo)
	if data.SafeOutputs != nil && data.SafeOutputs.FailureIssueRepo != "" {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_FAILURE_ISSUE_REPO: %q\n", data.SafeOutputs.FailureIssueRepo))
	}

	// Pass timeout minutes value so the failure handler can provide an actionable hint when timed out
	timeoutValue := strings.TrimPrefix(data.TimeoutMinutes, "timeout-minutes: ")
	if timeoutValue != "" {
		agentFailureEnvVars = append(agentFailureEnvVars, fmt.Sprintf("          GH_AW_TIMEOUT_MINUTES: %q\n", timeoutValue))
	}

	// Pass cache-memory availability flag
	if data.CacheMemoryConfig != nil && len(data.CacheMemoryConfig.Caches) > 0 {
		agentFailureEnvVars = append(agentFailureEnvVars, "          GH_AW_CACHE_MEMORY_ENABLED: \"true\"\n")
	}

	return agentFailureEnvVars, nil
}

// buildConclusionJobCondition builds the if-condition expression for the conclusion job.
// The condition ensures:
//  1. always() - runs even if agent fails
//  2. Agent was activated (not skipped) OR lockdown/stale-lock-file check failed in activation job
//  3. If add_comment job exists: it hasn't already created a comment (avoids duplicate updates)
func buildConclusionJobCondition(mainJobName string, safeOutputJobNames []string) ConditionNode {
	alwaysFunc := BuildFunctionCall("always")

	// Check that agent job was activated (not skipped)
	agentNotSkipped := BuildNotEquals(
		BuildPropertyAccess(fmt.Sprintf("needs.%s.result", constants.AgentJobName)),
		BuildStringLiteral("skipped"),
	)

	// Check if the lockdown check failed in the activation job
	lockdownCheckFailed := BuildEquals(
		BuildPropertyAccess(fmt.Sprintf("needs.%s.outputs.lockdown_check_failed", string(constants.ActivationJobName))),
		BuildStringLiteral("true"),
	)

	// Check if the frontmatter hash (stale lock file) check failed in the activation job
	staleLockFileFailed := BuildEquals(
		BuildPropertyAccess(fmt.Sprintf("needs.%s.outputs.stale_lock_file_failed", string(constants.ActivationJobName))),
		BuildStringLiteral("true"),
	)

	// Agent not skipped OR lockdown check failed OR stale lock file check failed
	agentNotSkippedOrActivationFailed := BuildOr(BuildOr(agentNotSkipped, lockdownCheckFailed), staleLockFileFailed)

	// Build the condition based on whether add_comment job exists
	if slices.Contains(safeOutputJobNames, "add_comment") {
		// If add_comment job exists, also check that it hasn't already created a comment.
		// This prevents duplicate updates when add_comment has already updated the activation comment.
		noAddCommentOutput := &NotNode{
			Child: BuildPropertyAccess("needs.add_comment.outputs.comment_id"),
		}
		return BuildAnd(
			BuildAnd(alwaysFunc, agentNotSkippedOrActivationFailed),
			noAddCommentOutput,
		)
	}

	// If add_comment job doesn't exist, just check the basic conditions
	return BuildAnd(alwaysFunc, agentNotSkippedOrActivationFailed)
}
