//go:build !integration

package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPluginsToSharedImportCodemod(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	assert.Equal(t, "plugins-to-shared-import", codemod.ID)
	assert.Equal(t, "Migrate plugins to shared Copilot plugins import", codemod.Name)
	assert.NotEmpty(t, codemod.Description)
	assert.Equal(t, "1.0.0", codemod.IntroducedIn)
	require.NotNil(t, codemod.Apply)
}

func TestPluginsToSharedImportCodemod_NoPlugins(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	content := `---
on: workflow_dispatch
engine: copilot
---

# No plugins`

	frontmatter := map[string]any{
		"on":     "workflow_dispatch",
		"engine": "copilot",
	}

	result, applied, err := codemod.Apply(content, frontmatter)

	require.NoError(t, err)
	assert.False(t, applied, "Codemod should not be applied when plugins is absent")
	assert.Equal(t, content, result, "Content should not be modified")
}

func TestPluginsToSharedImportCodemod_ArrayFormat(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	content := `---
on:
  issues:
    types: [opened]
engine: copilot
plugins:
  - github/test-plugin
  - acme/custom-tools
---

# Test workflow`

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{"types": []any{"opened"}},
		},
		"engine":  "copilot",
		"plugins": []any{"github/test-plugin", "acme/custom-tools"},
	}

	result, applied, err := codemod.Apply(content, frontmatter)

	require.NoError(t, err)
	assert.True(t, applied, "Codemod should have been applied")
	assertNoTopLevelPluginsKey(t, result)
	assert.Contains(t, result, "imports:", "imports key should be present")
	assert.Contains(t, result, "- uses: shared/copilot-plugins.md", "shared workflow import should be present")
	assert.Contains(t, result, "with:\n      plugins:", "nested plugins input under with should be preserved")
	assert.Contains(t, result, "plugins: [\"github/test-plugin\", \"acme/custom-tools\"]", "plugin list should be preserved")
}

func TestPluginsToSharedImportCodemod_ObjectFormat(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	content := `---
on:
  issues:
    types: [opened]
engine: copilot
plugins:
  repos:
    - github/test-plugin
    - acme/custom-tools
  github-token: ${{ secrets.MY_TOKEN }}
---

# Test workflow`

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{"types": []any{"opened"}},
		},
		"engine": "copilot",
		"plugins": map[string]any{
			"repos":        []any{"github/test-plugin", "acme/custom-tools"},
			"github-token": "${{ secrets.MY_TOKEN }}",
		},
	}

	result, applied, err := codemod.Apply(content, frontmatter)

	require.NoError(t, err)
	assert.True(t, applied, "Codemod should have been applied")
	assertNoTopLevelPluginsKey(t, result)
	assert.Contains(t, result, "imports:", "imports key should be present")
	assert.Contains(t, result, "- uses: shared/copilot-plugins.md", "shared workflow import should be present")
	assert.Contains(t, result, "with:\n      plugins:", "nested plugins input under with should be preserved")
	assert.Contains(t, result, "plugins: [\"github/test-plugin\", \"acme/custom-tools\"]", "repos list should map to plugins input")
	assert.Contains(t, result, "github-token: ${{ secrets.MY_TOKEN }}", "github-token should be preserved")
}

func TestPluginsToSharedImportCodemod_RemovesPluginsWhenImportAlreadyExists(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	content := `---
engine: copilot
imports:
  - uses: shared/copilot-plugins.md
    with:
      plugins: ["github/test-plugin"]
plugins:
  - github/test-plugin
---

# Test workflow`

	frontmatter := map[string]any{
		"engine": "copilot",
		"imports": []any{
			map[string]any{
				"uses": "shared/copilot-plugins.md",
				"with": map[string]any{
					"plugins": []any{"github/test-plugin"},
				},
			},
		},
		"plugins": []any{"github/test-plugin"},
	}

	result, applied, err := codemod.Apply(content, frontmatter)

	require.NoError(t, err)
	assert.True(t, applied, "Codemod should have been applied")
	assertNoTopLevelPluginsKey(t, result)
	assert.Equal(t, 1, strings.Count(result, "shared/copilot-plugins.md"), "Codemod should not add duplicate imports")
}

func TestHasCopilotPluginsSharedImport_AcceptsExtensionlessPath(t *testing.T) {
	frontmatter := map[string]any{
		"imports": []any{
			"shared/copilot-plugins",
		},
	}

	assert.True(t, hasCopilotPluginsSharedImport(frontmatter), "Extensionless shared/copilot-plugins import should be detected")
}

func assertNoTopLevelPluginsKey(t *testing.T, content string) {
	t.Helper()
	assert.NotRegexp(t, `(?m)^plugins:`, content, "top-level plugins key should be removed")
}

func TestPluginsToSharedImportCodemod_PreservesMarkdownBody(t *testing.T) {
	codemod := getPluginsToSharedImportCodemod()

	content := `---
engine: copilot
plugins:
  - github/test-plugin
---

# My workflow

Install the plugin and do work.`

	frontmatter := map[string]any{
		"engine":  "copilot",
		"plugins": []any{"github/test-plugin"},
	}

	result, applied, err := codemod.Apply(content, frontmatter)

	require.NoError(t, err)
	assert.True(t, applied, "Codemod should have been applied")
	assert.Contains(t, result, "# My workflow", "Markdown body should be preserved")
	assert.Contains(t, result, "Install the plugin and do work.", "Markdown body should be preserved")
}
