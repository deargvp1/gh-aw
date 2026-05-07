//go:build !integration

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRunsOn(t *testing.T) {
	tests := []struct {
		name        string
		frontmatter map[string]any
		wantErr     bool
		errorInMsg  string
		description string
	}{
		{
			name:        "no runs-on field",
			frontmatter: map[string]any{},
			wantErr:     false,
			description: "Missing runs-on should pass validation",
		},
		{
			name:        "ubuntu-latest string",
			frontmatter: map[string]any{"runs-on": "ubuntu-latest"},
			wantErr:     false,
			description: "ubuntu-latest should be allowed",
		},
		{
			name:        "windows-latest string",
			frontmatter: map[string]any{"runs-on": "windows-latest"},
			wantErr:     false,
			description: "windows-latest should be allowed",
		},
		{
			name:        "self-hosted string",
			frontmatter: map[string]any{"runs-on": "self-hosted"},
			wantErr:     false,
			description: "self-hosted should be allowed",
		},
		{
			name:        "macos-latest string",
			frontmatter: map[string]any{"runs-on": "macos-latest"},
			wantErr:     false,
			description: "macos-latest is allowed (Docker provided via Colima)",
		},
		{
			name:        "macos-14 string",
			frontmatter: map[string]any{"runs-on": "macos-14"},
			wantErr:     false,
			description: "macos-14 is allowed (Docker provided via Colima)",
		},
		{
			name:        "macos-13 string",
			frontmatter: map[string]any{"runs-on": "macos-13"},
			wantErr:     false,
			description: "macos-13 is allowed (Docker provided via Colima)",
		},
		{
			name:        "bare macos string",
			frontmatter: map[string]any{"runs-on": "macos"},
			wantErr:     false,
			description: "bare 'macos' runner label is allowed (Docker provided via Colima)",
		},
		{
			name:        "ubuntu array",
			frontmatter: map[string]any{"runs-on": []any{"self-hosted", "linux"}},
			wantErr:     false,
			description: "Array with linux runners should be allowed",
		},
		{
			name:        "macos in array",
			frontmatter: map[string]any{"runs-on": []any{"self-hosted", "macos-latest"}},
			wantErr:     false,
			description: "Array containing macos runner is allowed (Docker provided via Colima)",
		},
		{
			name: "object with linux labels",
			frontmatter: map[string]any{
				"runs-on": map[string]any{
					"group":  "ubuntu-runners",
					"labels": []any{"ubuntu-latest"},
				},
			},
			wantErr:     false,
			description: "Object form with linux labels should be allowed",
		},
		{
			name: "object with macos labels",
			frontmatter: map[string]any{
				"runs-on": map[string]any{
					"group":  "macos-runners",
					"labels": []any{"macos-14"},
				},
			},
			wantErr:     false,
			description: "Object form with macos labels is allowed (Docker provided via Colima)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRunsOn(tt.frontmatter, "test-workflow.md")

			if tt.wantErr {
				require.Error(t, err, "Test: %s - Expected error but got nil", tt.description)
				if tt.errorInMsg != "" {
					assert.Contains(t, err.Error(), tt.errorInMsg,
						"Error should contain '%s' for: %s", tt.errorInMsg, tt.description)
				}
			} else {
				assert.NoError(t, err, "Test: %s - Expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestExtractRunnerLabels(t *testing.T) {
	tests := []struct {
		name     string
		runsOn   any
		expected []string
	}{
		{
			name:     "string label",
			runsOn:   "ubuntu-latest",
			expected: []string{"ubuntu-latest"},
		},
		{
			name:     "array of labels",
			runsOn:   []any{"self-hosted", "linux"},
			expected: []string{"self-hosted", "linux"},
		},
		{
			name: "object with labels",
			runsOn: map[string]any{
				"labels": []any{"linux", "x64"},
			},
			expected: []string{"linux", "x64"},
		},
		{
			name: "object without labels",
			runsOn: map[string]any{
				"group": "my-group",
			},
			expected: nil,
		},
		{
			name:     "nil",
			runsOn:   nil,
			expected: nil,
		},
		{
			name:     "integer (unsupported type)",
			runsOn:   42,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRunnerLabels(tt.runsOn)
			assert.Equal(t, tt.expected, result)
		})
	}
}
