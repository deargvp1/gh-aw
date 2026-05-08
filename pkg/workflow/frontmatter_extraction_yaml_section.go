package workflow

import "strings"

// onSectionParseState tracks the current parsing position within the `on:` YAML section.
// Each field indicates whether the parser is currently inside a particular nested block.
type onSectionParseState struct {
	inPullRequest                bool
	inIssues                     bool
	inDiscussion                 bool
	inIssueComment               bool
	inDeploymentStatus           bool
	inWorkflowRun                bool
	inWorkflowRunConclusionArray bool
	inForksArray                 bool
	inSkipIfMatch                bool
	inSkipIfNoMatch              bool
	inSkipIfCheckFailing         bool
	inSkipRolesArray             bool
	inSkipBotsArray              bool
	inRolesArray                 bool
	inBotsArray                  bool
	inGitHubApp                  bool
	inOnSteps                    bool
	inOnPermissions              bool
	currentSection               string // "issues", "pull_request", "discussion", or "issue_comment"
}

// buildNativeLabelFilterSections scans the frontmatter `on:` map and returns a set of section
// keys (e.g. "pull_request", "issues") that use native GitHub label filtering (i.e. have the
// `__gh_aw_native_label_filter__: true` marker).  These sections must NOT have their `names:`
// field commented out.
func buildNativeLabelFilterSections(frontmatter map[string]any) map[string]bool {
	sections := make(map[string]bool)
	if onValue, exists := frontmatter["on"]; exists {
		if onMap, ok := onValue.(map[string]any); ok {
			for _, sectionKey := range []string{"issues", "pull_request", "discussion", "issue_comment"} {
				if sectionValue, hasSec := onMap[sectionKey]; hasSec {
					if sectionMap, ok := sectionValue.(map[string]any); ok {
						if marker, hasMarker := sectionMap["__gh_aw_native_label_filter__"]; hasMarker {
							if useNative, ok := marker.(bool); ok && useNative {
								sections[sectionKey] = true
								frontmatterLog.Printf("Section %s uses native label filtering", sectionKey)
							}
						}
					}
				}
			}
		}
	}
	return sections
}

// updateSectionEntry detects whether line is a section-header for pull_request, issues,
// discussion, issue_comment, deployment_status, or workflow_run.  When it is, the state is
// updated accordingly and the method returns true, signalling the caller to append the line
// to the result slice and continue to the next loop iteration.
// This check is skipped when inside on.permissions or on.steps to avoid false positives.
func (s *onSectionParseState) updateSectionEntry(line string) bool {
	if s.inOnPermissions || s.inOnSteps {
		return false
	}
	if strings.Contains(line, "pull_request:") {
		s.inPullRequest, s.inIssues, s.inDiscussion, s.inIssueComment = true, false, false, false
		s.inDeploymentStatus, s.inWorkflowRun, s.inWorkflowRunConclusionArray = false, false, false
		s.currentSection = "pull_request"
		return true
	}
	if strings.Contains(line, "issues:") {
		s.inIssues, s.inPullRequest, s.inDiscussion, s.inIssueComment = true, false, false, false
		s.inDeploymentStatus, s.inWorkflowRun, s.inWorkflowRunConclusionArray = false, false, false
		s.currentSection = "issues"
		return true
	}
	if strings.Contains(line, "discussion:") {
		s.inDiscussion, s.inPullRequest, s.inIssues, s.inIssueComment = true, false, false, false
		s.inDeploymentStatus, s.inWorkflowRun, s.inWorkflowRunConclusionArray = false, false, false
		s.currentSection = "discussion"
		return true
	}
	if strings.Contains(line, "issue_comment:") {
		s.inIssueComment, s.inPullRequest, s.inIssues, s.inDiscussion = true, false, false, false
		s.inDeploymentStatus, s.inWorkflowRun, s.inWorkflowRunConclusionArray = false, false, false
		s.currentSection = "issue_comment"
		return true
	}
	if strings.Contains(line, "deployment_status:") {
		s.inDeploymentStatus, s.inWorkflowRun = true, false
		s.inPullRequest, s.inIssues, s.inDiscussion, s.inIssueComment = false, false, false, false
		s.currentSection = ""
		return true
	}
	if strings.Contains(line, "workflow_run:") {
		s.inWorkflowRun, s.inDeploymentStatus = true, false
		s.inPullRequest, s.inIssues, s.inDiscussion, s.inIssueComment = false, false, false, false
		s.currentSection = ""
		return true
	}
	return false
}

