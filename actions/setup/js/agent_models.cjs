// @ts-check
/// <reference types="@actions/github-script" />

"use strict";

/**
 * agent_models.cjs
 *
 * Queries the agentic engine's models endpoint before agent execution and stores
 * the results in /tmp/gh-aw/agents.json for inclusion in the agent artifact.
 *
 * The JSON file follows the structure:
 *   { "<engineId>": { "version": "<version>", "models": <models-data> } }
 *
 * Primary API — usable from any Node.js context (driver, github-script, standalone):
 *   queryModels({ endpoint, token, engineId, engineVersion, agentsJsonPath?, stepSummaryPath?, logFn? })
 *
 * github-script convenience wrapper (uses core.* globals):
 *   main()   — reads env vars and delegates to queryModels()
 *
 * Required environment variables (set by the Go engine via GH_AW_MODELS_* vars):
 *   - GH_AW_MODELS_ROUTE:            Route path for the models endpoint (e.g. "/models")
 *   - GH_AW_MODELS_BASE_URL_ENV:     Name of the env var that holds the API base URL at runtime
 *                                    (e.g. "GITHUB_COPILOT_BASE_URL", "ANTHROPIC_BASE_URL")
 *   - GH_AW_MODELS_DEFAULT_BASE_URL: Fallback base URL when the above env var is not set
 *   - GH_AW_MODELS_TOKEN_ENV:        Name of the env var that holds the auth token
 *                                    (e.g. "COPILOT_GITHUB_TOKEN", "ANTHROPIC_API_KEY")
 *   - GH_AW_ENGINE_ID:               Agentic engine identifier (e.g. "copilot")
 */

const fs = require("fs");
const { getErrorMessage } = require("./error_helpers.cjs");

/** Default Copilot API base URL — used as last-resort fallback when no engine-specific default is provided. */
const DEFAULT_COPILOT_BASE_URL = "https://api.githubcopilot.com";

/** Path where model data is written so it is bundled in the agent artifact. */
const AGENTS_JSON_PATH = "/tmp/gh-aw/agents.json";

/** Request timeout in milliseconds for the models HTTP call. */
const REQUEST_TIMEOUT_MS = 15_000;

/**
 * Resolve the effective models endpoint URL from environment variables set by the Go engine.
 * Uses GH_AW_MODELS_BASE_URL_ENV to dynamically look up the runtime base URL, falling back to
 * GH_AW_MODELS_DEFAULT_BASE_URL and then DEFAULT_COPILOT_BASE_URL.
 *
 * @param {string} modelsRoute - Route path from GH_AW_MODELS_ROUTE (e.g. "/models")
 * @returns {string} Full models endpoint URL
 */
function resolveModelsEndpoint(modelsRoute) {
  const baseUrlEnvName = process.env.GH_AW_MODELS_BASE_URL_ENV;
  const runtimeBaseUrl = baseUrlEnvName ? process.env[baseUrlEnvName] : undefined;
  const defaultBaseUrl = process.env.GH_AW_MODELS_DEFAULT_BASE_URL || DEFAULT_COPILOT_BASE_URL;
  const baseUrl = (runtimeBaseUrl || defaultBaseUrl).replace(/\/$/, "");
  return baseUrl + modelsRoute;
}

/**
 * Resolve the authentication token from the env var named by GH_AW_MODELS_TOKEN_ENV,
 * falling back to COPILOT_GITHUB_TOKEN for backwards compatibility.
 *
 * @returns {string | undefined}
 */
function resolveModelsToken() {
  const tokenEnvName = process.env.GH_AW_MODELS_TOKEN_ENV;
  return tokenEnvName ? process.env[tokenEnvName] : process.env.COPILOT_GITHUB_TOKEN;
}

/**
 * Perform an HTTP GET request to the models endpoint and return the parsed JSON body.
 * Uses the Node.js built-in fetch API (available since Node 18, stable in Node 21+).
 *
 * @param {string} endpointUrl - Full URL of the models endpoint
 * @param {string} authToken   - Bearer token for the Authorization header
 * @returns {Promise<unknown>}  Parsed JSON response body
 */
