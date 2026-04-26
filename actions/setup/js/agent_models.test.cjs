// @ts-check
/// <reference types="@actions/github-script" />

import { describe, test, expect, beforeEach, afterEach, vi } from "vitest";
import { createRequire } from "module";
import fs from "fs";
import os from "os";
import path from "path";

const require = createRequire(import.meta.url);
const { main, queryModels, extractModelsList, buildModelsMarkdown, logModels, resolveModelsEndpoint, resolveModelsToken, AGENTS_JSON_PATH, DEFAULT_COPILOT_BASE_URL } = require("./agent_models.cjs");

// ---------------------------------------------------------------------------
// Sample API response fixtures
// ---------------------------------------------------------------------------

const MODELS_RESPONSE_WITH_MODELS_KEY = {
  models: [
    { id: "gpt-4o", display_name: "GPT-4o", vendor: "openai" },
    { id: "claude-3-5-sonnet", display_name: "Claude 3.5 Sonnet", vendor: "anthropic" },
  ],
};

const MODELS_RESPONSE_WITH_DATA_KEY = {
  data: [{ id: "o1", display_name: "o1", vendor: "openai" }],
};

const MODELS_RESPONSE_BARE_ARRAY = [{ id: "gemini-1.5-pro", display_name: "Gemini 1.5 Pro", vendor: "google" }];

