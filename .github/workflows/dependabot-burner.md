---
on: weekly
permissions:
  contents: read
  issues: read
  pull-requests: read
tools:
  cli-proxy: true
  github:
imports:
  - uses: shared/daily-issue-base.md
    with:
      title-prefix: "[dependabot-burner] "
      expires: "2d"
---
# Dependabot Burner

- Find all open Dependabot PRs.
- Create bundle issues, each for exactly **one runtime + one manifest file**.

{{#import shared/noop-reminder.md}}
