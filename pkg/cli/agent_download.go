package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/workflow"
)

var agentDownloadLog = logger.New("cli:agent_download")

// downloadAgentFileFromGitHub downloads the agentic-workflows dispatcher file from GitHub
func downloadAgentFileFromGitHub(verbose bool) (string, error) {
	agentDownloadLog.Print("Downloading agentic-workflows dispatcher from GitHub")

	// Determine the ref to use (tag for releases, main for dev builds)
	ref := "main"
	if workflow.IsRelease() {
		ref = GetVersion()
		agentDownloadLog.Printf("Using release tag: %s", ref)
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Using release version: "+ref))
		}
	} else {
		agentDownloadLog.Print("Using main branch for dev build")
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Using main branch (dev build)"))
		}
	}

	agentPaths := []string{
		".agents/agents/agentic-workflows.md",
		".github/agents/agentic-workflows.agent.md",
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for _, agentPath := range agentPaths {
		// Construct the raw GitHub URL
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/github/gh-aw/%s/%s", ref, agentPath)
		agentDownloadLog.Printf("Downloading from URL: %s", rawURL)

		// Download the file
		resp, err := client.Get(rawURL)
		if err != nil {
			return "", fmt.Errorf("failed to download agent file: %w", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			discardAndClose(resp)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			discardAndClose(resp)
			return "", fmt.Errorf("failed to download agent file: HTTP %d", resp.StatusCode)
		}

		// Read the content
		content, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read agent file content: %w", err)
		}

		contentStr := string(content)

		// Patch URLs to match the current version/ref
		patchedContent := patchAgentFileURLs(contentStr, ref)
		if patchedContent != contentStr && verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Patched URLs to use ref: "+ref))
		}

		agentDownloadLog.Printf("Successfully downloaded agent file (%d bytes)", len(patchedContent))
		return patchedContent, nil
	}

	// Fall back to gh CLI for authenticated access (e.g., private repos in codespaces)
	if isGHCLIAvailable() {
		agentDownloadLog.Print("Unauthenticated download returned 404, trying gh CLI for authenticated access")
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Retrying download with gh CLI authentication..."))
		}
		if ghContent, err := downloadAgentFileViaGHCLI(ref, agentPaths); err == nil {
			patchedContent := patchAgentFileURLs(ghContent, ref)
			agentDownloadLog.Printf("Successfully downloaded agent file via gh CLI (%d bytes)", len(patchedContent))
			return patchedContent, nil
		} else {
			agentDownloadLog.Printf("gh CLI fallback failed: %v", err)
		}
	}

	return "", errors.New("failed to download agent file from all known dispatcher paths")
}

// patchAgentFileURLs patches URLs in the agent file to use the correct ref
func patchAgentFileURLs(content, ref string) string {
	// Pattern 1: Convert local paths to GitHub URLs
	// `.github/aw/file.md` -> `https://github.com/github/gh-aw/blob/{ref}/.github/aw/file.md`
	content = strings.ReplaceAll(content, "`.github/aw/", fmt.Sprintf("`https://github.com/github/gh-aw/blob/%s/.github/aw/", ref))

	// Pattern 2: Update existing GitHub URLs to use the correct ref
	// https://github.com/github/gh-aw/blob/main/ -> https://github.com/github/gh-aw/blob/{ref}/
	if ref != "main" {
		content = strings.ReplaceAll(content, "/blob/main/", fmt.Sprintf("/blob/%s/", ref))
	}

	return content
}

// downloadAgentFileViaGHCLI downloads the agent file using the gh CLI with authentication.
// This is used as a fallback when the unauthenticated raw.githubusercontent.com download fails
// (e.g., for private repositories accessed from codespaces).
func downloadAgentFileViaGHCLI(ref string, agentPaths []string) (string, error) {
	for _, agentPath := range agentPaths {
		output, err := workflow.RunGH("Downloading agent file...", "api",
			"/repos/github/gh-aw/contents/"+agentPath+"?ref="+url.QueryEscape(ref),
			"--header", "Accept: application/vnd.github.raw")
		if err == nil {
			return string(output), nil
		}
	}

	return "", errors.New("gh api download failed for all known dispatcher paths")
}

func isGHCLIAvailable() bool {
	cmd := exec.Command("gh", "--version")
	return cmd.Run() == nil
}

// discardAndClose drains and closes an HTTP response body safely, including nil responses.
func discardAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
