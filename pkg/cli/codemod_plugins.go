package cli

import (
	"slices"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/sliceutil"
)

var pluginsCodemodLog = logger.New("cli:codemod_plugins")

// getPluginsToSharedImportCodemod creates a codemod that migrates the removed
// top-level `plugins:` field to an equivalent imports entry that uses
// `shared/copilot-plugins.md`.
func getPluginsToSharedImportCodemod() Codemod {
	return Codemod{
		ID:           "plugins-to-shared-import",
		Name:         "Migrate plugins to shared Copilot plugins import",
		Description:  "Removes top-level 'plugins' and adds an equivalent 'imports' entry using shared/copilot-plugins.md.",
		IntroducedIn: "1.0.0",
		Apply: func(content string, frontmatter map[string]any) (string, bool, error) {
			pluginsAny, hasPlugins := frontmatter["plugins"]
			if !hasPlugins {
				return content, false, nil
			}

			plugins, githubToken, ok := extractPluginsMigrationConfig(pluginsAny)
			if !ok || len(plugins) == 0 {
				pluginsCodemodLog.Print("Found plugins field but it was empty or invalid - skipping migration")
				return content, false, nil
			}

			alreadyImported := hasCopilotPluginsSharedImport(frontmatter)

			newContent, applied, err := applyFrontmatterLineTransform(content, func(lines []string) ([]string, bool) {
				result, modified := removeTopLevelBlock(lines, "plugins")
				if !modified {
					return lines, false
				}

				if alreadyImported {
					return result, true
				}

				return addCopilotPluginsImport(result, plugins, githubToken), true
			})
			if applied {
				if alreadyImported {
					pluginsCodemodLog.Print("Removed plugins field (shared/copilot-plugins.md import already present)")
				} else {
					pluginsCodemodLog.Printf("Migrated plugins to shared/copilot-plugins.md import with %d plugin(s)", len(plugins))
				}
			}
			return newContent, applied, err
		},
	}
}

func extractPluginsMigrationConfig(pluginsAny any) ([]string, string, bool) {
	switch plugins := pluginsAny.(type) {
	case []string:
		return sliceutil.Deduplicate(plugins), "", len(plugins) > 0
	case []any:
		values := make([]string, 0, len(plugins))
		for _, item := range plugins {
			value, ok := item.(string)
			if ok && strings.TrimSpace(value) != "" {
				values = append(values, value)
			}
		}
		return sliceutil.Deduplicate(values), "", len(values) > 0
	case map[string]any:
		reposAny, hasRepos := plugins["repos"]
		if !hasRepos {
			return nil, "", false
		}

		repos, _, ok := extractPluginsMigrationConfig(reposAny)
		if !ok || len(repos) == 0 {
			return nil, "", false
		}

		var githubToken string
		if tokenAny, hasToken := plugins["github-token"]; hasToken {
			if token, ok := tokenAny.(string); ok {
				githubToken = token
			}
		}

		return repos, githubToken, true
	default:
		return nil, "", false
	}
}

func hasCopilotPluginsSharedImport(frontmatter map[string]any) bool {
	importsAny, hasImports := frontmatter["imports"]
	if !hasImports {
		return false
	}

	switch imports := importsAny.(type) {
	case []string:
		return slices.ContainsFunc(imports, isCopilotPluginsImportPath)
	case []any:
		for _, entry := range imports {
			switch typed := entry.(type) {
			case string:
				if isCopilotPluginsImportPath(typed) {
					return true
				}
			case map[string]any:
				usesAny, hasUses := typed["uses"]
				if !hasUses {
					continue
				}
				uses, ok := usesAny.(string)
				if ok && isCopilotPluginsImportPath(uses) {
					return true
				}
			}
		}
	}

	return false
}

func isCopilotPluginsImportPath(path string) bool {
	trimmed := strings.TrimSpace(path)
	return trimmed == "shared/copilot-plugins.md" || trimmed == "shared/copilot-plugins"
}

func addCopilotPluginsImport(lines []string, plugins []string, githubToken string) []string {
	entry := []string{
		"  - uses: shared/copilot-plugins.md",
		"    with:",
		"      plugins: " + formatStringArrayInline(plugins),
	}
	if strings.TrimSpace(githubToken) != "" {
		entry = append(entry, "      github-token: "+githubToken)
	}

	importsIdx := -1
	importsEnd := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isTopLevelKey(line) && strings.HasPrefix(trimmed, "imports:") {
			importsIdx = i
			for j := i + 1; j < len(lines); j++ {
				if isTopLevelKey(lines[j]) {
					importsEnd = j
					break
				}
			}
			break
		}
	}

	if importsIdx >= 0 {
		result := make([]string, 0, len(lines)+len(entry))
		result = append(result, lines[:importsEnd]...)
		result = append(result, entry...)
		result = append(result, lines[importsEnd:]...)
		return result
	}

	insertAt := 0
	for i, line := range lines {
		if isTopLevelKey(line) && strings.HasPrefix(strings.TrimSpace(line), "engine:") {
			insertAt = i + 1
			break
		}
	}

	importBlock := make([]string, 0, 1+len(entry))
	importBlock = append(importBlock, "imports:")
	importBlock = append(importBlock, entry...)

	result := make([]string, 0, len(lines)+len(importBlock))
	result = append(result, lines[:insertAt]...)
	result = append(result, importBlock...)
	result = append(result, lines[insertAt:]...)
	return result
}

func removeTopLevelBlock(lines []string, blockName string) ([]string, bool) {
	blockIdx := -1
	blockEnd := len(lines)
	for i, line := range lines {
		if isTopLevelKey(line) && strings.HasPrefix(strings.TrimSpace(line), blockName+":") {
			blockIdx = i
			for j := i + 1; j < len(lines); j++ {
				if isTopLevelKey(lines[j]) {
					blockEnd = j
					break
				}
			}
			break
		}
	}

	if blockIdx == -1 {
		return lines, false
	}

	result := make([]string, 0, len(lines)-(blockEnd-blockIdx))
	result = append(result, lines[:blockIdx]...)
	result = append(result, lines[blockEnd:]...)
	return result, true
}