describe("agent_models", () => {
  describe("AGENTS_JSON_PATH constant", () => {
    test("points to /tmp/gh-aw/agents.json", () => {
      expect(AGENTS_JSON_PATH).toBe("/tmp/gh-aw/agents.json");
    });
  });

  // -------------------------------------------------------------------------
  // extractModelsList
  // -------------------------------------------------------------------------
  describe("extractModelsList", () => {
    test("extracts from { models: [...] } shape", () => {
      const result = extractModelsList(MODELS_RESPONSE_WITH_MODELS_KEY);
      expect(result).toHaveLength(2);
      expect(result[0]).toMatchObject({ id: "gpt-4o" });
    });

    test("extracts from { data: [...] } shape (OpenAI-compatible)", () => {
      const result = extractModelsList(MODELS_RESPONSE_WITH_DATA_KEY);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ id: "o1" });
    });

    test("extracts from bare array", () => {
      const result = extractModelsList(MODELS_RESPONSE_BARE_ARRAY);
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ id: "gemini-1.5-pro" });
    });

    test("returns empty array for null", () => {
      expect(extractModelsList(null)).toEqual([]);
    });

    test("returns empty array for empty object", () => {
      expect(extractModelsList({})).toEqual([]);
    });

    test("returns empty array for non-array models key", () => {
      expect(extractModelsList({ models: "not-an-array" })).toEqual([]);
    });
  });

  // -------------------------------------------------------------------------
  // buildModelsMarkdown
  // -------------------------------------------------------------------------
  describe("buildModelsMarkdown", () => {
    test("returns a markdown table for non-empty models", () => {
      const md = buildModelsMarkdown(MODELS_RESPONSE_WITH_MODELS_KEY.models);
      expect(md).toContain("| ID |");
      expect(md).toContain("gpt-4o");
      expect(md).toContain("GPT-4o");
      expect(md).toContain("openai");
      expect(md).toContain("claude-3-5-sonnet");
    });

    test("returns fallback message for empty list", () => {
      const md = buildModelsMarkdown([]);
      expect(md).toContain("No models");
    });

    test("handles models with display_name fallback to name", () => {
      const models = [{ id: "my-model", name: "My Model" }];
      const md = buildModelsMarkdown(models);
      expect(md).toContain("My Model");
    });

    test("handles models with owned_by as vendor fallback", () => {
      const models = [{ id: "x", display_name: "X", owned_by: "openai" }];
      const md = buildModelsMarkdown(models);
      expect(md).toContain("openai");
    });

    test("skips non-object entries gracefully", () => {
      const md = buildModelsMarkdown([null, undefined, "string"]);
      expect(md).toBeDefined();
    });
  });

  // -------------------------------------------------------------------------
  // logModels — now accepts a logFn callback (no core.* dependency)
  // -------------------------------------------------------------------------
  describe("logModels", () => {
    test("calls logFn with count and individual model IDs", () => {
      const logs = [];
      logModels(MODELS_RESPONSE_WITH_MODELS_KEY.models, "copilot", msg => logs.push(msg));
      expect(logs.some(l => l.includes("2"))).toBe(true);
      expect(logs.some(l => l.includes("gpt-4o"))).toBe(true);
      expect(logs.some(l => l.includes("claude-3-5-sonnet"))).toBe(true);
    });

    test("calls logFn with 0 for empty list", () => {
      const logs = [];
      logModels([], "copilot", msg => logs.push(msg));
      expect(logs.some(l => l.includes("0"))).toBe(true);
    });

    test("skips non-object entries", () => {
      const logs = [];
      logModels([null, "bad"], "copilot", msg => logs.push(msg));
      expect(logs).toHaveLength(1); // only the count line
    });
  });

  // -------------------------------------------------------------------------
  // resolveModelsEndpoint and resolveModelsToken — generic env var helpers
  // -------------------------------------------------------------------------
  describe("resolveModelsEndpoint", () => {
    let savedEnv;
    beforeEach(() => {
      savedEnv = { ...process.env };
    });
    afterEach(() => {
      process.env = savedEnv;
    });

    test("uses runtime env var when GH_AW_MODELS_BASE_URL_ENV points to a set var", () => {
      process.env.GH_AW_MODELS_BASE_URL_ENV = "ANTHROPIC_BASE_URL";
      process.env.ANTHROPIC_BASE_URL = "http://gateway:1234";
      delete process.env.GH_AW_MODELS_DEFAULT_BASE_URL;
      expect(resolveModelsEndpoint("/v1/models")).toBe("http://gateway:1234/v1/models");
    });

    test("falls back to GH_AW_MODELS_DEFAULT_BASE_URL when runtime var is absent", () => {
      process.env.GH_AW_MODELS_BASE_URL_ENV = "ANTHROPIC_BASE_URL";
      delete process.env.ANTHROPIC_BASE_URL;
      process.env.GH_AW_MODELS_DEFAULT_BASE_URL = "https://api.anthropic.com";
      expect(resolveModelsEndpoint("/v1/models")).toBe("https://api.anthropic.com/v1/models");
    });

    test("strips trailing slash from base URL", () => {
      delete process.env.GH_AW_MODELS_BASE_URL_ENV;
      process.env.GH_AW_MODELS_DEFAULT_BASE_URL = "https://api.example.com/";
      expect(resolveModelsEndpoint("/models")).toBe("https://api.example.com/models");
    });

    test("falls back to DEFAULT_COPILOT_BASE_URL when nothing is configured", () => {
      delete process.env.GH_AW_MODELS_BASE_URL_ENV;
      delete process.env.GH_AW_MODELS_DEFAULT_BASE_URL;
      expect(resolveModelsEndpoint("/models")).toBe(`${DEFAULT_COPILOT_BASE_URL}/models`);
    });
  });

  describe("resolveModelsToken", () => {
    let savedEnv;
    beforeEach(() => {
      savedEnv = { ...process.env };
    });
    afterEach(() => {
      process.env = savedEnv;
    });

    test("returns token from env var named by GH_AW_MODELS_TOKEN_ENV", () => {
      process.env.GH_AW_MODELS_TOKEN_ENV = "ANTHROPIC_API_KEY";
      process.env.ANTHROPIC_API_KEY = "sk-ant-test";
      expect(resolveModelsToken()).toBe("sk-ant-test");
    });

    test("falls back to COPILOT_GITHUB_TOKEN when GH_AW_MODELS_TOKEN_ENV is not set", () => {
      delete process.env.GH_AW_MODELS_TOKEN_ENV;
      process.env.COPILOT_GITHUB_TOKEN = "ghs_copilot_token";
      expect(resolveModelsToken()).toBe("ghs_copilot_token");
    });

    test("returns undefined when both token env vars are absent", () => {
      delete process.env.GH_AW_MODELS_TOKEN_ENV;
      delete process.env.COPILOT_GITHUB_TOKEN;
      expect(resolveModelsToken()).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // queryModels — core driver-context function
  // -------------------------------------------------------------------------
  describe("queryModels", () => {
    let tmpDir;

    beforeEach(() => {
      tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "agent-models-test-"));
    });

    afterEach(() => {
      fs.rmSync(tmpDir, { recursive: true, force: true });
    });

    test("emits warning and exits cleanly when endpoint is unreachable", async () => {
      const logs = [];
      await queryModels({
        endpoint: "https://127.0.0.1:1/models",
        token: "test-token",
        engineId: "copilot",
        engineVersion: "1.0.36",
        agentsJsonPath: path.join(tmpDir, "agents.json"),
        stepSummaryPath: null,
        logFn: msg => logs.push(msg),
      });
      expect(logs.some(l => l.includes("warning"))).toBe(true);
      expect(fs.existsSync(path.join(tmpDir, "agents.json"))).toBe(false);
    });

    test("writes agents.json and step summary when models are returned (mocked fs)", async () => {
      const agentsJsonPath = path.join(tmpDir, "agents.json");
      const summaryPath = path.join(tmpDir, "summary.md");
      const logs = [];

      // Use an in-process HTTP server to simulate a successful models response
      const http = require("http");
      const fakeModels = { models: [{ id: "test-model", display_name: "Test", vendor: "test" }] };
      const server = http.createServer((req, res) => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeModels));
      });
      await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
      const port = server.address().port;

      try {
        await queryModels({
          endpoint: `http://127.0.0.1:${port}/models`,
          token: "test-token",
          engineId: "copilot",
          engineVersion: "1.0.36",
          agentsJsonPath,
          stepSummaryPath: summaryPath,
          logFn: msg => logs.push(msg),
        });

        // agents.json written
        expect(fs.existsSync(agentsJsonPath)).toBe(true);
        const agentsData = JSON.parse(fs.readFileSync(agentsJsonPath, "utf8"));
        expect(agentsData).toHaveProperty("copilot");
        expect(agentsData.copilot.version).toBe("1.0.36");
        expect(agentsData.copilot.models).toEqual(fakeModels);

        // step summary written
        expect(fs.existsSync(summaryPath)).toBe(true);
        const summary = fs.readFileSync(summaryPath, "utf8");
        expect(summary).toContain("<details>");
        expect(summary).toContain("test-model");
      } finally {
        await new Promise(resolve => server.close(resolve));
      }
    });

    test("skips step summary when stepSummaryPath is null", async () => {
      const agentsJsonPath = path.join(tmpDir, "agents.json");
      const http = require("http");
      const fakeModels = { models: [] };
      const server = http.createServer((req, res) => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeModels));
      });
      await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
      const port = server.address().port;

      try {
        await queryModels({
          endpoint: `http://127.0.0.1:${port}/models`,
          token: "tok",
          engineId: "copilot",
          engineVersion: "1.0.36",
          agentsJsonPath,
          stepSummaryPath: null,
          logFn: () => {},
        });
        expect(fs.existsSync(agentsJsonPath)).toBe(true);
      } finally {
        await new Promise(resolve => server.close(resolve));
      }
    });
  });

  describe("DEFAULT_COPILOT_BASE_URL constant", () => {
    test("points to the Copilot API domain", () => {
      expect(DEFAULT_COPILOT_BASE_URL).toBe("https://api.githubcopilot.com");
    });
  });

  // -------------------------------------------------------------------------
  // main — github-script context wrapper
  // -------------------------------------------------------------------------
  describe("main", () => {
    let mockCore;
    let savedEnv;
    let tmpDir;

    beforeEach(() => {
      tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "agent-models-main-test-"));
      savedEnv = { ...process.env };

      mockCore = {
        info: vi.fn(),
        warning: vi.fn(),
        error: vi.fn(),
        setFailed: vi.fn(),
        summary: {
          addDetails: vi.fn().mockReturnThis(),
          write: vi.fn().mockResolvedValue(undefined),
        },
      };
      global.core = mockCore;
    });

    afterEach(() => {
      process.env = savedEnv;
      delete global.core;
      fs.rmSync(tmpDir, { recursive: true, force: true });
    });

    test("skips when GH_AW_MODELS_ROUTE is not set", async () => {
      delete process.env.GH_AW_MODELS_ROUTE;
      delete process.env.GH_AW_MODELS_TOKEN_ENV;
      delete process.env.COPILOT_GITHUB_TOKEN;

      await main();

      expect(mockCore.info).toHaveBeenCalledWith(expect.stringContaining("GH_AW_MODELS_ROUTE is not set"));
    });

    test("skips when auth token env var is not set", async () => {
      process.env.GH_AW_MODELS_ROUTE = "/models";
      delete process.env.GH_AW_MODELS_TOKEN_ENV;
      delete process.env.COPILOT_GITHUB_TOKEN;

      await main();

      expect(mockCore.info).toHaveBeenCalledWith(expect.stringContaining("Auth token env var is not set"));
    });

    test("logs warning when endpoint is unreachable", async () => {
      process.env.GH_AW_MODELS_ROUTE = "/models";
      process.env.GH_AW_MODELS_BASE_URL_ENV = "GITHUB_COPILOT_BASE_URL";
      process.env.GITHUB_COPILOT_BASE_URL = "https://127.0.0.1:1";
      process.env.GH_AW_MODELS_TOKEN_ENV = "COPILOT_GITHUB_TOKEN";
      process.env.COPILOT_GITHUB_TOKEN = "test-token";
      process.env.GH_AW_ENGINE_ID = "copilot";

      await main();

      expect(mockCore.info).toHaveBeenCalledWith(expect.stringContaining("warning: failed to query models endpoint"));
    });

    test("assembles URL from base URL env var and GH_AW_MODELS_ROUTE for any engine", async () => {
      const http = require("http");
      const fakeModels = { models: [] };
      const server = http.createServer((req, res) => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeModels));
      });
      await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
      const port = server.address().port;

      try {
        // Simulate Claude engine with ANTHROPIC_BASE_URL
        process.env.GH_AW_MODELS_ROUTE = "/v1/models";
        process.env.GH_AW_MODELS_BASE_URL_ENV = "ANTHROPIC_BASE_URL";
        process.env.ANTHROPIC_BASE_URL = `http://127.0.0.1:${port}`;
        process.env.GH_AW_MODELS_TOKEN_ENV = "ANTHROPIC_API_KEY";
        process.env.ANTHROPIC_API_KEY = "sk-ant-test";
        process.env.GH_AW_ENGINE_ID = "claude";

        await main();

        // Should have called info with the querying message (no warning)
        expect(mockCore.info).toHaveBeenCalledWith(expect.stringContaining("querying models from"));
      } finally {
        await new Promise(resolve => server.close(resolve));
      }
    });
  });
});
