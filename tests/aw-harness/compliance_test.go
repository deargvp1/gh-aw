//go:build !integration

// Package awharness contains compliance test stubs for the AW Harness specification
// (specs/aw-harness.md, Section 12). These tests are stubs that document the expected
// behavior once aw_harness.cjs is implemented.
//
// Current implementation status: aw_harness.cjs does not yet exist. All tests in this
// package are marked as pending via t.Skip() until the implementation is complete.
// See specs/aw-harness.md (Status of This Document) for implementation status tracking.
//
// Once aw_harness.cjs is implemented, remove the t.Skip() calls and supply real
// fixture config.json / prompt.txt files via testdata/.
package awharness_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const awHarnessPath = "../../actions/setup/js/aw_harness.cjs"

// harnessExists reports whether aw_harness.cjs has been built and placed in
// actions/setup/js/. All tests skip when the file is absent.
func harnessExists() bool {
	_, err := os.Stat(awHarnessPath)
	return err == nil
}

// TestT_AW_001_HarnessInvocationContract validates the basic invocation contract.
//
// Spec: T-AW-001 (specs/aw-harness.md, Section 12)
// Precondition: Valid config.json and prompt.txt at known paths; Pi SDK loadable;
//
//	at least one LLM provider credential in the environment.
//
// Stimulus: node aw_harness.cjs --config <path> --prompt <path>
// Expected: Exits with code 0; session_start and session_end JSONL events on stderr;
//
//	step summary contains valid Markdown if $GITHUB_STEP_SUMMARY is set.
func TestT_AW_001_HarnessInvocationContract(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	configPath := filepath.Join("testdata", "valid-config.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture config.json must exist")
	require.FileExists(t, promptPath, "test fixture prompt.txt must exist")

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "harness should exit with code 0 on success")
	assert.Contains(t, string(output), `"event":"session_start"`,
		"stderr must contain a session_start JSONL event")
	assert.Contains(t, string(output), `"event":"session_end"`,
		"stderr must contain a session_end JSONL event")
}