async function fetchModels(endpointUrl, authToken) {
  const response = await fetch(endpointUrl, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${authToken}`,
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });

  if (!response.ok) {
    const body = await response.text();
    throw new Error(`HTTP ${response.status} from ${endpointUrl}: ${body.slice(0, 200)}`);
  }

  return response.json();
}

/**
 * Extract the flat models list from the API response, which may arrive in several shapes:
 *   - { models: [...] }  — most common
 *   - { data: [...] }    — OpenAI-compatible list
 *   - [...]              — bare array
 *
 * @param {unknown} responseBody
 * @returns {unknown[]}
 */
function extractModelsList(responseBody) {
  if (Array.isArray(responseBody)) {
    return responseBody;
  }
  if (responseBody && typeof responseBody === "object") {
    const obj = /** @type {Record<string, unknown>} */ responseBody;
    if (Array.isArray(obj["models"])) {
      return /** @type {unknown[]} */ obj["models"];
    }
    if (Array.isArray(obj["data"])) {
      return /** @type {unknown[]} */ obj["data"];
    }
  }
  return [];
}

/**
 * Build a markdown table from a flat models list for the step summary.
 *
 * @param {unknown[]} models
 * @returns {string} Markdown table or a plain message when the list is empty
 */
function buildModelsMarkdown(models) {
  if (models.length === 0) {
    return "No models returned by the endpoint.";
  }

  const rows = ["| ID | Display name | Vendor |", "| --- | --- | --- |"];
  for (const m of models) {
    if (!m || typeof m !== "object") continue;
    const entry = /** @type {Record<string, unknown>} */ m;
    const id = String(entry["id"] || "");
    const name = String(entry["display_name"] || entry["name"] || "");
    const vendor = String(entry["vendor"] || entry["owned_by"] || "");
    rows.push(`| ${id} | ${name} | ${vendor} |`);
  }
  return rows.join("\n");
}

/**
 * Log individual models using the provided logging function.
 *
 * @param {unknown[]} models
 * @param {string}    engineId
 * @param {(msg: string) => void} logFn
 */
function logModels(models, engineId, logFn) {
  logFn(`[${engineId}] Available models (${models.length}):`);
  for (const m of models) {
    if (!m || typeof m !== "object") continue;
    const entry = /** @type {Record<string, unknown>} */ m;
    const id = String(entry["id"] || "?");
    const name = String(entry["display_name"] || entry["name"] || "");
    logFn(`  - ${id}${name ? ": " + name : ""}`);
  }
}

/**
 * Query the models endpoint, persist results to agents.json, and optionally append
 * a summary section to the step-summary file.  Callable from any Node.js context
 * (driver harness, github-script, standalone) without depending on global `core.*`.
 *
 * @param {{
 *   endpoint: string,
 *   token: string,
 *   engineId: string,
 *   engineVersion: string,
 *   agentsJsonPath?: string,
 *   stepSummaryPath?: string | null,
 *   logFn?: (msg: string) => void,
 * }} options
 * @returns {Promise<void>}
 */
async function queryModels({ endpoint, token, engineId, engineVersion, agentsJsonPath = AGENTS_JSON_PATH, stepSummaryPath = null, logFn = () => {} }) {
  logFn(`querying models from: ${endpoint} (engine=${engineId} version=${engineVersion})`);

  let modelsData;
  try {
    modelsData = await fetchModels(endpoint, token);
  } catch (error) {
    logFn(`warning: failed to query models endpoint: ${getErrorMessage(error)}`);
    return;
  }

  const modelsList = extractModelsList(modelsData);
  logModels(modelsList, engineId, logFn);

  // Write agents.json so the data is bundled in the agent artifact
  const agentsInfo = {
    [engineId]: {
      version: engineVersion,
      models: modelsData,
    },
  };

  try {
    fs.mkdirSync("/tmp/gh-aw", { recursive: true });
    fs.writeFileSync(agentsJsonPath, JSON.stringify(agentsInfo, null, 2) + "\n");
    logFn(`wrote models info to ${agentsJsonPath}`);
  } catch (error) {
    logFn(`warning: failed to write ${agentsJsonPath}: ${getErrorMessage(error)}`);
  }

  // Append a collapsible section to the step summary file
  if (stepSummaryPath) {
    const markdown = buildModelsMarkdown(modelsList);
    const section = `\n<details>\n<summary>Available Models (${engineId} ${engineVersion})</summary>\n\n${markdown}\n</details>\n`;
    try {
      fs.appendFileSync(stepSummaryPath, section);
    } catch (error) {
      logFn(`warning: failed to write models step summary: ${getErrorMessage(error)}`);
    }
  }
}

/**
 * Main entry point — called by the compiler-generated github-script step.
 * Reads engine identity from GH_AW_MODELS_* env vars set by the Go engine.
 * Exits cleanly (non-fatal) when required env vars are absent.
 */
async function main() {
  const modelsRoute = process.env.GH_AW_MODELS_ROUTE;
  if (!modelsRoute) {
    core.info("GH_AW_MODELS_ROUTE is not set — skipping models query");
    return;
  }

  const authToken = resolveModelsToken();
  if (!authToken) {
    core.info("Auth token env var is not set — skipping models query");
    return;
  }

  const endpoint = resolveModelsEndpoint(modelsRoute);
  const engineId = process.env.GH_AW_ENGINE_ID || "copilot";
  const engineVersion = process.env.GH_AW_ENGINE_VERSION || "unknown";

  await queryModels({
    endpoint,
    token: authToken,
    engineId,
    engineVersion,
    stepSummaryPath: process.env.GITHUB_STEP_SUMMARY || null,
    logFn: msg => core.info(msg),
  });
}

module.exports = {
  main,
  queryModels,
  fetchModels,
  extractModelsList,
  buildModelsMarkdown,
  logModels,
  resolveModelsEndpoint,
  resolveModelsToken,
  AGENTS_JSON_PATH,
  REQUEST_TIMEOUT_MS,
  DEFAULT_COPILOT_BASE_URL,
};

if (require.main === module) {
  // Standalone mode: invoked directly from a shell preamble (e.g. node agent_models.cjs).
  // Reads all config from GH_AW_MODELS_* environment variables; no github-script context needed.
  const modelsRoute = process.env.GH_AW_MODELS_ROUTE;
  if (!modelsRoute) process.exit(0); // Nothing configured — silent skip

  const authToken = resolveModelsToken();
  if (!authToken) process.exit(0); // Token unavailable (e.g. secret excluded in AWF sandbox)

  const endpoint = resolveModelsEndpoint(modelsRoute);
  const engineId = process.env.GH_AW_ENGINE_ID || "unknown";

  queryModels({
    endpoint,
    token: authToken,
    engineId,
    engineVersion: "unknown",
    stepSummaryPath: process.env.GITHUB_STEP_SUMMARY || null,
    logFn: msg => process.stderr.write(`[agent_models] ${msg}\n`),
  }).catch(e => {
    process.stderr.write(`[agent_models] warning: ${e.message || e}\n`);
  });
}
