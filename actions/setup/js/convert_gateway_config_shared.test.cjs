// @ts-check
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createRequire } from "module";
import fs from "fs";
import os from "os";
import path from "path";

const req = createRequire(import.meta.url);

// Load shim so global.core is available
req("./shim.cjs");

const { rewriteUrl, filterAndTransformServers, logCLIFilters, logServerStats, writeSecureOutput, loadGatewayContext } = req("./convert_gateway_config_shared.cjs");

// ---------------------------------------------------------------------------
// rewriteUrl
// ---------------------------------------------------------------------------

describe("rewriteUrl", () => {
  it("rewrites the host and port while preserving the server path", () => {
    const url = "http://localhost:8080/mcp/my-server";
    expect(rewriteUrl(url, "http://host.docker.internal:80")).toBe("http://host.docker.internal:80/mcp/my-server");
  });

  it("handles a URL with a different original host", () => {
    expect(rewriteUrl("http://127.0.0.1:3000/mcp/tool", "http://example.com:9000")).toBe("http://example.com:9000/mcp/tool");
  });

  it("returns the url unchanged when it does not match the expected pattern", () => {
    const url = "https://example.com/api/v1";
    expect(rewriteUrl(url, "http://host.docker.internal:80")).toBe(url);
  });

  it("handles an empty server name path segment", () => {
    const url = "http://localhost:1234/mcp/";
    expect(rewriteUrl(url, "http://proxy:42")).toBe("http://proxy:42/mcp/");
  });
});

// ---------------------------------------------------------------------------
// filterAndTransformServers
// ---------------------------------------------------------------------------

describe("filterAndTransformServers", () => {
  const servers = {
    alpha: { url: "http://localhost/mcp/alpha", headers: {} },
    beta: { url: "http://localhost/mcp/beta", headers: {} },
    gamma: { url: "http://localhost/mcp/gamma", headers: {} },
  };

  it("returns all servers when cliServers is empty", () => {
    const result = filterAndTransformServers(servers, new Set(), (_name, entry) => entry);
    expect(Object.keys(result)).toEqual(["alpha", "beta", "gamma"]);
  });

  it("excludes CLI-mounted servers", () => {
    const result = filterAndTransformServers(servers, new Set(["beta"]), (_name, entry) => entry);
    expect(Object.keys(result)).toEqual(["alpha", "gamma"]);
    expect(result).not.toHaveProperty("beta");
  });

  it("excludes multiple CLI-mounted servers", () => {
    const result = filterAndTransformServers(servers, new Set(["alpha", "gamma"]), (_name, entry) => entry);
    expect(Object.keys(result)).toEqual(["beta"]);
  });

  it("applies the transform function to each included server", () => {
    const result = filterAndTransformServers(servers, new Set(), (name, entry) => ({ ...entry, name }));
    expect(result["alpha"]).toMatchObject({ name: "alpha" });
    expect(result["gamma"]).toMatchObject({ name: "gamma" });
  });

  it("passes a shallow copy of the entry to transformServer", () => {
    /** @type {Record<string, unknown> | null} */
    let received = null;
    filterAndTransformServers({ single: { a: 1 } }, new Set(), (_name, entry) => {
      received = entry;
      return entry;
    });
    expect(received).toEqual({ a: 1 });
    // Mutating the copy must not affect the original
    if (received) received["a"] = 99;
    expect(servers["alpha"]["url"]).toBe("http://localhost/mcp/alpha");
  });

  it("returns an empty object when all servers are filtered", () => {
    const result = filterAndTransformServers(servers, new Set(["alpha", "beta", "gamma"]), (_name, e) => e);
    expect(result).toEqual({});
  });

  it("returns an empty object when servers is empty", () => {
    const result = filterAndTransformServers({}, new Set(), (_name, e) => e);
    expect(result).toEqual({});
  });
});

// ---------------------------------------------------------------------------
// logCLIFilters
// ---------------------------------------------------------------------------