// updateSectionExit checks whether the current line exits an active section block.
// It handles the pull_request/issues/discussion/issue_comment group and the
// deployment_status and workflow_run sections.
func (s *onSectionParseState) updateSectionExit(line string) {
	notIndented := strings.TrimSpace(line) != "" &&
		!strings.HasPrefix(line, "    ") &&
		!strings.HasPrefix(line, "\t")

	if (s.inPullRequest || s.inIssues || s.inDiscussion || s.inIssueComment) && notIndented {
		s.inPullRequest, s.inIssues, s.inDiscussion, s.inIssueComment = false, false, false, false
		s.inForksArray = false
		s.currentSection = ""
	}
	if s.inDeploymentStatus && notIndented {
		s.inDeploymentStatus = false
	}
	if s.inWorkflowRun && notIndented {
		s.inWorkflowRun = false
		s.inWorkflowRunConclusionArray = false
	}
}

// notInEventSections returns true when the parser is NOT inside one of the four
// section-level event blocks (pull_request / issues / discussion / issue_comment).
func (s *onSectionParseState) notInEventSections() bool {
	return !s.inPullRequest && !s.inIssues && !s.inDiscussion && !s.inIssueComment
}

// updateObjectEntry detects entry into forks, skip-roles, skip-bots, roles, bots,
// steps, permissions, skip-if-match, skip-if-no-match, skip-if-check-failing, and
// github-app object/array blocks and updates the tracking state accordingly.
func (s *onSectionParseState) updateObjectEntry(trimmedLine string) {
	if s.inPullRequest && strings.HasPrefix(trimmedLine, "forks:") {
		s.inForksArray = true
	}
	if s.notInEventSections() {
		if strings.HasPrefix(trimmedLine, "skip-roles:") {
			s.inSkipRolesArray = true
		}
		if strings.HasPrefix(trimmedLine, "skip-bots:") {
			s.inSkipBotsArray = true
		}
		if strings.HasPrefix(trimmedLine, "roles:") {
			s.inRolesArray = true
		}
		if strings.HasPrefix(trimmedLine, "bots:") {
			s.inBotsArray = true
		}
		if strings.HasPrefix(trimmedLine, "steps:") {
			s.inOnSteps = true
		}
		if !s.inOnPermissions && strings.HasPrefix(trimmedLine, "permissions:") {
			s.inOnPermissions = true
		}
		if !s.inSkipIfMatch {
			if (strings.HasPrefix(trimmedLine, "skip-if-match:") && trimmedLine == "skip-if-match:") ||
				(strings.HasPrefix(trimmedLine, "# skip-if-match:") && strings.Contains(trimmedLine, "pre-activation job")) {
				s.inSkipIfMatch = true
			}
		}
		if !s.inSkipIfNoMatch {
			if (strings.HasPrefix(trimmedLine, "skip-if-no-match:") && trimmedLine == "skip-if-no-match:") ||
				(strings.HasPrefix(trimmedLine, "# skip-if-no-match:") && strings.Contains(trimmedLine, "pre-activation job")) {
				s.inSkipIfNoMatch = true
			}
		}
		if !s.inSkipIfCheckFailing {
			if trimmedLine == "skip-if-check-failing:" ||
				(strings.HasPrefix(trimmedLine, "# skip-if-check-failing:") && strings.Contains(trimmedLine, "pre-activation job")) {
				s.inSkipIfCheckFailing = true
			}
		}
		if !s.inGitHubApp {
			if (strings.HasPrefix(trimmedLine, "github-app:") && trimmedLine == "github-app:") ||
				(strings.HasPrefix(trimmedLine, "# github-app:") && strings.Contains(trimmedLine, "pre-activation job")) {
				s.inGitHubApp = true
			}
		}
	}
}

