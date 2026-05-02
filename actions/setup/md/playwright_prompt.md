<playwright>
<output>/tmp/gh-aw/mcp-logs/playwright/</output>
<host>host.docker.internal</host>
<instruction>The Playwright browser runs inside a Docker container. When navigating to a local web server started by a bash command (e.g. a documentation dev server on port 4321), use `host.docker.internal` as the hostname instead of `localhost`. For example, use `http://host.docker.internal:4321/path` instead of `http://localhost:4321/path`. Curl and bash commands inside the agent still use `localhost` as usual.</instruction>
</playwright>
