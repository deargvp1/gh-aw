---
# Copilot Plugins - Shared Workflow
# Migrate legacy `plugins:` frontmatter to an importable shared workflow.
#
# This shared workflow creates a dedicated "copilot_plugins" job that packs plugin
# dependencies with microsoft/apm-action and uploads the bundle as an artifact.
# The agent job restores the bundle in jobs.agent.pre-steps (before checkout).
#
# Usage:
#   imports:
#     - uses: shared/copilot-plugins.md
#       with:
#         plugins:
#           - github/my-copilot-plugin
#           - github/awesome-copilot/plugins/context-engineering

import-schema:
  plugins:
    type: array
    items:
      type: string
    required: true
    description: >
      List of plugin package references to install.
      Format: owner/repo or owner/repo/path/to/plugin.
  github-token:
    type: string
    required: false
    description: >
      Optional GitHub token expression used by APM when fetching private plugin repositories.
      If not provided, falls back to GH_AW_PLUGINS_TOKEN, GH_AW_GITHUB_TOKEN, then GITHUB_TOKEN.

jobs:
  copilot_plugins:
    runs-on: ubuntu-slim
    needs: [activation]
    permissions: {}
    steps:
      - name: Prepare Copilot plugin list
        id: copilot_plugins_prep
        env:
          AW_COPILOT_PLUGINS: '${{ github.aw.import-inputs.plugins }}'
        run: |
          DEPS=$(echo "$AW_COPILOT_PLUGINS" | jq -r '.[] | "- " + .')
          {
            echo "deps<<COPILOTPLUGINS"
            printf '%s\n' "$DEPS"
            echo "COPILOTPLUGINS"
          } >> "$GITHUB_OUTPUT"

      - name: Pack Copilot plugins
        id: copilot_plugins_pack
        uses: microsoft/apm-action@v1.4.1
        env:
          GITHUB_TOKEN: ${{ github.aw.import-inputs.github-token || secrets.GH_AW_PLUGINS_TOKEN || secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}
        with:
          dependencies: ${{ steps.copilot_plugins_prep.outputs.deps }}
          isolated: 'true'
          pack: 'true'
          archive: 'true'
          target: copilot
          working-directory: /tmp/gh-aw/apm-workspace

      - name: Upload Copilot plugin bundle artifact
        if: success()
        uses: actions/upload-artifact@v7.0.1
        with:
          name: ${{ needs.activation.outputs.artifact_prefix }}copilot-plugins
          path: ${{ steps.copilot_plugins_pack.outputs.bundle-path }}
          retention-days: '1'

  agent:
    pre-steps:
      - name: Download Copilot plugin bundle artifact
        uses: actions/download-artifact@v8.0.1
        with:
          name: ${{ needs.activation.outputs.artifact_prefix }}copilot-plugins
          path: /tmp/gh-aw/copilot-plugins-bundle

      - name: Find Copilot plugin bundle path
        id: copilot_plugins_bundle
        run: echo "path=$(find /tmp/gh-aw/copilot-plugins-bundle -name '*.tar.gz' | head -1)" >> "$GITHUB_OUTPUT"

      - name: Restore Copilot plugins before checkout
        uses: microsoft/apm-action@v1.4.1
        with:
          bundle: ${{ steps.copilot_plugins_bundle.outputs.path }}
---