// indentOf returns the number of leading spaces/tabs in line.
func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

// updateObjectExit detects when the parser exits one of the tracked object/array
// blocks by encountering a sibling key at the same indentation level.
func (s *onSectionParseState) updateObjectExit(line, trimmedLine string) {
	if strings.TrimSpace(line) == "" {
		return
	}

	lineIndent := indentOf(line)

	// Exit skip-if-match block
	if s.inSkipIfMatch &&
		!strings.HasPrefix(trimmedLine, "skip-if-match:") &&
		!strings.HasPrefix(trimmedLine, "# skip-if-match:") &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "#") {
		s.inSkipIfMatch = false
	}

	// Exit skip-if-no-match block
	if s.inSkipIfNoMatch &&
		!strings.HasPrefix(trimmedLine, "skip-if-no-match:") &&
		!strings.HasPrefix(trimmedLine, "# skip-if-no-match:") &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "#") {
		s.inSkipIfNoMatch = false
	}

	// Exit skip-if-check-failing block
	if s.inSkipIfCheckFailing &&
		!strings.HasPrefix(trimmedLine, "skip-if-check-failing:") &&
		!strings.HasPrefix(trimmedLine, "# skip-if-check-failing:") &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "#") {
		s.inSkipIfCheckFailing = false
	}

	// Exit github-app block
	if s.inGitHubApp &&
		!strings.HasPrefix(trimmedLine, "github-app:") &&
		!strings.HasPrefix(trimmedLine, "# github-app:") &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "#") {
		s.inGitHubApp = false
	}

	// Exit forks array
	if s.inForksArray && s.inPullRequest &&
		lineIndent == 4 && !strings.HasPrefix(trimmedLine, "-") && !strings.HasPrefix(trimmedLine, "forks:") {
		s.inForksArray = false
	}

	// Exit skip-roles array
	if s.inSkipRolesArray &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "-") &&
		!strings.HasPrefix(trimmedLine, "skip-roles:") && !strings.HasPrefix(trimmedLine, "#") {
		s.inSkipRolesArray = false
	}

	// Exit skip-bots array
	if s.inSkipBotsArray &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "-") &&
		!strings.HasPrefix(trimmedLine, "skip-bots:") && !strings.HasPrefix(trimmedLine, "#") {
		s.inSkipBotsArray = false
	}

	// Exit roles array
	if s.inRolesArray &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "-") &&
		!strings.HasPrefix(trimmedLine, "roles:") && !strings.HasPrefix(trimmedLine, "#") {
		s.inRolesArray = false
	}

	// Exit bots array
	if s.inBotsArray &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "-") &&
		!strings.HasPrefix(trimmedLine, "bots:") && !strings.HasPrefix(trimmedLine, "#") {
		s.inBotsArray = false
	}

	// Exit on.steps array
	if s.inOnSteps &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "-") &&
		!strings.HasPrefix(trimmedLine, "steps:") && !strings.HasPrefix(trimmedLine, "#") {
		s.inOnSteps = false
	}

	// Exit on.permissions block
	if s.inOnPermissions &&
		!strings.HasPrefix(trimmedLine, "permissions:") &&
		!strings.HasPrefix(trimmedLine, "# permissions:") &&
		lineIndent == 2 && !strings.HasPrefix(trimmedLine, "#") {
		s.inOnPermissions = false
	}
}