describe("logCLIFilters", () => {
  beforeEach(() => {
    vi.spyOn(global.core, "info").mockImplementation(() => {});
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("logs server names when cliServers is non-empty", () => {
    logCLIFilters(new Set(["safeoutputs", "mcpscripts"]));
    expect(global.core.info).toHaveBeenCalledWith(expect.stringContaining("safeoutputs"));
    expect(global.core.info).toHaveBeenCalledWith(expect.stringContaining("mcpscripts"));
  });

  it("does not log when cliServers is empty", () => {
    logCLIFilters(new Set());
    expect(global.core.info).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// logServerStats
// ---------------------------------------------------------------------------

describe("logServerStats", () => {
  beforeEach(() => {
    vi.spyOn(global.core, "info").mockImplementation(() => {});
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("logs included and filtered counts", () => {
    const servers = { a: {}, b: {}, c: {} };
    logServerStats(servers, 2);
    expect(global.core.info).toHaveBeenCalledWith(expect.stringMatching(/2 included.*1 filtered/));
  });

  it("reports zero filtered when all servers are included", () => {
    logServerStats({ x: {}, y: {} }, 2);
    expect(global.core.info).toHaveBeenCalledWith(expect.stringMatching(/2 included.*0 filtered/));
  });
});

// ---------------------------------------------------------------------------
// writeSecureOutput
// ---------------------------------------------------------------------------

describe("writeSecureOutput", () => {
  /** @type {string} */
  let tmpDir;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "jsweep-test-"));
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  it("creates parent directories and writes the file", () => {
    const outputPath = path.join(tmpDir, "nested", "dir", "output.json");
    writeSecureOutput(outputPath, '{"ok":true}');
    expect(fs.existsSync(outputPath)).toBe(true);
    expect(fs.readFileSync(outputPath, "utf8")).toBe('{"ok":true}');
  });

  it("writes the file with 0o600 permissions", () => {
    const outputPath = path.join(tmpDir, "secure.json");
    writeSecureOutput(outputPath, "secret");
    const mode = fs.statSync(outputPath).mode & 0o777;
    expect(mode).toBe(0o600);
  });
});

// ---------------------------------------------------------------------------
// loadGatewayContext
// ---------------------------------------------------------------------------

describe("loadGatewayContext", () => {
  /** @type {string} */
  let tmpDir;
  /** @type {string} */
  let gatewayFile;
  /** @type {Record<string, string | undefined>} */
  let savedEnv;

  const MANAGED_ENV = ["MCP_GATEWAY_OUTPUT", "MCP_GATEWAY_DOMAIN", "MCP_GATEWAY_PORT", "GH_AW_MCP_CLI_SERVERS"];

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "jsweep-test-"));
    gatewayFile = path.join(tmpDir, "gateway.json");

    savedEnv = Object.fromEntries(MANAGED_ENV.map(k => [k, process.env[k]]));

    process.env.MCP_GATEWAY_OUTPUT = gatewayFile;
    process.env.MCP_GATEWAY_DOMAIN = "host.docker.internal";
    process.env.MCP_GATEWAY_PORT = "80";
    delete process.env.GH_AW_MCP_CLI_SERVERS;

    vi.spyOn(global.core, "info").mockImplementation(() => {});
    vi.spyOn(global.core, "error").mockImplementation(() => {});
    vi.spyOn(process, "exit").mockImplementation(
      /** @param {number | undefined} _code */ _code => {
        throw new Error("process.exit called");
      }
    );
  });

  afterEach(() => {
    vi.restoreAllMocks();
    fs.rmSync(tmpDir, { recursive: true, force: true });
    for (const [k, v] of Object.entries(savedEnv)) {
      if (v === undefined) delete process.env[k];
      else process.env[k] = v;
    }
  });

  it("returns expected context fields for a valid gateway file", () => {
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: { tool1: { url: "http://localhost/mcp/tool1" } } }));
    const ctx = loadGatewayContext();
    expect(ctx.domain).toBe("host.docker.internal");
    expect(ctx.port).toBe("80");
    expect(ctx.urlPrefix).toBe("http://host.docker.internal:80");
    expect(ctx.gatewayOutput).toBe(gatewayFile);
    expect(ctx.cliServers).toBeInstanceOf(Set);
    expect(ctx.servers).toHaveProperty("tool1");
    expect(ctx.extraEnv).toEqual({});
  });

  it("parses GH_AW_MCP_CLI_SERVERS into a Set", () => {
    process.env.GH_AW_MCP_CLI_SERVERS = JSON.stringify(["safeoutputs", "mcpscripts"]);
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: {} }));
    const ctx = loadGatewayContext();
    expect(ctx.cliServers.has("safeoutputs")).toBe(true);
    expect(ctx.cliServers.has("mcpscripts")).toBe(true);
  });

  it("treats missing GH_AW_MCP_CLI_SERVERS as an empty set", () => {
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: {} }));
    const ctx = loadGatewayContext();
    expect(ctx.cliServers.size).toBe(0);
  });

  it("returns empty servers when mcpServers is missing from the gateway file", () => {
    fs.writeFileSync(gatewayFile, JSON.stringify({}));
    const ctx = loadGatewayContext();
    expect(ctx.servers).toEqual({});
  });

  it("returns empty servers when mcpServers is an array", () => {
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: [] }));
    const ctx = loadGatewayContext();
    expect(ctx.servers).toEqual({});
  });

  it("resolves extra required env vars", () => {
    process.env.MY_TOKEN = "abc123";
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: {} }));
    const ctx = loadGatewayContext({ extraRequiredEnv: ["MY_TOKEN"] });
    expect(ctx.extraEnv["MY_TOKEN"]).toBe("abc123");
    delete process.env.MY_TOKEN;
  });

  it("calls process.exit when MCP_GATEWAY_OUTPUT is missing", () => {
    delete process.env.MCP_GATEWAY_OUTPUT;
    expect(() => loadGatewayContext()).toThrow("process.exit called");
  });

  it("calls process.exit when gateway output file does not exist", () => {
    process.env.MCP_GATEWAY_OUTPUT = path.join(tmpDir, "nonexistent.json");
    expect(() => loadGatewayContext()).toThrow("process.exit called");
  });

  it("calls process.exit when MCP_GATEWAY_DOMAIN is missing", () => {
    fs.writeFileSync(gatewayFile, JSON.stringify({ mcpServers: {} }));
    delete process.env.MCP_GATEWAY_DOMAIN;
    expect(() => loadGatewayContext()).toThrow("process.exit called");
  });
});
