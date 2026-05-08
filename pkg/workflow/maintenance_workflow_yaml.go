package workflow

import (
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var maintenanceWorkflowYAMLLog = logger.New("workflow:maintenance_workflow_yaml")

// maintenanceWorkflowParams holds all pre-resolved parameters for building maintenance workflow sections.
// It is created once in buildMaintenanceWorkflowYAML and passed to each section builder.
type maintenanceWorkflowParams struct {
	cronSchedule        string
	scheduleDesc        string
	minExpiresDays      int
	runsOnValue         string
	actionMode          ActionMode
	version             string
	actionTag           string
	resolver            ActionSHAResolver
	configuredRunsOn    RunsOnValue
	defaultBranch       string
	disableLabelTrigger bool
	// setupActionRef is pre-resolved to avoid calling ResolveSetupActionReference multiple times.
	setupActionRef string
}

// buildMaintenanceWorkflowYAML generates the complete YAML content for the
// agentics-maintenance.yml workflow. It is called by GenerateMaintenanceWorkflow
// after the cron schedule and setup parameters have been resolved.
func buildMaintenanceWorkflowYAML(
	cronSchedule, scheduleDesc string,
	minExpiresDays int,
	runsOnValue string,
	actionMode ActionMode,
	version, actionTag string,
	resolver ActionSHAResolver,
	configuredRunsOn RunsOnValue,
	defaultBranch string,
	disableLabelTrigger bool,
) string {
	maintenanceWorkflowYAMLLog.Printf("Building maintenance workflow YAML: actionMode=%s minExpiresDays=%d cronSchedule=%q defaultBranch=%q disableLabelTrigger=%v", actionMode, minExpiresDays, cronSchedule, defaultBranch, disableLabelTrigger)

	p := maintenanceWorkflowParams{
		cronSchedule:        cronSchedule,
		scheduleDesc:        scheduleDesc,
		minExpiresDays:      minExpiresDays,
		runsOnValue:         runsOnValue,
		actionMode:          actionMode,
		version:             version,
		actionTag:           actionTag,
		resolver:            resolver,
		configuredRunsOn:    configuredRunsOn,
		defaultBranch:       defaultBranch,
		disableLabelTrigger: disableLabelTrigger,
		setupActionRef:      ResolveSetupActionReference(actionMode, version, actionTag, resolver),
	}

	customInstructions := `Alternative regeneration methods:
  make recompile

Or use the gh-aw CLI directly:
  ./gh-aw compile --validate --verbose

The workflow is generated when any workflow uses the 'expires' field
in create-discussions, create-issues, or create-pull-request safe-outputs configuration.
Schedule frequency is automatically determined by the shortest expiration time.`

	header := GenerateWorkflowHeader("", "pkg/workflow/maintenance_workflow.go", customInstructions)

	var yaml strings.Builder
	yaml.WriteString(header)
	yaml.WriteString(buildMaintenanceWorkflowOnSection(p))
	yaml.WriteString(buildMaintenanceCloseExpiredEntitiesJob(p))
	yaml.WriteString(buildMaintenanceCleanupCacheMemoryJob(p))
	yaml.WriteString(buildMaintenanceRunOperationJob(p))
	yaml.WriteString(buildMaintenanceUpdatePullRequestBranchesJob(p))
	yaml.WriteString(buildMaintenanceApplySafeOutputsJob(p))
	yaml.WriteString(buildMaintenanceCreateLabelsJob(p))
	yaml.WriteString(buildMaintenanceActivityReportJob(p))
	yaml.WriteString(buildMaintenanceCloseAgenticWorkflowsIssuesJob(p))
	yaml.WriteString(buildMaintenanceValidateWorkflowsJob(p))
	if !disableLabelTrigger {
		yaml.WriteString(buildMaintenanceLabelTriggerJobs(p))
	}
	if actionMode == ActionModeDev {
		maintenanceWorkflowYAMLLog.Printf("Adding dev-only jobs: compile-workflows and secret-validation")
		yaml.WriteString(buildMaintenanceDevModeJobs(p))
	}
	return yaml.String()
}
