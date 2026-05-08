package workflow

import "math"

// extractSafeOutputsEntityHandlers parses entity-related safe-output handlers
// (create-issue, create-agent-session, project management, discussions, comments, etc.)
// from outputMap and populates config accordingly.
func (c *Compiler) extractSafeOutputsEntityHandlers(outputMap map[string]any, config *SafeOutputsConfig) {
	if issuesConfig := c.parseIssuesConfig(outputMap); issuesConfig != nil {
		safeOutputsConfigLog.Print("Configured create-issue output handler")
		config.CreateIssues = issuesConfig
	}
	if agentSessionConfig := c.parseAgentSessionConfig(outputMap); agentSessionConfig != nil {
		config.CreateAgentSessions = agentSessionConfig
	}
	if updateProjectConfig := c.parseUpdateProjectConfig(outputMap); updateProjectConfig != nil {
		config.UpdateProjects = updateProjectConfig
	}
	if createProjectConfig := c.parseCreateProjectsConfig(outputMap); createProjectConfig != nil {
		config.CreateProjects = createProjectConfig
	}
	if createProjectStatusUpdateConfig := c.parseCreateProjectStatusUpdateConfig(outputMap); createProjectStatusUpdateConfig != nil {
		config.CreateProjectStatusUpdates = createProjectStatusUpdateConfig
	}
	if discussionsConfig := c.parseDiscussionsConfig(outputMap); discussionsConfig != nil {
		config.CreateDiscussions = discussionsConfig
	}
	if closeDiscussionsConfig := c.parseCloseDiscussionsConfig(outputMap); closeDiscussionsConfig != nil {
		config.CloseDiscussions = closeDiscussionsConfig
	}
	if closeIssuesConfig := c.parseCloseIssuesConfig(outputMap); closeIssuesConfig != nil {
		config.CloseIssues = closeIssuesConfig
	}
	if closePullRequestsConfig := c.parseClosePullRequestsConfig(outputMap); closePullRequestsConfig != nil {
		config.ClosePullRequests = closePullRequestsConfig
	}
	if markPRReadyConfig := c.parseMarkPullRequestAsReadyForReviewConfig(outputMap); markPRReadyConfig != nil {
		config.MarkPullRequestAsReadyForReview = markPRReadyConfig
	}
	if commentsConfig := c.parseCommentsConfig(outputMap); commentsConfig != nil {
		config.AddComments = commentsConfig
	}
}

// extractSafeOutputsPRAndCodeHandlers parses pull-request and code-scanning safe-output handlers
// (create-pull-request, review comments, review submission, code scanning alerts, etc.)
// from outputMap and populates config accordingly.
func (c *Compiler) extractSafeOutputsPRAndCodeHandlers(outputMap map[string]any, config *SafeOutputsConfig) {
	if pullRequestsConfig := c.parsePullRequestsConfig(outputMap); pullRequestsConfig != nil {
		safeOutputsConfigLog.Print("Configured create-pull-request output handler")
		config.CreatePullRequests = pullRequestsConfig
	}
	if prReviewCommentsConfig := c.parsePullRequestReviewCommentsConfig(outputMap); prReviewCommentsConfig != nil {
		config.CreatePullRequestReviewComments = prReviewCommentsConfig
	}
	if submitPRReviewConfig := c.parseSubmitPullRequestReviewConfig(outputMap); submitPRReviewConfig != nil {
		config.SubmitPullRequestReview = submitPRReviewConfig
	}
	if replyToPRReviewCommentConfig := c.parseReplyToPullRequestReviewCommentConfig(outputMap); replyToPRReviewCommentConfig != nil {
		config.ReplyToPullRequestReviewComment = replyToPRReviewCommentConfig
	}
	if resolvePRReviewThreadConfig := c.parseResolvePullRequestReviewThreadConfig(outputMap); resolvePRReviewThreadConfig != nil {
		config.ResolvePullRequestReviewThread = resolvePRReviewThreadConfig
	}
	if securityReportsConfig := c.parseCodeScanningAlertsConfig(outputMap); securityReportsConfig != nil {
		config.CreateCodeScanningAlerts = securityReportsConfig
	}
	if autofixCodeScanningAlertConfig := c.parseAutofixCodeScanningAlertConfig(outputMap); autofixCodeScanningAlertConfig != nil {
		config.AutofixCodeScanningAlert = autofixCodeScanningAlertConfig
	}
}

