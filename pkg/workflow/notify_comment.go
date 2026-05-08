package workflow

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var notifyCommentLog = logger.New("workflow:notify_comment")

// buildConclusionJob creates a job that handles workflow completion tasks
// This job is generated when safe-outputs are configured and handles:
// - Updating status comments (if status-comment: true)
// - Processing noop messages
// - Handling agent failures
// - Recording missing tools
// This job runs when:
// 1. always() - runs even if agent fails
// 2. Agent job was not skipped
// 3. NO add_comment output was produced by the agent (avoids duplicate updates)
// This job depends on all safe output jobs to ensure it runs last
func (c *Compiler) buildConclusionJob(data *WorkflowData, mainJobName string, safeOutputJobNames []string) (*Job, error) {
	notifyCommentLog.Printf("Building conclusion job: main_job=%s, safe_output_jobs_count=%d", mainJobName, len(safeOutputJobNames))

	// Always create this job when safe-outputs exist (because noop is always enabled)
	if data.SafeOutputs == nil {
		notifyCommentLog.Printf("Skipping job: no safe-outputs configured")
		return nil, nil // No safe-outputs configured, no need for conclusion job
	}

	// Build the job steps
	var steps []string

	// Add setup step to copy scripts
	setupActionRef := c.resolveActionReference("./actions/setup", data)
	if setupActionRef != "" || c.actionMode.IsScript() {
		// For dev mode (local action path), checkout the actions folder first
		steps = append(steps, c.generateCheckoutActionsFolder(data)...)

		// Conclusion/notify job depends on activation, reuse its trace ID
		notifyTraceID := fmt.Sprintf("${{ needs.%s.outputs.setup-trace-id }}", constants.ActivationJobName)
		steps = append(steps, c.generateSetupStep(data, setupActionRef, SetupActionDestination, false, notifyTraceID)...)
	}

	// Add GitHub App token minting step if app is configured
	if data.SafeOutputs.GitHubApp != nil {
		permissions := ComputePermissionsForSafeOutputs(data.SafeOutputs)
		var appTokenFallbackRepo string
		if hasWorkflowCallTrigger(data.On) {
			appTokenFallbackRepo = "${{ needs.activation.outputs.target_repo_name }}"
		}
		steps = append(steps, c.buildGitHubAppTokenMintStep(data.SafeOutputs.GitHubApp, permissions, appTokenFallbackRepo)...)
	}

	// Add artifact download steps once (shared by noop and conclusion steps)
	steps = append(steps, buildAgentOutputDownloadSteps(artifactPrefixExprForDownstreamJob(data))...)

	// Add noop processing step if noop is configured
	steps = append(steps, c.buildConclusionJobNoOpSteps(data, mainJobName)...)

	// Add detection runs logging step if threat detection is enabled
	steps = append(steps, c.buildConclusionJobDetectionRunsSteps(data, mainJobName)...)

	// Add missing_tool processing step if missing-tool is configured
	steps = append(steps, c.buildConclusionJobMissingToolSteps(data, mainJobName)...)

	// Add report_incomplete processing step if report-incomplete is configured
	steps = append(steps, c.buildConclusionJobReportIncompleteSteps(data, mainJobName)...)

	// Serialize messages config once for reuse in both handler steps below.
	var messagesJSON string
	if data.SafeOutputs != nil && data.SafeOutputs.Messages != nil {
		if jsonStr, jsonErr := serializeMessagesConfig(data.SafeOutputs.Messages); jsonErr != nil {
			notifyCommentLog.Printf("Warning: failed to serialize messages config: %v", jsonErr)
		} else {
			messagesJSON = jsonStr
		}
	}

	// Build and add agent failure handling step
	agentFailureEnvVars, err := c.buildConclusionJobAgentFailureEnvVars(data, mainJobName, safeOutputJobNames, messagesJSON)
	if err != nil {
		return nil, err
	}
	agentFailureSteps := c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
		StepName:      "Handle agent failure",
		StepID:        "handle_agent_failure",
		MainJobName:   mainJobName,
		CustomEnvVars: agentFailureEnvVars,
		Script:        "const { main } = require('${{ runner.temp }}/gh-aw/actions/handle_agent_failure.cjs'); await main();",
		ScriptFile:    "handle_agent_failure.cjs",
		CustomToken:   "",
		StepCondition: "always()",
	})
	steps = append(steps, agentFailureSteps...)

	// Build environment variables for the conclusion script
	var customEnvVars []string
	customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_COMMENT_ID: ${{ needs.%s.outputs.comment_id }}\n", constants.ActivationJobName))
	customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_COMMENT_REPO: ${{ needs.%s.outputs.comment_repo }}\n", constants.ActivationJobName))
	customEnvVars = append(customEnvVars, "          GH_AW_RUN_URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}\n")
	customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_WORKFLOW_NAME: %q\n", data.Name))
	if data.TrackerID != "" {
		customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_TRACKER_ID: %q\n", data.TrackerID))
	}
	customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_AGENT_CONCLUSION: ${{ needs.%s.result }}\n", mainJobName))

	// Pass safe_outputs job result so the conclusion script can detect when safe outputs failed
	if slices.Contains(safeOutputJobNames, string(constants.SafeOutputsJobName)) {
		customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_SAFE_OUTPUTS_RESULT: ${{ needs.%s.result }}\n", constants.SafeOutputsJobName))
		notifyCommentLog.Print("Added safe_outputs job result environment variable to conclusion job")
	}

	// Pass detection conclusion and reason if threat detection is enabled
	if IsDetectionJobEnabled(data.SafeOutputs) {
		customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_DETECTION_CONCLUSION: ${{ needs.%s.outputs.detection_conclusion }}\n", constants.DetectionJobName))
		customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_DETECTION_REASON: ${{ needs.%s.outputs.detection_reason }}\n", constants.DetectionJobName))
		notifyCommentLog.Print("Added detection conclusion and reason environment variables to conclusion job")
	}

	// Pass assignment error count to the conclusion step
	if data.SafeOutputs != nil && data.SafeOutputs.AssignToAgent != nil {
		customEnvVars = append(customEnvVars, "          GH_AW_ASSIGNMENT_ERROR_COUNT: ${{ needs.safe_outputs.outputs.assign_to_agent_assignment_error_count }}\n")
	}

	// Pass custom messages config if present (JSON computed once above, reused here)
	if messagesJSON != "" {
		customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_SAFE_OUTPUT_MESSAGES: %q\n", messagesJSON))
	}

	// Pass safe output job information for link generation
	if len(safeOutputJobNames) > 0 {
		safeOutputJobsJSON, jobURLEnvVars := buildSafeOutputJobsEnvVars(safeOutputJobNames)
		if safeOutputJobsJSON != "" {
			customEnvVars = append(customEnvVars, fmt.Sprintf("          GH_AW_SAFE_OUTPUT_JOBS: %q\n", safeOutputJobsJSON))
			customEnvVars = append(customEnvVars, jobURLEnvVars...)
			notifyCommentLog.Printf("Added safe output jobs info for %d job(s)", len(safeOutputJobNames))
		}
	}

	// Only add the conclusion update step if status comments are explicitly enabled
	var token string
	if data.SafeOutputs != nil && data.SafeOutputs.AddComments != nil {
		token = data.SafeOutputs.AddComments.GitHubToken
	}
	if data.StatusComment != nil && *data.StatusComment {
		scriptSteps := c.buildGitHubScriptStepWithoutDownload(data, GitHubScriptStepConfig{
			StepName:      "Update reaction comment with completion status",
			StepID:        "conclusion",
			MainJobName:   mainJobName,
			CustomEnvVars: customEnvVars,
			Script:        getNotifyCommentErrorScript(),
			ScriptFile:    "notify_comment_error.cjs",
			CustomToken:   token,
		})
		steps = append(steps, scriptSteps...)
	}

	// Add GitHub App token invalidation step if app is configured
	if data.SafeOutputs.GitHubApp != nil {
		notifyCommentLog.Print("Adding GitHub App token invalidation step to conclusion job")
		steps = append(steps, c.buildGitHubAppTokenInvalidationStep()...)
	}

	// In script mode, explicitly add a cleanup step
	if c.actionMode.IsScript() {
		steps = append(steps, c.generateScriptModeCleanupStep())
	}

	// Build the if-condition for this job
	condition := buildConclusionJobCondition(mainJobName, safeOutputJobNames)

	// Build dependencies - this job depends on all safe output jobs to ensure it runs last
	needs := []string{mainJobName, string(constants.ActivationJobName)}
	needs = append(needs, safeOutputJobNames...)

	// When threat detection is enabled, the conclusion job also depends on the detection job
	if IsDetectionJobEnabled(data.SafeOutputs) {
		needs = append(needs, string(constants.DetectionJobName))
		notifyCommentLog.Print("Added detection job dependency to conclusion job")
	}

	notifyCommentLog.Printf("Job built successfully: dependencies_count=%d", len(needs))

	// Create outputs for the job
	outputs := map[string]string{}
	if data.SafeOutputs.NoOp != nil {
		outputs["noop_message"] = "${{ steps.noop.outputs.noop_message }}"
	}
	if data.SafeOutputs.MissingTool != nil {
		outputs["tools_reported"] = "${{ steps.missing_tool.outputs.tools_reported }}"
		outputs["total_count"] = "${{ steps.missing_tool.outputs.total_count }}"
	}
	if data.SafeOutputs.ReportIncomplete != nil {
		outputs["incomplete_count"] = "${{ steps.report_incomplete.outputs.incomplete_count }}"
	}

	// Compute permissions based on configured safe outputs (principle of least privilege)
	permissions := ComputePermissionsForSafeOutputs(data.SafeOutputs)

	// Build concurrency config for the conclusion job using the workflow ID
	var concurrency string
	if data.WorkflowID != "" {
		group := "gh-aw-conclusion-" + data.WorkflowID
		if data.ConcurrencyJobDiscriminator != "" {
			notifyCommentLog.Printf("Appending job discriminator to conclusion job concurrency group: %s", data.ConcurrencyJobDiscriminator)
			group = fmt.Sprintf("%s-%s", group, data.ConcurrencyJobDiscriminator)
		}
		concurrency = c.indentYAMLLines(fmt.Sprintf("concurrency:\n  group: %q\n  cancel-in-progress: false", group), "    ")
		notifyCommentLog.Printf("Configuring conclusion job concurrency group: %s", group)
	}

	job := &Job{
		Name:        "conclusion",
		If:          RenderCondition(condition),
		RunsOn:      c.formatFrameworkJobRunsOn(data),
		Environment: c.indentYAMLLines(resolveSafeOutputsEnvironment(data), "    "),
		Permissions: permissions.RenderToYAML(),
		Concurrency: concurrency,
		Steps:       steps,
		Needs:       needs,
		Outputs:     outputs,
	}

	return job, nil
}

