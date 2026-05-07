---
# Standard Repo-Memory Configuration
# Provides a standardized repo-memory setup for workflows that need historical data persistence.
#
# Usage:
#   imports:
#     - uses: shared/repo-memory-standard.md
#       with:
#         branch-name: "memory/my-workflow"
#         description: "Historical my-workflow analysis results"
#
# Optional overrides:
#   - file-glob: ["*.json", "*.md"]      # custom file patterns (default: JSON/JSONL/CSV/MD)
#   - max-file-size: 204800              # custom file size limit (default: 100KB)
#   - max-patch-size: 51200             # custom patch size limit (default: 10KB)

import-schema:
  branch-name:
    type: string
    required: true
    description: "Branch name for repo-memory storage (e.g. memory/my-workflow)"
  description:
    type: string
    required: true
    description: "Human-readable description of what is stored"
  file-glob:
    type: array
    required: false
    default:
      - "*.json"
      - "*.jsonl"
      - "*.csv"
      - "*.md"
    description: "File glob patterns for allowed files (default: JSON, JSONL, CSV, MD files)"
    items:
      type: string
  max-file-size:
    type: integer
    default: 102400
    description: "Max file size in bytes (default: 100KB)"
  max-patch-size:
    type: integer
    default: 10240
    description: "Max total patch size in bytes per push (default: 10KB, max: 100KB)"

tools:
  repo-memory:
    branch-name: ${{ github.aw.import-inputs.branch-name }}
    description: ${{ github.aw.import-inputs.description }}
    file-glob: ${{ github.aw.import-inputs.file-glob }}
    max-file-size: ${{ github.aw.import-inputs.max-file-size }}
    max-patch-size: ${{ github.aw.import-inputs.max-patch-size }}
---