// extractSafeOutputsDomainAndManagementHandlers parses domain-filtering, label management,
// assignment, issue/PR update, upload, and workflow dispatch safe-output handlers
// from outputMap and populates config accordingly.
func (c *Compiler) extractSafeOutputsDomainAndManagementHandlers(outputMap map[string]any, config *SafeOutputsConfig) {
	// Allowed domains and GitHub references
	if allowedDomains, exists := outputMap["allowed-domains"]; exists {
		if domainsArray, ok := allowedDomains.([]any); ok {
			var domainStrings []string
			for _, domain := range domainsArray {
				if domainStr, ok := domain.(string); ok {
					domainStrings = append(domainStrings, domainStr)
				}
			}
			config.AllowedDomains = domainStrings
			safeOutputsConfigLog.Printf("Configured allowed-domains with %d domain(s)", len(domainStrings))
		}
	}
	if allowGitHubRefs, exists := outputMap["allowed-github-references"]; exists {
		if refsArray, ok := allowGitHubRefs.([]any); ok {
			refStrings := []string{}
			for _, ref := range refsArray {
				if refStr, ok := ref.(string); ok {
					refStrings = append(refStrings, refStr)
				}
			}
			config.AllowGitHubReferences = refStrings
		}
	}

	// Label and assignment handlers
	if addLabelsConfig := c.parseAddLabelsConfig(outputMap); addLabelsConfig != nil {
		config.AddLabels = addLabelsConfig
	}
	if removeLabelsConfig := c.parseRemoveLabelsConfig(outputMap); removeLabelsConfig != nil {
		config.RemoveLabels = removeLabelsConfig
	}
	if addReviewerConfig := c.parseAddReviewerConfig(outputMap); addReviewerConfig != nil {
		config.AddReviewer = addReviewerConfig
	}
	if assignMilestoneConfig := c.parseAssignMilestoneConfig(outputMap); assignMilestoneConfig != nil {
		config.AssignMilestone = assignMilestoneConfig
	}
	if assignToAgentConfig := c.parseAssignToAgentConfig(outputMap); assignToAgentConfig != nil {
		config.AssignToAgent = assignToAgentConfig
	}
	if assignToUserConfig := c.parseAssignToUserConfig(outputMap); assignToUserConfig != nil {
		config.AssignToUser = assignToUserConfig
	}
	if unassignFromUserConfig := c.parseUnassignFromUserConfig(outputMap); unassignFromUserConfig != nil {
		config.UnassignFromUser = unassignFromUserConfig
	}

	// Issue/PR/discussion update handlers
	if updateIssuesConfig := c.parseUpdateIssuesConfig(outputMap); updateIssuesConfig != nil {
		config.UpdateIssues = updateIssuesConfig
	}
	if updateDiscussionsConfig := c.parseUpdateDiscussionsConfig(outputMap); updateDiscussionsConfig != nil {
		config.UpdateDiscussions = updateDiscussionsConfig
	}
	if updatePullRequestsConfig := c.parseUpdatePullRequestsConfig(outputMap); updatePullRequestsConfig != nil {
		config.UpdatePullRequests = updatePullRequestsConfig
	}
	if mergePullRequestConfig := c.parseMergePullRequestConfig(outputMap); mergePullRequestConfig != nil {
		config.MergePullRequest = mergePullRequestConfig
	}
	if pushToBranchConfig := c.parsePushToPullRequestBranchConfig(outputMap); pushToBranchConfig != nil {
		config.PushToPullRequestBranch = pushToBranchConfig
	}

	// Asset/artifact/release and misc handlers
	if uploadAssetsConfig := c.parseUploadAssetConfig(outputMap); uploadAssetsConfig != nil {
		config.UploadAssets = uploadAssetsConfig
	}
	if uploadArtifactConfig := c.parseUploadArtifactConfig(outputMap); uploadArtifactConfig != nil {
		config.UploadArtifact = uploadArtifactConfig
	}
	if updateReleaseConfig := c.parseUpdateReleaseConfig(outputMap); updateReleaseConfig != nil {
		config.UpdateRelease = updateReleaseConfig
	}
	if linkSubIssueConfig := c.parseLinkSubIssueConfig(outputMap); linkSubIssueConfig != nil {
		config.LinkSubIssue = linkSubIssueConfig
	}
	if hideCommentConfig := c.parseHideCommentConfig(outputMap); hideCommentConfig != nil {
		config.HideComment = hideCommentConfig
	}
	if setIssueTypeConfig := c.parseSetIssueTypeConfig(outputMap); setIssueTypeConfig != nil {
		config.SetIssueType = setIssueTypeConfig
	}
	if setIssueFieldConfig := c.parseSetIssueFieldConfig(outputMap); setIssueFieldConfig != nil {
		config.SetIssueField = setIssueFieldConfig
	}

	// Workflow dispatch/call handlers
	if dispatchWorkflowConfig := c.parseDispatchWorkflowConfig(outputMap); dispatchWorkflowConfig != nil {
		config.DispatchWorkflow = dispatchWorkflowConfig
	}
	if dispatchRepositoryConfig := c.parseDispatchRepositoryConfig(outputMap); dispatchRepositoryConfig != nil {
		config.DispatchRepository = dispatchRepositoryConfig
	}
	if callWorkflowConfig := c.parseCallWorkflowConfig(outputMap); callWorkflowConfig != nil {
		config.CallWorkflow = callWorkflowConfig
	}
}

