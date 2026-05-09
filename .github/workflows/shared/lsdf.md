---
# L-SDF Setup
# Shared configuration for installing and using lsdf-core in workflows.
#
# Usage:
#   imports:
#     - shared/lsdf.md
#
# This import provides:
# - Automatic lsdf-core installation via pipx
# - `lsdf` available in PATH
# - Prompt guidance for LSDF-first repository navigation

tools:
  bash:
    - "lsdf *"

steps:
  - name: Install lsdf-core
    run: |
      python3 -m pip install --user --upgrade pipx
      python3 -m pipx ensurepath
      export PATH="${HOME}/.local/bin:${PATH}"
      pipx install --force lsdf-core
      echo "${HOME}/.local/bin" >> "$GITHUB_PATH"

  - name: Verify lsdf installation
    run: |
      export PATH="${HOME}/.local/bin:${PATH}"
      lsdf --help >/dev/null
      lsdf --version
---

<!--
## L-SDF (Latent-Structured Documentation Format)

This shared configuration installs `lsdf-core` so workflows can use compact index maps (`project.lsdf`, `INDEX.lsdf`, `INDEX.detail.lsdf`) for token-efficient codebase navigation.

Links:
- GitHub: https://github.com/ec1980/lsdf-core
- PyPI: https://pypi.org/project/lsdf-core/
-->

You have access to the `lsdf` CLI.

Use LSDF-first navigation whenever possible:
1. Read `project.lsdf` first (if present).
2. Read the nearest `INDEX.lsdf` for structure.
3. Read `INDEX.detail.lsdf` when signatures/call edges are needed.
4. Open raw source files only after LSDF map review.

If the repository does not have LSDF files yet and your task benefits from structural indexing, initialize and generate indices before deep analysis:

```bash
lsdf init
lsdf gen . --recursive
```

After structural code edits, regenerate indices for changed areas:

```bash
lsdf gen . --recursive
```
