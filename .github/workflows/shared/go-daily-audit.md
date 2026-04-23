---
# Go Daily Audit Base — bundles daily-audit-base + go-source-analysis + go-make
# for Go code quality audit workflows.
#
# Usage:
#   imports:
#     - uses: shared/go-daily-audit.md
#       with:
#         title-prefix: "[my-workflow] "
#         expires: "1d"      # optional, default: 3d

import-schema:
  title-prefix:
    type: string
    required: true
    description: "Title prefix for created discussions, e.g. '[daily-report] '"
  expires:
    type: string
    default: "3d"
    description: "How long to keep discussions before expiry (e.g. 1d, 3d, 7d)"

imports:
  - uses: shared/daily-audit-base.md
    with:
      title-prefix: "${{ github.aw.import-inputs.title-prefix }}"
      expires: "${{ github.aw.import-inputs.expires }}"
  - shared/go-source-analysis.md
  - shared/go-make.md
---