// systemSafeOutputJobNames contains job names that are built-in system jobs and should not be
// treated as custom safe output job types in the GH_AW_SAFE_OUTPUT_JOBS mapping.
// The safe output handler manager uses this mapping to determine which message types are
// handled by custom job steps (and therefore should be silently skipped rather than flagged
// as "no handler loaded").
var systemSafeOutputJobNames = map[string]bool{
	"safe_outputs":  true, // consolidated safe outputs job
	"upload_assets": true, // upload assets job
}

// buildSafeOutputJobsEnvVars creates environment variables for safe output job URLs
// Returns both a JSON mapping and the actual environment variable declarations.
// The mapping includes:
//   - Built-in jobs with known URL outputs (e.g., create_issue → issue_url)
//   - Custom safe-output jobs (from safe-outputs.jobs) with an empty URL key, so the handler
//     manager knows those message types are handled by a dedicated job step and should be
//     skipped gracefully rather than reported as "No handler loaded".
func buildSafeOutputJobsEnvVars(jobNames []string) (string, []string) {
	// Map job names to their expected URL output keys
	jobOutputMapping := make(map[string]string)
	var envVars []string

	for _, jobName := range jobNames {
		var urlKey string
		switch jobName {
		case "create_issue":
			urlKey = "issue_url"
		case "add_comment":
			urlKey = "comment_url"
		case "create_pull_request":
			urlKey = "pull_request_url"
		case "create_discussion":
			urlKey = "discussion_url"
		case "create_pr_review_comment":
			urlKey = "review_comment_url"
		case "close_issue":
			urlKey = "issue_url"
		case "close_pull_request":
			urlKey = "pull_request_url"
		case "close_discussion":
			urlKey = "discussion_url"
		case "create_agent_session":
			urlKey = "task_url"
		case "push_to_pull_request_branch":
			urlKey = "commit_url"
		default:
			if !systemSafeOutputJobNames[jobName] {
				// Custom safe-output job: include in the mapping with an empty URL key so the
				// handler manager can silently skip messages of this type.
				jobOutputMapping[jobName] = ""
			}
			continue
		}

		jobOutputMapping[jobName] = urlKey

		// Add environment variable for this job's URL output
		envVarName := fmt.Sprintf("GH_AW_OUTPUT_%s_%s",
			toEnvVarCase(jobName),
			toEnvVarCase(urlKey))
		envVars = append(envVars,
			fmt.Sprintf("          %s: ${{ needs.%s.outputs.%s }}\n",
				envVarName, jobName, urlKey))
	}

	if len(jobOutputMapping) == 0 {
		return "", nil
	}

	jsonBytes, err := json.Marshal(jobOutputMapping)
	if err != nil {
		notifyCommentLog.Printf("Warning: failed to marshal safe output jobs info: %v", err)
		return "", nil
	}

	return string(jsonBytes), envVars
}

