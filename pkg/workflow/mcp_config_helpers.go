package workflow

import (
	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var mcpConfigHelpersLog = logger.New("workflow:mcp_config_helpers")

// getWellKnownContainer returns the appropriate container configuration for well-known commands.
// This enables automatic containerization of stdio MCP servers based on their command.
func getWellKnownContainer(command string) *WellKnownContainer {
	wellKnownContainers := map[string]*WellKnownContainer{
		"npx": {
			Image:      constants.DefaultNodeAlpineLTSImage,
			Entrypoint: "npx",
		},
		"uvx": {
			Image:      constants.DefaultPythonAlpineLTSImage,
			Entrypoint: "uvx",
		},
	}

	container := wellKnownContainers[command]
	if container != nil {
		mcpConfigHelpersLog.Printf("Found well-known container for command: command=%s, image=%s", command, container.Image)
	} else {
		mcpConfigHelpersLog.Printf("No well-known container found for command: %s", command)
	}
	return container
}