// classifyForCommentOut determines whether the current line should be commented out in the
// compiled YAML output and returns (shouldComment, commentReason).
// result is passed read-only to allow lookback at previously processed lines.
func (s *onSectionParseState) classifyForCommentOut(
	trimmedLine, line string,
	nativeLabelFilterSections map[string]bool,
	result []string,
) (bool, string) {
	// Top-level fields (not inside pull_request / issues / discussion / issue_comment)
	if s.notInEventSections() {
		switch {
		case strings.HasPrefix(trimmedLine, "manual-approval:"):
			return true, " # Manual approval processed as environment field in activation job"
		case strings.HasPrefix(trimmedLine, "stop-after:"):
			return true, " # Stop-after processed as stop-time check in pre-activation job"
		case strings.HasPrefix(trimmedLine, "skip-if-match:"):
			return true, " # Skip-if-match processed as search check in pre-activation job"
		case s.inSkipIfMatch && (strings.HasPrefix(trimmedLine, "query:") || strings.HasPrefix(trimmedLine, "max:") || strings.HasPrefix(trimmedLine, "scope:")):
			return true, ""
		case strings.HasPrefix(trimmedLine, "skip-if-no-match:"):
			return true, " # Skip-if-no-match processed as search check in pre-activation job"
		case s.inSkipIfNoMatch && (strings.HasPrefix(trimmedLine, "query:") || strings.HasPrefix(trimmedLine, "min:") || strings.HasPrefix(trimmedLine, "scope:")):
			return true, ""
		case strings.HasPrefix(trimmedLine, "skip-if-check-failing:"):
			return true, " # Skip-if-check-failing processed as check status gate in pre-activation job"
		case s.inSkipIfCheckFailing && (strings.HasPrefix(trimmedLine, "include:") || strings.HasPrefix(trimmedLine, "exclude:") || strings.HasPrefix(trimmedLine, "branch:") || strings.HasPrefix(trimmedLine, "allow-pending:") || strings.HasPrefix(trimmedLine, "-")):
			return true, ""
		case strings.HasPrefix(trimmedLine, "skip-roles:"):
			return true, " # Skip-roles processed as role check in pre-activation job"
		case s.inSkipRolesArray && strings.HasPrefix(trimmedLine, "-"):
			return true, " # Skip-roles processed as role check in pre-activation job"
		case strings.HasPrefix(trimmedLine, "skip-bots:"):
			return true, " # Skip-bots processed as bot check in pre-activation job"
		case s.inSkipBotsArray && strings.HasPrefix(trimmedLine, "-"):
			return true, " # Skip-bots processed as bot check in pre-activation job"
		case strings.HasPrefix(trimmedLine, "roles:"):
			return true, " # Roles processed as role check in pre-activation job"
		case s.inRolesArray && strings.HasPrefix(trimmedLine, "-"):
			return true, " # Roles processed as role check in pre-activation job"
		case strings.HasPrefix(trimmedLine, "bots:"):
			return true, " # Bots processed as bot check in pre-activation job"
		case s.inBotsArray && strings.HasPrefix(trimmedLine, "-"):
			return true, " # Bots processed as bot check in pre-activation job"
		case strings.HasPrefix(trimmedLine, "steps:"):
			return true, " # Steps injected into pre-activation job"
		case s.inOnSteps:
			return true, ""
		case strings.HasPrefix(trimmedLine, "permissions:"):
			return true, " # Permissions applied to pre-activation job"
		case s.inOnPermissions:
			return true, ""
		case strings.HasPrefix(trimmedLine, "reaction:"):
			return true, " # Reaction processed as activation job step"
		case strings.HasPrefix(trimmedLine, "github-token:"):
			return true, " # GitHub token used for reactions and status comments in activation"
		case strings.HasPrefix(trimmedLine, "github-app:"):
			return true, " # GitHub App used to mint token for reactions and status comments in activation"
		case s.inGitHubApp && isGitHubAppNestedField(trimmedLine):
			return true, ""
		case strings.HasPrefix(trimmedLine, "stale-check:"):
			return true, " # Stale-check processed as frontmatter hash check step in activation job"
		}
	}

	// Section-specific fields (inside pull_request / issues / discussion / issue_comment / etc.)
	return s.classifyEventSectionLine(trimmedLine, line, nativeLabelFilterSections, result)
}