// toEnvVarCase converts a string to uppercase environment variable case
func toEnvVarCase(s string) string {
	// Convert to uppercase and keep underscores
	var result strings.Builder
	for _, ch := range s {
		if ch >= 'a' && ch <= 'z' {
			result.WriteRune(ch - 32) // Convert to uppercase
		} else if ch >= 'A' && ch <= 'Z' {
			result.WriteRune(ch)
		} else if ch == '_' {
			result.WriteString("_")
		}
	}
	return result.String()
}

// getEngineAPIHosts returns the primary AI inference API hostnames for the given engine and
// workflow data. These are the hosts that appear in the firewall audit log when the engine
// makes authenticated API calls. The returned slice is used to populate GH_AW_ENGINE_API_HOSTS
// so the failure handler can detect credential authentication rejections without relying solely
// on hardcoded host patterns.
//
// Resolution order (per engine):
//   - engine.api-target (explicit GHES / enterprise override) takes precedence
//   - Default public API hostname(s) for the engine
func getEngineAPIHosts(data *WorkflowData, engine CodingAgentEngine) []string {
	if engine == nil {
		return nil
	}

	// Explicit api-target overrides the engine-specific default for all engine types.
	if data != nil && data.EngineConfig != nil && data.EngineConfig.APITarget != "" {
		return []string{data.EngineConfig.APITarget}
	}

	switch engine.(type) {
	case *CopilotEngine:
		// Return the full set of known Copilot inference endpoints so that any variant
		// (enterprise, business, individual, or the routing hub) is covered.
		return []string{
			"api.enterprise.githubcopilot.com",
			"api.githubcopilot.com",
			"api.business.githubcopilot.com",
			"api.individual.githubcopilot.com",
		}
	case *ClaudeEngine:
		return []string{"api.anthropic.com"}
	case *CodexEngine:
		return []string{"api.openai.com"}
	case *GeminiEngine:
		return []string{DefaultGeminiAPITarget}
	default:
		// Custom or unknown engine — no known API hosts without explicit api-target.
		return nil
	}
}