// extractSafeOutputsDefaultHandlers parses the four default-enabled signal handlers
// (missing-tool, missing-data, noop, report-incomplete) from outputMap.
// Each handler is enabled by default when the key is absent, or disabled with explicit false.
func (c *Compiler) extractSafeOutputsDefaultHandlers(outputMap map[string]any, config *SafeOutputsConfig) {
	// Handle missing-tool (parse configuration if present, or enable by default)
	if missingToolConfig := c.parseMissingToolConfig(outputMap); missingToolConfig != nil {
		config.MissingTool = missingToolConfig
	} else if _, exists := outputMap["missing-tool"]; !exists {
		trueVal := "true"
		config.MissingTool = &MissingToolConfig{CreateIssue: &trueVal, TitlePrefix: "", Labels: nil}
	}

	// Handle missing-data (parse configuration if present, or enable by default)
	if missingDataConfig := c.parseMissingDataConfig(outputMap); missingDataConfig != nil {
		config.MissingData = missingDataConfig
	} else if _, exists := outputMap["missing-data"]; !exists {
		trueVal := "true"
		config.MissingData = &MissingDataConfig{CreateIssue: &trueVal, TitlePrefix: "", Labels: nil}
	}

	// Handle noop (parse configuration if present, or enable by default as fallback)
	if noopConfig := c.parseNoOpConfig(outputMap); noopConfig != nil {
		config.NoOp = noopConfig
	} else if _, exists := outputMap["noop"]; !exists {
		config.NoOp = &NoOpConfig{}
		config.NoOp.Max = defaultIntStr(1)
		trueVal := "true"
		config.NoOp.ReportAsIssue = &trueVal
	}

	// Handle report-incomplete (parse configuration if present, or enable by default)
	if reportIncompleteConfig := c.parseReportIncompleteConfig(outputMap); reportIncompleteConfig != nil {
		config.ReportIncomplete = reportIncompleteConfig
	} else if _, exists := outputMap["report-incomplete"]; !exists {
		trueVal := "true"
		config.ReportIncomplete = &ReportIncompleteConfig{CreateIssue: &trueVal, TitlePrefix: "", Labels: nil}
	}
}

