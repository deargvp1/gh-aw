//go:build !integration

package cli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeBase64FileContent(t *testing.T) {
	tests := []struct {
		name     string
		input    func() string // build the raw API-style input
		expected string
		wantErr  bool
	}{
		{
			name: "plain base64 without newlines",
			input: func() string {
				return base64.StdEncoding.EncodeToString([]byte("hello world"))
			},
			expected: "hello world",
		},
		{
			name: "GitHub API style with embedded newlines every 60 chars",
			input: func() string {
				encoded := base64.StdEncoding.EncodeToString([]byte("hello world"))
				// Simulate GitHub API line-wrapping at 60 characters
				var sb strings.Builder
				for i, c := range encoded {
					if i > 0 && i%60 == 0 {
						sb.WriteByte('\n')
					}
					sb.WriteRune(c)
				}
				return sb.String()
			},
			expected: "hello world",
		},
		{
			name: "leading and trailing whitespace stripped",
			input: func() string {
				return "  " + base64.StdEncoding.EncodeToString([]byte("trim me")) + "\n"
			},
			expected: "trim me",
		},
		{
			name: "binary content round-trips correctly",
			input: func() string {
				data := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
				return base64.StdEncoding.EncodeToString(data)
			},
			expected: string([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE}),
		},
		{
			name:    "invalid base64 returns error",
			input:   func() string { return "!!!not-valid-base64!!!" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeBase64FileContent(tt.input())
			if tt.wantErr {
				assert.Error(t, err, "expected an error for invalid base64 input")
				return
			}
			require.NoError(t, err, "unexpected error decoding base64 content")
			assert.Equal(t, tt.expected, string(got), "decoded content should match expected")
		})
	}
}

func TestWorkflowContentCandidatePaths(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "short workflow name tries common workflow directories",
			path:     "test-workflow.md",
			expected: []string{"test-workflow.md", "workflows/test-workflow.md", ".github/workflows/test-workflow.md"},
		},
		{
			name:     "root workflows path falls back to github workflows path",
			path:     "workflows/test-workflow.md",
			expected: []string{"workflows/test-workflow.md", ".github/workflows/test-workflow.md"},
		},
		{
			name:     "github workflows path falls back to root workflows path",
			path:     ".github/workflows/test-workflow.md",
			expected: []string{".github/workflows/test-workflow.md", "workflows/test-workflow.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, workflowContentCandidatePaths(tt.path), "candidate paths should match")
		})
	}
}

func TestDownloadWorkflowContent_TriesAlternateWorkflowDirectory(t *testing.T) {
	originalAPIFn := downloadWorkflowContentAPIFn
	defer func() {
		downloadWorkflowContentAPIFn = originalAPIFn
	}()

	encodedContent := base64.StdEncoding.EncodeToString([]byte("workflow content"))
	var requestedPaths []string
	downloadWorkflowContentAPIFn = func(_ context.Context, repo, path, ref string) ([]byte, error) {
		assert.Equal(t, "githubnext/agentic-ops", repo, "repo should be preserved")
		assert.Equal(t, "main", ref, "ref should be preserved")
		requestedPaths = append(requestedPaths, path)

		if path == "workflows/copilot-token-audit.md" {
			return nil, errors.New("HTTP 404: Not Found")
		}
		if path == ".github/workflows/copilot-token-audit.md" {
			return []byte(encodedContent), nil
		}

		return nil, fmt.Errorf("unexpected path: %s", path)
	}

	content, err := downloadWorkflowContent(context.Background(), "githubnext/agentic-ops", "workflows/copilot-token-audit.md", "main", false)
	require.NoError(t, err, "alternate workflow directory should be tried after a 404")
	assert.Equal(t, []byte("workflow content"), content, "downloaded content should match decoded workflow")
	assert.Equal(t,
		[]string{"workflows/copilot-token-audit.md", ".github/workflows/copilot-token-audit.md"},
		requestedPaths,
		"download should try the alternate workflow directory when the original path is missing",
	)
}
