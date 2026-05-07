// This file provides validation for the runs-on field in agentic workflows.
//
// # Runner Type Validation
//
// This file validates that the runs-on field in workflow frontmatter does not
// specify runner types that are incompatible with agentic workflows. Specifically,
// macOS runners require Docker to be available for the Agent Workflow Firewall
// containers. On GitHub-hosted macOS runners Docker can be provided via Colima
// (a lightweight macOS Docker runtime); the generated workflow installs it
// automatically before pulling AWF container images.
//
// # Validation Functions
//
//   - validateRunsOn() - Validates the runs-on field for unsupported runner types
//   - extractRunnerLabels() - Extracts individual runner labels from runs-on value
//
// # When to Add Validation Here
//
// Add validation to this file when:
//   - Adding new runner type restrictions
//   - Detecting additional unsupported runner configurations
//   - Improving error messages for runner selection

package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/gh-aw/pkg/console"
)

var runsOnValidationLog = newValidationLogger("runs_on")

// macOSRunnerFAQURL is the URL to the FAQ entry explaining macOS runner support via Colima.
const macOSRunnerFAQURL = "https://github.github.com/gh-aw/reference/faq/#why-are-macos-runners-not-supported"

// validateRunsOn validates the runs-on field for runner types that require
// special setup in agentic workflows. macOS runners are supported via Docker
// installed through Colima; a warning is emitted to remind authors of this
// dependency. All other runner types are allowed without restriction.
//
// Returns nil in all cases (macOS support is allowed with a warning).
func validateRunsOn(frontmatter map[string]any, markdownPath string) error {
	runsOn, exists := frontmatter["runs-on"]
	if !exists {
		return nil
	}

	runsOnValidationLog.Printf("Validating runs-on configuration")

	labels := extractRunnerLabels(runsOn)
	for _, label := range labels {
		lower := strings.ToLower(label)
		if strings.HasPrefix(lower, "macos-") || lower == "macos" {
			warningMsg := fmt.Sprintf(
				"runner '%s' requires Docker to be available for the AWF containers. "+
					"On GitHub-hosted macOS runners Docker is installed automatically via Colima "+
					"(install_docker_macos.sh) before image downloads. "+
					"Network firewalling runs inside the Colima Linux VM where iptables is available. "+
					"See %s for details.",
				label, macOSRunnerFAQURL)
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(warningMsg))
			runsOnValidationLog.Printf("macOS runner detected with warning: %s", label)
			return nil
		}
	}

	runsOnValidationLog.Printf("runs-on validation passed")
	return nil
}

// extractRunnerLabels extracts individual runner label strings from a runs-on value.
// Handles all supported GitHub Actions runs-on forms:
//   - string: "ubuntu-latest"
//   - array: ["self-hosted", "linux"]
//   - object with labels: {group: "...", labels: ["linux"]}
func extractRunnerLabels(runsOn any) []string {
	var labels []string

	switch v := runsOn.(type) {
	case string:
		labels = append(labels, v)
	case []any:
		labels = parseStringSliceAny(v, nil)
	case map[string]any:
		if labelsVal, ok := v["labels"]; ok {
			labels = parseStringSliceAny(labelsVal, nil)
		}
	}

	return labels
}