// extractSafeOutputsGlobalConfig parses all global/cross-cutting configuration fields
// (patch limits, threat detection, messaging, flags, steps, infrastructure)
// from outputMap and populates config accordingly.
func (c *Compiler) extractSafeOutputsGlobalConfig(outputMap map[string]any, config *SafeOutputsConfig) {
	// Handle staged flag
	if staged, exists := outputMap["staged"]; exists {
		if stagedBool, ok := staged.(bool); ok {
			config.Staged = stagedBool
		}
	}

	// Handle env configuration
	if env, exists := outputMap["env"]; exists {
		if envMap, ok := env.(map[string]any); ok {
			config.Env = make(map[string]string)
			for key, value := range envMap {
				if valueStr, ok := value.(string); ok {
					config.Env[key] = valueStr
				}
			}
		}
	}

	// Handle github-token configuration
	if githubToken, exists := outputMap["github-token"]; exists {
		if githubTokenStr, ok := githubToken.(string); ok {
			config.GitHubToken = githubTokenStr
		}
	}

	// Handle max-patch-size configuration
	if maxPatchSize, exists := outputMap["max-patch-size"]; exists {
		switch v := maxPatchSize.(type) {
		case int:
			if v >= 1 {
				config.MaximumPatchSize = v
			}
		case int64:
			if v >= 1 {
				config.MaximumPatchSize = int(v)
			}
		case uint64:
			if v >= 1 {
				config.MaximumPatchSize = int(v)
			}
		case float64:
			intVal := int(v)
			if v != float64(intVal) {
				safeOutputsConfigLog.Printf("max-patch-size: float value %.2f truncated to integer %d", v, intVal)
			}
			if intVal >= 1 {
				config.MaximumPatchSize = intVal
			}
		}
	}
	if config.MaximumPatchSize == 0 {
		config.MaximumPatchSize = 1024 // Default to 1MB = 1024 KB
	}

	// Handle max-patch-files configuration
	if maxPatchFiles, exists := outputMap["max-patch-files"]; exists {
		switch v := maxPatchFiles.(type) {
		case int:
			if v >= 1 {
				config.MaximumPatchFiles = v
			}
		case int64:
			if v >= 1 {
				if v > int64(math.MaxInt) {
					safeOutputsConfigLog.Printf("max-patch-files: int64 value %d exceeds platform int range, clamping to %d", v, math.MaxInt)
					config.MaximumPatchFiles = math.MaxInt
				} else {
					config.MaximumPatchFiles = int(v)
				}
			}
		case uint64:
			if v >= 1 {
				if v > uint64(math.MaxInt) {
					safeOutputsConfigLog.Printf("max-patch-files: uint64 value %d exceeds platform int range, clamping to %d", v, math.MaxInt)
					config.MaximumPatchFiles = math.MaxInt
				} else {
					config.MaximumPatchFiles = int(v)
				}
			}
		case float64:
			if v != v || v > float64(math.MaxInt) || v < float64(math.MinInt) {
				safeOutputsConfigLog.Printf("max-patch-files: float value %.2f is out of range, ignoring", v)
				break
			}
			intVal := int(v)
			if v != float64(intVal) {
				safeOutputsConfigLog.Printf("max-patch-files: float value %.2f truncated to integer %d", v, intVal)
			}
			if intVal >= 1 {
				config.MaximumPatchFiles = intVal
			}
		}
	}
	if config.MaximumPatchFiles == 0 {
		config.MaximumPatchFiles = 100 // Default to 100 unique files
	}

	// Handle threat-detection
	if threatDetectionConfig := c.parseThreatDetectionConfig(outputMap); threatDetectionConfig != nil {
		config.ThreatDetection = threatDetectionConfig
	}

	// Handle runs-on configuration
	if runsOn, exists := outputMap["runs-on"]; exists {
		if runsOnStr, ok := runsOn.(string); ok {
			config.RunsOn = runsOnStr
		}
	}

	// Handle messages configuration
	if messages, exists := outputMap["messages"]; exists {
		if messagesMap, ok := messages.(map[string]any); ok {
			config.Messages = parseMessagesConfig(messagesMap)
		}
	}

	// Handle activation-comments at safe-outputs top level (templatable boolean)
	if err := preprocessBoolFieldAsString(outputMap, "activation-comments", safeOutputsConfigLog); err != nil {
		safeOutputsConfigLog.Printf("activation-comments: %v", err)
	}
	if activationComments, exists := outputMap["activation-comments"]; exists {
		if activationCommentsStr, ok := activationComments.(string); ok && activationCommentsStr != "" {
			if config.Messages == nil {
				config.Messages = &SafeOutputMessagesConfig{}
			}
			config.Messages.ActivationComments = activationCommentsStr
		}
	}

	// Handle mentions configuration
	if mentions, exists := outputMap["mentions"]; exists {
		config.Mentions = parseMentionsConfig(mentions)
	}

	// Handle global footer flag
	if footer, exists := outputMap["footer"]; exists {
		if footerBool, ok := footer.(bool); ok {
			config.Footer = &footerBool
			safeOutputsConfigLog.Printf("Global footer control: %t", footerBool)
		}
	}

	// Handle group-reports flag
	if groupReports, exists := outputMap["group-reports"]; exists {
		if groupReportsBool, ok := groupReports.(bool); ok {
			config.GroupReports = groupReportsBool
			safeOutputsConfigLog.Printf("Group reports control: %t", groupReportsBool)
		}
	}

	// Handle report-failure-as-issue flag
	if reportFailureAsIssue, exists := outputMap["report-failure-as-issue"]; exists {
		if reportFailureAsIssueBool, ok := reportFailureAsIssue.(bool); ok {
			config.ReportFailureAsIssue = &reportFailureAsIssueBool
			safeOutputsConfigLog.Printf("Report failure as issue: %t", reportFailureAsIssueBool)
		}
	}

	// Handle failure-issue-repo (repository for failure issues, format: "owner/repo")
	if failureIssueRepo, exists := outputMap["failure-issue-repo"]; exists {
		if failureIssueRepoStr, ok := failureIssueRepo.(string); ok && failureIssueRepoStr != "" {
			config.FailureIssueRepo = failureIssueRepoStr
			safeOutputsConfigLog.Printf("Failure issue repo: %s", failureIssueRepoStr)
		}
	}

	// Handle max-bot-mentions (templatable integer)
	if err := preprocessIntFieldAsString(outputMap, "max-bot-mentions", safeOutputsConfigLog); err != nil {
		safeOutputsConfigLog.Printf("max-bot-mentions: %v", err)
	} else if maxBotMentions, exists := outputMap["max-bot-mentions"]; exists {
		if maxBotMentionsStr, ok := maxBotMentions.(string); ok {
			config.MaxBotMentions = &maxBotMentionsStr
		}
	}

	// Handle steps (user-provided steps injected after checkout/setup, before safe-output code)
	if steps, exists := outputMap["steps"]; exists {
		if stepsList, ok := steps.([]any); ok {
			config.Steps = stepsList
			safeOutputsConfigLog.Printf("Configured %d user-provided steps for safe-outputs", len(stepsList))
		}
	}

	// Handle id-token permission override ("write" to force-add, "none" to disable auto-detection)
	if idToken, exists := outputMap["id-token"]; exists {
		if idTokenStr, ok := idToken.(string); ok {
			if idTokenStr == "write" || idTokenStr == "none" {
				config.IDToken = &idTokenStr
				safeOutputsConfigLog.Printf("Configured id-token permission override: %s", idTokenStr)
			} else {
				safeOutputsConfigLog.Printf("Warning: unrecognized safe-outputs id-token value %q (expected \"write\" or \"none\"); ignoring", idTokenStr)
			}
		}
	}

	// Handle concurrency-group configuration
	if concurrencyGroup, exists := outputMap["concurrency-group"]; exists {
		if concurrencyGroupStr, ok := concurrencyGroup.(string); ok && concurrencyGroupStr != "" {
			config.ConcurrencyGroup = concurrencyGroupStr
			safeOutputsConfigLog.Printf("Configured concurrency-group for safe-outputs job: %s", concurrencyGroupStr)
		}
	}

	// Handle needs configuration
	if needsValue, exists := outputMap["needs"]; exists {
		if needsArray, ok := needsValue.([]any); ok {
			for _, need := range needsArray {
				if needStr, ok := need.(string); ok && needStr != "" {
					config.Needs = append(config.Needs, needStr)
				}
			}
			if len(config.Needs) > 0 {
				safeOutputsConfigLog.Printf("Configured %d explicit safe-outputs needs dependency(ies)", len(config.Needs))
			}
		}
	}

	// Handle environment configuration (override for safe-outputs job)
	config.Environment = c.extractTopLevelYAMLSection(outputMap, "environment")
	if config.Environment != "" {
		safeOutputsConfigLog.Printf("Configured environment override for safe-outputs job: %s", config.Environment)
	}

	// Handle jobs (safe-jobs must be under safe-outputs)
	if jobs, exists := outputMap["jobs"]; exists {
		if jobsMap, ok := jobs.(map[string]any); ok {
			cc := &Compiler{}
			config.Jobs = cc.parseSafeJobsConfig(jobsMap)
		}
	}

	// Handle scripts (inline handlers that run in the safe-output handler loop)
	if scripts, exists := outputMap["scripts"]; exists {
		if scriptsMap, ok := scripts.(map[string]any); ok {
			config.Scripts = parseSafeScriptsConfig(scriptsMap)
			safeOutputsConfigLog.Printf("Configured %d custom safe-output script(s)", len(config.Scripts))
		}
	}

	// Handle actions (custom GitHub Actions mounted as safe output tools)
	if actions, exists := outputMap["actions"]; exists {
		if actionsMap, ok := actions.(map[string]any); ok {
			config.Actions = parseActionsConfig(actionsMap)
			safeOutputsConfigLog.Printf("Configured %d custom safe-output action(s)", len(config.Actions))
		}
	}

	// Handle app configuration for GitHub App token minting
	if app, exists := outputMap["github-app"]; exists {
		if appMap, ok := app.(map[string]any); ok {
			config.GitHubApp = parseAppConfig(appMap)
		}
	}
}
