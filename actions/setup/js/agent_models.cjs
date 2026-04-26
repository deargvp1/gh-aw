// @ts-check
/// <reference types="@actions/github-script" />

"use strict";

/**
 * agent_models.cjs
 *
 * Queries the agentic engine's /models endpoint before agent execution and stores
 * the results in /tmp/gh-aw/agents.json for inclusion in the agent artifact.
 *
 * The JSON file follows the structure:
 *   { "<engineId>": { "version": "<version>", "models": <models-data> } }
 *
 * Required environment variables:
 *   - GH_AW_MODELS_ENDPOINT: Full URL of the models endpoint
 *     (e.g. "https://api.githubcopilot.com/models")
 *   - COPILOT_GITHUB_TOKEN: Bearer token used to authenticate the request
 *   - GH_AW_ENGINE_ID: Agentic engine identifier (e.g. "copilot")
 *   - GH_AW_ENGINE_VERSION: Version string of the engine CLI
 *
 * If GH_AW_MODELS_ENDPOINT or COPILOT_GITHUB_TOKEN is unset the step exits
 * cleanly without writing anything (the compiler marks the step continue-on-error
 * so failures are non-fatal for the overall agent run).
 */

const fs = require("fs");
const https = require("https");
const http = require("http");
const { getErrorMessage } = require("./error_helpers.cjs");

/** Path where model data is written so it is bundled in the agent artifact. */
const AGENTS_JSON_PATH = "/tmp/gh-aw/agents.json";

/** Request timeout in milliseconds for the models HTTP call. */
const REQUEST_TIMEOUT_MS = 15_000;

/**
 * Perform an HTTP GET request to the models endpoint and return the parsed JSON body.
 *
 * @param {string} endpointUrl - Full URL of the models endpoint
 * @param {string} authToken   - Bearer token for the Authorization header
 * @returns {Promise<unknown>}  Parsed JSON response body
 */
function fetchModels(endpointUrl, authToken) {
  return new Promise((resolve, reject) => {
    let url;
    try {
      url = new URL(endpointUrl);
    } catch {
      reject(new Error(`Invalid models endpoint URL: ${endpointUrl}`));
      return;
    }

    const options = {
      hostname: url.hostname,
      port: url.port || (url.protocol === "https:" ? 443 : 80),
      path: url.pathname + (url.search || ""),
      method: "GET",
      headers: {
        Authorization: `Bearer ${authToken}`,
        Accept: "application/json",
        "Content-Type": "application/json",
      },
    };

    const requester = url.protocol === "https:" ? https : http;
    const req = requester.request(options, res => {
      let body = "";
      res.on(
        "data",
        /** @param {Buffer} chunk */ chunk => {
          body += chunk.toString();
        }
      );
      res.on("end", () => {
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode} from ${endpointUrl}: ${body.slice(0, 200)}`));
          return;
        }
        try {
          resolve(JSON.parse(body));
        } catch {
          reject(new Error(`Failed to parse models response as JSON: ${body.slice(0, 200)}`));
        }
      });
    });

    req.setTimeout(REQUEST_TIMEOUT_MS, () => {
      req.destroy();
      reject(new Error(`Request to ${endpointUrl} timed out after ${REQUEST_TIMEOUT_MS}ms`));
    });

    req.on("error", reject);
    req.end();
  });
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
 * Log individual models to core.info for easy scanning in the Actions log.
 *
 * @param {unknown[]} models
 * @param {string}    engineId
 */
function logModels(models, engineId) {
  core.info(`[${engineId}] Available models (${models.length}):`);
  for (const m of models) {
    if (!m || typeof m !== "object") continue;
    const entry = /** @type {Record<string, unknown>} */ m;
    const id = String(entry["id"] || "?");
    const name = String(entry["display_name"] || entry["name"] || "");
    core.info(`  - ${id}${name ? ": " + name : ""}`);
  }
}

/**
 * Main entry point — called by the compiler-generated github-script step.
 * Exits cleanly (non-fatal) when required env vars are absent.
 */
async function main() {
  const modelsEndpoint = process.env.GH_AW_MODELS_ENDPOINT;
  if (!modelsEndpoint) {
    core.info("GH_AW_MODELS_ENDPOINT is not set — skipping models query");
    return;
  }

  const authToken = process.env.COPILOT_GITHUB_TOKEN;
  if (!authToken) {
    core.info("COPILOT_GITHUB_TOKEN is not set — skipping models query");
    return;
  }

  const engineId = process.env.GH_AW_ENGINE_ID || "copilot";
  const engineVersion = process.env.GH_AW_ENGINE_VERSION || "unknown";

  core.info(`Querying models from: ${modelsEndpoint} (engine=${engineId} version=${engineVersion})`);

  let modelsData;
  try {
    modelsData = await fetchModels(modelsEndpoint, authToken);
  } catch (error) {
    core.warning(`Failed to query models endpoint: ${getErrorMessage(error)}`);
    return;
  }

  const modelsList = extractModelsList(modelsData);
  logModels(modelsList, engineId);

  // Write agents.json so the data is bundled in the agent artifact
  const agentsInfo = {
    [engineId]: {
      version: engineVersion,
      models: modelsData,
    },
  };

  try {
    fs.mkdirSync("/tmp/gh-aw", { recursive: true });
    fs.writeFileSync(AGENTS_JSON_PATH, JSON.stringify(agentsInfo, null, 2) + "\n");
    core.info(`Wrote models info to ${AGENTS_JSON_PATH}`);
  } catch (error) {
    core.warning(`Failed to write ${AGENTS_JSON_PATH}: ${getErrorMessage(error)}`);
  }

  // Add collapsible section to step summary
  const markdown = buildModelsMarkdown(modelsList);
  try {
    await core.summary.addDetails(`Available Models (${engineId} ${engineVersion})`, "\n\n" + markdown + "\n").write();
  } catch (error) {
    core.warning(`Failed to write models step summary: ${getErrorMessage(error)}`);
  }
}

module.exports = {
  main,
  fetchModels,
  extractModelsList,
  buildModelsMarkdown,
  logModels,
  AGENTS_JSON_PATH,
  REQUEST_TIMEOUT_MS,
};

if (require.main === module) {
  main().catch(err => {
    // eslint-disable-next-line no-console
    console.error(`[agent_models] unexpected error: ${err.message}`);
    process.exit(1);
  });
}