// classifyEventSectionLine handles classification of lines within event-trigger sections.
func (s *onSectionParseState) classifyEventSectionLine(
	trimmedLine, line string,
	nativeLabelFilterSections map[string]bool,
	result []string,
) (bool, string) {
	if s.inPullRequest && strings.Contains(trimmedLine, "draft:") {
		return true, " # Draft filtering applied via job conditions"
	}
	if s.inPullRequest && strings.HasPrefix(trimmedLine, "forks:") {
		return true, " # Fork filtering applied via job conditions"
	}
	if s.inForksArray && strings.HasPrefix(trimmedLine, "-") {
		return true, " # Fork filtering applied via job conditions"
	}
	if s.inDeploymentStatus && strings.HasPrefix(trimmedLine, "state:") {
		return true, " # State filtering compiled into if condition"
	}
	if s.inDeploymentStatus && strings.HasPrefix(trimmedLine, "-") {
		return true, " # State filtering compiled into if condition"
	}
	if s.inWorkflowRun && strings.HasPrefix(trimmedLine, "conclusion:") {
		s.inWorkflowRunConclusionArray = true
		return true, " # Conclusion filtering compiled into if condition"
	}
	if s.inWorkflowRunConclusionArray && strings.HasPrefix(trimmedLine, "-") {
		return true, " # Conclusion filtering compiled into if condition"
	}
	if s.inWorkflowRun && !strings.HasPrefix(trimmedLine, "-") && strings.Contains(trimmedLine, ":") {
		s.inWorkflowRunConclusionArray = false
	}
	if (s.inPullRequest || s.inIssues || s.inDiscussion || s.inIssueComment) && strings.HasPrefix(trimmedLine, "lock-for-agent:") {
		return true, " # Lock-for-agent processed as issue locking in activation job"
	}
	if (s.inPullRequest || s.inIssues || s.inDiscussion || s.inIssueComment) && strings.HasPrefix(trimmedLine, "names:") {
		if !nativeLabelFilterSections[s.currentSection] {
			return true, " # Label filtering applied via job conditions"
		}
	}
	// Check if current line is a names array item (lookback logic)
	if (s.inPullRequest || s.inIssues || s.inDiscussion || s.inIssueComment) && line != "" {
		if !nativeLabelFilterSections[s.currentSection] {
			return classifyNamesArrayItem(trimmedLine, result)
		}
	}
	return false, ""
}

// classifyNamesArrayItem checks whether trimmedLine is an array item that belongs to a
// "names:" field that was commented out.  It performs a lookback over result to find the
// most recent preceding non-empty line.
func classifyNamesArrayItem(trimmedLine string, result []string) (bool, string) {
	for i := len(result) - 1; i >= 0; i-- {
		prevLine := result[i]
		prevTrimmed := strings.TrimSpace(prevLine)
		if prevTrimmed == "" {
			continue
		}
		if strings.Contains(prevTrimmed, "names:") && strings.Contains(prevTrimmed, "# Label filtering") {
			if strings.HasPrefix(trimmedLine, "-") {
				return true, " # Label filtering applied via job conditions"
			}
			break
		}
		if !strings.HasPrefix(prevTrimmed, "#") || !strings.Contains(prevTrimmed, "Label filtering") {
			break
		}
		if strings.HasPrefix(prevTrimmed, "# -") && strings.Contains(prevTrimmed, "Label filtering") {
			if strings.HasPrefix(trimmedLine, "-") {
				return true, " # Label filtering applied via job conditions"
			}
			continue
		}
		break
	}
	return false, ""
}
