// This file provides command-line interface functionality for gh-aw.
// This file (logs_utils.go) contains utility functions used by the logs command.
//
// Key responsibilities:
//   - Discovering agentic workflow names from .lock.yml files
//   - Utility functions for slice operations

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/workflow"
)

var logsUtilsLog = logger.New("cli:logs_utils")

// getAgenticWorkflowNames reads all .lock.yml files and extracts their workflow names
func getAgenticWorkflowNames(verbose bool) ([]string, error) {
	logsUtilsLog.Print("Discovering agentic workflow names from .lock.yml files")
	var workflowNames []string

	// Look for .lock.yml files in .github/workflows directory
	workflowsDir := ".github/workflows"
	if _, err := os.Stat(workflowsDir); os.IsNotExist(err) {
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage("No .github/workflows directory found"))
		}
		return workflowNames, nil
	}

	files, err := filepath.Glob(filepath.Join(workflowsDir, "*.lock.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob .lock.yml files: %w", err)
	}

	logsUtilsLog.Printf("Found %d .lock.yml file(s) in %s", len(files), workflowsDir)

	for _, file := range files {
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Reading workflow file: "+file))
		}

		content, err := os.ReadFile(file)
		if err != nil {
			if verbose {
				fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to read %s: %v", file, err)))
			}
			continue
		}

		// Extract the workflow name using simple string parsing
		lines := strings.SplitSeq(string(content), "\n")
		for line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "name:") {
				// Parse the name field
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[1])
					// Remove quotes if present
					name = strings.Trim(name, `"'`)
					if name != "" {
						workflowNames = append(workflowNames, name)
						logsUtilsLog.Printf("Discovered workflow name: %s (from %s)", name, file)
						if verbose {
							fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Found agentic workflow: "+name))
						}
						break
					}
				}
			}
		}
	}

	if verbose {
		fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("Found %d agentic workflows", len(workflowNames))))
	}

	return workflowNames, nil
}

// resolveExcludeWorkflows resolves a list of workflow names (IDs or display names) to their
// canonical display names for matching against WorkflowRun.WorkflowName.
// If a name cannot be resolved (e.g., no .lock.yml files present or the workflow is from another
// repo), the raw value is kept for case-insensitive matching.
func resolveExcludeWorkflows(excludes []string, verbose bool) []string {
	if len(excludes) == 0 {
		return nil
	}

	resolved := make([]string, 0, len(excludes))
	for _, name := range excludes {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		// Try to resolve to canonical display name via lock files.
		displayName, err := workflow.FindWorkflowName(name)
		if err == nil && displayName != "" {
			logsUtilsLog.Printf("Resolved exclude workflow '%s' -> '%s'", name, displayName)
			resolved = append(resolved, displayName)
		} else {
			// Cannot resolve - keep the raw value for case-insensitive matching.
			logsUtilsLog.Printf("Could not resolve exclude workflow '%s' (using raw value for matching)", name)
			if verbose {
				fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("Exclude: using '%s' for matching (could not resolve to a known workflow)", name)))
			}
			resolved = append(resolved, name)
		}
	}
	return resolved
}

// isWorkflowExcluded reports whether a workflow display name matches any entry in the exclude list.
// Matching is case-insensitive to be resilient to capitalization differences.
func isWorkflowExcluded(workflowName string, excludes []string) bool {
	lowerName := strings.ToLower(workflowName)
	for _, ex := range excludes {
		if strings.ToLower(ex) == lowerName {
			return true
		}
	}
	return false
}