// TestT_AW_002_ExtensionLoading validates extension loading with partial failure.
//
// Spec: T-AW-002 (specs/aw-harness.md, Section 12)
// Precondition: config.json references one valid extension and one invalid (missing) path;
//
//	harness.extensions-required is false (default).
//
// Stimulus: Invoke harness with this configuration.
// Expected: Valid extension loaded; missing extension triggers stderr warning;
//
//	harness does NOT abort; exit code 0.
func TestT_AW_002_ExtensionLoading(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	configPath := filepath.Join("testdata", "config-with-missing-extension.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "harness should not abort on missing optional extension (exit code 0)")
	assert.Contains(t, string(output), "warning",
		"stderr must contain a warning about the missing extension")
}

// TestT_AW_003_BudgetGate validates that the harness enforces the hard token budget.
//
// Spec: T-AW-003 (specs/aw-harness.md, Section 12)
// Precondition: config.json sets harness.budget.max-effective-tokens to a very low
//
//	value so budget is exceeded immediately.
//
// Stimulus: Invoke harness; agent session begins consuming tokens.
// Expected: budget_exceeded JSONL event emitted; harness exits with code 1.
func TestT_AW_003_BudgetGate(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	configPath := filepath.Join("testdata", "config-budget-1-token.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "harness must exit non-zero when budget is exceeded")
	assert.Equal(t, 1, exitErr.ExitCode(), "harness must exit with code 1 on budget exhaustion")
	assert.Contains(t, string(output), `"event":"budget_exceeded"`,
		"stderr must contain a budget_exceeded JSONL event")
}

// TestT_AW_004_ModelResolution validates that the provider-setup extension resolves
// model aliases using credentials from the environment.
//
// Spec: T-AW-004 (specs/aw-harness.md, Section 12)
// Precondition: config.json specifies a model alias; ANTHROPIC_API_KEY (or stub) is set.
// Stimulus: Start the harness; AgentSession is created.
// Expected: Provider registered without hard-coded URLs; session starts without error.
func TestT_AW_004_ModelResolution(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" && os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("no LLM provider credential in environment; set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GITHUB_TOKEN")
	}

	configPath := filepath.Join("testdata", "config-with-model-alias.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "harness should start without a provider resolution error")
	// Ensure no hard-coded API key appears in JSONL output
	assert.NotContains(t, string(output), "sk-ant-",
		"Anthropic API key MUST NOT appear in JSONL output")
	assert.NotContains(t, string(output), "sk-",
		"OpenAI API key MUST NOT appear in JSONL output")
}

// TestT_AW_005_SessionTermination validates clean exit after normal session completion.
//
// Spec: T-AW-005 (specs/aw-harness.md, Section 12)
// Precondition: Harness running; Pi SDK session completes normally.
// Stimulus: AgentSession reaches natural end.
// Expected: context-provenance.jsonl written; session_end emitted; exit code 0.
func TestT_AW_005_SessionTermination(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	configPath := filepath.Join("testdata", "valid-config.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	// Set GITHUB_STEP_SUMMARY to capture step summary output
	summaryFile := filepath.Join(t.TempDir(), "step-summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "harness must exit with code 0 on normal session completion")
	assert.Contains(t, string(output), `"event":"session_end"`,
		"harness must emit session_end JSONL event")

	// Validate context-provenance.jsonl was written
	provenancePath := "/tmp/gh-aw/context-provenance.jsonl"
	assert.FileExists(t, provenancePath, "harness must write context-provenance.jsonl on session end")

	// Validate step summary was written
	summaryContent, readErr := os.ReadFile(summaryFile)
	require.NoError(t, readErr, "step summary file must be readable")
	assert.Contains(t, string(summaryContent), "AW Harness Run",
		"step summary must contain a header identifying the workflow and model")
}

// TestT_AW_006_PiSDKFailureToLoad validates that the harness handles a missing Pi SDK gracefully.
//
// Spec: T-AW-006 (specs/aw-harness.md, Section 12)
// Precondition: Pi SDK is unavailable (bundle corrupted or require path wrong).
// Stimulus: Invoke the harness.
// Expected: Structured JSONL error event on stderr; human-readable error to GITHUB_STEP_SUMMARY;
//
//	exit code 2; no AgentSession created.
func TestT_AW_006_PiSDKFailureToLoad(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	// Use a config that forces a bad Pi SDK require path
	configPath := filepath.Join("testdata", "config-bad-sdk-path.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	summaryFile := filepath.Join(t.TempDir(), "step-summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "harness must exit non-zero when Pi SDK fails to load")
	assert.Equal(t, 2, exitErr.ExitCode(),
		"harness must exit with code 2 (invocation error) for SDK load failure, not code 1 (session failure)")
	assert.Contains(t, string(output), `"event":"error"`,
		"stderr must contain a structured JSONL error event identifying the SDK load failure")

	summaryContent, readErr := os.ReadFile(summaryFile)
	require.NoError(t, readErr, "step summary must be written even on SDK failure")
	assert.Contains(t, string(summaryContent), "Pi SDK",
		"step summary must identify the failed module")
}

// TestT_AW_007_ExtensionCrashIsolation validates that a crashing user extension is isolated.
//
// Spec: T-AW-007 (specs/aw-harness.md, Section 12)
// Precondition: config.json references a user extension that throws on initialization;
//
//	harness.extensions-required is false.
//
// Stimulus: Invoke the harness.
// Expected: Extension crash does not abort harness; warning on stderr; session proceeds;
//
//	exit code 0 (if session completes successfully without the crashed extension).
func TestT_AW_007_ExtensionCrashIsolation(t *testing.T) {
	if !harnessExists() {
		t.Skip("aw_harness.cjs not yet implemented — see specs/aw-harness.md Implementation Status")
	}

	configPath := filepath.Join("testdata", "config-with-crashing-extension.json")
	promptPath := filepath.Join("testdata", "simple-prompt.txt")
	require.FileExists(t, configPath, "test fixture must exist")

	cmd := exec.Command("node", awHarnessPath, "--config", configPath, "--prompt", promptPath)
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err,
		"harness must NOT exit non-zero due to a user extension crash alone (exit code 0)")
	assert.Contains(t, string(output), "warning",
		"stderr must contain a warning identifying the crashing extension and its error")
}
